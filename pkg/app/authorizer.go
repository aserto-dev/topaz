package app

import (
	"context"
	"net/http"

	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	azOpenAPI "github.com/aserto-dev/openapi-authorizer/publish/authorizer"
	"github.com/aserto-dev/topaz/pkg/app/impl"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/service/builder"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

const (
	authorizerService = "authorizer"
)

type Authorizer struct {
	Resolver         *resolvers.Resolvers
	AuthorizerServer *impl.AuthorizerServer

	cfg  *builder.API
	opts []grpc.ServerOption
}

var _ builder.ServiceTypes = (*Authorizer)(nil)

func NewAuthorizer(
	ctx context.Context,
	cfg *builder.API,
	commonConfig *config.Common,
	authorizerOpts []grpc.ServerOption,
	logger *zerolog.Logger,
) (*Authorizer,
	error,
) {
	if cfg.GRPC.Certs.HasCert() {
		tlsCreds, err := cfg.GRPC.Certs.ServerCredentials()
		if err != nil {
			return nil, errors.Wrap(err, "failed to calculate tls config")
		}

		tlsAuth := grpc.Creds(tlsCreds)
		authorizerOpts = append(authorizerOpts, tlsAuth)
	}

	authResolvers := resolvers.New()

	authServer := impl.NewAuthorizerServer(ctx, logger, commonConfig, authResolvers)

	return &Authorizer{
		cfg:              cfg,
		opts:             authorizerOpts,
		Resolver:         authResolvers,
		AuthorizerServer: authServer,
	}, nil
}

func (e *Authorizer) AvailableServices() []string {
	return []string{authorizerService}
}

func (e *Authorizer) GetGRPCRegistrations(services ...string) builder.GRPCRegistrations {
	return func(server *grpc.Server) {
		authz.RegisterAuthorizerServer(server, e.AuthorizerServer)
	}
}

func (e *Authorizer) GetGatewayRegistration(port string, services ...string) builder.HandlerRegistrations {
	return func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
		if err := authz.RegisterAuthorizerHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
			return err
		}

		if len(services) > 0 {
			if err := mux.HandlePath(http.MethodGet, authorizerOpenAPISpec, azOpenAPIHandler); err != nil {
				return err
			}
		}

		return nil
	}
}

func (e *Authorizer) Cleanups() []func() {
	return nil
}

const (
	authorizerOpenAPISpec string = "/authorizer/openapi.json"
)

func azOpenAPIHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	azOpenAPI.OpenApiHandler(w, r)
}
