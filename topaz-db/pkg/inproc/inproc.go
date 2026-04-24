package inproc

import (
	"context"
	"net"

	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
	dsw "github.com/aserto-dev/topaz/api/directory/v4/writer"

	"github.com/aserto-dev/topaz/internal/eds"
	"github.com/aserto-dev/topaz/internal/eds/pkg/directory"
	"github.com/aserto-dev/topaz/internal/grpc/middlewares/gerr"
	"github.com/rs/zerolog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize int = 1024 * 1024

func NewServer(ctx context.Context, logger *zerolog.Logger, cfg *directory.Config) (*grpc.ClientConn, func()) {
	listener := bufconn.Listen(bufSize)

	dsLogger := logger.With().Str("component", "ds").Logger()

	inProcDirectory, err := eds.New(ctx, cfg, &dsLogger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to start edge directory server")
	}

	errMiddleware := gerr.NewErrorMiddleware()
	s := grpc.NewServer(
		grpc.UnaryInterceptor(errMiddleware.Unary()),
		grpc.StreamInterceptor(errMiddleware.Stream()),
	)

	dsr.RegisterReaderServer(s, inProcDirectory.Reader3())
	dsw.RegisterWriterServer(s, inProcDirectory.Writer3())

	go func() {
		if err := s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), //nolint:staticcheck
	)
	if err != nil {
		panic(err)
	}

	return conn, s.GracefulStop
}
