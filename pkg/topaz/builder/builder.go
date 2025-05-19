package builder

import (
	"context"
	"slices"

	gorilla "github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/console"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/middleware"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/aserto-dev/topaz/pkg/topaz/config"
)

type TopazServices interface {
	Directory() *directory.Service
	Authorizer() *authorizer.Service
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
	grpcServer, err := b.buildGRPC(cfg)
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
			if err := b.registerGateway(ctx, service, gwMux, addr, creds); err != nil {
				return nil, err
			}
		}

		httpServer.AttachGateway("/api/", gwMux)
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

	router := server.router.PathPrefix("/debug/").Subrouter()
	debug.RegisterHandlers(ctx, router)

	return server, nil
}

func (b *serverBuilder) buildGRPC(cfg *servers.Server) (*grpcServer, error) {
	if !cfg.GRPC.HasListener() {
		return noGRPC, nil
	}

	server, err := newGRPCServer(&cfg.GRPC, b.middleware.unary(), b.middleware.stream())
	if err != nil {
		return nil, err
	}

	for _, service := range cfg.Services {
		b.registerService(server.Server, service)
	}

	if !cfg.GRPC.NoReflection {
		reflection.Register(server)
	}

	return server, nil
}

func (b *serverBuilder) buildHTTP(cfg *servers.HTTPServer) (*httpServer, error) {
	if !cfg.HasListener() {
		return noHTTP, nil
	}

	return newHTTPServer(cfg)
}

func (b *serverBuilder) registerService(server *grpc.Server, service servers.ServiceName) {
	switch service {
	case servers.Service.Access:
		b.services.Directory().RegisterAccessServer(server)
	case servers.Service.Reader:
		b.services.Directory().RegisterReaderServer(server)
	case servers.Service.Writer:
		b.services.Directory().RegisterWriterServer(server)
	case servers.Service.Authorizer:
		b.services.Authorizer().RegisterAuthorizerServer(server)
	case servers.Service.Model:
		b.services.Directory().RegisterModelServer(server)
	case servers.Service.Importer:
		b.services.Directory().RegisterImporterServer(server)
	case servers.Service.Exporter:
		b.services.Directory().RegisterExporterServer(server)
	case servers.Service.Console:
		// No gateway for the console.
	default:
		panic(errors.Errorf("unknown service %q", service))
	}
}

func (b *serverBuilder) registerGateway(
	ctx context.Context,
	service servers.ServiceName,
	mux *runtime.ServeMux,
	addr string,
	opts ...grpc.DialOption,
) error {
	switch service {
	case servers.Service.Access:
		return b.services.Directory().RegisterAccessGateway(ctx, mux, addr, opts...)
	case servers.Service.Reader:
		return b.services.Directory().RegisterReaderGateway(ctx, mux, addr, opts...)
	case servers.Service.Writer:
		return b.services.Directory().RegisterWriterGateway(ctx, mux, addr, opts...)
	case servers.Service.Authorizer:
		return b.services.Authorizer().RegisterAuthorizerGateway(ctx, mux, addr, opts...)
	case servers.Service.Model:
		return b.services.Directory().RegisterModelGateway(ctx, mux, addr, opts...)
	case servers.Service.Importer, servers.Service.Exporter, servers.Service.Console:
		return nil
	default:
		panic(errors.Errorf("unknown service %q", service))
	}
}

func (b *serverBuilder) registerConsole(router *gorilla.Router) {
	// The config endpoint can be called without authentication but we attach the auth middleware because
	// if an api key is included in the request, we do want to validate it.
	// This is all part of somewhat odd behavior in the console that really needs to be rethought.
	router.Handle("/api/v2/config", b.middleware.auth.Handler(console.ConfigHandler(b.cfg)))

	console.RegisterAppHandlers(router)
}
