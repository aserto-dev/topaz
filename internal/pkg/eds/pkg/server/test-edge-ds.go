package server

import (
	"context"
	"net"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"

	"github.com/aserto-dev/aserto-grpc/middlewares/gerr"
	"github.com/aserto-dev/topaz/internal/pkg/eds"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/directory"
	"github.com/rs/zerolog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type TestEdgeClient struct {
	V3 ClientV3
}

type ClientV3 struct {
	Model    dsm3.ModelClient
	Reader   dsr3.ReaderClient
	Writer   dsw3.WriterClient
	Importer dsi3.ImporterClient
	Exporter dse3.ExporterClient
}

const bufferSize int = 1024 * 1024

func NewTestEdgeServer(ctx context.Context, logger *zerolog.Logger, cfg *directory.Config) (*TestEdgeClient, func()) {
	listener := bufconn.Listen(bufferSize)

	edgeDSLogger := logger.With().Str("component", "api.edge-directory").Logger()

	edgeDirServer, err := eds.New(context.Background(), cfg, &edgeDSLogger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to start edge directory server")
	}

	errMiddleware := gerr.NewErrorMiddleware()
	s := grpc.NewServer(
		grpc.UnaryInterceptor(errMiddleware.Unary()),
		grpc.StreamInterceptor(errMiddleware.Stream()),
	)

	dsm3.RegisterModelServer(s, edgeDirServer.Model3())
	dsr3.RegisterReaderServer(s, edgeDirServer.Reader3())
	dsw3.RegisterWriterServer(s, edgeDirServer.Writer3())
	dse3.RegisterExporterServer(s, edgeDirServer.Exporter3())
	dsi3.RegisterImporterServer(s, edgeDirServer.Importer3())

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

	client := TestEdgeClient{
		V3: ClientV3{
			Model:    dsm3.NewModelClient(conn),
			Reader:   dsr3.NewReaderClient(conn),
			Writer:   dsw3.NewWriterClient(conn),
			Importer: dsi3.NewImporterClient(conn),
			Exporter: dse3.NewExporterClient(conn),
		},
	}

	return &client, s.Stop
}
