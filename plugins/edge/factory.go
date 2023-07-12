package edge

import (
	"bytes"
	"context"
	"strings"

	"github.com/aserto-dev/sidecar/plugins/dummy"
	topaz "github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/mitchellh/mapstructure"
	"github.com/open-policy-agent/opa/plugins"
	"github.com/open-policy-agent/opa/util"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type PluginFactory struct {
	ctx    context.Context
	cfg    *topaz.Config
	logger *zerolog.Logger
}

func NewPluginFactory(ctx context.Context, cfg *topaz.Config, logger *zerolog.Logger) PluginFactory {
	return PluginFactory{
		ctx:    ctx,
		cfg:    cfg,
		logger: logger,
	}
}

func (f PluginFactory) New(m *plugins.Manager, config interface{}) plugins.Plugin {
	cfg := config.(*Config)
	cfg.TenantID = strings.Split(m.ID, "/")[0]

	if !cfg.Enabled {
		return &dummy.Dummy{
			Manager: m,
			Name:    PluginName,
		}
	}

	return newEdgePlugin(f.logger, cfg, f.cfg, m)
}

func (PluginFactory) Validate(m *plugins.Manager, config []byte) (interface{}, error) {
	parsedConfig := Config{}
	v := viper.New()
	v.SetConfigType("json")
	err := v.ReadConfig(bytes.NewReader(config))
	if err != nil {
		return nil, errors.Wrap(err, "error parsing edge directory config")
	}
	err = v.UnmarshalExact(&parsedConfig, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
	if err != nil {
		return nil, errors.Wrap(err, "error parsing edge directory config")
	}

	return &parsedConfig, util.Unmarshal(config, &parsedConfig)
}
