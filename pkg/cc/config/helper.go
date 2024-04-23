package config

import (
	builder "github.com/aserto-dev/service-host"
	"github.com/samber/lo"
)

type currentConfig struct {
	*Loader
	err error
}

func GetConfig(configFilePath string) *currentConfig {
	cfg, err := LoadConfiguration(configFilePath)
	if err != nil {
		return &currentConfig{Loader: nil, err: err}
	}

	return &currentConfig{Loader: cfg, err: nil}
}

func (c *currentConfig) Ports() ([]string, error) {
	if c.err != nil {
		return []string{}, c.err
	}

	return c.GetPorts()
}

func (c *currentConfig) Services() ([]string, error) {
	if c.err != nil {
		return []string{}, c.err
	}

	return lo.MapToSlice(c.Configuration.APIConfig.Services, func(k string, v *builder.API) string {
		return k
	}), nil
}

func (c *currentConfig) HealthService() (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.Configuration.APIConfig.Health.ListenAddress, nil
}
