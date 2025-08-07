package builder

import (
	"context"
	"net"

	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var noGRPC = &grpcServer{Server: new(grpc.Server)}

type grpcServer struct {
	*grpc.Server

	listenAddr string
}

func newGRPCServer(cfg *servers.GRPCServer, opts ...grpc.ServerOption) (*grpcServer, error) {
	creds, err := cfg.Certs.ServerCredentials()
	if err != nil {
		return nil, err
	}

	opts = append(opts,
		grpc.Creds(creds),
		grpc.ConnectionTimeout(cfg.ConnectionTimeout),
	)

	return &grpcServer{
		Server:     grpc.NewServer(opts...),
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

	zerolog.Ctx(ctx).Info().Msgf("Starting %s gRPC server", s.listenAddr)

	runner.Go(func() error {
		e := s.Serve(listener)
		return e
	})

	return nil
}

func (s *grpcServer) Stop(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}

	shutdown := make(chan struct{}, 1)

	go func() {
		s.GracefulStop()
		close(shutdown)
	}()

	select {
	case <-ctx.Done():
		s.Server.Stop()
	case <-shutdown:
	}

	return nil
}

func (s *grpcServer) Enabled() bool {
	return len(s.GetServiceInfo()) > 0
}
