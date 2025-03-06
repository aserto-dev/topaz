package decisionlog

import (
	"bytes"

	decisionlog "github.com/aserto-dev/topaz/decision_log"
	"github.com/mitchellh/mapstructure"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/util"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Factory struct {
	logger decisionlog.DecisionLogger
}

func NewFactory(logger decisionlog.DecisionLogger) Factory {
	return Factory{
		logger: logger,
	}
}

func (f Factory) New(m *plugins.Manager, config interface{}) plugins.Plugin {
	cfg, ok := config.(*Config)
	if !ok {
		panic("cast failed")
	}
	return newDecisionLogger(cfg, m, f.logger)
}

func (Factory) Validate(m *plugins.Manager, config []byte) (interface{}, error) {
	parsedConfig := Config{}
	v := viper.New()
	v.SetConfigType("json")
	err := v.ReadConfig(bytes.NewReader(config))
	if err != nil {
		return nil, errors.Wrap(err, "error parsing decision logs config")
	}
	err = v.UnmarshalExact(&parsedConfig, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
	if err != nil {
		return nil, errors.Wrap(err, "error parsing decision logs config")
	}

	return &parsedConfig, util.Unmarshal(config, &parsedConfig)
}
