package config

import (
	"fmt"
	"strings"

	"github.com/aserto-dev/aserto-management/controller"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Common           `json:",squash"`   // nolint:staticcheck // squash is used by mapstructure
	Auth             AuthnConfig        `json:"auth"`
	DecisionLogger   DecisionLogConfig  `json:"decision_logger"`
	ControllerConfig *controller.Config `json:"controller"`
}

type DecisionLogConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type AuthnConfig struct {
	APIKeys map[string]string `json:"api_keys"`
	Options CallOptions       `json:"options"`
}

type CallOptions struct {
	Default   Options           `json:"default"`
	Overrides []OptionOverrides `json:"overrides"`
}

type Options struct {

	// API Key for machine-to-machine communication, internal to Aserto
	EnableAPIKey bool `json:"enable_api_key"`
	// Allows calls without any form of authentication
	EnableAnonymous bool `json:"enable_anonymous"`
}

type OptionOverrides struct {
	// API paths to override
	Paths []string `json:"paths"`
	// Override options
	Override Options `json:"override"`
}

func (co *CallOptions) ForPath(path string) *Options {
	for _, override := range co.Overrides {
		for _, prefix := range override.Paths {
			if strings.HasPrefix(strings.ToLower(path), prefix) {
				return &override.Override
			}
		}
	}

	return &co.Default
}

func defaults(v *viper.Viper) {
}

func (c *Config) validation() error {
	if c.Command.Mode == CommandModeRun && c.OPA.InstanceID == "" {
		return errors.New("opa.instance_id not set")
	}
	if len(c.OPA.Config.Bundles) > 1 {
		return errors.New("opa.config.bundles - too many bundles")
	}

	if len(c.Services) == 0 {
		return errors.New("no api services configured")
	}

	for key := range c.Services {
		if _, ok := ServiceTypeMap()[key]; !ok {
			return errors.New(fmt.Sprintf("unknown service type configuration %s", key))
		}
	}

	setDefaultCallsAuthz(c)

	if len(c.Auth.APIKeys) > 0 {
		c.Auth.Options.Default.EnableAPIKey = true
	} else {
		c.Auth.Options.Default.EnableAnonymous = true
	}

	return nil
}

func setDefaultCallsAuthz(cfg *Config) {
	if len(cfg.Auth.Options.Overrides) == 0 {
		infoPath := OptionOverrides{
			Paths:    []string{"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"},
			Override: Options{EnableAPIKey: false, EnableAnonymous: true},
		}
		cfg.Auth.Options.Overrides = append(cfg.Auth.Options.Overrides, infoPath)
	}

	// We unfortunately have to use a pipe '|' delimiter for these keys in the config file
	// and fix them up once we load the config, because of this bug:
	// https://github.com/spf13/viper/issues/324
	// Keys also become lowercase
	for i := 0; i < len(cfg.Auth.Options.Overrides); i++ {
		for j := 0; j < len(cfg.Auth.Options.Overrides[i].Paths); j++ {
			cfg.Auth.Options.Overrides[i].Paths[j] = strings.ToLower(strings.ReplaceAll(cfg.Auth.Options.Overrides[i].Paths[j], "|", "."))
		}
	}
}
