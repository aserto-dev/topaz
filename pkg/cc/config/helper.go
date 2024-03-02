package config

import (
	"path/filepath"

	builder "github.com/aserto-dev/service-host"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/samber/lo"
)

type currentConfig struct {
	*Loader
	lErr error
}

func CurrentConfig() *currentConfig {
	cfg, err := LoadConfiguration(filepath.Join(cc.GetTopazCfgDir(), "config.yaml"))
	if err != nil {
		return &currentConfig{Loader: nil, lErr: err}
	}

	return &currentConfig{Loader: cfg, lErr: nil}
}

func (c *currentConfig) Ports() ([]string, error) {
	if c.lErr != nil {
		return []string{}, c.lErr
	}

	return c.GetPorts()
}

func (c *currentConfig) Services() ([]string, error) {
	if c.lErr != nil {
		return []string{}, c.lErr
	}

	return lo.MapToSlice(c.Configuration.APIConfig.Services, func(k string, v *builder.API) string {
		return k
	}), nil
}

func (c *currentConfig) HealthService() (string, error) {
	if c.lErr != nil {
		return "", c.lErr
	}

	return c.Configuration.APIConfig.Health.ListenAddress, nil
}
