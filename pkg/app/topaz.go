package app

import (
	"github.com/aserto-dev/certs"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	builder "github.com/aserto-dev/service-host"
	"github.com/aserto-dev/topaz/pkg/app/impl"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
)

type Topaz struct {
	Resolver         *resolvers.Resolvers
	AuthorizerServer *impl.AuthorizerServer

	cfg  *builder.API
	opts []grpc.ServerOption
}

const (
	authorizerService = "authorizer"
)

func NewTopaz(cfg *builder.API, commonConfig *config.Common, authorizerOpts []grpc.ServerOption, logger *zerolog.Logger) (ServiceTypes, error) {
	if cfg.GRPC.Certs.TLSCertPath != "" {
		tlsCreds, err := certs.GRPCServerTLSCreds(cfg.GRPC.Certs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to calculate tls config")
		}

		tlsAuth := grpc.Creds(tlsCreds)
		authorizerOpts = append(authorizerOpts, tlsAuth)
	}
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		return nil, err
	}
	authorizerOpts = append(authorizerOpts, grpc.StatsHandler(&ocgrpc.ServerHandler{}))

	authResolvers := resolvers.New()

	authServer := impl.NewAuthorizerServer(logger, commonConfig, authResolvers)

	return &Topaz{
		cfg:              cfg,
		opts:             authorizerOpts,
		Resolver:         authResolvers,
		AuthorizerServer: authServer,
	}, nil
}

func (e *Topaz) AvailableServices() []string {
	return []string{authorizerService}
}

func (e *Topaz) GetGRPCRegistrations(services ...string) builder.GRPCRegistrations {
	return func(server *grpc.Server) {
		authz.RegisterAuthorizerServer(server, e.AuthorizerServer)
	}
}

func (e *Topaz) GetGatewayRegistration(services ...string) builder.HandlerRegistrations {
	return authz.RegisterAuthorizerHandlerFromEndpoint
}
