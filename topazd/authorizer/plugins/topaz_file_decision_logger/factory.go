package topaz_file_decision_logger

import (
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/util"
)

type PluginFactory struct{}

var _ plugins.Factory = (*PluginFactory)(nil)

func NewFactory() PluginFactory {
	return PluginFactory{}
}

func (f PluginFactory) New(manager *plugins.Manager, cfg any) plugins.Plugin {
	c, ok := cfg.(Config)
	if !ok {
		// panic as the plugins.Factory interface definition of New does ot provide an error return,
		// nor does the OPA implementation not handle nil plugins.
		// NOTE that Validate() is called before New, mitigating the risk of the panic occurring.
		panic("failed to parse authzen logger plugin config")
	}

	return &Plugin{
		config:  &c,
		manager: manager,
	}
}

func (p PluginFactory) Validate(manager *plugins.Manager, cfg []byte) (any, error) {
	var parsedConfig Config
	if err := util.Unmarshal(cfg, &parsedConfig); err != nil {
		return nil, err
	}

	return parsedConfig, nil
}

func Factory(m *plugins.Manager, cfg any) plugins.Plugin {
	defaultConfig := defaultConfig()

	if c, ok := cfg.(*Config); ok {
		defaultConfig = c
	}

	return &Plugin{manager: m, config: defaultConfig}
}
