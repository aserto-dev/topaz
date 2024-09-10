package middlewares

import (
	"context"
	"strings"

	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type PolicyInstanceMiddleware struct {
	policyName string
	logger     *zerolog.Logger
}

func NewInstanceMiddleware(cfg *config.Config, logger *zerolog.Logger) *PolicyInstanceMiddleware {
	details := strings.Split(*cfg.OPA.Config.Discovery.Resource, "/")

	return &PolicyInstanceMiddleware{
		policyName: details[0],
		logger:     logger,
	}
}

var _ grpcutil.Middleware = &PolicyInstanceMiddleware{}

// If the unary operation is an Is request attach configured instance information to request.
func (m *PolicyInstanceMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		request, ok := req.(*authorizer.IsRequest)
		if ok {
			request.PolicyInstance = &api.PolicyInstance{
				Name: m.policyName,
			}
		}
		return handler(ctx, req)
	}
}

// passthrough as Is call is Unary type operation.
func (m *PolicyInstanceMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		wrapped := grpcmiddleware.WrapServerStream(stream)
		wrapped.WrappedContext = ctx
		return handler(srv, wrapped)
	}
}
