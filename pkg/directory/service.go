package directory

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	ds "github.com/aserto-dev/go-edge-ds/pkg/directory"
	dsa1 "github.com/authzen/access.go/api/access/v1"

	"github.com/aserto-dev/topaz/pkg/config"
)

type Service struct {
	*ds.Directory
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

func (s *Service) RegisterAccessServer(server *grpc.Server) {
	if s.Directory != nil {
		dsa1.RegisterAccessServer(server, s.Access1())
	}
}

func (s *Service) RegisterAccessGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) error {
	if s.Directory != nil {
		return dsa1.RegisterAccessHandlerFromEndpoint(ctx, mux, endpoint, opts)
	}

	return nil
}

func (s *Service) RegisterReaderServer(server *grpc.Server) {
	if s.Directory != nil {
		dsr3.RegisterReaderServer(server, s.Reader3())
	}
}

func (s *Service) RegisterReaderGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) error {
	if s.Directory != nil {
		return dsr3.RegisterReaderHandlerFromEndpoint(ctx, mux, endpoint, opts)
	}

	return nil
}

func (s *Service) RegisterWriterServer(server *grpc.Server) {
	if s.Directory != nil {
		dsw3.RegisterWriterServer(server, s.Writer3())
	}
}

func (s *Service) RegisterWriterGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) error {
	if s.Directory != nil {
		return dsw3.RegisterWriterHandlerFromEndpoint(ctx, mux, endpoint, opts)
	}

	return nil
}

func (s *Service) RegisterModelServer(server *grpc.Server) {
	if s.Directory != nil {
		dsm3.RegisterModelServer(server, s.Model3())
	}
}

func (s *Service) RegisterModelGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) error {
	if s.Directory != nil {
		return dsm3.RegisterModelHandlerFromEndpoint(ctx, mux, endpoint, opts)
	}

	return nil
}

func (s *Service) RegisterImporterServer(server *grpc.Server) {
	if s.Directory != nil {
		dsi3.RegisterImporterServer(server, s.Importer3())
	}
}

func (s *Service) RegisterExporterServer(server *grpc.Server) {
	if s.Directory != nil {
		dse3.RegisterExporterServer(server, s.Exporter3())
	}
}
