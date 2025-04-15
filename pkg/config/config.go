package config

import (
	"fmt"
	"io"
	"text/template"

	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/config/directory"
	"github.com/aserto-dev/topaz/pkg/config/handler"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/metrics"
	"github.com/aserto-dev/topaz/pkg/services"
	"github.com/samber/lo"
	"github.com/spf13/viper"

	"github.com/Masterminds/sprig/v3"
)

const Version int = 3

type Config struct {
	Version        int                   `json:"version"`
	Logging        logger.Config         `json:"logging"`
	Authentication authentication.Config `json:"authentication,omitempty"`
	Debug          debug.Config          `json:"debug,omitempty"`
	Health         health.Config         `json:"health,omitempty"`
	Metrics        metrics.Config        `json:"metrics,omitempty"`
	Services       services.Config       `json:"services"`
	Directory      directory.Config      `json:"directory"`
	Authorizer     authorizer.Config     `json:"authorizer"`
}

var _ handler.Config = (*Config)(nil)

//nolint:mnd  // this is where default values are defined.
func (c *Config) Defaults() map[string]any {
	services := services.Config{"topaz": {}}

	return lo.Assign(
		map[string]any{
			"version":                3,
			"logging.prod":           false,
			"logging.log_level":      "info",
			"logging.grpc_log_level": "info",
		},
		handler.PrefixKeys("authentication", c.Authentication.Defaults()),
		handler.PrefixKeys("debug", c.Debug.Defaults()),
		handler.PrefixKeys("health", c.Health.Defaults()),
		handler.PrefixKeys("metrics", c.Metrics.Defaults()),
		handler.PrefixKeys("services", services.Defaults()),
		handler.PrefixKeys("directory", c.Directory.Defaults()),
	)
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

func (c *Config) Generate(w io.Writer) error {
	cfgV3 := ConfigV3{Version: c.Version, Logging: c.Logging}

	if err := cfgV3.Generate(w); err != nil {
		return err
	}

	if err := c.Authentication.Generate(w); err != nil {
		return err
	}

	if err := c.Debug.Generate(w); err != nil {
		return err
	}

	if err := c.Health.Generate(w); err != nil {
		return err
	}

	if err := c.Metrics.Generate(w); err != nil {
		return err
	}

	if err := c.Services.Generate(w); err != nil {
		return err
	}

	if err := c.Directory.Generate(w); err != nil {
		return err
	}

	if err := c.Authorizer.Generate(w); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(w)

	return nil
}

func SetDefaults(v *viper.Viper) {
	c := Config{}

	for key, value := range c.Defaults() {
		v.SetDefault(key, value)
	}
}

type ConfigV3 struct {
	Version int           `json:"version" yaml:"version"`
	Logging logger.Config `json:"logging" yaml:"logging"`
}

var _ handler.Config = (*ConfigV3)(nil)

func (c *ConfigV3) Defaults() map[string]any {
	return map[string]any{}
}

func (c *ConfigV3) Validate() (bool, error) {
	return true, nil
}

func (c *ConfigV3) Generate(w io.Writer) error {
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
