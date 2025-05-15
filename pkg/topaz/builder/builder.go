package builder

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/middleware"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/aserto-dev/topaz/pkg/topaz/config"
)

type TopazServices interface {
	Directory() *directory.Service
	Authorizer() *authorizer.Service
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

		httpServer.AttachGateway("/api", gwMux)
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

	svc := health.New(cfg)
	svc.RegisterHealthServer(server.Server)

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
	default:
		panic(errors.Errorf("unknown service %q", service))
	}
}
