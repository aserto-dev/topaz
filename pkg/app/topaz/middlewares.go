package topaz

import (
	"context"

	"github.com/aserto-dev/aserto-grpc/grpcclient"
	public_grpcutil "github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/gerr"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/metrics"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/request"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/tracing"
	"github.com/aserto-dev/topaz/pkg/app/auth"
	"github.com/aserto-dev/topaz/pkg/app/impl"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func Middlewares(
	ctx context.Context,
	logger *zerolog.Logger,
	cfg *config.Config,
	authzServer *impl.AuthorizerServer,
	dirServer *impl.DirectoryServer,
	dop grpcclient.DialOptionsProvider) (public_grpcutil.Middlewares, error) {

	authmiddleware, err := auth.NewAPIKeyAuthMiddleware(ctx, &cfg.Auth, logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth middleware")
	}
	return metrics.NewMiddlewares(
		cfg.API.Metrics,
		request.NewRequestIDMiddleware(),
		tracing.NewTracingMiddleware(logger),
		gerr.NewErrorMiddleware(),
		NewTenantIDMiddleware(cfg),
		authmiddleware,
	), nil
}

type TenantIDIDMiddleware struct {
	tenantID string
}

func NewTenantIDMiddleware(cfg *config.Config) *TenantIDIDMiddleware {
	return &TenantIDIDMiddleware{
		tenantID: extractTenantID(cfg),
	}
}

var _ public_grpcutil.Middleware = &TenantIDIDMiddleware{}

func (m *TenantIDIDMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := public_grpcutil.ContextWithTenantID(ctx, m.tenantID)
		result, err := handler(newCtx, req)
		return result, err
	}
}

func (m *TenantIDIDMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		newCtx := public_grpcutil.ContextWithTenantID(ctx, m.tenantID)
		wrapped := grpcmiddleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

func extractTenantID(cfg *config.Config) string {
	tenantService, ok := cfg.OPA.Config.Services["aserto-tenant"].(map[string]interface{})
	if !ok {
		return cfg.OPA.InstanceID
	}

	headers, ok := tenantService["headers"].(map[string]interface{})
	if !ok {
		return cfg.OPA.InstanceID
	}

	tenantID, ok := headers["aserto-tenant-id"].(string)
	if !ok {
		return cfg.OPA.InstanceID
	}
	return tenantID
}
