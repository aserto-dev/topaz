package v3

import (
	"context"

	dsm "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Model struct {
	logger *zerolog.Logger
}

var _ dsm.ModelServer = (*Model)(nil)

func NewModel(logger *zerolog.Logger, _ *bdb.BoltDB) *Model {
	return &Model{logger: logger}
}

var _ = dsm.ModelServer(&Model{})

func (s *Model) GetManifest(req *dsm.GetManifestRequest, stream dsm.Model_GetManifestServer) error {
	return status.Error(codes.Unimplemented, "model.GetManifest is deprecated, use reader.GetManifest")
}

func (s *Model) SetManifest(stream dsm.Model_SetManifestServer) error {
	return status.Error(codes.Unimplemented, "model.SetManifest is deprecated, use writer.SetManifest")
}

func (s *Model) DeleteManifest(ctx context.Context, req *dsm.DeleteManifestRequest) (*dsm.DeleteManifestResponse, error) {
	return &dsm.DeleteManifestResponse{}, status.Error(codes.Unimplemented, "model.DeleteManifest is deprecated, use writer.DeleteManifest")
}
