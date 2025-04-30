package topaz

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/grpc"

	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/app/impl"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/middleware"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/aserto-dev/topaz/resolvers"
)

type server struct {
	grpc *grpc.Server
	http *http.Server
}

var _ Server = (*server)(nil)

func (s *server) Run(ctx context.Context) error {
	return nil
}

func newServers(ctx context.Context, cfg *Config) ([]Server, error) {
	services, err := newTopazServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	servers := make([]Server, 0, len(cfg.Servers)+countTrue(cfg.Health.Enabled, cfg.Metrics.Enabled))

	for name, serverCfg := range cfg.Servers {
		srvr, err := buildServer(ctx, serverCfg, services)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build server %q", name)
		}

		servers = append(servers, srvr)
	}

	return nil, nil
}

//nolint:ireturn  // Factory function.
func buildServer(ctx context.Context, cfg *servers.Server, services *topazServices) (Server, error) {
	return nil, nil
}

func grpcMiddleware(logger *zerolog.Logger, cfg *Config) ([]middleware.GRPC, error) {
	return nil, nil
}

type topazServices struct {
	directory  *directory.Directory
	authorizer *impl.AuthorizerServer
	console    *app.ConsoleService
}

func newTopazServices(ctx context.Context, cfg *Config) (*topazServices, error) {
	dir, err := directory.NewDirectory(ctx, &cfg.Directory)
	if err != nil {
		return nil, err
	}

	return &topazServices{
		directory:  dir,
		authorizer: newAuthorizer(ctx, &cfg.Authorizer),
		console:    app.NewConsole(),
	}, nil
}

func newAuthorizer(ctx context.Context, cfg *authorizer.Config) *impl.AuthorizerServer {
	return impl.NewAuthorizerServer(ctx, zerolog.Ctx(ctx), resolvers.New(), cfg.JWT.AcceptableTimeSkew)
}

func countTrue(vals ...bool) int {
	return lo.Reduce(vals,
		func(count int, val bool, _ int) int {
			return count + lo.Ternary(val, 1, 0)
		},
		0,
	)
}
