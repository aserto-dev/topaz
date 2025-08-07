package directory

import (
	"context"
	"net/http"

	gorilla "github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	dsm3stream "github.com/aserto-dev/go-directory/pkg/gateway/model/v3"
	ds "github.com/aserto-dev/go-edge-ds/pkg/directory"
	dsOpenAPI "github.com/aserto-dev/openapi-directory/publish/directory"
	dsa1 "github.com/authzen/access.go/api/access/v1"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/aserto-dev/topaz/pkg/service"
)

type Service struct {
	server *ds.Directory
}

func New(ctx context.Context, cfg *Config) (*Service, error) {
	if cfg.Store.Provider != BoltDBStorePlugin {
		return nil, errors.Wrap(config.ErrConfig, "only boltdb provider is currently supported")
	}

	dir, err := ds.New(ctx, (*ds.Config)(&cfg.Store.Bolt), zerolog.Ctx(ctx))
	if err != nil {
		return nil, err
	}

	return &Service{dir}, nil
}

func (s *Service) Access() service.Registrar {
	return s.registrar(
		s.registerAccessServer,
		dsa1.RegisterAccessHandlerFromEndpoint,
		s.registerAccessWellKnownHandler,
	)
}

func (s *Service) Reader() service.Registrar {
	return s.registrar(
		s.registerReaderServer,
		dsr3.RegisterReaderHandlerFromEndpoint,
		service.NoHTTP,
	)
}

func (s *Service) Writer() service.Registrar {
	return s.registrar(
		s.registerWriterServer,
		dsw3.RegisterWriterHandlerFromEndpoint,
		service.NoHTTP,
	)
}

func (s *Service) Model() service.Registrar {
	return s.registrar(
		s.registerModelServer,
		s.registerModelGateway,
		service.NoHTTP,
	)
}

func (s *Service) Importer() service.Registrar {
	return s.registrar(
		s.registerImporterServer,
		service.NoGateway,
		service.NoHTTP,
	)
}

func (s *Service) Exporter() service.Registrar {
	return s.registrar(
		s.registerExporterServer,
		service.NoGateway,
		service.NoHTTP,
	)
}

func (s *Service) OpenAPIHandler(port string, services ...string) http.HandlerFunc {
	return dsOpenAPI.OpenAPIHandler(port, services...)
}

func (s *Service) registrar(g service.GRPCRegistrar, gw service.GatewayRegistrar, h service.HTTPRegistrar) service.Registrar {
	if s.server == nil {
		return service.Noop
	}

	return service.NewImpl(g, gw, h)
}

func (s *Service) registerAccessServer(server *grpc.Server) {
	dsa1.RegisterAccessServer(server, s.server.Access1())
}

func (s *Service) registerAccessWellKnownHandler(ctx context.Context, cfg *servers.HTTPServer, router *gorilla.Router) {
	baseURL, err := cfg.BaseURL()
	if err != nil {
		zerolog.Ctx(ctx).Fatal().Err(err).Msg("unable to register access service well-known handler.")
		return
	}

	router.HandleFunc(AuthZENConfiguration, WellKnownConfigHandler(baseURL)).Methods(http.MethodGet)
}

func (s *Service) registerReaderServer(server *grpc.Server) {
	dsr3.RegisterReaderServer(server, s.server.Reader3())
}

func (s *Service) registerWriterServer(server *grpc.Server) {
	dsw3.RegisterWriterServer(server, s.server.Writer3())
}

func (s *Service) registerModelServer(server *grpc.Server) {
	dsm3.RegisterModelServer(server, s.server.Model3())
}

func (s *Service) registerModelGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	if err := dsm3.RegisterModelHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		return err
	}

	return dsm3stream.RegisterModelStreamHandlersFromEndpoint(ctx, mux, endpoint, opts)
}

func (s *Service) registerImporterServer(server *grpc.Server) {
	dsi3.RegisterImporterServer(server, s.server.Importer3())
}

func (s *Service) registerExporterServer(server *grpc.Server) {
	dse3.RegisterExporterServer(server, s.server.Exporter3())
}
