package v3

import (
	dse "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rs/zerolog"
)

type Exporter struct {
	logger *zerolog.Logger
}

var _ dse.ExporterServer = (*Exporter)(nil)

func NewExporter(logger *zerolog.Logger, _ *bdb.BoltDB) *Exporter {
	return &Exporter{logger: logger}
}

func (s *Exporter) Export(req *dse.ExportRequest, stream dse.Exporter_ExportServer) error {
	return status.Error(codes.Unimplemented, "exporter.Export is deprecated, use reader.Export instead")
}
