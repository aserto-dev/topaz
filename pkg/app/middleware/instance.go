package middleware

import (
	"context"
	"net/textproto"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

type CtxKey string

var (
	// TODO: figure out what the directory needs as a tenant ID.
	InstanceIDHeader = CtxKey(textproto.CanonicalMIMEHeaderKey("Aserto-Tenant-Id"))
)

type InstanceIDMiddleware struct {
	instanceID string
}

func NewInstanceIDMiddleware(cfg *config.Common) *InstanceIDMiddleware {
	return &InstanceIDMiddleware{
		instanceID: cfg.OPA.InstanceID,
	}
}

func (m *InstanceIDMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := context.WithValue(ctx, InstanceIDHeader, m.instanceID)
		result, err := handler(newCtx, req)
		return result, err
	}
}

func (m *InstanceIDMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		newCtx := context.WithValue(ctx, InstanceIDHeader, m.instanceID)
		wrapped := grpcmiddleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

func (m *InstanceIDMiddleware) AsGRPCOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(m.Unary()),
		grpc.StreamInterceptor(m.Stream()),
	}
}
