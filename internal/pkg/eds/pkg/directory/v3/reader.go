package v3

import (
	"context"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/validator"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/ds"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/x"
	"github.com/pkg/errors"

	"github.com/go-http-utils/headers"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
	grpcmd "google.golang.org/grpc/metadata"
)

type Reader struct {
	dsr3.UnimplementedReaderServer

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
func (s *Reader) GetObject(ctx context.Context, req *dsr3.GetObjectRequest) (*dsr3.GetObjectResponse, error) {
	resp := &dsr3.GetObjectResponse{}

	if err := validator.GetObjectRequest(req); err != nil {
		return resp, err
	}

	objIdent := ds.ObjectIdentifier(&dsc3.ObjectIdentifier{ObjectType: req.GetObjectType(), ObjectId: req.GetObjectId()})
	if err := objIdent.Validate(s.store.MC()); err != nil {
		return resp, err
	}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		obj, err := bdb.Get[dsc3.Object](ctx, tx, bdb.ObjectsPath, objIdent.Key())
		if err != nil {
			return err
		}

		inMD, _ := grpcmd.FromIncomingContext(ctx)
		// optimistic concurrency check
		if lo.Contains(inMD.Get(headers.IfNoneMatch), obj.GetEtag()) {
			_ = grpc.SetHeader(ctx, grpcmd.Pairs("x-http-code", "304"))

			return nil
		}

		if req.GetWithRelations() {
			// incoming object relations of object instance
			// (result.type == incoming.subject.type && result.key == incoming.subject.key)
			incoming, err := bdb.Scan[dsc3.Relation](ctx, tx, bdb.RelationsSubPath, ds.Object(obj).Key())
			if err != nil {
				return err
			}

			resp.Relations = append(resp.Relations, incoming...)

			// outgoing object relations of object instance
			// (result.type == outgoing.object.type && result.key == outgoing.object.key)
			outgoing, err := bdb.Scan[dsc3.Relation](ctx, tx, bdb.RelationsObjPath, ds.Object(obj).Key())
			if err != nil {
				return err
			}

			resp.Relations = append(resp.Relations, outgoing...)

			s.logger.Trace().Msg("get object with relations")
		}

		resp.Result = obj

		resp.Page = &dsc3.PaginationResponse{}

		return nil
	})

	return resp, err
}

// GetObjectMany, get multiple object instances by type+id, in a single request.
func (s *Reader) GetObjectMany(ctx context.Context, req *dsr3.GetObjectManyRequest) (*dsr3.GetObjectManyResponse, error) {
	resp := &dsr3.GetObjectManyResponse{Results: []*dsc3.Object{}}

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
			obj, err := bdb.Get[dsc3.Object](ctx, tx, bdb.ObjectsPath, ds.ObjectIdentifier(i).Key())
			if err != nil {
				return err
			}

			resp.Results = append(resp.GetResults(), obj)
		}

		return nil
	})

	return resp, err
}

// GetObjects, gets (all) object instances, optionally filtered by object type, as a paginated array of objects.
func (s *Reader) GetObjects(ctx context.Context, req *dsr3.GetObjectsRequest) (*dsr3.GetObjectsResponse, error) {
	resp := &dsr3.GetObjectsResponse{Results: []*dsc3.Object{}, Page: &dsc3.PaginationResponse{}}

	if err := validator.GetObjectsRequest(req); err != nil {
		return resp, err
	}

	if req.GetPage() == nil {
		req.Page = &dsc3.PaginationRequest{Size: x.MaxPageSize}
	}

	opts := []bdb.ScanOption{
		bdb.WithPageSize(req.GetPage().GetSize()),
		bdb.WithPageToken(req.GetPage().GetToken()),
	}

	if req.GetObjectType() != "" {
		oid := ds.ObjectIdentifier(&dsc3.ObjectIdentifier{ObjectType: req.GetObjectType()})
		if err := ds.ObjectSelector(oid.ObjectIdentifier).Validate(s.store.MC()); err != nil {
			return resp, err
		}

		opts = append(opts, bdb.WithKeyFilter(oid.Key()))
	}

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		iter, err := bdb.NewPageIterator[dsc3.Object](ctx, tx, bdb.ObjectsPath, opts...)
		if err != nil {
			return err
		}

		iter.Next()

		resp.Results = iter.Value()
		resp.Page = &dsc3.PaginationResponse{NextToken: iter.NextToken()}

		return nil
	})

	return resp, err
}

// GetRelation, get a single relation instance based on subject, relation, object filter.
func (s *Reader) GetRelation(ctx context.Context, req *dsr3.GetRelationRequest) (*dsr3.GetRelationResponse, error) {
	resp := &dsr3.GetRelationResponse{
		Result:  &dsc3.Relation{},
		Objects: map[string]*dsc3.Object{},
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
		relations, err := bdb.Scan[dsc3.Relation](ctx, tx, path, filter.Bytes())
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

		inMD, _ := grpcmd.FromIncomingContext(ctx)
		if lo.Contains(inMD.Get(headers.IfNoneMatch), dbRel.GetEtag()) {
			_ = grpc.SetHeader(ctx, grpcmd.Pairs("x-http-code", "304"))

			return nil
		}

		if req.GetWithObjects() {
			relations := []*dsc3.Relation{resp.GetResult()}
			resp.Objects = s.getWithObjects(ctx, tx, relations)
		}

		return nil
	})

	return resp, err
}

