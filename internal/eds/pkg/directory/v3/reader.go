package v3

import (
	"context"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
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
	dsr.UnimplementedReaderServer

	logger *zerolog.Logger
	store  *bdb.BoltDB
}

func NewReader(logger *zerolog.Logger, store *bdb.BoltDB) *Reader {
	return &Reader{
		logger: logger,
		store:  store,
	}
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

		resp.Result = ds.PatchObjectRead(obj)

		return nil
	})

	return resp, err
}

// GetObjectMany, get multiple object instances by type+id, in a single request.
func (s *Reader) GetObjectMany(ctx context.Context, req *dsr.GetObjectManyRequest) (*dsr.GetObjectManyResponse, error) {
	resp := &dsr.GetObjectManyResponse{Results: []*dsc.Object{}}

	if err := validator.GetObjectManyRequest(req); err != nil {
		return resp, err
	}

	// validate all object identifiers first.
	for _, i := range req.GetParam() {
		if err := ds.ObjectIdentifier(i).Validate(s.store.MC()); err != nil {
			return resp, err
		}
	}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		for _, i := range req.GetParam() {
			obj, err := bdb.Get[dsc.Object](ctx, tx, bdb.ObjectsPath, ds.ObjectIdentifier(i).Key())
			if err != nil {
				return err
			}

			resp.Results = append(resp.GetResults(), ds.PatchObjectRead(obj))
		}

		return nil
	})

	return resp, err
}

// GetObjects, gets (all) object instances, optionally filtered by object type, as a paginated array of objects.
func (s *Reader) GetObjects(ctx context.Context, req *dsr.GetObjectsRequest) (*dsr.GetObjectsResponse, error) {
	resp := &dsr.GetObjectsResponse{Results: []*dsc.Object{}, Page: &dsc.PaginationResponse{}}

	if err := validator.GetObjectsRequest(req); err != nil {
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
			return ds.PatchObjectRead(x)
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
	resp := &dsr.GetRelationsResponse{
		Results: []*dsc.Relation{},
		Objects: map[string]*dsc.Object{},
		Page:    &dsc.PaginationResponse{},
	}

	if err := validator.GetRelationsRequest(req); err != nil {
		return resp, err
	}

	if req.GetPage() == nil {
		req.Page = &dsc.PaginationRequest{Size: x.MaxPageSize}
	}

	getRelations := ds.GetRelations(req)
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

// CheckRelation, check if subject has the specified relation to a resource (object).
func (s *Reader) CheckRelation(ctx context.Context, req *dsr.CheckRelationRequest) (*dsr.CheckRelationResponse, error) {
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

func (*Reader) getWithObjects(ctx context.Context, tx *bolt.Tx, relations []*dsc.Relation) map[string]*dsc.Object {
	objects := map[string]*dsc.Object{}

	for _, r := range relations {
		rel := ds.Relation(r)

		sub, err := bdb.Get[dsc.Object](ctx, tx, bdb.ObjectsPath, ds.ObjectIdentifier(rel.Subject()).Key())
		if err != nil {
			sub = &dsc.Object{Type: rel.SubjectType, Id: rel.SubjectId}
		}

		objects[ds.Object(sub).StrKey()] = sub

		obj, err := bdb.Get[dsc.Object](ctx, tx, bdb.ObjectsPath, ds.ObjectIdentifier(rel.Object()).Key())
		if err != nil {
			obj = &dsc.Object{Type: rel.ObjectType, Id: rel.ObjectId}
		}

		objects[ds.Object(obj).StrKey()] = obj
	}

	return objects
}
