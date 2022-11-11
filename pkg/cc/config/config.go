package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/directory"
)

// CommandMode -- enum type
type CommandMode int

var (
	DefaultTLSGenDir = os.ExpandEnv("$HOME/.config/topaz/certs")
	CertificateSets  = []string{"grpc", "gateway"}
)

// CommandMode -- enum constants
const (
	CommandModeUnknown CommandMode = 0 + iota
	CommandModeRun
	CommandModeBuild
)

// Config holds the configuration for the app.
type Common struct {
	Logging struct {
		Prod           bool          `json:"prod"`
		LogLevel       string        `json:"log_level"`
		LogLevelParsed zerolog.Level `json:"-"`
	} `json:"logging"`

	Command struct {
		Mode CommandMode
	} `json:"-"`

	API struct {
		GRPC struct {
			ListenAddress string `json:"listen_address"`
			// Default connection timeout is 120 seconds
			// https://godoc.org/google.golang.org/grpc#ConnectionTimeout
			ConnectionTimeoutSeconds uint32               `json:"connection_timeout_seconds"`
			Certs                    certs.TLSCredsConfig `json:"certs"`
		} `json:"grpc"`
		Gateway struct {
			ListenAddress  string               `json:"listen_address"`
			AllowedOrigins []string             `json:"allowed_origins"`
			Certs          certs.TLSCredsConfig `json:"certs"`
			HTTP           bool                 `json:"http"`
		} `json:"gateway"`
		Health struct {
			ListenAddress string `json:"listen_address"`
		} `json:"health"`
	} `json:"api"`

	JWT struct {
		// Specifies the duration in which exp (Expiry) and nbf (Not Before)
		// claims may differ by. This value should be positive.
		AcceptableTimeSkewSeconds int `json:"acceptable_time_skew_seconds"`
	} `json:"jwt"`

	// Directory configuration
	Directory directory.Config `json:"directory_service"`

	// Default OPA configuration
	OPA runtime.Config `json:"opa"`
}

// LoggerConfig is a basic Config copy that gets loaded before everything else,
// so we can log during resolving configuration
type LoggerConfig Config

// Path represents the path to a configuration file
type Path string

// Overrider is a func that mutates configuration
type Overrider func(*Config)

// NewConfig creates the configuration by reading env & files
func NewConfig(configPath Path, log *zerolog.Logger, overrides Overrider, certsGenerator *certs.Generator) (*Config, error) { // nolint:funlen // default list of values can be long
	newLogger := log.With().Str("component", "config").Logger()
	log = &newLogger
	v := viper.New()

	file := "config.yaml"
	if configPath != "" {
		exists, err := fileExists(string(configPath))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to determine if config file '%s' exists", configPath)
		}

		if !exists {
			return nil, errors.Errorf("config file '%s' doesn't exist", configPath)
		}

		file = string(configPath)
	}

	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetConfigFile(file)
	v.SetEnvPrefix("TOPAZ")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Set defaults
	v.SetDefault("jwt.acceptable_time_skew_seconds", 5)
	for _, svc := range CertificateSets {
		v.SetDefault(fmt.Sprintf("api.%s.certs.tls_key_path", svc), filepath.Join(DefaultTLSGenDir, svc+".key"))
		v.SetDefault(fmt.Sprintf("api.%s.certs.tls_cert_path", svc), filepath.Join(DefaultTLSGenDir, svc+".crt"))
		v.SetDefault(fmt.Sprintf("api.%s.certs.tls_ca_cert_path", svc), filepath.Join(DefaultTLSGenDir, svc+"-ca.crt"))
	}
	v.SetDefault("api.grpc.connection_timeout_seconds", 120)
	v.SetDefault("api.grpc.listen_address", "0.0.0.0:8282")
	v.SetDefault("api.gateway.listen_address", "0.0.0.0:8383")
	v.SetDefault("api.gateway.http", false)
	v.SetDefault("api.health.listen_address", "0.0.0.0:8484")
	v.SetDefault("opa.max_plugin_wait_time_seconds", "30")

	defaults(v)

	configExists, err := fileExists(file)
	if err != nil {
		return nil, errors.Wrapf(err, "filesystem error")
	}

	if configExists {
		if err := v.ReadInConfig(); err != nil {
			return nil, errors.Wrapf(err, "failed to read config file '%s'", file)
		}
	}
	v.AutomaticEnv()

	cfg := new(Config)

	err = v.UnmarshalExact(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config file")
	}

	if overrides != nil {
		overrides(cfg)
	}

	// This is where validation of config happens
	err = func() error {
		var err error

		if cfg.Logging.LogLevel == "" {
			cfg.Logging.LogLevelParsed = zerolog.InfoLevel
		} else {
			cfg.Logging.LogLevelParsed, err = zerolog.ParseLevel(cfg.Logging.LogLevel)
			if err != nil {
				return errors.Wrapf(err, "logging.log_level failed to parse")
			}
		}

		if cfg.JWT.AcceptableTimeSkewSeconds < 0 {
			return errors.New("jwt.acceptable_time_skew_seconds must be positive or 0")
		}

		return cfg.validation()
	}()

	if err != nil {
		return nil, errors.Wrap(err, "failed to validate config file")
	}

	if certsGenerator != nil {
		err = cfg.setupCerts(log, certsGenerator)
		if err != nil {
			return nil, errors.Wrap(err, "failed to setup certs")
		}
	}

	return cfg, nil
}

