package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/runtime"
	builder "github.com/aserto-dev/service-host"
)

// CommandMode -- enum type.
type CommandMode int

var (
	DefaultTLSGenDir = os.ExpandEnv("$HOME/.config/topaz/certs")
	CertificateSets  = []string{"grpc", "gateway"}
)

// CommandMode -- enum constants.
const (
	CommandModeUnknown CommandMode = 0 + iota
	CommandModeRun
	CommandModeBuild
)

type ServicesConfig struct {
	Health struct {
		ListenAddress string                `json:"listen_address"`
		Certificates  *certs.TLSCredsConfig `json:"certs"`
	} `json:"health"`
	Metrics struct {
		ListenAddress string                `json:"listen_address"`
		Certificates  *certs.TLSCredsConfig `json:"certs"`
		ZPages        bool                  `json:"zpages"`
	} `json:"metrics"`
	Services map[string]*builder.API `json:"services"`
}

// Config holds the configuration for the app.
type Common struct {
	Version int           `json:"version"`
	Logging logger.Config `json:"logging"`

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

	v.SetDefault("opa.max_plugin_wait_time_seconds", "30")

	v.SetDefault("remote_directory.address", "0.0.0.0:9292")
	v.SetDefault("remote_directory.insecure", "true")

	defaults(v)

	configExists, err := fileExists(file)
	if err != nil {
		return nil, errors.Wrapf(err, "filesystem error")
	}

	if configExists {
		buf, err := os.ReadFile(v.ConfigFileUsed())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read config file '%s'", file)
		}
		subBuf := subEnvVars(string(buf))
		r := bytes.NewReader([]byte(subBuf))

		if err := v.ReadConfig(r); err != nil {
			return nil, errors.Wrapf(err, "failed to parse config file '%s'", file)
		}

		err = validateVersion(v.GetInt("version"))
		if err != nil {
			return nil, err
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

	// This is where validation of config happens.
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

	err = setDefaultCerts(cfg)
	if err != nil {
		return nil, err
	}

	if certsGenerator != nil {
		err = cfg.setupCerts(log, certsGenerator)
		if err != nil {
			return nil, errors.Wrap(err, "failed to setup certs")
		}
	}

	return cfg, nil
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
	existingFiles := []string{}
	for serviceName, config := range c.APIConfig.Services {
		log.Info().Msgf("setting up certs for %s", serviceName)
		for _, file := range []string{
			config.GRPC.Certs.TLSCACertPath,
			config.GRPC.Certs.TLSCertPath,
			config.GRPC.Certs.TLSKeyPath,
			config.Gateway.Certs.TLSCACertPath,
			config.Gateway.Certs.TLSCertPath,
			config.Gateway.Certs.TLSKeyPath,
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
				CommonName:       fmt.Sprintf("%s-grpc", serviceName),
				CertKeyPath:      config.GRPC.Certs.TLSKeyPath,
				CertPath:         config.GRPC.Certs.TLSCertPath,
				CACertPath:       config.GRPC.Certs.TLSCACertPath,
				DefaultTLSGenDir: DefaultTLSGenDir,
			})
			if err != nil {
				return errors.Wrap(err, "failed to generate grpc certs")
			}

			err = certsGenerator.MakeDevCert(&certs.CertGenConfig{
				CommonName:       fmt.Sprintf("%s-gateway", serviceName),
				CertKeyPath:      config.Gateway.Certs.TLSKeyPath,
				CertPath:         config.Gateway.Certs.TLSCertPath,
				CACertPath:       config.Gateway.Certs.TLSCACertPath,
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

var envRegex = regexp.MustCompile(`(?U:\${.*})`)

// subEnvVars will look for any environment variables in the passed in string
// with the syntax of ${VAR_NAME} and replace that string with ENV[VAR_NAME].
func subEnvVars(s string) string {
	updatedConfig := envRegex.ReplaceAllStringFunc(s, func(s string) string {
		// Trim off the '${' and '}'
		if len(s) <= 3 {
			// This should never happen..
			return ""
		}
		varName := s[2 : len(s)-1]

		// Lookup the variable in the environment. We play by
		// bash rules.. if its undefined we'll treat it as an
		// empty string instead of raising an error.
		return os.Getenv(varName)
	})

	return updatedConfig
}

func setDefaultCerts(cfg *Config) error {
	for srvName, config := range cfg.APIConfig.Services {
		if config.GRPC.ListenAddress == "" {
			return errors.New(fmt.Sprintf("%s must have a grpc listen address specified", srvName))
		}
		if config.GRPC.Certs.TLSCACertPath == "" || config.GRPC.Certs.TLSCertPath == "" || config.GRPC.Certs.TLSKeyPath == "" {
			config.GRPC.Certs.TLSKeyPath = filepath.Join(DefaultTLSGenDir, fmt.Sprintf("%s_grpc.key", srvName))
			config.GRPC.Certs.TLSCertPath = filepath.Join(DefaultTLSGenDir, fmt.Sprintf("%s_grpc.crt", srvName))
			config.GRPC.Certs.TLSCACertPath = filepath.Join(DefaultTLSGenDir, fmt.Sprintf("%s_grpc-ca.crt", srvName))
		}
		if config.Gateway.Certs.TLSCACertPath == "" || config.Gateway.Certs.TLSCertPath == "" || config.Gateway.Certs.TLSKeyPath == "" {
			config.Gateway.Certs.TLSKeyPath = filepath.Join(DefaultTLSGenDir, fmt.Sprintf("%s_gateway.key", srvName))
			config.Gateway.Certs.TLSCertPath = filepath.Join(DefaultTLSGenDir, fmt.Sprintf("%s_gateway.crt", srvName))
			config.Gateway.Certs.TLSCACertPath = filepath.Join(DefaultTLSGenDir, fmt.Sprintf("%s_gateway-ca.crt", srvName))
		}
	}
	return nil
}
