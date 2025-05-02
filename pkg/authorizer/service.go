package authorizer

import (
	"context"

	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/aserto-dev/topaz/pkg/app/impl"
	"github.com/aserto-dev/topaz/resolvers"
)

type Service struct {
	*impl.AuthorizerServer
}

func New(ctx context.Context, cfg *Config) *Service {
	return &Service{impl.NewAuthorizerServer(ctx, zerolog.Ctx(ctx), resolvers.New(), cfg.JWT.AcceptableTimeSkew)}
}

func (s *Service) RegisterAuthorizerServer(server *grpc.Server) {
	authz.RegisterAuthorizerServer(server, s)
}

func (s *Service) RegisterAuthorizerGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) error {
	return authz.RegisterAuthorizerHandlerFromEndpoint(ctx, mux, endpoint, opts)
}
