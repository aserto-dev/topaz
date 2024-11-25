package config

import (
	"fmt"
	"io"
	"os"

	"github.com/aserto-dev/certs"
	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/pkg/debug"
	builder "github.com/aserto-dev/topaz/pkg/service/builder"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// CommandMode -- enum type.
type CommandMode int

var CertificateSets = []string{"grpc", "gateway"}

// CommandMode -- enum constants.
const (
	CommandModeUnknown CommandMode = 0 + iota
	CommandModeRun
	CommandModeBuild
)

type ServicesConfig struct {
	Health struct {
		ListenAddress string            `json:"listen_address"`
		Certificates  *client.TLSConfig `json:"certs"`
	} `json:"health"`
	Metrics struct {
		ListenAddress string            `json:"listen_address"`
		Certificates  *client.TLSConfig `json:"certs"`
		ZPages        bool              `json:"zpages"`
	} `json:"metrics"`
	Services map[string]*builder.API `json:"services"`
}

// Config holds the configuration for the app.
type Common struct {
	Version      int           `json:"version"`
	Logging      logger.Config `json:"logging"`
	DebugService debug.Config  `json:"debug_service"`

	Command struct {
		Mode CommandMode
	} `json:"-"`

	APIConfig ServicesConfig `json:"api"`

	JWT struct {
		// Specifies the duration in which exp (Expiry) and nbf (Not Before)
		// claims may differ by. This value should be positive.
		AcceptableTimeSkewSeconds int `json:"acceptable_time_skew_seconds"`
	} `json:"jwt"`

	// Directory configuration
	Edge directory.Config `json:"directory"`

	// Authorizer directory resolver configuration
	DirectoryResolver client.Config `json:"remote_directory"`

	// Default OPA configuration
	OPA runtime.Config `json:"opa"`
}

// LoggerConfig is a basic Config copy that gets loaded before everything else,
// so we can log during resolving configuration.
type LoggerConfig Config

// Path represents the path to a configuration file.
type Path string

// Overrider is a func that mutates configuration.
type Overrider func(*Config)

// NewConfig creates the configuration by reading env & files.
func NewConfig(configPath Path, log *zerolog.Logger, overrides Overrider, certsGenerator *certs.Generator) (*Config, error) { // nolint:funlen // default list of values can be long
	newLogger := log.With().Str("component", "config").Logger()
	log = &newLogger

	file := "config.yaml"
	if configPath != "" {
		exists, err := FileExists(string(configPath))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to determine if config file '%s' exists", configPath)
		}

		if !exists {
			return nil, errors.Errorf("config file '%s' doesn't exist", configPath)
		}

		file = string(configPath)
	}

	configExists, err := FileExists(file)
	if err != nil {
		return nil, errors.Wrapf(err, "filesystem error")
	}

	configLoader := new(Loader)

	if configExists {
		configLoader, err = LoadConfiguration(file)
		if err != nil {
			return nil, err
		}
		if configLoader.HasTopazDir {
			log.Warn().Msg("This configuration file still uses TOPAZ_DIR environment variable. Please change to using the new TOPAZ_DB_DIR and TOPAZ_CERTS_DIR environment variables.")
		}

		err = validateVersion(configLoader.Configuration.Version)
		if err != nil {
			return nil, err
		}
	}

	if overrides != nil {
		overrides(configLoader.Configuration)
	}

	// This is where validation of config happens.
	err = func() error {
		var err error

		if configLoader.Configuration.Logging.LogLevel == "" {
			configLoader.Configuration.Logging.LogLevelParsed = zerolog.InfoLevel
		} else {
			configLoader.Configuration.Logging.LogLevelParsed, err = zerolog.ParseLevel(configLoader.Configuration.Logging.LogLevel)
			if err != nil {
				return errors.Wrapf(err, "logging.log_level failed to parse")
			}
		}

		if configLoader.Configuration.JWT.AcceptableTimeSkewSeconds < 0 {
			return errors.New("jwt.acceptable_time_skew_seconds must be positive or 0")
		}

		return configLoader.Configuration.validation()
	}()
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate config file")
	}

	if certsGenerator != nil {
		err = configLoader.Configuration.setupCerts(log, certsGenerator)
		if err != nil {
			return nil, errors.Wrap(err, "failed to setup certs")
		}
	}

	return configLoader.Configuration, nil
}

// NewLoggerConfig creates a new LoggerConfig.
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
	commonName := "topaz"

	existingFiles := []string{}
	for serviceName, config := range c.APIConfig.Services {
		for _, file := range []string{
			config.GRPC.Certs.CA,
			config.GRPC.Certs.Cert,
			config.GRPC.Certs.Key,
			config.Gateway.Certs.CA,
			config.Gateway.Certs.Cert,
			config.Gateway.Certs.Key,
		} {
			exists, err := FileExists(file)
			if err != nil {
				return errors.Wrapf(err, "failed to determine if file '%s' exists (%s)", file, serviceName)
			}

			if !exists {
				continue
			}

			existingFiles = append(existingFiles, file)
		}

		if len(existingFiles) == 0 {
			if config.GRPC.Certs.HasCert() && config.GRPC.Certs.HasCA() {
				err := certsGenerator.MakeDevCert(&certs.CertGenConfig{
					CommonName:  fmt.Sprintf("%s-grpc", commonName),
					CertKeyPath: config.GRPC.Certs.Key,
					CertPath:    config.GRPC.Certs.Cert,
					CertCAPath:  config.GRPC.Certs.CA,
				})
				if err != nil {
					return errors.Wrapf(err, "failed to generate grpc certs (%s)", serviceName)
				}
				log.Info().Str("service", serviceName).Msg("gRPC certs configured")
			}

			if config.Gateway.Certs.HasCert() && config.Gateway.Certs.HasCA() {
				err := certsGenerator.MakeDevCert(&certs.CertGenConfig{
					CommonName:  fmt.Sprintf("%s-gateway", commonName),
					CertKeyPath: config.Gateway.Certs.Key,
					CertPath:    config.Gateway.Certs.Cert,
					CertCAPath:  config.Gateway.Certs.CA,
				})
				if err != nil {
					return errors.Wrapf(err, "failed to generate gateway certs (%s)", serviceName)
				}
				log.Info().Str("service", serviceName).Msg("gateway certs configured")
			}
		}
	}
	return nil
}

func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}
