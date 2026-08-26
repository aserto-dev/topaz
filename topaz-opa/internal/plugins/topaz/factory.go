package topaz

import (
	"sync/atomic"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/topaz-opa/internal/config"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/util"
)

type PluginFactory struct{}

func NewFactory() *PluginFactory {
	// set a default config for the plugin, in case there is no OPA config section for the plugin.
	SetConfig(Config{Enabled: false})

	return &PluginFactory{}
}

var (
	_    plugins.Factory = (*PluginFactory)(nil)
	aCfg atomic.Pointer[Config]
)

func SetConfig(c Config) {
	aCfg.Store(&c)
}

func GetConfig() Config {
	c := aCfg.Load()
	return *c
}

func (p *PluginFactory) New(manager *plugins.Manager, cfg any) plugins.Plugin {
	log := manager.Logger()

	c, ok := cfg.(Config)
	if !ok {
		log.Error("failed to parse topaz plugin config")
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
	log := manager.Logger()

	var parsedConfig Config
	if err := util.Unmarshal(cfg, &parsedConfig); err != nil {
		log.Error("failed to unmarshal topaz plugin config (%v)", err)
		return nil, err
	}

	SetConfig(parsedConfig)
	log.Info("topaz plugin enabled = %t", parsedConfig.Enabled)

	return parsedConfig, nil
}

func Factory(m *plugins.Manager, cfg any) plugins.Plugin {
	defaultConfig := &Config{
		Enabled: false,
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
		RequestTimeout: config.Duration(DefaultRequestTimeout),
	}

	if c, ok := cfg.(*Config); ok {
		defaultConfig = c
	}

	return &Plugin{manager: m, config: defaultConfig}
}
