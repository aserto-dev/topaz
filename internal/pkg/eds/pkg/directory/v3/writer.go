package v3

import (
	"context"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/go-directory/pkg/validator"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/ds"

	"github.com/go-http-utils/headers"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Writer struct {
	dsw3.UnimplementedWriterServer

	logger *zerolog.Logger
	store  *bdb.BoltDB
}

func NewWriter(logger *zerolog.Logger, store *bdb.BoltDB) *Writer {
	return &Writer{
		logger: logger,
		store:  store,
	}
}

// SetObject.
func (s *Writer) SetObject(ctx context.Context, req *dsw3.SetObjectRequest) (*dsw3.SetObjectResponse, error) {
	resp := &dsw3.SetObjectResponse{}

	if err := validator.SetObjectRequest(req); err != nil {
		return resp, err
	}

	obj := ds.Object(req.GetObject())
	if err := obj.Validate(s.store.MC()); err != nil {
		// The object violates the model.
		return resp, err
	}

	etag := obj.Hash()

	err := s.store.DB().Update(func(tx *bolt.Tx) error {
		updObj, err := ds.UpdateMetadataObject(ctx, tx, bdb.ObjectsPath, obj.Key(), req.GetObject())
		if err != nil {
			return err
		}

		// optimistic concurrency check
		ifMatchHeader := metautils.ExtractIncoming(ctx).Get(headers.IfMatch)
		// if the updReq.Etag == "" this means the this is an insert
		if ifMatchHeader != "" && updObj.GetEtag() != "" && ifMatchHeader != updObj.GetEtag() {
			return derr.ErrHashMismatch.Msgf("for object with type [%s] and id [%s]", updObj.GetType(), updObj.GetId())
		}

		if etag == updObj.GetEtag() {
			s.logger.Trace().Bytes("key", ds.Object(req.GetObject()).Key()).Str("etag-equal", etag).Msg("set_object")

			resp.Result = updObj

			return nil
		}

		updObj.Etag = etag

		objType, err := bdb.Set(ctx, tx, bdb.ObjectsPath, obj.Key(), updObj)
		if err != nil {
			return err
		}

		resp.Result = objType

		return nil
	})

	return resp, err
}

func (s *Writer) DeleteObject(ctx context.Context, req *dsw3.DeleteObjectRequest) (*dsw3.DeleteObjectResponse, error) {
	resp := &dsw3.DeleteObjectResponse{}

	if err := validator.DeleteObjectRequest(req); err != nil {
		return resp, err
	}

	objIdent := ds.ObjectIdentifier(&dsc3.ObjectIdentifier{ObjectType: req.GetObjectType(), ObjectId: req.GetObjectId()})

	if err := objIdent.Validate(s.store.MC()); err != nil {
		return resp, err
	}

	err := s.store.DB().Update(func(tx *bolt.Tx) error {
		objIdent := ds.ObjectIdentifier(&dsc3.ObjectIdentifier{ObjectType: req.GetObjectType(), ObjectId: req.GetObjectId()})

		// optimistic concurrency check
		ifMatchHeader := metautils.ExtractIncoming(ctx).Get(headers.IfMatch)
		if ifMatchHeader != "" {
			obj := &dsc3.Object{Type: req.GetObjectType(), Id: req.GetObjectId()}

			updObj, err := ds.UpdateMetadataObject(ctx, tx, bdb.ObjectsPath, ds.Object(obj).Key(), obj)
			if err != nil {
				return err
			}

			if ifMatchHeader != updObj.GetEtag() {
				return derr.ErrHashMismatch.Msgf("for object with type [%s] and id [%s]", updObj.GetType(), updObj.GetId())
			}
		}

		if err := bdb.Delete(ctx, tx, bdb.ObjectsPath, objIdent.Key()); err != nil {
			return err
		}

		if req.GetWithRelations() {
			// incoming object relations of object instance (result.type == incoming.subject.type && result.key == incoming.subject.key)
			if err := s.deleteRelations(ctx, bdb.RelationsSubPath, tx, objIdent.ObjectIdentifier); err != nil {
				return err
			}
			// outgoing object relations of object instance (result.type == outgoing.object.type && result.key == outgoing.object.key)
			if err := s.deleteRelations(ctx, bdb.RelationsObjPath, tx, objIdent.ObjectIdentifier); err != nil {
				return err
			}
		}

		resp.Result = &emptypb.Empty{}

		return nil
	})

	return resp, err
}

// SetRelation.
func (s *Writer) SetRelation(ctx context.Context, req *dsw3.SetRelationRequest) (*dsw3.SetRelationResponse, error) {
	resp := &dsw3.SetRelationResponse{}

	if err := validator.SetRelationRequest(req); err != nil {
		return resp, err
	}

	relation := ds.Relation(req.GetRelation())
	if err := relation.Validate(s.store.MC()); err != nil {
		return resp, err
	}

	etag := relation.Hash()

	err := s.store.DB().Update(func(tx *bolt.Tx) error {
		updRel, err := ds.UpdateMetadataRelation(ctx, tx, bdb.RelationsObjPath, relation.ObjKey(), req.GetRelation())
		if err != nil {
			return err
		}

		// optimistic concurrency check
		ifMatchHeader := metautils.ExtractIncoming(ctx).Get(headers.IfMatch)
		// if the updReq.Etag == "" this means the this is an insert
		if ifMatchHeader != "" && updRel.GetEtag() != "" && ifMatchHeader != updRel.GetEtag() {
			return derr.ErrHashMismatch.Msgf("for relation with objectType [%s], objectId [%s], relation [%s], subjectType [%s], SubjectId [%s]",
				updRel.GetObjectType(), updRel.GetObjectId(), updRel.GetRelation(), updRel.GetSubjectType(), updRel.GetSubjectId(),
			)
		}

		if etag == updRel.GetEtag() {
			s.logger.Trace().Bytes("key", ds.Relation(req.GetRelation()).ObjKey()).Str("etag-equal", etag).Msg("set_relation")

			resp.Result = updRel

			return nil
		}

		updRel.Etag = etag

		objRel, err := bdb.Set(ctx, tx, bdb.RelationsObjPath, relation.ObjKey(), updRel)
		if err != nil {
			return err
		}

		if _, err := bdb.Set(ctx, tx, bdb.RelationsSubPath, relation.SubKey(), updRel); err != nil {
			return err
		}

		resp.Result = objRel

		return nil
	})

	return resp, err
}

func (s *Writer) DeleteRelation(ctx context.Context, req *dsw3.DeleteRelationRequest) (*dsw3.DeleteRelationResponse, error) {
	resp := &dsw3.DeleteRelationResponse{}

	if err := validator.DeleteRelationRequest(req); err != nil {
		return resp, err
	}

	rel := &dsc3.Relation{
		ObjectType:      req.GetObjectType(),
		ObjectId:        req.GetObjectId(),
		Relation:        req.GetRelation(),
		SubjectType:     req.GetSubjectType(),
		SubjectId:       req.GetSubjectId(),
		SubjectRelation: req.GetSubjectRelation(),
	}

	rid := ds.Relation(rel)
	if err := rid.Validate(s.store.MC()); err != nil {
		return resp, err
	}

	err := s.store.DB().Update(func(tx *bolt.Tx) error {
		// optimistic concurrency check
		ifMatchHeader := metautils.ExtractIncoming(ctx).Get(headers.IfMatch)
		if ifMatchHeader != "" {
			updRel, err := ds.UpdateMetadataRelation(ctx, tx, bdb.RelationsObjPath, rid.ObjKey(), rel)
			if err != nil {
				return err
			}

			if ifMatchHeader != updRel.GetEtag() {
				return derr.ErrHashMismatch.Msgf("for relation with objectType [%s], objectId [%s], relation [%s], subjectType [%s], SubjectId [%s]",
					rel.GetObjectType(), rel.GetObjectId(), rel.GetRelation(), rel.GetSubjectType(), rel.GetSubjectId(),
				)
			}
		}

		if err := bdb.Delete(ctx, tx, bdb.RelationsObjPath, rid.ObjKey()); err != nil {
			return err
		}

		if err := bdb.Delete(ctx, tx, bdb.RelationsSubPath, rid.SubKey()); err != nil {
			return err
		}

		resp.Result = &emptypb.Empty{}

		return nil
	})

	return resp, err
}

func (*Writer) deleteRelations(ctx context.Context, path bdb.Path, tx *bolt.Tx, oid *dsc3.ObjectIdentifier) error {
	objIdent := ds.ObjectIdentifier(oid)

	iter, err := bdb.NewScanIterator[dsc3.Relation](
		ctx, tx, path,
		bdb.WithKeyFilter(append(objIdent.Key(), ds.InstanceSeparator)),
	)
	if err != nil {
		return err
	}

	for iter.Next() {
		rel := ds.Relation(iter.Value())
		if err := bdb.Delete(ctx, tx, bdb.RelationsObjPath, rel.ObjKey()); err != nil {
			return err
		}

		if err := bdb.Delete(ctx, tx, bdb.RelationsSubPath, rel.SubKey()); err != nil {
			return err
		}
	}

	return nil
}
