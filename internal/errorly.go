package errorly

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"context"

	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// VERSION respects semantic versioning
const VERSION = "0.1"

// ConfigurationPath is the path to the file the configration will be located
// at.
const ConfigurationPath = "errorly.yaml"

// Configuration represents the configuration for Errorly
type Configuration struct {
	Host          string `json:"host" yaml:"host"`
	SessionSecret string `json:"secret" yaml:"secret"`

	Postgres *pg.Options `json:"postgres" yaml:"postgres"`

	// Postgres struct {
	// 	// Network               string
	// 	Addr     string `json:"addr" yaml:"addr"`
	// 	User     string `json:"user" yaml:"user"`
	// 	Password string `json:"password" yaml:"password"`
	// 	Database string `json:"database" yaml:"database"`
	// 	// ApplicationName       string
	// 	// TLSConfig             *tls.Config
	// 	// DialTimeout           time.Duration
	// 	// ReadTimeout           time.Duration
	// 	// WriteTimeout          time.Duration
	// 	// MaxRetries            int
	// 	// RetryStatementTimeout bool
	// 	// MinRetryBackoff       time.Duration
	// 	// MaxRetryBackoff       time.Duration
	// 	// PoolSize              int
	// 	// MinIdleConns          int
	// 	// MaxConnAge            time.Duration
	// 	// PoolTimeout           time.Duration
	// 	// IdleTimeout           time.Duration
	// 	// IdleCheckFrequency    time.Duration
	// } `json:"postgres" yaml:"postgres"`

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

	Postgres *pg.DB
	Router   *mux.Router
	Store    *sessions.CookieStore
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

	configuration, err := er.LoadConfiguration(ConfigurationPath)
	if err != nil {
		return nil, xerrors.Errorf("new errorly: %w", err)
	}
	if configuration.Host == "" {
		return nil, xerrors.New("No host provided")
	}
	er.Configuration = configuration

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

// Open starts the web worker
func (er *Errorly) Open() (err error) {
	return
}
