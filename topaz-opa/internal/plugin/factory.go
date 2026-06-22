package plugin

import (
	"sync/atomic"

	"github.com/aserto-dev/topaz/topaz-opa/internal/config"

	"github.com/aserto-dev/go-aserto"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/util"
)

type PluginFactory struct{}

var (
	_   plugins.Factory = (*PluginFactory)(nil)
	cfg atomic.Value
)

func SetConfig(c config.Config) {
	cfg.Store(c)
}

func GetConfig() config.Config {
	return cfg.Load().(config.Config) //nolint:forcetypeassert
}

func (p *PluginFactory) New(manager *plugins.Manager, cfg any) plugins.Plugin {
	c, ok := cfg.(config.Config)
	if !ok {
		// panic as the plugins.Factory interface definition of New does ot provide an error return,
		// nor does the OPA implementation not handle nil plugins.
		// NOTE that Validate() is called before New, mitigating the risk of the panic occurring.
		panic("failed to parse topaz plugin config")
	}

	return &Plugin{
		config:  &c,
		manager: manager,
	}
}

func (p *PluginFactory) Validate(manager *plugins.Manager, cfg []byte) (any, error) {
	var parsedConfig config.Config
	if err := util.Unmarshal(cfg, &parsedConfig); err != nil {
		return nil, err
	}

	SetConfig(parsedConfig)

	return parsedConfig, nil
}

func Factory(m *plugins.Manager, cfg any) plugins.Plugin {
	defaultConfig := &config.Config{
		Enabled: true,
		Connection: aserto.Config{
			Address:        "localhost:9292",
			APIKey:         "",
			Token:          "",
			ClientCertPath: "",
			ClientKeyPath:  "",
			CACertPath:     "",
			Insecure:       true,
			NoTLS:          false,
			NoProxy:        false,
			Headers:        map[string]string{},
		},
		RequestTimeout:          config.Duration(config.DefaultRequestTimeout),
		EnableDirectoryBuiltIns: true,
		EnableAccessBuiltIns:    true,
	}

	if c, ok := cfg.(*config.Config); ok {
		defaultConfig = c
	}

	return &Plugin{manager: m, config: defaultConfig}
}
