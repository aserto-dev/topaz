package instance

import (
	"context"

	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

type CtxKey string

var (
	// TODO: figure out what the directory needs as a tenant ID.
	InstanceIDHeader = grpcutil.HeaderAsertoTenantID
)

type IDMiddleware struct {
	instanceID string
}

func NewInstanceIDMiddleware(cfg *config.Common) *IDMiddleware {
	return &IDMiddleware{
		instanceID: cfg.OPA.InstanceID,
	}
}

func (m *IDMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := context.WithValue(ctx, InstanceIDHeader, m.instanceID)
		result, err := handler(newCtx, req)
		return result, err
	}
}

func (m *IDMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		newCtx := context.WithValue(ctx, InstanceIDHeader, m.instanceID)
		wrapped := grpcmiddleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

func (m *IDMiddleware) AsGRPCOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(m.Unary()),
		grpc.StreamInterceptor(m.Stream()),
	}
}
