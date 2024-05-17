package inproc

import (
	"context"
	"net"

	dse "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"

	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/gerr"
	eds "github.com/aserto-dev/go-edge-ds"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/rs/zerolog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func NewServer(ctx context.Context, logger *zerolog.Logger, cfg *directory.Config) (*grpc.ClientConn, func()) {
	buffer := 1024 * 1024
	listener := bufconn.Listen(buffer)

	dsLogger := logger.With().Str("component", "ds").Logger()

	inProcDirectory, err := eds.New(context.Background(), cfg, &dsLogger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to start edge directory server")
	}

	errMiddleware := gerr.NewErrorMiddleware()
	s := grpc.NewServer(
		grpc.UnaryInterceptor(errMiddleware.Unary()),
		grpc.StreamInterceptor(errMiddleware.Stream()),
	)

	dsm.RegisterModelServer(s, inProcDirectory.Model3())
	dsr.RegisterReaderServer(s, inProcDirectory.Reader3())
	dsw.RegisterWriterServer(s, inProcDirectory.Writer3())
	dse.RegisterExporterServer(s, inProcDirectory.Exporter3())
	dsi.RegisterImporterServer(s, inProcDirectory.Importer3())

	go func() {
		if err := s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	// nolint: staticcheck // bufConn does not seem to work with the default DNS provided by grpc.NewClient.
	conn, _ := grpc.DialContext(ctx, "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

	return conn, s.GracefulStop
}
