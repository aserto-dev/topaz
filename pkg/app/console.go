package app

import (
	"context"
	"embed"

	builder "github.com/aserto-dev/service-host"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

//nolint:all
//go:embed console
var console embed.FS

type ConsoleService struct{}

func NewConsole() ServiceTypes {
	return &ConsoleService{}
}

func (e *ConsoleService) AvailableServices() []string {
	return []string{"console"}
}

func (e *ConsoleService) GetGRPCRegistrations(services ...string) builder.GRPCRegistrations {
	return func(server *grpc.Server) {
	}
}

func (e *ConsoleService) GetGatewayRegistration(services ...string) builder.HandlerRegistrations {
	return func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
		return nil
	}
}

func (e *ConsoleService) Cleanups() []func() {
	return nil
}
