package builder

import (
	"context"
	"net/http"
	"slices"

	gorilla "github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/config/v3"
	"github.com/aserto-dev/topaz/pkg/console"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/middleware"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/aserto-dev/topaz/pkg/service"
)

type TopazServices interface {
	Registrar(ctx context.Context, name servers.ServiceName) service.Registrar
	Authorizer() *authorizer.Service
	Directory() *directory.Service
	Health() *health.Service
}

type serverBuilder struct {
	cfg      *config.Config
	services TopazServices

	middleware *middlewares
}

func NewServerBuilder(logger *zerolog.Logger, cfg *config.Config, services TopazServices) *serverBuilder {
	return &serverBuilder{
		cfg:      cfg,
		services: services,
		middleware: &middlewares{
			auth:    authentication.New(&cfg.Authentication),
			logging: middleware.NewLogging(logger),
		},
	}
}

func (b *serverBuilder) Build(ctx context.Context, cfg *servers.Server) (*server, error) {
	grpcServer, err := b.buildGRPC(ctx, cfg)
	if err != nil {
		return nil, err
	}

	httpServer, err := b.buildHTTP(&cfg.HTTP)
	if err != nil {
		return nil, err
	}

	// The console http routes need to be attached before the gateway because routes are matched
	// in the order they are registered and the gateway attaches to the '/api' prefix which would
	// match against the console's config endpoint ('/api/v2/config').
	if slices.Contains(cfg.Services, servers.Service.Console) {
		b.registerConsole(httpServer.router)
	}

	if grpcServer.Enabled() && httpServer.Enabled() {
		// wire up grpc-gateway.
		addr := "dns:///" + cfg.GRPC.ListenAddress
		gwMux := gatewayMux(cfg.HTTP.AllowedHeaders)

		creds, err := cfg.GRPC.ClientCredentials()
		if err != nil {
			return nil, err
		}

		for _, service := range cfg.Services {
			registrar := b.services.Registrar(ctx, service)

			if err := registrar.RegisterGateway(ctx, gwMux, addr, []grpc.DialOption{creds}); err != nil {
				return nil, errors.Wrapf(err, "failed to register gateway for service %q", service)
			}

			if err := registrar.RegisterHTTP(ctx, &cfg.HTTP, httpServer.router); err != nil {
				return nil, errors.Wrapf(err, "failed to register HTTP handlers for service %q", service)
			}
		}

		b.registerOpenAPI(httpServer, cfg)
		httpServer.AttachGateway(routes.GatewayPrefix, gwMux)
	}

	for _, service := range cfg.Services {
		b.services.Health().SetServingStatus(string(service), healthpb.HealthCheckResponse_SERVING)
	}

	return &server{grpc: grpcServer, http: httpServer}, nil
}

func (b *serverBuilder) BuildHealth(cfg *health.Config) (*grpcServer, error) {
	if !cfg.Enabled {
		return noGRPC, nil
	}

	server, err := newGRPCServer(&cfg.GRPCServer)
	if err != nil {
		return nil, err
	}

	b.services.Health().RegisterHealthServer(server.Server)
	reflection.Register(server)

	return server, nil
}

func (b *serverBuilder) BuildDebug(ctx context.Context, cfg *debug.Config) (*httpServer, error) {
	if !cfg.Enabled {
		return noHTTP, nil
	}

	server, err := newHTTPServer(&cfg.HTTPServer)
	if err != nil {
		return nil, err
	}

	router := server.router.PathPrefix(routes.DebugPrefix).Subrouter()
	debug.RegisterHandlers(ctx, router)

	return server, nil
}

func (b *serverBuilder) buildGRPC(ctx context.Context, cfg *servers.Server) (*grpcServer, error) {
	if cfg.GRPC.IsEmptyAddress() {
		return noGRPC, nil
	}

	server, err := newGRPCServer(&cfg.GRPC, b.middleware.unary(), b.middleware.stream())
	if err != nil {
		return nil, err
	}

	for _, service := range cfg.Services {
		b.services.Registrar(ctx, service).RegisterGRPC(server.Server)
	}

	if !cfg.GRPC.NoReflection {
		reflection.Register(server)
	}

	return server, nil
}

func (b *serverBuilder) buildHTTP(cfg *servers.HTTPServer) (*httpServer, error) {
	if cfg.IsEmptyAddress() {
		return noHTTP, nil
	}

	return newHTTPServer(cfg)
}

func (b *serverBuilder) registerConsole(router *gorilla.Router) {
	// The config endpoint can be called without authentication but we attach the auth middleware because
	// if an api key is included in the request, we do want to validate it.
	// This is all part of somewhat odd behavior in the console that really needs to be rethought.
	router.Handle(routes.Config, b.middleware.auth.Handler(console.ConfigHandler(b.cfg))).
		Methods(http.MethodGet)

	console.RegisterAppHandlers(router)
}

func (b *serverBuilder) registerOpenAPI(httpServer *httpServer, cfg *servers.Server) {
	if slices.Contains(cfg.Services, servers.Service.Authorizer) {
		httpServer.router.
			HandleFunc(routes.OpenAPIAuthorizer, b.services.Authorizer().OpenAPIHandler()).
			Methods(http.MethodGet)
	}

	dsServices := lo.FilterMap(cfg.Services, func(service servers.ServiceName, _ int) (string, bool) {
		return string(service), slices.Contains(servers.DirectoryServices, service)
	})

	if len(dsServices) > 0 {
		httpServer.router.
			HandleFunc(routes.OpenAPIDirectory, b.services.Directory().OpenAPIHandler(cfg.HTTP.Port(), dsServices...)).
			Methods(http.MethodGet)
	}
}
