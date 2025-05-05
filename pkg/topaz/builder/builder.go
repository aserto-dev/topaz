package builder

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/middleware"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/aserto-dev/topaz/pkg/topaz/config"
)

type serverBuilder struct {
	cfg      *config.Config
	services *topazServices

	middleware *middlewares
}

func NewServerBuilder(logger *zerolog.Logger, cfg *config.Config, services *topazServices) *serverBuilder {
	return &serverBuilder{
		cfg:      cfg,
		services: services,
		middleware: &middlewares{
			auth:    authentication.New(&cfg.Authentication),
			logging: middleware.NewLogging(logger),
		},
	}
}

func (b *serverBuilder) Build(ctx context.Context, cfg *servers.Server) (*Server, error) {
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

	return &Server{grpc: grpcServer, http: httpServer}, nil
}

func (b *serverBuilder) buildGRPC(cfg *servers.Server) (*grpcServer, error) {
	if !cfg.GRPC.HasListener() {
		return noGRPC, nil
	}

	server, err := newGRPCServer(&cfg.GRPC, b.middleware)
	if err != nil {
		return nil, err
	}

	// TODO: register reflection service. Need to add a config option.

	for _, service := range cfg.Services {
		b.registerService(server.Server, service)
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
		b.services.directory.RegisterAccessServer(server)
	case servers.Service.Reader:
		b.services.directory.RegisterReaderServer(server)
	case servers.Service.Writer:
		b.services.directory.RegisterWriterServer(server)
	case servers.Service.Authorizer:
		b.services.authorizer.RegisterAuthorizerServer(server)
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
		return b.services.directory.RegisterAccessGateway(ctx, mux, addr, opts...)
	case servers.Service.Reader:
		return b.services.directory.RegisterReaderGateway(ctx, mux, addr, opts...)
	case servers.Service.Writer:
		return b.services.directory.RegisterWriterGateway(ctx, mux, addr, opts...)
	case servers.Service.Authorizer:
		return b.services.authorizer.RegisterAuthorizerGateway(ctx, mux, addr, opts...)
	default:
		panic(errors.Errorf("unknown service %q", service))
	}
}

type topazServices struct {
	directory  *directory.Service
	authorizer *authorizer.Service
	console    *app.ConsoleService
}

func NewTopazServices(ctx context.Context, cfg *config.Config) (*topazServices, error) {
	dir, err := directory.New(ctx, &cfg.Directory)
	if err != nil {
		return nil, err
	}

	return &topazServices{
		directory:  dir,
		authorizer: authorizer.New(ctx, &cfg.Authorizer),
		console:    app.NewConsole(),
	}, nil
}
