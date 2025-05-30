package edge

import (
	"bytes"
	"context"
	"strings"

	topaz "github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/plugins/noop"
	"github.com/go-viper/mapstructure/v2"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/util"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type PluginFactory struct {
	ctx    context.Context
	cfg    *topaz.Config
	logger *zerolog.Logger
}

var _ plugins.Factory = (*PluginFactory)(nil)

func NewPluginFactory(ctx context.Context, cfg *topaz.Config, logger *zerolog.Logger) PluginFactory {
	return PluginFactory{
		ctx:    ctx,
		cfg:    cfg,
		logger: logger,
	}
}

func (f PluginFactory) New(m *plugins.Manager, config any) plugins.Plugin {
	cfg, _ := config.(*Config)
	if cfg.TenantID == "" {
		cfg.TenantID = strings.Split(m.ID, "/")[0]
	}

	if !cfg.Enabled {
		return &noop.Noop{
			Manager: m,
			Name:    PluginName,
		}
	}

	return newEdgePlugin(f.logger, cfg, f.cfg, m)
}

func (PluginFactory) Validate(m *plugins.Manager, config []byte) (any, error) {
	parsedConfig := Config{}

	v := viper.New()
	v.SetConfigType("json")

	if err := v.ReadConfig(bytes.NewReader(config)); err != nil {
		return nil, errors.Wrap(err, "error parsing edge directory config")
	}

	if err := v.UnmarshalExact(&parsedConfig, func(dc *mapstructure.DecoderConfig) { dc.TagName = "json" }); err != nil {
		return nil, errors.Wrap(err, "error parsing edge directory config")
	}

	return &parsedConfig, util.Unmarshal(config, &parsedConfig)
}
