package inproc

import (
	"context"
	"net"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"

	dsr4 "github.com/aserto-dev/go-directory/aserto/directory/reader/v4"
	dsw4 "github.com/aserto-dev/go-directory/aserto/directory/writer/v4"

	"github.com/aserto-dev/aserto-grpc/middlewares/gerr"
	"github.com/aserto-dev/topaz/internal/eds"
	"github.com/aserto-dev/topaz/internal/eds/pkg/directory"
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

	dsr4.RegisterReaderServer(s, inProcDirectory.Reader4())
	dsw4.RegisterWriterServer(s, inProcDirectory.Writer4())

	dsm3.RegisterModelServer(s, inProcDirectory.Model3())
	dsr3.RegisterReaderServer(s, inProcDirectory.Reader3())
	dsw3.RegisterWriterServer(s, inProcDirectory.Writer3())
	dse3.RegisterExporterServer(s, inProcDirectory.Exporter3())
	dsi3.RegisterImporterServer(s, inProcDirectory.Importer3())

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
