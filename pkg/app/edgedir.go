package app

import (
	"context"
	"net/http"
	"strconv"

	dse2 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v2"
	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi2 "github.com/aserto-dev/go-directory/aserto/directory/importer/v2"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw2 "github.com/aserto-dev/go-directory/aserto/directory/writer/v2"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	dsm3stream "github.com/aserto-dev/go-directory/pkg/gateway/model/v3"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	dsOpenAPI "github.com/aserto-dev/openapi-directory/publish/directory"
	builder "github.com/aserto-dev/service-host"
	"github.com/aserto-dev/topaz/pkg/rapidoc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/samber/lo"
	"google.golang.org/grpc"
)

type EdgeDir struct {
	dir *directory.Directory
}

const (
	modelService    = "model"
	readerService   = "reader"
	writerService   = "writer"
	exporterService = "exporter"
	importerService = "importer"
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
	return []string{modelService, readerService, writerService, exporterService, importerService}
}

func (e *EdgeDir) GetGRPCRegistrations(services ...string) builder.GRPCRegistrations {
	return func(server *grpc.Server) {
		if lo.Contains(services, modelService) {
			dsm3.RegisterModelServer(server, e.dir.Model3())
		}
		if lo.Contains(services, readerService) {
			if e.dir.Config().EnableV2 {
				dsr2.RegisterReaderServer(server, e.dir.Reader2())
			}
			dsr3.RegisterReaderServer(server, e.dir.Reader3())
		}
		if lo.Contains(services, writerService) {
			if e.dir.Config().EnableV2 {
				dsw2.RegisterWriterServer(server, e.dir.Writer2())
			}
			dsw3.RegisterWriterServer(server, e.dir.Writer3())
		}
		if lo.Contains(services, importerService) {
			if e.dir.Config().EnableV2 {
				dsi2.RegisterImporterServer(server, e.dir.Importer2())
			}
			dsi3.RegisterImporterServer(server, e.dir.Importer3())
		}
		if lo.Contains(services, exporterService) {
			if e.dir.Config().EnableV2 {
				dse2.RegisterExporterServer(server, e.dir.Exporter2())
			}
			dse3.RegisterExporterServer(server, e.dir.Exporter3())
		}
	}
}

func (e *EdgeDir) GetGatewayRegistration(services ...string) builder.HandlerRegistrations {
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
			err := dsr3.RegisterReaderHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
		}
		if lo.Contains(services, writerService) {
			err := dsw3.RegisterWriterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
		}
		if lo.Contains(services, importerService) {
			err := dsi3.RegisterImporterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
		}
		if lo.Contains(services, exporterService) {
			err := dse3.RegisterExporterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
		}

		if len(services) > 0 {
			if err := mux.HandlePath(http.MethodGet, directoryOpenAPISpec, dsOpenAPIHandler); err != nil {
				return err
			}
			if err := mux.HandlePath(http.MethodGet, directoryOpenAPIDocs, dsOpenAPIDocsHandler()); err != nil {
				return err
			}
		}

		return nil
	}
}

const (
	directoryOpenAPISpec string = "/directory/openapi.json"
	directoryOpenAPIDocs string = "/directory/docs"
)

func dsOpenAPIHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	buf, err := dsOpenAPI.Static().ReadFile("openapi.json")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.FormatInt(int64(len(buf)), 10))
	_, _ = w.Write(buf)
}

func dsOpenAPIDocsHandler() runtime.HandlerFunc {
	h := rapidoc.Handler(&rapidoc.Opts{
		Path:    directoryOpenAPIDocs,
		SpecURL: directoryOpenAPISpec,
		Title:   "Aserto Directory HTTP API",
	}, nil)

	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		h.ServeHTTP(w, r)
	}
}
