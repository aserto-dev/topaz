package app

import (
	"context"

	"github.com/aserto-dev/go-directory/aserto/directory/exporter/v2"
	"github.com/aserto-dev/go-directory/aserto/directory/importer/v2"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v2"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	builder "github.com/aserto-dev/service-host"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type EdgeDir struct {
	dir *directory.Directory
}

const (
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

func (e *EdgeDir) AvailableServices() []string {
	return []string{readerService, writerService, exporterService, importerService}
}

func (e *EdgeDir) GetGRPCRegistrations(services ...string) builder.GRPCRegistrations {
	return func(server *grpc.Server) {
		if contains(services, "reader") {
			reader.RegisterReaderServer(server, e.dir.Reader2())
		}
		if contains(services, "writer") {
			writer.RegisterWriterServer(server, e.dir.Writer2())
		}
		if contains(services, "importer") {
			importer.RegisterImporterServer(server, e.dir.Importer2())
		}
		if contains(services, "exporter") {
			exporter.RegisterExporterServer(server, e.dir.Exporter2())
		}
	}
}

func (e *EdgeDir) GetGatewayRegistration(services ...string) builder.HandlerRegistrations {
	return func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
		// nolint: gocritic temporary disabled until 0.30 schema release/integration.
		// if contains(registeredServices, "reader") {
		// 	err := reader.RegisterReaderHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		// if contains(registeredServices, "writer") {
		// 	err := writer.RegisterWriterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		// if contains(registeredServices, "importer") {
		// 	err := importer.RegisterImporterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		// if contains(registeredServices, "exporter") {
		// 	err := exporter.RegisterExporterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		return nil
	}
}
