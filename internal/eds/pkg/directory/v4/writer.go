package v4

import (
	"context"

	dsw4 "github.com/aserto-dev/go-directory/aserto/directory/writer/v4"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rs/zerolog"
)

type Writer struct {
	logger *zerolog.Logger
	store  *bdb.BoltDB
}

var _ = dsw4.WriterServer(&Writer{})

func NewWriter(logger *zerolog.Logger, store *bdb.BoltDB) *Writer {
	return &Writer{
		logger: logger,
		store:  store,
	}
}

func (w *Writer) SetManifest(ctx context.Context, req *dsw4.SetManifestRequest) (*dsw4.SetManifestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetManifest not implemented")
}

func (w *Writer) DeleteManifest(ctx context.Context, req *dsw4.DeleteManifestRequest) (*dsw4.DeleteManifestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteManifest not implemented")
}

func (w *Writer) SetObject(ctx context.Context, req *dsw4.SetObjectRequest) (*dsw4.SetObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetObject not implemented")
}

func (w *Writer) DeleteObject(ctx context.Context, req *dsw4.DeleteObjectRequest) (*dsw4.DeleteObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteObject not implemented")
}

func (w *Writer) SetRelation(ctx context.Context, req *dsw4.SetRelationRequest) (*dsw4.SetRelationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetRelation not implemented")
}

func (w *Writer) DeleteRelation(ctx context.Context, req *dsw4.DeleteRelationRequest) (*dsw4.DeleteRelationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRelation not implemented")
}

func (w *Writer) Import(grpc.BidiStreamingServer[dsw4.ImportRequest, dsw4.ImportResponse]) error {
	return status.Errorf(codes.Unimplemented, "method Import not implemented")
}
