package decisionlog

import (
	"bytes"

	"github.com/aserto-dev/topaz/decisionlog"
	"github.com/go-viper/mapstructure/v2"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/util"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type PluginFactory struct {
	logger decisionlog.DecisionLogger
}

var _ plugins.Factory = (*PluginFactory)(nil)

func NewFactory(logger decisionlog.DecisionLogger) PluginFactory {
	return PluginFactory{
		logger: logger,
	}
}

func (f PluginFactory) New(m *plugins.Manager, config any) plugins.Plugin {
	cfg, ok := config.(*Config)
	if !ok {
		return &DecisionLogsPlugin{}
	}

	return newDecisionLogger(cfg, m, f.logger)
}

func (PluginFactory) Validate(m *plugins.Manager, config []byte) (any, error) {
	parsedConfig := Config{}

	v := viper.New()
	v.SetConfigType("json")

	err := v.ReadConfig(bytes.NewReader(config))
	if err != nil {
		return nil, errors.Wrap(err, "error parsing decision logs config")
	}

	if err := v.UnmarshalExact(&parsedConfig, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	}); err != nil {
		return nil, errors.Wrap(err, "error parsing decision logs config")
	}

	return &parsedConfig, util.Unmarshal(config, &parsedConfig)
}
