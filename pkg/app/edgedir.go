package app

import (
	"context"
	"net/http"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	dsm3stream "github.com/aserto-dev/go-directory/pkg/gateway/model/v3"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	dsOpenAPI "github.com/aserto-dev/openapi-directory/publish/directory"
	"github.com/aserto-dev/topaz/pkg/service/builder"

	dsa1 "github.com/authzen/access.go/api/access/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/samber/lo"
	"google.golang.org/grpc"
)

type EdgeDir struct {
	dir *directory.Directory
}

const (
	EnvTopazAuthZEN = "TOPAZ_AUTHZEN"
)

const (
	modelService    = "model"
	readerService   = "reader"
	writerService   = "writer"
	exporterService = "exporter"
	importerService = "importer"
	accessService   = "access"
)

func NewEdgeDir(edge *directory.Directory) (ServiceTypes, error) {
	return &EdgeDir{
		dir: edge,
	}, nil
}

func (e *EdgeDir) Cleanups() []func() {
	if e.dir != nil {
		return []func(){e.dir.Close}
	}
	return nil
}

func (e *EdgeDir) AvailableServices() []string {
	return []string{modelService, readerService, writerService, exporterService, importerService, accessService}
}

func (e *EdgeDir) GetGRPCRegistrations(services ...string) builder.GRPCRegistrations {
	return func(server *grpc.Server) {
		if lo.Contains(services, modelService) {
			dsm3.RegisterModelServer(server, e.dir.Model3())
		}
		if lo.Contains(services, readerService) {
			dsr3.RegisterReaderServer(server, e.dir.Reader3())
			dsa1.RegisterAccessServer(server, e.dir.Access1())
		}
		if lo.Contains(services, writerService) {
			dsw3.RegisterWriterServer(server, e.dir.Writer3())
		}
		if lo.Contains(services, importerService) {
			dsi3.RegisterImporterServer(server, e.dir.Importer3())
		}
		if lo.Contains(services, exporterService) {
			dse3.RegisterExporterServer(server, e.dir.Exporter3())
		}
	}
}

func (e *EdgeDir) GetGatewayRegistration(port string, services ...string) builder.HandlerRegistrations {
	return func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
		if lo.Contains(services, modelService) {
			err := dsm3.RegisterModelHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
			if err := dsm3stream.RegisterModelStreamHandlersFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
				return err
			}
		}
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
	directoryOpenAPISpec string = "/directory/openapi.json"
)

func dsOpenAPIHandler(port string, services ...string) func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	handler := dsOpenAPI.OpenAPIHandler(port, services...)
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		handler(w, r)
	}
}
