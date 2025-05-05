package builder

import (
	"context"

	"github.com/pkg/errors"
)

type Runner interface {
	Go(f func() error)
}

type Server struct {
	grpc *grpcServer
	http *httpServer
}

func (s *Server) Start(ctx context.Context, runner Runner) error {
	if err := s.grpc.Start(ctx, runner); err != nil {
		return errors.Wrapf(err, "failed to start grpc server on %q", s.grpc.listenAddr)
	}

	return s.http.Start(ctx, runner)
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}
