package builder

import (
	"context"
	"net"

	"github.com/aserto-dev/topaz/pkg/servers"
	"google.golang.org/grpc"
)

var noGRPC = &grpcServer{Server: new(grpc.Server)}

type grpcServer struct {
	*grpc.Server

	listenAddr string
}

func newGRPCServer(cfg *servers.GRPCServer, mw *middlewares) (*grpcServer, error) {
	creds, err := cfg.Certs.ServerCredentials()
	if err != nil {
		return nil, err
	}

	return &grpcServer{
		Server:     grpc.NewServer(grpc.Creds(creds), mw.unary(), mw.stream()),
		listenAddr: cfg.ListenAddress,
	}, nil
}

func (s *grpcServer) Start(ctx context.Context, runner Runner) error {
	if !s.Enabled() {
		// No services registered, nothing to do.
		return nil
	}

	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", s.listenAddr)
	if err != nil {
		return err
	}

	runner.Go(func() error {
		return s.Serve(listener)
	})

	return nil
}

func (s *grpcServer) Enabled() bool {
	return len(s.GetServiceInfo()) > 0
}
