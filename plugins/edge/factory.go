package edge

import (
	"bytes"
	"context"
	"strings"

	"github.com/aserto-dev/topaz/pkg/app"
	topaz "github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/plugins/noop"
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
	app *app.Topaz
}

func NewPluginFactory(ctx context.Context, cfg *topaz.Config, logger *zerolog.Logger, topazApp *app.Topaz) PluginFactory {
	return PluginFactory{
		ctx:    ctx,
		cfg:    cfg,
		logger: logger,
		app: topazApp,
	}
}

func (f PluginFactory) New(m *plugins.Manager, config interface{}) plugins.Plugin {
	cfg := config.(*Config)
	if cfg.TenantID == "" {
		cfg.TenantID = strings.Split(m.ID, "/")[0]
	}
	if !cfg.Enabled {
		return &noop.Noop{
			Manager: m,
			Name:    PluginName,
		}
	}

	return newEdgePlugin(f.logger, cfg, f.cfg, m,f.app)
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
