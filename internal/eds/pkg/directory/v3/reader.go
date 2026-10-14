package v3

import (
	"context"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dse "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/validator"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/eds/pkg/ds"
	"github.com/aserto-dev/topaz/internal/eds/pkg/x"
	"github.com/pkg/errors"

	"github.com/go-http-utils/headers"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Reader struct {
	logger *zerolog.Logger
	store  *bdb.BoltDB
}

var _ dsr.ReaderServer = (*Reader)(nil)

func NewReader(logger *zerolog.Logger, store *bdb.BoltDB) *Reader {
	return &Reader{
		logger: logger,
		store:  store,
	}
}

// GetManifest, returns the manifest document (yaml).
func (s *Reader) GetManifest(ctx context.Context, req *dsr.GetManifestRequest) (*dsr.GetManifestResponse, error) {
	resp := &dsr.GetManifestResponse{}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		man, err := bdb.Get[dsc.Manifest](ctx, tx, bdb.ManifestPathV3, bdb.ManifestKeyV3)
		if err != nil {
			return err
		}

		resp.Manifest = man

		return nil
	})

	return resp, err
}

// GetModel, returns in the model representation of the manifest.
func (s *Reader) GetModel(ctx context.Context, req *dsr.GetModelRequest) (*dsr.GetModelResponse, error) {
	resp := &dsr.GetModelResponse{}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		mod, err := bdb.Get[dsc.Model](ctx, tx, bdb.ModelPathV3, bdb.ModelKeyV3)
		if err != nil {
			return err
		}

		resp.Model = mod

		return nil
	})

	return resp, err
}

// GetObject, get single object instance.
func (s *Reader) GetObject(ctx context.Context, req *dsr.GetObjectRequest) (*dsr.GetObjectResponse, error) {
	resp := &dsr.GetObjectResponse{}

	if err := validator.GetObjectRequest(req); err != nil {
		return resp, err
	}

	objIdent := ds.ObjectIdentifier(&dsc.ObjectIdentifier{ObjectType: req.GetObjectType(), ObjectId: req.GetObjectId()})
	if err := objIdent.Validate(s.store.MC()); err != nil {
		return resp, err
	}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		obj, err := bdb.Get[dsc.Object](ctx, tx, bdb.ObjectsPath, objIdent.Key())
		if err != nil {
			return err
		}

		inMD, _ := metadata.FromIncomingContext(ctx)
		// optimistic concurrency check
		if lo.Contains(inMD.Get(headers.IfNoneMatch), obj.GetEtag()) {
			_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "304"))

			return nil
		}

		if req.GetWithRelations() {
			// incoming object relations of object instance
			// (result.type == incoming.subject.type && result.key == incoming.subject.key)
			incoming, err := bdb.Scan[dsc.Relation](ctx, tx, bdb.RelationsSubPath, ds.Object(obj).Key())
			if err != nil {
				return err
			}

			resp.Relations = append(resp.Relations, incoming...)

			// outgoing object relations of object instance
			// (result.type == outgoing.object.type && result.key == outgoing.object.key)
			outgoing, err := bdb.Scan[dsc.Relation](ctx, tx, bdb.RelationsObjPath, ds.Object(obj).Key())
			if err != nil {
				return err
			}

			resp.Relations = append(resp.Relations, outgoing...)

			s.logger.Trace().Msg("get object with relations")
		}

		resp.Result = obj // ds.PatchObjectRead(obj)

		return nil
	})

	return resp, err
}

// GetObjectMany, get multiple object instances by type+id, in a single request.
func (s *Reader) GetObjectMany(ctx context.Context, req *dsr.GetObjectManyRequest) (*dsr.GetObjectManyResponse, error) {
	resp := &dsr.GetObjectManyResponse{Results: []*dsc.Object{}}
	return resp, status.Error(codes.Unimplemented, "reader.GetObjectMany is deprecated, use reader.Batch instead")
}

// GetObjects, gets (all) object instances, optionally filtered by object type, as a paginated array of objects.
func (s *Reader) GetObjects(ctx context.Context, req *dsr.GetObjectsRequest) (*dsr.GetObjectsResponse, error) {
	resp := &dsr.GetObjectsResponse{Results: []*dsc.Object{}, Page: &dsc.PaginationResponse{}}
	return resp, status.Error(codes.Unimplemented, "reader.GetObjects is deprecated, use reader.ListObjects instead")
}

