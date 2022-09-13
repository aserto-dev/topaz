package server

import (
	"context"
	"net/http"

	"github.com/aserto-dev/topaz/pkg/app/impl"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	dir "github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
)

// GRPCRegistrations represents a function that can register API implementations to the GRPC server.
type GRPCRegistrations func(server *grpc.Server)

// HandlerRegistrations represents a function that can register handlers for the GRPC Gateway.
type HandlerRegistrations func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error

// HTTPRouteRegistrations represents a function that can register any custom http handler for HTTP server.
type HTTPRouteRegistrations func(mux *http.ServeMux)

func CoreServiceRegistrations(
	implAuthorizerServer *impl.AuthorizerServer,
	implDirectoryServer *impl.DirectoryServer,
) GRPCRegistrations {
	return func(srv *grpc.Server) {
		authz.RegisterAuthorizerServer(srv, implAuthorizerServer)
		dir.RegisterDirectoryServer(srv, implDirectoryServer)
	}
}