// GetRelations, gets paginated set of relation instances based on subject, relation, object filter.
func (s *Reader) GetRelations(ctx context.Context, req *dsr3.GetRelationsRequest) (*dsr3.GetRelationsResponse, error) {
	resp := &dsr3.GetRelationsResponse{
		Results: []*dsc3.Relation{},
		Objects: map[string]*dsc3.Object{},
		Page:    &dsc3.PaginationResponse{},
	}

	if err := validator.GetRelationsRequest(req); err != nil {
		return resp, err
	}

	if req.GetPage() == nil {
		req.Page = &dsc3.PaginationRequest{Size: x.MaxPageSize}
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
		iter, err := bdb.NewScanIterator[dsc3.Relation](ctx, tx, path, opts...)
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
func (s *Reader) Check(ctx context.Context, req *dsr3.CheckRequest) (*dsr3.CheckResponse, error) {
	resp := &dsr3.CheckResponse{}

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
func (s *Reader) Checks(ctx context.Context, req *dsr3.ChecksRequest) (*dsr3.ChecksResponse, error) {
	resp := &dsr3.ChecksResponse{}

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

// CheckPermission, check if subject is permitted to access resource (object).
//
//nolint:dupl
func (s *Reader) CheckPermission(ctx context.Context, req *dsr3.CheckPermissionRequest) (*dsr3.CheckPermissionResponse, error) {
	resp := &dsr3.CheckPermissionResponse{}

	if err := validator.CheckPermissionRequest(req); err != nil {
		return resp, err
	}

	if err := ds.CheckPermission(req).Validate(s.store.MC()); err != nil {
		return resp, err
	}

	check := ds.Check(&dsr3.CheckRequest{
		ObjectType:  req.GetObjectType(),
		ObjectId:    req.GetObjectId(),
		Relation:    req.GetPermission(),
		SubjectType: req.GetSubjectType(),
		SubjectId:   req.GetSubjectId(),
		Trace:       req.GetTrace(),
	})

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		var err error

		r, err := check.Exec(ctx, tx, s.store.MC())
		if err == nil {
			resp.Check = r.GetCheck()
			resp.Trace = r.GetTrace()
		}

		return err
	})

	return resp, err
}

// CheckRelation, check if subject has the specified relation to a resource (object).
//
//nolint:dupl
func (s *Reader) CheckRelation(ctx context.Context, req *dsr3.CheckRelationRequest) (*dsr3.CheckRelationResponse, error) {
	resp := &dsr3.CheckRelationResponse{}

	if err := validator.CheckRelationRequest(req); err != nil {
		return resp, err
	}

	if err := ds.CheckRelation(req).Validate(s.store.MC()); err != nil {
		return resp, err
	}

	check := ds.Check(&dsr3.CheckRequest{
		ObjectType:  req.GetObjectType(),
		ObjectId:    req.GetObjectId(),
		Relation:    req.GetRelation(),
		SubjectType: req.GetSubjectType(),
		SubjectId:   req.GetSubjectId(),
		Trace:       req.GetTrace(),
	})

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		var err error

		r, err := check.Exec(ctx, tx, s.store.MC())
		if err == nil {
			resp.Check = r.GetCheck()
			resp.Trace = r.GetTrace()
		}

		return err
	})

	return resp, err
}

// GetGraph, return graph of connected objects and relations for requested anchor subject/object.
func (s *Reader) GetGraph(ctx context.Context, req *dsr3.GetGraphRequest) (*dsr3.GetGraphResponse, error) {
	resp := &dsr3.GetGraphResponse{}

	if err := validator.GetGraphRequest(req); err != nil {
		return &dsr3.GetGraphResponse{}, err
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

func (*Reader) getWithObjects(ctx context.Context, tx *bolt.Tx, relations []*dsc3.Relation) map[string]*dsc3.Object {
	objects := map[string]*dsc3.Object{}

	for _, r := range relations {
		rel := ds.Relation(r)

		sub, err := bdb.Get[dsc3.Object](ctx, tx, bdb.ObjectsPath, ds.ObjectIdentifier(rel.Subject()).Key())
		if err != nil {
			sub = &dsc3.Object{Type: rel.SubjectType, Id: rel.SubjectId}
		}

		objects[ds.Object(sub).StrKey()] = sub

		obj, err := bdb.Get[dsc3.Object](ctx, tx, bdb.ObjectsPath, ds.ObjectIdentifier(rel.Object()).Key())
		if err != nil {
			obj = &dsc3.Object{Type: rel.ObjectType, Id: rel.ObjectId}
		}

		objects[ds.Object(obj).StrKey()] = obj
	}

	return objects
}
