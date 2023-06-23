package builder

import (
	"context"
	"net"
	"net/http"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"google.golang.org/grpc"
)

type GRPCRegistrations func(server *grpc.Server)

type Server struct {
	Config        *directory.API
	Server        *grpc.Server
	Listener      net.Listener
	Registrations GRPCRegistrations
	Gateway       Gateway
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
	Certs  *certs.TLSCredsConfig
}
