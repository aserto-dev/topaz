package v3

import (
	dsi "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Importer struct {
	logger *zerolog.Logger
}

var _ dsi.ImporterServer = (*Importer)(nil)

func NewImporter(logger *zerolog.Logger, _ *bdb.BoltDB) *Importer {
	return &Importer{logger: logger}
}

func (s *Importer) Import(stream dsi.Importer_ImportServer) error {
	return status.Error(codes.Unimplemented, "importer.Import is deprecated, use writer.Import instead")
}
