package topaz_file_decision_logger

import (
	"context"

	"github.com/aserto-dev/topaz/topazd/authorizer/plugins/noop"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/util"
	"github.com/rs/zerolog"
)

type PluginFactory struct {
	logger *zerolog.Logger
}

var _ plugins.Factory = (*PluginFactory)(nil)

func NewFactory(ctx context.Context) PluginFactory {
	newLogger := zerolog.Ctx(ctx).With().Str("component", PluginName).Logger()

	return PluginFactory{
		logger: &newLogger,
	}
}

func (f PluginFactory) New(pm *plugins.Manager, config any) plugins.Plugin {
	cfg, ok := config.(Config)
	if !ok {
		// panic as the plugins.Factory interface definition of New does ot provide an error return,
		// nor does the OPA implementation not handle nil plugins.
		// NOTE that Validate() is called before New, mitigating the risk of the panic occurring.
		panic("failed to parse authzen logger plugin config")
	}

	if !cfg.Enabled {
		return &noop.Noop{
			Manager: pm,
			Name:    PluginName,
		}
	}

	return &Plugin{
		config:  &cfg,
		logger:  f.logger,
		manager: pm,
	}
}

func (p PluginFactory) Validate(pm *plugins.Manager, cfg []byte) (any, error) {
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
