package builder

import (
	"context"
	"net"
	"net/http"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type ServiceInterface interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type GRPCRegistrations func(server *grpc.Server)

type HandlerRegistrations func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error

type Server struct {
	Config        *directory.API
	Server        *grpc.Server
	Listener      net.Listener
	Registrations GRPCRegistrations
	Gateway       Gateway
	Health        *HealthServer
}

func (s *Server) Start(ctx context.Context) error {
	return s.Server.Serve(s.Listener)
}

func (s *Server) Stop(ctx context.Context) error {
	s.Server.GracefulStop()
	return nil
}

type Gateway struct {
	Server *http.Server
	Mux    *http.ServeMux
	Certs  *certs.TLSCredsConfig
}
