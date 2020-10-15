package errorly

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"context"

	idgenerator "github.com/TheRockettek/Errorly-Web/pkg/dictionary/idgenerator"
	"github.com/go-pg/pg/v10"
	"github.com/gorilla/sessions"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// VERSION respects semantic versioning
const VERSION = "0.2"

// ConfigurationPath is the path to the file the configration will be located
// at.
const ConfigurationPath = "errorly.yaml"

// ErrMissingSecret is raised when no/invalid secret is specified for cookie signing
var ErrMissingSecret = xerrors.New("Invalid secret '%s' provided")

// Configuration represents the configuration for Errorly
type Configuration struct {
	Host          string `json:"host" yaml:"host"`
	SessionSecret string `json:"secret" yaml:"secret"`

	Postgres *pg.Options    `json:"postgres" yaml:"postgres"`
	OAuth    *oauth2.Config `json:"oauth" yaml:"oauth"`

	Logging struct {
		ConsoleLoggingEnabled bool `json:"console_logging" yaml:"console_logging"`
		FileLoggingEnabled    bool `json:"file_logging" yaml:"file_logging"`

		EncodeAsJSON bool `json:"encode_as_json" yaml:"encode_as_json"` // Make the framework log as json

		Directory  string `json:"directory" yaml:"directory"`     // Directory to log into
		Filename   string `json:"filename" yaml:"filename"`       // Name of logfile
		MaxSize    int    `json:"max_size" yaml:"max_size"`       /// Size in MB before a new file
		MaxBackups int    `json:"max_backups" yaml:"max_backups"` // Number of files to keep
		MaxAge     int    `json:"max_age" yaml:"max_age"`         // Number of days to keep a logfile
	} `json:"logging" yaml:"logging"`
}

// Errorly represents the global application state
type Errorly struct {
	ctx    context.Context
	cancel func()

	Configuration *Configuration `json:"configuration"`

	Logger zerolog.Logger `json:"-"`

	Postgres      *pg.DB
	PostgressConn *pg.Conn

	Router *MethodRouter
	Store  *sessions.CookieStore
	IDGen  *idgenerator.IDGenerator
}

// NewErrorly creates an Errorly instance.
func NewErrorly(logger io.Writer) (er *Errorly, err error) {

	_ = pg.Options{}

	ctx, cancel := context.WithCancel(context.Background())
	er = &Errorly{
		ctx:    ctx,
		cancel: cancel,

		Logger: zerolog.New(logger).With().Timestamp().Logger(),
	}

	// Load Configuration
	configuration, err := er.LoadConfiguration(ConfigurationPath)
	if err != nil {
		return nil, xerrors.Errorf("new errorly: %w", err)
	}
	if configuration.Host == "" {
		return nil, xerrors.New("No host provided")
	}
	er.Configuration = configuration

	// Create logging writers
	var writers []io.Writer
	if er.Configuration.Logging.ConsoleLoggingEnabled {
		writers = append(writers, logger)
	}
	if er.Configuration.Logging.FileLoggingEnabled {
		if err := os.MkdirAll(er.Configuration.Logging.Directory, 0744); err != nil {
			log.Error().Err(err).Str("path", er.Configuration.Logging.Directory).Msg("Unable to create log directory")
		} else {
			lumber := &lumberjack.Logger{
				Filename:   path.Join(er.Configuration.Logging.Directory, er.Configuration.Logging.Filename),
				MaxBackups: er.Configuration.Logging.MaxBackups,
				MaxSize:    er.Configuration.Logging.MaxSize,
				MaxAge:     er.Configuration.Logging.MaxAge,
			}
			if er.Configuration.Logging.EncodeAsJSON {
				writers = append(writers, lumber)
			} else {
				writers = append(writers, zerolog.ConsoleWriter{
					Out:        lumber,
					TimeFormat: time.Stamp,
					NoColor:    true,
				})
			}
		}
	}
	mw := io.MultiWriter(writers...)
	er.Logger = zerolog.New(mw).With().Timestamp().Logger()
	er.Logger.Info().Msg("Logging configured")

	return
}

// LoadConfiguration loads the configuration for RestTunnel
func (er *Errorly) LoadConfiguration(path string) (configuration *Configuration, err error) {
	er.Logger.Debug().Msg("Loading configuration")

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return configuration, xerrors.Errorf("load configuration readfile: %w", err)
	}

	configuration = &Configuration{}
	err = yaml.Unmarshal(file, &configuration)
	if err != nil {
		return configuration, xerrors.Errorf("load configuration unmarshal: %w", err)
	}

	return
}

// HandleRequest handles incoming HTTP requests
func (er *Errorly) HandleRequest(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Request.URI().Path())

	defer func() {
		fmt.Printf("%s %s %s %d\n",
			ctx.RemoteAddr(),
			ctx.Request.Header.Method(),
			ctx.Request.URI().Path(),
			ctx.Response.StatusCode())
	}()

	if strings.HasPrefix(path, "/static") {
		root, _ := os.Getwd()
		_, filename := filepath.Split(path)
		filepath := filepath.Join(root, "web/static", filename)

		if _, err := os.Stat(filepath); err != nil {
			if os.IsNotExist(err) {
				ctx.SetStatusCode(404)
			} else {
				ctx.SetStatusCode(500)
			}
		} else {
			ctx.SendFile(filepath)
		}
		return
	}

	if path == "/" {
		body, _ := ioutil.ReadFile("web/spa.html")
		ctx.Write(body)
		ctx.SetContentType("text/html")
		ctx.SetStatusCode(200)
		// ctx.SendFile("web/spa.html")
		return
	}

	fasthttpadaptor.NewFastHTTPHandler(er.Router)(ctx)
}

// Open starts the web worker
func (er *Errorly) Open() (err error) {

	var secret string
	secret = er.Configuration.SessionSecret
	if secret == "" {
		secret = os.Getenv("ERRORLY_SECRET")
		if secret == "" {
			return xerrors.Errorf(ErrMissingSecret.Error(), secret)
		}
	}
	if len(secret) != 32 {
		return xerrors.Errorf(ErrMissingSecret.Error(), secret)
	}

	er.Postgres = pg.Connect(er.Configuration.Postgres)
	er.PostgressConn = er.Postgres.Conn()
	defer er.PostgressConn.Close()
	defer er.Postgres.Close()

	if err := er.PostgressConn.Ping(er.ctx); err != nil {
		return err
	}
	er.Logger.Info().Msg("Connected to postgres")

	er.Store = sessions.NewCookieStore([]byte(er.Configuration.SessionSecret))
	er.IDGen = idgenerator.NewIDGenerator(epoch, 0)

	er.Logger.Debug().Msg("Creating schema")
	err = createSchema(er.Postgres)
	if err != nil {
		return err
	}
	er.Logger.Debug().Msg("Created schema")

	er.Logger.Debug().Msg("Creating endpoints")
	er.Router = createEndpoints(er)
	er.Logger.Debug().Msg("Created endpoints")

	fmt.Printf("Serving on %s (Press CTRL+C to quit)\n", er.Configuration.Host)
	err = fasthttp.ListenAndServe(er.Configuration.Host, er.HandleRequest)

	return
}
