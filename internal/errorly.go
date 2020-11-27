package errorly

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	idgenerator "github.com/TheRockettek/Errorly-Web/pkg/idgenerator"
	"github.com/go-pg/pg/v10"
	"github.com/gorilla/sessions"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// VERSION respects semantic versioning.
const VERSION = "0.2"

// ConfigurationPath is the path to the file the configration will be located
// at.
const ConfigurationPath = "errorly.yaml"

const cacheDuration = time.Hour * 24 * 30

// ErrMissingSecret is raised when no/invalid secret is specified for cookie signing.
var ErrMissingSecret = xerrors.New("Invalid secret '%s' provided")

// Configuration represents the configuration for Errorly.
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
		MaxSize    int    `json:"max_size" yaml:"max_size"`       // Size in MB before a new file
		MaxBackups int    `json:"max_backups" yaml:"max_backups"` // Number of files to keep
		MaxAge     int    `json:"max_age" yaml:"max_age"`         // Number of days to keep a logfile
	} `json:"logging" yaml:"logging"`
}

// Errorly represents the global application state.
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

	distHandler fasthttp.RequestHandler
	fs          *fasthttp.FS
}

type dbLogger struct{}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	a, _ := q.FormattedQuery()
	println(gotils.B2S(a))

	return nil
}

// NewErrorly creates an Errorly instance.
func NewErrorly(logger io.Writer, level zerolog.Level) (er *Errorly, err error) {
	_ = pg.Options{}

	ctx, cancel := context.WithCancel(context.Background())
	er = &Errorly{
		ctx:    ctx,
		cancel: cancel,

		Logger: zerolog.New(logger).With().Timestamp().Logger(),
	}

	// Load Configuration.
	configuration, err := er.LoadConfiguration(ConfigurationPath)
	if err != nil {
		return nil, xerrors.Errorf("new errorly: %w", err)
	}

	if configuration.Host == "" {
		return nil, xerrors.New("No host provided")
	}

	er.Configuration = configuration

	// Create logging writers.
	var writers []io.Writer

	if er.Configuration.Logging.ConsoleLoggingEnabled {
		writers = append(writers, logger)
	}

	if er.Configuration.Logging.FileLoggingEnabled {
		if err := os.MkdirAll(er.Configuration.Logging.Directory, 0o744); err != nil {
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
	er.Logger = zerolog.New(mw).With().Timestamp().Logger().Level(level)
	er.Logger.Info().Msg("Logging configured")

	er.fs = &fasthttp.FS{
		Root:            "web/dist",
		Compress:        true,
		CompressBrotli:  true,
		AcceptByteRange: true,
		CacheDuration:   cacheDuration,
		PathNotFound:    fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {}),
	}
	er.distHandler = er.fs.NewRequestHandler()

	return er, err
}

// LoadConfiguration loads the configuration for RestTunnel.
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

// HandleRequest handles incoming HTTP requests.
func (er *Errorly) HandleRequest(ctx *fasthttp.RequestCtx) {
	start := time.Now()

	var processingMS int64

	defer func() {
		var log *zerolog.Event

		statusCode := ctx.Response.StatusCode()

		switch {
		case (statusCode >= 400 && statusCode <= 499):
			log = er.Logger.Warn()
		case (statusCode >= 500 && statusCode <= 599):
			log = er.Logger.Error()
		default:
			log = er.Logger.Info()
		}

		log.Msgf("%s %s %s %d %d %dms",
			ctx.RemoteAddr(),
			ctx.Request.Header.Method(),
			ctx.Request.URI().PathOriginal(),
			statusCode,
			len(ctx.Response.Body()),
			processingMS,
		)
	}()

	fasthttp.CompressHandlerBrotliLevel(func(ctx *fasthttp.RequestCtx) {
		fasthttpadaptor.NewFastHTTPHandler(er.Router)(ctx)
		if ctx.Response.StatusCode() != http.StatusNotFound {
			ctx.SetContentType("application/json;charset=utf8")

			return
		}
		// If there is no URL in router then try serving from the dist
		// folder.
		if ctx.Response.StatusCode() == http.StatusNotFound && gotils.B2S(ctx.Path()) != "/" {
			ctx.Response.Reset()
			er.distHandler(ctx)
		}
		// If there is no URL in router or in dist then send index.html
		if ctx.Response.StatusCode() == http.StatusNotFound {
			ctx.Response.Reset()
			ctx.SendFile("web/dist/index.html")
		}
	}, fasthttp.CompressBrotliDefaultCompression, fasthttp.CompressDefaultCompression)(ctx)

	processingMS = time.Since(start).Milliseconds()

	ctx.Response.Header.Set("X-Elapsed", strconv.FormatInt(processingMS, 10))
}

// Open starts the web worker.
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

	if er.Logger.GetLevel() <= 0 {
		er.Postgres.AddQueryHook(dbLogger{})
	}

	er.PostgressConn = er.Postgres.Conn()
	defer er.PostgressConn.Close()
	defer er.Postgres.Close()

	if err := er.PostgressConn.Ping(er.ctx); err != nil {
		return xerrors.Errorf("Failed to ping postgres: %w", err)
	}

	er.Logger.Info().Msg("Connected to postgres")

	er.Store = sessions.NewCookieStore([]byte(er.Configuration.SessionSecret))
	er.IDGen = idgenerator.NewIDGenerator(epoch, 0)

	// removeStaleEntries(er.Postgres)

	// er.Logger.Debug().Msg("Creating schema")
	// err = createSchema(er.Postgres)
	// if err != nil {
	// 	return err
	// }
	// er.Logger.Debug().Msg("Created schema")

	er.Logger.Debug().Msg("Creating endpoints")
	er.Router = createEndpoints(er)
	er.Logger.Debug().Msg("Created endpoints")

	fmt.Printf("Serving on %s (Press CTRL+C to quit)\n", er.Configuration.Host)
	err = fasthttp.ListenAndServe(er.Configuration.Host, er.HandleRequest)

	if err != nil {
		return xerrors.Errorf("Failed to listen and serve: %w", err)
	}

	return nil
}
