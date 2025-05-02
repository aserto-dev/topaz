package topaz

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/aserto-dev/logger"

	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/pkg/topaz/config"
)

type Topaz struct {
	servers  []Server
	errGroup *errgroup.Group
}

func NewTopaz(ctx context.Context, configPath string, configOverrides ...config.ConfigOverride) (*Topaz, error) {
	cfgBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read config file %q", configPath)
	}

	cfg, err := config.NewConfig(bytes.NewReader(cfgBytes), configOverrides...)
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log, err := logger.NewLogger(os.Stdout, os.Stderr, &cfg.Logging)
	if err != nil {
		return nil, err
	}

	if strings.Contains(string(cfgBytes), x.EnvTopazDir) {
		log.Warn().Msg("This configuration file uses the obsolete TOPAZ_DIR environment variable.")
		log.Warn().Msg("Please update to use the new TOPAZ_DB_DIR and TOPAZ_CERTS_DIR environment variables.")
	}

	servers, err := newServers(log.WithContext(ctx), cfg)
	if err != nil {
		return nil, err
	}

	return &Topaz{
		servers: servers,
	}, nil
}

func (t *Topaz) Start(ctx context.Context) (context.Context, error) {
	t.errGroup, ctx = errgroup.WithContext(ctx)

	for _, server := range t.servers {
		if err := server.Start(ctx, t.errGroup); err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

func newServers(ctx context.Context, cfg *config.Config) ([]Server, error) {
	services, err := newTopazServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	builder := newServerBuilder(zerolog.Ctx(ctx), cfg, services)
	servers := make([]Server, 0, len(cfg.Servers)+countTrue(cfg.Health.Enabled, cfg.Metrics.Enabled))

	for name, serverCfg := range cfg.Servers {
		srvr, err := builder.Build(ctx, serverCfg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build server %q", name)
		}

		servers = append(servers, srvr)
	}

	return servers, nil
}
