package app

import (
	"context"
	"net/http"
	"strconv"

	"github.com/aserto-dev/certs"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	azOpenAPI "github.com/aserto-dev/openapi-authorizer/publish/authorizer"
	builder "github.com/aserto-dev/service-host"
	"github.com/aserto-dev/topaz/pkg/app/impl"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/rapidoc"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

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
	consoleService    = "console"
)

func NewAuthorizer(cfg *builder.API, commonConfig *config.Common, authorizerOpts []grpc.ServerOption, logger *zerolog.Logger) (ServiceTypes, error) {
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
	return func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
		if err := authz.RegisterAuthorizerHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
			return err
		}

		if len(services) > 0 {
			if err := mux.HandlePath(http.MethodGet, authorizerOpenAPISpec, azOpenAPIHandler); err != nil {
				return err
			}
			if err := mux.HandlePath(http.MethodGet, authorizerOpenAPIDocs, azOpenAPIDocsHandler()); err != nil {
				return err
			}
		}

		return nil
	}
}

func (e *Topaz) Cleanups() []func() {
	return nil
}

const (
	authorizerOpenAPISpec string = "/authorizer/openapi.json"
	authorizerOpenAPIDocs string = "/authorizer/docs"
)

func azOpenAPIHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	buf, err := azOpenAPI.Static().ReadFile("openapi.json")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.FormatInt(int64(len(buf)), 10))
	_, _ = w.Write(buf)
}

func azOpenAPIDocsHandler() runtime.HandlerFunc {
	h := rapidoc.Handler(&rapidoc.Opts{
		Path:    authorizerOpenAPIDocs,
		SpecURL: authorizerOpenAPISpec,
		Title:   "Aserto Directory HTTP API",
	}, nil)

	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		h.ServeHTTP(w, r)
	}
}
