package server

import (
	"context"
	"net"

	dse "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"

	"github.com/aserto-dev/topaz/internal/eds"
	"github.com/aserto-dev/topaz/internal/eds/pkg/directory"
	"github.com/aserto-dev/topaz/internal/grpc/middlewares/gerr"
	"github.com/rs/zerolog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type TestEdgeClient struct {
	V3 ClientV3
}

type ClientV3 struct {
	Model    dsm.ModelClient
	Reader   dsr.ReaderClient
	Writer   dsw.WriterClient
	Importer dsi.ImporterClient
	Exporter dse.ExporterClient
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

	dsm.RegisterModelServer(s, edgeDirServer.Model3())
	dsr.RegisterReaderServer(s, edgeDirServer.Reader3())
	dsw.RegisterWriterServer(s, edgeDirServer.Writer3())
	dse.RegisterExporterServer(s, edgeDirServer.Exporter3())
	dsi.RegisterImporterServer(s, edgeDirServer.Importer3())

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
			Model:    dsm.NewModelClient(conn),
			Reader:   dsr.NewReaderClient(conn),
			Writer:   dsw.NewWriterClient(conn),
			Importer: dsi.NewImporterClient(conn),
			Exporter: dse.NewExporterClient(conn),
		},
	}

	return &client, s.Stop
}