func (s *Reader) ListObjects(ctx context.Context, req *dsr.ListObjectsRequest) (*dsr.ListObjectsResponse, error) {
	resp := &dsr.ListObjectsResponse{Results: []*dsc.Object{}, Page: &dsc.PaginationResponse{}}

	if err := validator.ListObjectsRequest(req); err != nil {
		return resp, err
	}

	if req.GetPage() == nil {
		req.Page = &dsc.PaginationRequest{Size: x.MaxPageSize}
	}

	opts := []bdb.ScanOption{
		bdb.WithPageSize(req.GetPage().GetSize()),
		bdb.WithPageToken(req.GetPage().GetToken()),
	}

	if req.GetObjectType() != "" {
		oid := ds.ObjectIdentifier(&dsc.ObjectIdentifier{ObjectType: req.GetObjectType()})
		if err := ds.ObjectSelector(oid.ObjectIdentifier).Validate(s.store.MC()); err != nil {
			return resp, err
		}

		opts = append(opts, bdb.WithKeyFilter(oid.Key()))
	}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		iter, err := bdb.NewPageIterator[dsc.Object](ctx, tx, bdb.ObjectsPath, opts...)
		if err != nil {
			return err
		}

		iter.Next()

		resp.Results = lo.Map(iter.Value(), func(x *dsc.Object, _ int) *dsc.Object {
			return x // ds.PatchObjectRead(x)
		})

		resp.Page = &dsc.PaginationResponse{NextToken: iter.NextToken()}

		return nil
	})

	return resp, err
}

// GetRelation, get a single relation instance based on subject, relation, object filter.
func (s *Reader) GetRelation(ctx context.Context, req *dsr.GetRelationRequest) (*dsr.GetRelationResponse, error) {
	resp := &dsr.GetRelationResponse{
		Result:  &dsc.Relation{},
		Objects: map[string]*dsc.Object{},
	}

	if err := validator.GetRelationRequest(req); err != nil {
		return resp, err
	}

	getRelation := ds.GetRelation(req)
	if err := getRelation.Validate(s.store.MC()); err != nil {
		return resp, err
	}

	filter := ds.RelationIdentifierBuffer()
	defer ds.ReturnRelationIdentifierBuffer(filter)

	path, err := getRelation.PathAndFilter(filter)
	if err != nil {
		return resp, err
	}

	err = s.store.DB().View(func(tx *bolt.Tx) error {
		relations, err := bdb.Scan[dsc.Relation](ctx, tx, path, filter.Bytes())
		if err != nil {
			return err
		}

		if len(relations) == 0 {
			return bdb.ErrKeyNotFound
		}

		if len(relations) != 1 {
			return bdb.ErrMultipleResults
		}

		dbRel := relations[0]
		resp.Result = dbRel

		inMD, _ := metadata.FromIncomingContext(ctx)
		if lo.Contains(inMD.Get(headers.IfNoneMatch), dbRel.GetEtag()) {
			_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "304"))

			return nil
		}

		if req.GetWithObjects() {
			relations := []*dsc.Relation{resp.GetResult()}
			resp.Objects = s.getWithObjects(ctx, tx, relations)
		}

		return nil
	})

	return resp, err
}

// GetRelations, gets paginated set of relation instances based on subject, relation, object filter.
func (s *Reader) GetRelations(ctx context.Context, req *dsr.GetRelationsRequest) (*dsr.GetRelationsResponse, error) {
	resp := &dsr.GetRelationsResponse{Results: []*dsc.Relation{}, Objects: map[string]*dsc.Object{}, Page: &dsc.PaginationResponse{}}
	return resp, status.Error(codes.Unimplemented, "reader.GetRelations is deprecated, use reader.ListRelations instead")
}

func (s *Reader) ListRelations(ctx context.Context, req *dsr.ListRelationsRequest) (*dsr.ListRelationsResponse, error) {
	resp := &dsr.ListRelationsResponse{
		Results: []*dsc.Relation{},
		Objects: map[string]*dsc.Object{},
		Page:    &dsc.PaginationResponse{},
	}

	if err := validator.ListRelationsRequest(req); err != nil {
		return resp, err
	}

	if req.GetPage() == nil {
		req.Page = &dsc.PaginationRequest{Size: x.MaxPageSize}
	}

	getRelations := ds.ListRelations(req)
	if err := getRelations.Validate(s.store.MC()); err != nil {
		return resp, err
	}

	keyFilter := ds.RelationIdentifierBuffer()
	defer ds.ReturnRelationIdentifierBuffer(keyFilter)

	path, valueFilter := getRelations.RelationValueFilter(keyFilter)

	opts := []bdb.ScanOption{
		bdb.WithPageToken(req.GetPage().GetToken()),
		bdb.WithKeyFilter(keyFilter.Bytes()),
	}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		iter, err := bdb.NewScanIterator[dsc.Relation](ctx, tx, path, opts...)
		if err != nil {
			return err
		}

		for iter.Next() {
			if !valueFilter(iter.Value()) {
				continue
			}

			resp.Results = append(resp.GetResults(), iter.Value())

			if int64(req.GetPage().GetSize()) == int64(len(resp.GetResults())) {
				if iter.Next() {
					resp.Page.NextToken = iter.Key()
				}

				break
			}
		}

		if req.GetWithObjects() {
			resp.Objects = s.getWithObjects(ctx, tx, resp.GetResults())
		}

		return nil
	})

	return resp, err
}

