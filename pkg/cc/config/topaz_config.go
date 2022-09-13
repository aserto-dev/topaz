package config

import (
	"strings"

	authn_config "github.com/aserto-dev/aserto-grpc/authn/config"
	"github.com/aserto-dev/topaz/decision_log/logger/file"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Common         `json:",squash"`    // nolint:staticcheck // squash is used by mapstructure
	Auth           authn_config.Config `json:"auth"`
	DecisionLogger file.Config         `json:"decision_logger"`
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
		infoPath := authn_config.OptionOverrides{
			Paths:    []string{"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"},
			Override: authn_config.Options{NeedsTenant: false, EnableAPIKey: false, EnableMachineKey: false, EnableAuth0Token: false, EnableAnonymous: true},
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
