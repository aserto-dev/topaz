package middlewares

import (
	"context"

	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/header"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

type TenantIDIDMiddleware struct {
	tenantID string
}

func NewTenantIDMiddleware(cfg *config.Config) *TenantIDIDMiddleware {
	return &TenantIDIDMiddleware{
		tenantID: extractTenantID(cfg),
	}
}

var _ grpcutil.Middleware = &TenantIDIDMiddleware{}

func (m *TenantIDIDMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := header.ContextWithTenantID(ctx, m.tenantID)
		result, err := handler(newCtx, req)
		return result, err
	}
}

func (m *TenantIDIDMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		newCtx := header.ContextWithTenantID(ctx, m.tenantID)
		wrapped := grpcmiddleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

// rely on OPA instance ID to be set to tenant id.
func extractTenantID(cfg *config.Config) string {
	return cfg.OPA.InstanceID
}
