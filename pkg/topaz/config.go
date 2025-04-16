package topaz

import (
	"fmt"
	"io"
	"text/template"

	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/metrics"
	"github.com/aserto-dev/topaz/pkg/services"
	"github.com/samber/lo"

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

var _ config.Section = (*Config)(nil)

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
		config.PrefixKeys("authentication", c.Authentication.Defaults()),
		config.PrefixKeys("debug", c.Debug.Defaults()),
		config.PrefixKeys("health", c.Health.Defaults()),
		config.PrefixKeys("metrics", c.Metrics.Defaults()),
		config.PrefixKeys("services", services.Defaults()),
		config.PrefixKeys("directory", c.Directory.Defaults()),
	)
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Serialize(w io.Writer) error {
	cfgV3 := ConfigV3{Version: c.Version, Logging: c.Logging}

	if err := cfgV3.Serialize(w); err != nil {
		return err
	}

	if err := c.Authentication.Serialize(w); err != nil {
		return err
	}

	if err := c.Debug.Serialize(w); err != nil {
		return err
	}

	if err := c.Health.Serialize(w); err != nil {
		return err
	}

	if err := c.Metrics.Serialize(w); err != nil {
		return err
	}

	if err := c.Services.Serialize(w); err != nil {
		return err
	}

	if err := c.Directory.Serialize(w); err != nil {
		return err
	}

	if err := c.Authorizer.Serialize(w); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(w)

	return nil
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