// NewLoggerConfig creates a new LoggerConfig
func NewLoggerConfig(configPath Path, overrides Overrider) (*logger.Config, error) {
	discardLogger := zerolog.New(io.Discard)

	cfg, err := NewConfig(configPath, &discardLogger, overrides, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new config")
	}

	lCfg := logger.Config{
		Prod:           cfg.Logging.Prod,
		LogLevel:       cfg.Logging.LogLevel,
		LogLevelParsed: cfg.Logging.LogLevelParsed,
	}
	return &lCfg, nil
}

func (c *Config) setupCerts(log *zerolog.Logger, certsGenerator *certs.Generator) error {
	existingFiles := []string{}
	for _, file := range []string{
		c.API.GRPC.Certs.TLSCACertPath,
		c.API.GRPC.Certs.TLSCertPath,
		c.API.GRPC.Certs.TLSKeyPath,
		c.API.Gateway.Certs.TLSCACertPath,
		c.API.Gateway.Certs.TLSCertPath,
		c.API.Gateway.Certs.TLSKeyPath,
	} {
		exists, err := fileExists(file)
		if err != nil {
			return errors.Wrapf(err, "failed to determine if file '%s' exists", file)
		}

		if !exists {
			continue
		}

		existingFiles = append(existingFiles, file)
	}

	if len(existingFiles) == 0 {
		err := certsGenerator.MakeDevCert(&certs.CertGenConfig{
			CommonName:       "authorizer-grpc",
			CertKeyPath:      c.API.GRPC.Certs.TLSKeyPath,
			CertPath:         c.API.GRPC.Certs.TLSCertPath,
			CACertPath:       c.API.GRPC.Certs.TLSCACertPath,
			DefaultTLSGenDir: DefaultTLSGenDir,
		})
		if err != nil {
			return errors.Wrap(err, "failed to generate grpc certs")
		}

		err = certsGenerator.MakeDevCert(&certs.CertGenConfig{
			CommonName:       "authorizer-gateway",
			CertKeyPath:      c.API.Gateway.Certs.TLSKeyPath,
			CertPath:         c.API.Gateway.Certs.TLSCertPath,
			CACertPath:       c.API.Gateway.Certs.TLSCACertPath,
			DefaultTLSGenDir: DefaultTLSGenDir,
		})
		if err != nil {
			return errors.Wrap(err, "failed to generate gateway certs")
		}
	} else {
		msg := zerolog.Arr()
		for _, f := range existingFiles {
			msg.Str(f)
		}
		log.Info().Array("existing-files", msg).Msg("some cert files already exist, skipping generation")
	}

	return nil
}

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}
