package edge

import (
	"bytes"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/util"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	client "github.com/aserto-dev/go-aserto"

	"github.com/aserto-dev/topaz/topazd/authorizer/plugins/noop"
)

type HealthReporter interface {
	SetServingStatus(service string, servingStatus healthpb.HealthCheckResponse_ServingStatus)
}

type PluginFactory struct {
	dsCfg  *client.Config
	logger *zerolog.Logger
	health HealthReporter
}

var _ plugins.Factory = (*PluginFactory)(nil)

func NewPluginFactory(cfg *client.Config, logger *zerolog.Logger, health HealthReporter) *PluginFactory {
	return &PluginFactory{
		dsCfg:  cfg,
		logger: logger,
		health: health,
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

	return newEdgePlugin(f.logger, cfg, f.dsCfg, m, f.health)
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
