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

	"github.com/aserto-dev/topaz/pkg/config/v3"
	sbuilder "github.com/aserto-dev/topaz/pkg/topaz/builder"
)

type Topaz struct {
	Logger   *zerolog.Logger
	services *topazServices
	servers  []sbuilder.Server
	errGroup *errgroup.Group
}

func NewTopaz(ctx context.Context, cfg *config.Config) (*Topaz, error) {
	log, err := logger.NewLogger(os.Stdout, os.Stderr, &cfg.Logging)
	if err != nil {
		return nil, err
	}

	ctx = log.WithContext(ctx)

	services, err := newTopazServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	servers, err := newServers(ctx, cfg, services)
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
	servers := make([]sbuilder.Server, 0, len(cfg.Servers)+countTrue(cfg.Debug.Enabled, cfg.Health.Enabled, cfg.Metrics.Enabled))

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

	if cfg.Debug.Enabled {
		debugSrvr, err := builder.BuildDebug(ctx, &cfg.Debug)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build debug server")
		}

		servers = append(servers, debugSrvr)
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