// Check, if subject is permitted to access resource (object).
func (s *Reader) Check(ctx context.Context, req *dsr.CheckRequest) (*dsr.CheckResponse, error) {
	resp := &dsr.CheckResponse{}

	if err := validator.CheckRequest(req); err != nil {
		resp.Check = false
		resp.Context = ds.SetContextWithReason(err)

		return resp, nil
	}

	check := ds.Check(req)
	if err := check.Validate(s.store.MC()); err != nil {
		resp.Check = false

		if err := errors.Unwrap(err); err != nil {
			resp.Context = ds.SetContextWithReason(err)
			return resp, nil
		}

		resp.Context = ds.SetContextWithReason(err)

		return resp, nil
	}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		var err error

		resp, err = check.Exec(ctx, tx, s.store.MC())

		return err
	})
	if err != nil {
		resp.Context = ds.SetContextWithReason(err)
	}

	return resp, nil
}

// Checks, execute multiple check requests in parallel.
func (s *Reader) Checks(ctx context.Context, req *dsr.ChecksRequest) (*dsr.ChecksResponse, error) {
	resp := &dsr.ChecksResponse{}

	checks := ds.Checks(req)
	if err := checks.Validate(s.store.MC()); err != nil {
		return resp, err
	}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		var err error

		resp, err = checks.Exec(ctx, tx, s.store.MC())

		return err
	})
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// CheckPermission is obsolete, use Check instead.
func (s *Reader) CheckPermission(_ context.Context, _ *dsr.CheckPermissionRequest) (*dsr.CheckPermissionResponse, error) {
	return &dsr.CheckPermissionResponse{}, status.Error(codes.Unimplemented, "check permission is obsolete, use check instead")
}

// CheckRelation is obsolete, use Check instead.
func (s *Reader) CheckRelation(_ context.Context, _ *dsr.CheckRelationRequest) (*dsr.CheckRelationResponse, error) {
	return &dsr.CheckRelationResponse{}, status.Error(codes.Unimplemented, "check relation is obsolete, use check instead")
}

// GetGraph, return graph of connected objects and relations for requested anchor subject/object.
func (s *Reader) GetGraph(ctx context.Context, req *dsr.GetGraphRequest) (*dsr.GetGraphResponse, error) {
	resp := &dsr.GetGraphResponse{}

	if err := validator.GetGraphRequest(req); err != nil {
		return &dsr.GetGraphResponse{}, err
	}

	getGraph := ds.GetGraph(req)
	if err := getGraph.Validate(s.store.MC()); err != nil {
		return resp, err
	}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		var err error

		results, err := getGraph.Exec(ctx, tx, s.store.MC())
		if err != nil {
			return err
		}

		resp = results

		return nil
	})

	return resp, err
}

// Export, exports manifest, objects and relations instances as a stream.
func (s *Reader) Export(req *dsr.ExportRequest, stream grpc.ServerStreamingServer[dsr.ExportResponse]) error {
	logger := s.logger.With().Str("method", "Export").Interface("req", req).Logger()

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		// stats mode, short circuits when enabled
		if req.GetOptions()&uint32(dse.Option_OPTION_STATS) != 0 {
			if err := exportStats(tx, stream, req.GetOptions()); err != nil {
				logger.Error().Err(err).Msg("export_stats")
				return err
			}

			return nil
		}

		if req.GetOptions()&uint32(dse.Option_OPTION_DATA_OBJECTS) != 0 {
			if err := exportObjects(tx, stream); err != nil {
				logger.Error().Err(err).Msg("export_objects")
				return err
			}
		}

		if req.GetOptions()&uint32(dse.Option_OPTION_DATA_RELATIONS) != 0 {
			if err := exportRelations(tx, stream); err != nil {
				logger.Error().Err(err).Msg("export_relations")
				return err
			}
		}

		return nil
	})

	return err
}

// Batch, allows sending multiple reader requests as a single request, executed within a single transaction scope.
func (s *Reader) Batch(ctx context.Context, req *dsr.BatchRequest) (*dsr.BatchResponse, error) {
	return nil, status.Error(codes.Unimplemented, "reader.Batch not implemented")
}

func (*Reader) getWithObjects(ctx context.Context, tx *bolt.Tx, relations []*dsc.Relation) map[string]*dsc.Object {
	objects := map[string]*dsc.Object{}

	for _, r := range relations {
		rel := ds.Relation(r)

		sub, err := bdb.Get[dsc.Object](ctx, tx, bdb.ObjectsPath, ds.ObjectIdentifier(rel.Subject()).Key())
		if err != nil {
			sub = &dsc.Object{ObjectType: rel.SubjectType, ObjectId: rel.SubjectId}
		}

		objects[ds.Object(sub).StrKey()] = sub

		obj, err := bdb.Get[dsc.Object](ctx, tx, bdb.ObjectsPath, ds.ObjectIdentifier(rel.Object()).Key())
		if err != nil {
			obj = &dsc.Object{ObjectType: rel.ObjectType, ObjectId: rel.ObjectId}
		}

		objects[ds.Object(obj).StrKey()] = obj
	}

	return objects
}
