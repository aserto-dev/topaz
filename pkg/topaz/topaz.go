package topaz

import (
	"context"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"

	"github.com/aserto-dev/logger"

	sbuilder "github.com/aserto-dev/topaz/pkg/topaz/builder"
	"github.com/aserto-dev/topaz/pkg/topaz/config"
)

type Topaz struct {
	Logger   *zerolog.Logger
	services *topazServices
	servers  []sbuilder.Server
	errGroup *errgroup.Group
}

func NewTopaz(ctx context.Context, configPath string, configOverrides ...config.ConfigOverride) (*Topaz, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read config file %q", configPath)
	}

	defer f.Close()

	cfg, err := config.NewConfig(f, configOverrides...)
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

	services, err := newTopazServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	servers, err := newServers(log.WithContext(ctx), cfg, services)
	if err != nil {
		return nil, err
	}

	return &Topaz{
		Logger:   log,
		services: services,
		servers:  servers,
	}, nil
}

func (t *Topaz) Start(ctx context.Context) (context.Context, error) {
	t.errGroup, ctx = errgroup.WithContext(t.Logger.WithContext(ctx))

	for _, server := range t.servers {
		if err := server.Start(ctx, t.errGroup); err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

func (t *Topaz) Stop(ctx context.Context) error {
	ctx = t.Logger.WithContext(ctx)

	for _, server := range t.servers {
		if err := server.Stop(ctx); err != nil {
			zerolog.Ctx(ctx).Err(err).Msg("error while stopping server")
		}
	}

	return multierror.Append(t.errGroup.Wait(), t.services.Close(ctx))
}

func newServers(ctx context.Context, cfg *config.Config, services *topazServices) ([]sbuilder.Server, error) {
	builder := sbuilder.NewServerBuilder(zerolog.Ctx(ctx), cfg, services)
	servers := make([]sbuilder.Server, 0, len(cfg.Servers)+countTrue(cfg.Health.Enabled, cfg.Metrics.Enabled))

	for name, serverCfg := range cfg.Servers {
		srvr, err := builder.Build(ctx, serverCfg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build server %q", name)
		}

		servers = append(servers, srvr)
	}

	if cfg.Health.Enabled {
		healthSrvr, err := builder.BuildHealth(&cfg.Health)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build health server")
		}

		servers = append(servers, healthSrvr)
	}

	return servers, nil
}

func countTrue(vals ...bool) int {
	return lo.Reduce(vals,
		func(count int, val bool, _ int) int {
			return count + lo.Ternary(val, 1, 0)
		},
		0,
	)
}
