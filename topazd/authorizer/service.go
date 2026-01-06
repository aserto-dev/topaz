package authorizer

import (
	"context"
	"net/http"
	"time"

	gorilla "github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	client "github.com/aserto-dev/go-aserto"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	azOpenAPI "github.com/aserto-dev/openapi-authorizer/publish/authorizer"
	"github.com/aserto-dev/self-decision-logger/logger/self"

	"github.com/aserto-dev/topaz/topazd/authorizer/decisionlog"
	"github.com/aserto-dev/topaz/topazd/authorizer/decisionlog/logger/file"
	"github.com/aserto-dev/topaz/topazd/authorizer/impl"
	"github.com/aserto-dev/topaz/topazd/authorizer/plugins/edge"
	"github.com/aserto-dev/topaz/topazd/servers"
	"github.com/aserto-dev/topaz/topazd/x"
)

type Service struct {
	*impl.AuthorizerServer

	close func(context.Context) error
}

func New(ctx context.Context, cfg *Config, edgeFactory *edge.PluginFactory, dsCfg *client.Config) (*Service, error) {
	var closer x.Closer

	dsConn, err := dsCfg.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create directory client")
	}

	closer = append(closer, x.CloserErr(dsConn.Close))

	decisionLogger, err := newDecisionLogger(ctx, &cfg.DecisionLogger)
	if err != nil {
		_ = closer.Close(ctx)
		return nil, errors.Wrap(err, "failed to create decision logger")
	}

	closer = append(closer, x.CloserFunc(decisionLogger.Shutdown))

	rtResolver, err := NewRuntimeResolver(ctx, cfg, decisionLogger, dsConn, edgeFactory)
	if err != nil {
		_ = closer.Close(ctx)
		return nil, errors.Wrap(err, "failed to create runtime resolver")
	}

	if err := rtResolver.Start(ctx); err != nil {
		_ = closer.Close(ctx)
		return nil, errors.Wrap(err, "failed to start runtime resolver")
	}

	closer = append(closer, rtResolver.Stop)

	return &Service{
		impl.NewAuthorizerServer(ctx, dsConn, rtResolver, cfg.JWT.AcceptableTimeSkew, cfg.OPA.PolicyInstance()),
		closer.Close,
	}, nil
}

func (s *Service) Close(ctx context.Context) error {
	if s.close == nil {
		return nil
	}

	return s.close(ctx)
}

func (s *Service) RegisterGRPC(server *grpc.Server) {
	if s.AuthorizerServer != nil {
		authz.RegisterAuthorizerServer(server, s)
	}
}

func (s *Service) RegisterGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	if s.AuthorizerServer != nil {
		return authz.RegisterAuthorizerHandlerFromEndpoint(ctx, mux, endpoint, opts)
	}

	return nil
}

func (s *Service) RegisterHTTP(_ context.Context, _ *servers.HTTPServer, _ *gorilla.Router) error {
	return nil
}

func (s *Service) OpenAPIHandler() http.HandlerFunc {
	return azOpenAPI.OpenApiHandler
}

const (
	keepaliveTime    = 30 * time.Second // send pings every 30 seconds if there is no activity
	keepaliveTimeout = 5 * time.Second  // wait 5 seconds for ping ack before considering the connection dead
)

func newDecisionLogger(ctx context.Context, cfg *DecisionLoggerConfig) (decisionlog.DecisionLogger, error) {
	if !cfg.Enabled {
		return noLogger, nil
	}

	switch cfg.Provider {
	case SelfDecisionLoggerPlugin:
		return self.NewFromConfig(
			ctx,
			(*self.Config)(&cfg.Self),
			zerolog.Ctx(ctx),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    keepaliveTime,
				Timeout: keepaliveTimeout,
			}))
	case FileDecisionLoggerPlugin:
		return file.New(ctx, (*file.Config)(&cfg.File), zerolog.Ctx(ctx))
	}

	return noLogger, nil
}

type noopLogger struct{}

var noLogger decisionlog.DecisionLogger = noopLogger{}

func (noopLogger) Log(d *api.Decision) error {
	return nil
}

func (noopLogger) Shutdown() {
}
