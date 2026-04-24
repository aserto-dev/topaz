package app

import (
	"context"
	"net/http"

	dsr3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
	dsw3 "github.com/aserto-dev/topaz/api/directory/v4/writer"
	dsOpenAPI "github.com/aserto-dev/topaz/api/openapi"
	"github.com/aserto-dev/topaz/internal/eds/pkg/directory"
	"github.com/aserto-dev/topaz/topazd/service/builder"

	dsa1 "github.com/authzen/access.go/api/access/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/samber/lo"
	"google.golang.org/grpc"
)

const (
	readerService = "reader"
	writerService = "writer"
	accessService = "access"
)

type EdgeDir struct {
	dir *directory.Directory
}

var _ builder.ServiceTypes = (*EdgeDir)(nil)

func NewEdgeDir(edge *directory.Directory) (*EdgeDir, error) {
	return &EdgeDir{
		dir: edge,
	}, nil
}

func (e *EdgeDir) Close() {
	if e.dir != nil {
		e.dir.Close()
	}
}

func (e *EdgeDir) AvailableServices() []string {
	return []string{readerService, writerService, accessService}
}

func (e *EdgeDir) GetGRPCRegistrations(services ...string) builder.GRPCRegistrations {
	return func(server *grpc.Server) {
		// if lo.Contains(services, modelService) {
		// 	dsm3.RegisterModelServer(server, e.dir.Model3())
		// }

		if lo.Contains(services, readerService) {
			dsr3.RegisterReaderServer(server, e.dir.Reader3())
			dsa1.RegisterAccessServer(server, e.dir.Access1())
		}

		if lo.Contains(services, writerService) {
			dsw3.RegisterWriterServer(server, e.dir.Writer3())
		}

		// if lo.Contains(services, importerService) {
		// 	dsi3.RegisterImporterServer(server, e.dir.Importer3())
		// }

		// if lo.Contains(services, exporterService) {
		// 	dse3.RegisterExporterServer(server, e.dir.Exporter3())
		// }
	}
}

func (e *EdgeDir) GetGatewayRegistration(port string, services ...string) builder.HandlerRegistrations {
	return func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
		// if lo.Contains(services, modelService) {
		// 	err := dsm3.RegisterModelHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	// if err := dsm3stream.RegisterModelStreamHandlersFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		// 	// 	return err
		// 	// }
		// }

		if lo.Contains(services, readerService) {
			{
				err := dsr3.RegisterReaderHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
				if err != nil {
					return err
				}
			}
			{
				err := dsa1.RegisterAccessHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
				if err != nil {
					return err
				}
			}
		}

		if lo.Contains(services, writerService) {
			err := dsw3.RegisterWriterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
		}

		if len(services) > 0 {
			if err := mux.HandlePath(http.MethodGet, directoryOpenAPISpec, dsOpenAPIHandler(port, services...)); err != nil {
				return err
			}
		}

		return nil
	}
}

const (
	directoryOpenAPISpec string = "api/directory/openapi/directory.openapi.json"
)

func dsOpenAPIHandler(port string, services ...string) func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	handler := dsOpenAPI.OpenAPIHandler(port, services...)

	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		handler(w, r)
	}
}
