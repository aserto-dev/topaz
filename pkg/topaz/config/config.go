package config

import (
	"fmt"
	"io"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-viper/mapstructure/v2"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/aserto-dev/logger"

	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/metrics"
	"github.com/aserto-dev/topaz/pkg/servers"
)

const Version int = 3

var ErrVersionMismatch = errors.Wrap(config.ErrConfig, "unsupported configuration version")

type Config struct {
	Version        int                   `json:"version"`
	Logging        logger.Config         `json:"logging"`
	Authentication authentication.Config `json:"authentication,omitempty"`
	Debug          debug.Config          `json:"debug,omitempty"`
	Health         health.Config         `json:"health,omitempty"`
	Metrics        metrics.Config        `json:"metrics,omitempty"`
	Servers        servers.Config        `json:"servers"`
	Directory      directory.Config      `json:"directory"`
	Authorizer     authorizer.Config     `json:"authorizer"`
}

var _ config.Section = (*Config)(nil)

type ConfigOverride func(*Config)

func NewConfig(r io.Reader, overrides ...ConfigOverride) (*Config, error) {
	var cfg Config

	v := config.NewViper()
	v.SetEnvPrefix("TOPAZ")
	v.AutomaticEnv()
	v.SetDefaults(&cfg)

	v.ReadConfig(r)

	if err := checkVersion(v); err != nil {
		return nil, err
	}

	if err := v.UnmarshalExact(&cfg, useJSONTags); err != nil {
		return nil, err
	}

	for _, override := range overrides {
		override(&cfg)
	}

	return &cfg, nil
}

func useJSONTags(dc *mapstructure.DecoderConfig) { dc.TagName = "json" }

//nolint:mnd  // this is where default values are defined.
func (c *Config) Defaults() map[string]any {
	srvrs := servers.Config{"topaz": {}}

	return lo.Assign(
		map[string]any{
			"version":                3,
			"logging.prod":           false,
			"logging.log_level":      "info",
			"logging.grpc_log_level": "warn",
		},
		config.PrefixKeys("authentication", c.Authentication.Defaults()),
		config.PrefixKeys("debug", c.Debug.Defaults()),
		config.PrefixKeys("health", c.Health.Defaults()),
		config.PrefixKeys("metrics", c.Metrics.Defaults()),
		config.PrefixKeys("servers", srvrs.Defaults()),
		config.PrefixKeys("directory", c.Directory.Defaults()),
		config.PrefixKeys("authorizer", c.Authorizer.Defaults()),
	)
}

const logLevelError = "invalid value %q in logging.log_level. expected one of [trace, debug, info, warn, error, fatal, panic]"

func (c *Config) Validate() error {
	// logging settings validation.
	if err := (&c.Logging).ParseLogLevel(zerolog.Disabled); err != nil {
		return errors.Wrapf(config.ErrConfig, logLevelError, c.Logging.LogLevel)
	}

	var errs error

	// Check that all config sections are valid before validating consistency across sections.
	for _, validator := range c.sections() {
		if err := validator.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs != nil {
		return errs
	}

	// All sections are valid. Check that they are consistent with each other.
	if c.Servers.DirectoryEnabled() == c.Directory.IsRemote() {
		errs = multierror.Append(errs, errors.Wrap(config.ErrConfig, "directory must be either remote or exposed locally"))
	}

	return errs
}

func (c *Config) Serialize(w io.Writer) error {
	cfgV3 := ConfigV3{Version: c.Version, Logging: c.Logging}

	if err := cfgV3.Serialize(w); err != nil {
		return err
	}

	for _, serializer := range c.sections() {
		if err := serializer.Serialize(w); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintln(w)

	return nil
}

func checkVersion(v config.Viper) error {
	var cfg struct {
		Version int `json:"version"`
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return err
	}

	if cfg.Version != Version {
		return errors.Wrapf(ErrVersionMismatch, "expected %d, got %d", Version, cfg.Version)
	}

	return nil
}

func (c *Config) sections() []config.Section {
	return []config.Section{&c.Authentication, &c.Debug, &c.Health, &c.Metrics, &c.Servers, &c.Directory, &c.Authorizer}
}

type ConfigV3 struct {
	Version int           `json:"version" yaml:"version"`
	Logging logger.Config `json:"logging" yaml:"logging"`
}

var _ config.Section = (*ConfigV3)(nil)

func (c *ConfigV3) Defaults() map[string]any {
	return map[string]any{}
}

func (c *ConfigV3) Validate() error {
	return nil
}

func (c *ConfigV3) Serialize(w io.Writer) error {
	{
		tmpl := template.Must(template.New("base").Funcs(sprig.FuncMap()).Parse(templateConfigHeader))

		// tmpl, err := template.
		// 	New("CONFIG_HEADER").
		// 	Funcs(sprig.TxtFuncMap()).
		// 	Parse(templateConfigHeader)
		// if err != nil {
		// 	return err
		// }

		if err := tmpl.Execute(w, c); err != nil {
			return err
		}
	}
	{
		tmpl, err := template.New("LOGGER").Parse(templateLogger)
		if err != nil {
			return err
		}

		var funcMap template.FuncMap = map[string]interface{}{}
		tmpl = tmpl.Funcs(sprig.TxtFuncMap()).Funcs(funcMap)

		if err := tmpl.Execute(w, c.Logging); err != nil {
			return err
		}
	}

	return nil
}

const templateConfigHeader = `
# yaml-language-server: $schema=https://topaz.sh/schema/config.json
---
# config schema version.
version: {{ .Version }}
`

const templateLogger = `
# logger settings.
logging:
  prod: {{ .Prod }}
  log_level: {{ .LogLevel }}
  grpc_log_level: {{ .GrpcLogLevel }}
`
