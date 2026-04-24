package v3

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	"github.com/aserto-dev/topaz/api/directory/pkg/validator"
	dsc3 "github.com/aserto-dev/topaz/api/directory/v4"
	dsw3 "github.com/aserto-dev/topaz/api/directory/v4/writer"
	aerr "github.com/aserto-dev/topaz/errors"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/eds/pkg/ds"

	"github.com/go-http-utils/headers"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ dsw3.WriterServer = &Writer{}

type Writer struct {
	logger *zerolog.Logger
	store  *bdb.BoltDB
}

func NewWriter(logger *zerolog.Logger, store *bdb.BoltDB) *Writer {
	return &Writer{
		logger: logger,
		store:  store,
	}
}

// SetManifest
func (s *Writer) SetManifest(ctx context.Context, req *dsw3.SetManifestRequest) (*dsw3.SetManifestResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method SetManifest not implemented")
}

// DeleteManifest
func (s *Writer) DeleteManifest(ctx context.Context, req *dsw3.DeleteManifestRequest) (*dsw3.DeleteManifestResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteManifest not implemented")
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
			return derr.ErrHashMismatch.Msgf("for object with type [%s] and id [%s]", updObj.GetObjectType(), updObj.GetObjectId())
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
			obj := &dsc3.Object{ObjectType: req.GetObjectType(), ObjectId: req.GetObjectId()}

			updObj, err := ds.UpdateMetadataObject(ctx, tx, bdb.ObjectsPath, ds.Object(obj).Key(), obj)
			if err != nil {
				return err
			}

			if ifMatchHeader != updObj.GetEtag() {
				return derr.ErrHashMismatch.Msgf("for object with type [%s] and id [%s]", updObj.GetObjectType(), updObj.GetObjectId())
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

const (
	object   string = "object"
	relation string = "relation"
)

type counters map[string]*dsw3.ImportCounter

func (s *Writer) Import(stream dsw3.Writer_ImportServer) error {
	ctx := stream.Context()

	ctr := counters{
		object:   {Type: object},
		relation: {Type: relation},
	}

	importErr := s.store.DB().Batch(func(tx *bolt.Tx) error {
		for {
			select {
			case <-ctx.Done(): // exit if context is done
				return nil
			default:
			}

			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				s.logger.Trace().Msg("import stream EOF")

				for _, c := range ctr {
					_ = stream.Send(&dsw3.ImportResponse{Msg: &dsw3.ImportResponse_Counter{Counter: c}})
				}

				// backwards compatible response.
				return stream.Send(&dsw3.ImportResponse{
					Object:   ctr[object],
					Relation: ctr[relation],
				})
			}

			if err != nil {
				s.logger.Trace().Str("err", err.Error()).Msg("cannot receive req")
				continue
			}

			if err := s.handleImportRequest(ctx, tx, req, ctr); err != nil {
				if stat, ok := status.FromError(err); ok {
					status := &dsw3.ImportStatus{
						Code: uint32(stat.Code()),
						Msg:  stat.Message(),
						Req:  req,
					}

					if err := stream.Send(&dsw3.ImportResponse{Msg: &dsw3.ImportResponse_Status{Status: status}}); err != nil {
						s.logger.Err(err).Msg("failed to send import status")
					}
				}
			}
		}
	})

	return importErr
}

func (s *Writer) handleImportRequest(ctx context.Context, tx *bolt.Tx, req *dsw3.ImportRequest, ctr counters) error {
	switch m := req.GetMsg().(type) {
	case *dsw3.ImportRequest_Object:
		if req.GetOpCode() == dsw3.Opcode_OPCODE_SET {
			err := s.objectSetHandler(ctx, tx, m.Object)
			ctr[object] = updateCounter(ctr[object], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsw3.Opcode_OPCODE_DELETE {
			err := s.objectDeleteHandler(ctx, tx, m.Object)
			ctr[object] = updateCounter(ctr[object], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsw3.Opcode_OPCODE_DELETE_WITH_RELATIONS {
			err := s.objectDeleteWithRelationsHandler(ctx, tx, m.Object)
			ctr[object] = updateCounter(ctr[object], req.GetOpCode(), err)

			return err
		}

		return derr.ErrUnknownOpCode.Msgf("%s - %d", req.GetOpCode().String(), int32(req.GetOpCode()))

	case *dsw3.ImportRequest_Relation:
		if req.GetOpCode() == dsw3.Opcode_OPCODE_SET {
			err := s.relationSetHandler(ctx, tx, m.Relation)
			ctr[relation] = updateCounter(ctr[relation], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsw3.Opcode_OPCODE_DELETE {
			err := s.relationDeleteHandler(ctx, tx, m.Relation)
			ctr[relation] = updateCounter(ctr[relation], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsw3.Opcode_OPCODE_DELETE_WITH_RELATIONS {
			return derr.ErrInvalidOpCode.Msgf("%s for type relation", req.GetOpCode().String())
		}

		return derr.ErrUnknownOpCode.Msgf("%s - %d", req.GetOpCode().String(), int32(req.GetOpCode()))

	default:
		return derr.ErrUnknown.Msgf("import request")
	}
}

func (s *Writer) objectSetHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Object) error {
	s.logger.Debug().Interface("object", req).Msg("ImportObject")

	if req == nil {
		return derr.ErrInvalidObject.Msg("nil")
	}

	if err := validator.Object(req); err != nil {
		return err
	}

	obj := ds.Object(req)
	if err := obj.Validate(s.store.MC()); err != nil {
		return modelValidateError(err)
	}

	etag := obj.Hash()

	updReq, err := ds.UpdateMetadataObject(ctx, tx, bdb.ObjectsPath, obj.Key(), req)
	if err != nil {
		return err
	}

	if etag == updReq.GetEtag() {
		s.logger.Trace().Bytes("key", obj.Key()).Str("etag-equal", etag).Msg("ImportObject")
		return nil
	}

	updReq.Etag = etag

	if _, err := bdb.Set[dsc3.Object](ctx, tx, bdb.ObjectsPath, ds.Object(updReq).Key(), updReq); err != nil {
		return derr.ErrInvalidObject.Msg("set")
	}

	return nil
}

func (s *Writer) objectDeleteHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Object) error {
	s.logger.Debug().Interface("object", req).Msg("ImportObject")

	if req == nil {
		return derr.ErrInvalidObject.Msg("nil")
	}

	if err := validator.Object(req); err != nil {
		return err
	}

	obj := ds.Object(req)
	if err := obj.Validate(s.store.MC()); err != nil {
		return modelValidateError(err)
	}

	if err := bdb.Delete(ctx, tx, bdb.ObjectsPath, obj.Key()); err != nil {
		return derr.ErrInvalidObject.Msg("delete")
	}

	return nil
}

func (s *Writer) objectDeleteWithRelationsHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Object) error {
	s.logger.Debug().Interface("object", req).Msg("ImportObject")

	if req == nil {
		return derr.ErrInvalidObject.Msg("nil")
	}

	if err := validator.Object(req); err != nil {
		return err
	}

	obj := ds.Object(req)
	if err := obj.Validate(s.store.MC()); err != nil {
		return modelValidateError(err)
	}

	if err := bdb.Delete(ctx, tx, bdb.ObjectsPath, obj.Key()); err != nil {
		return derr.ErrInvalidObject.Msg("delete")
	}

	// incoming object relations of object instance (result.type == incoming.subject.type && result.key == incoming.subject.key)
	if err := s.deleteObjectRelations(ctx, tx, bdb.RelationsSubPath, req); err != nil {
		return err
	}

	// outgoing object relations of object instance (result.type == outgoing.object.type && result.key == outgoing.object.key)
	if err := s.deleteObjectRelations(ctx, tx, bdb.RelationsObjPath, req); err != nil {
		return err
	}

	return nil
}

func (*Writer) deleteObjectRelations(ctx context.Context, tx *bolt.Tx, path bdb.Path, obj *dsc3.Object) error {
	iter, err := bdb.NewScanIterator[dsc3.Relation](
		ctx, tx, path,
		bdb.WithKeyFilter(append(ds.Object(obj).Key(), ds.InstanceSeparator)),
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

func (s *Writer) relationSetHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Relation) error {
	s.logger.Debug().Interface("relation", req).Msg("ImportRelation")

	if req == nil {
		return derr.ErrInvalidRelation.Msg("nil")
	}

	if err := validator.Relation(req); err != nil {
		return err
	}

	rel := ds.Relation(req)
	if err := rel.Validate(s.store.MC()); err != nil {
		return modelValidateError(err)
	}

	etag := rel.Hash()

	updReq, err := ds.UpdateMetadataRelation(ctx, tx, bdb.RelationsObjPath, rel.ObjKey(), req)
	if err != nil {
		return err
	}

	if etag == updReq.GetEtag() {
		s.logger.Trace().Bytes("key", rel.ObjKey()).Str("etag-equal", etag).Msg("ImportRelation")
		return nil
	}

	updReq.Etag = etag

	if _, err := bdb.Set[dsc3.Relation](ctx, tx, bdb.RelationsObjPath, ds.Relation(updReq).ObjKey(), updReq); err != nil {
		return derr.ErrInvalidRelation.Msg("set")
	}

	if _, err := bdb.Set[dsc3.Relation](ctx, tx, bdb.RelationsSubPath, ds.Relation(updReq).SubKey(), updReq); err != nil {
		return derr.ErrInvalidRelation.Msg("set")
	}

	return nil
}

func (s *Writer) relationDeleteHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Relation) error {
	s.logger.Debug().Interface("relation", req).Msg("ImportRelation")

	if req == nil {
		return derr.ErrInvalidRelation.Msg("nil")
	}

	if err := validator.Relation(req); err != nil {
		return err
	}

	rel := ds.Relation(req)
	if err := rel.Validate(s.store.MC()); err != nil {
		return modelValidateError(err)
	}

	if err := bdb.Delete(ctx, tx, bdb.RelationsObjPath, rel.ObjKey()); err != nil {
		return derr.ErrInvalidRelation.Msg("delete")
	}

	if err := bdb.Delete(ctx, tx, bdb.RelationsSubPath, rel.SubKey()); err != nil {
		return derr.ErrInvalidRelation.Msg("delete")
	}

	return nil
}

func updateCounter(c *dsw3.ImportCounter, opCode dsw3.Opcode, err error) *dsw3.ImportCounter {
	c.Recv++

	switch {
	case err != nil:
		c.Error++
	case opCode == dsw3.Opcode_OPCODE_SET:
		c.Set++
	case opCode == dsw3.Opcode_OPCODE_DELETE:
		c.Delete++
	case opCode == dsw3.Opcode_OPCODE_DELETE_WITH_RELATIONS:
		c.Delete++
	}

	return c
}

func modelValidateError(e error) error {
	var x *aerr.AsertoError
	if ok := errors.As(e, &x); ok {
		dataMsg, ok := x.Fields()[aerr.MessageKey].(string)
		if ok {
			if x.Message != "" {
				x.Message = fmt.Sprintf("%q: %s", dataMsg, x.Message)
			} else {
				x.Message = dataMsg
			}
		}
	}

	return e
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
