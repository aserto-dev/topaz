package v4

import (
	"context"

	dsr4 "github.com/aserto-dev/go-directory/aserto/directory/reader/v4"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rs/zerolog"
)

type Reader struct {
	logger *zerolog.Logger
	store  *bdb.BoltDB
}

var _ = dsr4.ReaderServer(&Reader{})

func NewReader(logger *zerolog.Logger, store *bdb.BoltDB) *Reader {
	return &Reader{
		logger: logger,
		store:  store,
	}
}

func (r *Reader) GetManifest(ctx context.Context, req *dsr4.GetManifestRequest) (*dsr4.GetManifestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetManifest not implemented")
}

func (r *Reader) GetModel(ctx context.Context, req *dsr4.GetModelRequest) (*dsr4.GetModelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModel not implemented")
}

func (r *Reader) GetObject(ctx context.Context, req *dsr4.GetObjectRequest) (*dsr4.GetObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetObject not implemented")
}

func (r *Reader) GetObjects(ctx context.Context, req *dsr4.GetObjectsRequest) (*dsr4.GetObjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetObjects not implemented")
}

func (r *Reader) ListObjects(ctx context.Context, req *dsr4.ListObjectsRequest) (*dsr4.ListObjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListObjects not implemented")
}

func (r *Reader) GetRelation(ctx context.Context, req *dsr4.GetRelationRequest) (*dsr4.GetRelationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRelation not implemented")
}

func (r *Reader) GetRelations(ctx context.Context, req *dsr4.GetRelationsRequest) (*dsr4.GetRelationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRelations not implemented")
}

func (r *Reader) ListRelations(ctx context.Context, req *dsr4.ListRelationsRequest) (*dsr4.ListRelationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRelations not implemented")
}

func (r *Reader) Check(ctx context.Context, req *dsr4.CheckRequest) (*dsr4.CheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Check not implemented")
}

func (r *Reader) Checks(ctx context.Context, req *dsr4.ChecksRequest) (*dsr4.ChecksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Checks not implemented")
}

func (r *Reader) GetGraph(ctx context.Context, req *dsr4.GetGraphRequest) (*dsr4.GetGraphResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGraph not implemented")
}

func (r *Reader) Export(req *dsr4.ExportRequest, stream grpc.ServerStreamingServer[dsr4.ExportResponse]) error {
	return status.Errorf(codes.Unimplemented, "method Export not implemented")
}
