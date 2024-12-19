package config

import (
	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/config/handler"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/metrics"
	"github.com/aserto-dev/topaz/pkg/services"
	"github.com/spf13/viper"
)

type ConfigHandler interface {
	SetDefaults(v *viper.Viper, p ...string)
	Validate() (bool, error)
}

const Version int = 3

type Config struct {
	Version        int                   `json:"version" yaml:"version"`
	Logging        logger.Config         `json:"logging" yaml:"logging"`
	Authentication authentication.Config `json:"authentication,omitempty" yaml:"authentication,omitempty"`
	Debug          debug.Config          `json:"debug,omitempty" yaml:"debug,omitempty"`
	Health         health.Config         `json:"health,omitempty" yaml:"health,omitempty"`
	Metrics        metrics.Config        `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	Services       services.Config       `json:"services" yaml:"services"`
	Authorizer     authorizer.Config     `json:"authorizer" yaml:"authorizer"`
	Directory      directory.Config      `json:"directory" yaml:"directory"`
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

	c.Authorizer.SetDefaults(v, []string{"authorizer"}...)
	c.Directory.SetDefaults(v, []string{"directory"}...)
}

func (c *Config) Validate() (bool, error) {
	return true, nil
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
