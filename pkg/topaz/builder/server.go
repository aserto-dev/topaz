package builder

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type Runner interface {
	Go(f func() error)
}

type Server interface {
	Start(ctx context.Context, runner Runner) error
	Stop(ctx context.Context) error
}

type server struct {
	grpc *grpcServer
	http *httpServer
}

func (s *server) Start(ctx context.Context, runner Runner) error {
	if err := s.grpc.Start(ctx, runner); err != nil {
		return errors.Wrapf(err, "failed to start grpc server on %q", s.grpc.listenAddr)
	}

	if err := s.http.Start(ctx, runner); err != nil {
		_ = s.grpc.Stop(ctx)
		return errors.Wrapf(err, "failed to start http server on %q", s.http.Addr)
	}

	return nil
}

func (s *server) Stop(ctx context.Context) error {
	var errs error

	if err := s.http.Stop(ctx); err != nil {
		errs = multierror.Append(errs, errors.Wrapf(err, "failed to stop http server on %q", s.http.Addr))
	}

	if err := s.grpc.Stop(ctx); err != nil {
		errs = multierror.Append(errs, errors.Wrapf(err, "failed to stop grpc server on %q", s.grpc.listenAddr))
	}

	return errs
}
