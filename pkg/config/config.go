package config

import (
	"encoding/json"
	"os"
	"text/template"

	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/config/handler"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/metrics"
	"github.com/aserto-dev/topaz/pkg/services"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/viper"
)

const Version int = 3

type Config struct {
	Version        int                   `json:"version" yaml:"version"`
	Logging        logger.Config         `json:"logging" yaml:"logging"`
	Authentication authentication.Config `json:"authentication,omitempty" yaml:"authentication,omitempty"`
	Debug          debug.Config          `json:"debug,omitempty" yaml:"debug,omitempty"`
	Health         health.Config         `json:"health,omitempty" yaml:"health,omitempty"`
	Metrics        metrics.Config        `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	Services       services.Config       `json:"services" yaml:"services"`
	Directory      directory.Config      `json:"directory" yaml:"directory"`
	Authorizer     authorizer.Config     `json:"authorizer" yaml:"authorizer"`
}

var _ = handler.Config(&Config{})

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault("version", 3)
	v.SetDefault("logging.prod", false)
	v.SetDefault("logging.log_level", "info")
	v.SetDefault("logging.grpc_log_level", "info")

	c.Authentication.SetDefaults(v, []string{"authentication"}...)

	c.Debug.SetDefaults(v, []string{"debug"}...)
	c.Health.SetDefaults(v, []string{"health"}...)
	c.Metrics.SetDefaults(v, []string{"metrics"}...)

	c.Services = map[string]*services.Service{"topaz": {}}
	c.Services.SetDefaults(v, []string{"services"}...)

	// c.Authorizer.SetDefaults(v, []string{"authorizer"}...)
	c.Directory.SetDefaults(v, []string{"directory"}...)
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

func (c *Config) Generate(w *os.File) error {
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

	return nil
}

type ConfigV3 struct {
	Version int           `json:"version" yaml:"version"`
	Logging logger.Config `json:"logging" yaml:"logging"`
}

var _ = handler.Config(&ConfigV3{})

func (c *ConfigV3) SetDefaults(v *viper.Viper, p ...string) {
}

func (c *ConfigV3) Validate() (bool, error) {
	return true, nil
}

func (c *ConfigV3) Generate(w *os.File) error {
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

func (c *ConfigV3) data() map[string]any {
	b, err := json.Marshal(c)
	if err != nil {
		return nil
	}

	v := map[string]any{}
	if err := json.Unmarshal(b, &v); err != nil {
		return nil
	}

	return v
}
