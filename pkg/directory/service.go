package directory

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

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
	dsa1.RegisterAccessServer(server, s.Access1())
}

func (s *Service) RegisterAccessGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) error {
	return dsa1.RegisterAccessHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

func (s *Service) RegisterReaderServer(server *grpc.Server) {
	dsr3.RegisterReaderServer(server, s.Reader3())
}

func (s *Service) RegisterReaderGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) error {
	return dsr3.RegisterReaderHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

func (s *Service) RegisterWriterServer(server *grpc.Server) {
	dsw3.RegisterWriterServer(server, s.Writer3())
}

func (s *Service) RegisterWriterGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) error {
	return dsw3.RegisterWriterHandlerFromEndpoint(ctx, mux, endpoint, opts)
}
