package v3

import (
	"context"
	"errors"
	"fmt"

	aerr "github.com/aserto-dev/errors"
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsw "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/go-directory/pkg/validator"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/eds/pkg/ds"

	bolt "go.etcd.io/bbolt"
)

const (
	object   string = "object"
	relation string = "relation"
)

type counters map[string]*dsw.ImportCounter

func (s *Writer) handleImportRequest(ctx context.Context, tx *bolt.Tx, req *dsw.ImportRequest, ctr counters) error {
	switch m := req.GetMsg().(type) {

	case *dsw.ImportRequest_Relation:
		if req.GetOpCode() == dsw.Opcode_OPCODE_SET {
			err := s.relationSetHandler(ctx, tx, m.Relation)
			ctr[relation] = updateCounter(ctr[relation], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsw.Opcode_OPCODE_DELETE {
			err := s.relationDeleteHandler(ctx, tx, m.Relation)
			ctr[relation] = updateCounter(ctr[relation], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsw.Opcode_OPCODE_DELETE_WITH_RELATIONS {
			return derr.ErrInvalidOpCode.Msgf("%s for type relation", req.GetOpCode().String())
		}

		return derr.ErrUnknownOpCode.Msgf("%s - %d", req.GetOpCode().String(), int32(req.GetOpCode()))

	case *dsw.ImportRequest_Object:
		if req.GetOpCode() == dsw.Opcode_OPCODE_SET {
			err := s.objectSetHandler(ctx, tx, m.Object)
			ctr[object] = updateCounter(ctr[object], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsw.Opcode_OPCODE_DELETE {
			err := s.objectDeleteHandler(ctx, tx, m.Object)
			ctr[object] = updateCounter(ctr[object], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsw.Opcode_OPCODE_DELETE_WITH_RELATIONS {
			err := s.objectDeleteWithRelationsHandler(ctx, tx, m.Object)
			ctr[object] = updateCounter(ctr[object], req.GetOpCode(), err)

			return err
		}

		return derr.ErrUnknownOpCode.Msgf("%s - %d", req.GetOpCode().String(), int32(req.GetOpCode()))

	default:
		return derr.ErrUnknown.Msgf("import request")
	}
}

func (s *Writer) objectSetHandler(ctx context.Context, tx *bolt.Tx, req *dsc.Object) error {
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

	if _, err := bdb.Set[dsc.Object](ctx, tx, bdb.ObjectsPath, ds.Object(updReq).Key(), updReq); err != nil {
		return derr.ErrInvalidObject.Msg("set")
	}

	return nil
}

func (s *Writer) objectDeleteHandler(ctx context.Context, tx *bolt.Tx, req *dsc.Object) error {
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

func (s *Writer) objectDeleteWithRelationsHandler(ctx context.Context, tx *bolt.Tx, req *dsc.Object) error {
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

func (*Writer) deleteObjectRelations(ctx context.Context, tx *bolt.Tx, path bdb.Path, obj *dsc.Object) error {
	iter, err := bdb.NewScanIterator[dsc.Relation](
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

func (s *Writer) relationSetHandler(ctx context.Context, tx *bolt.Tx, req *dsc.Relation) error {
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

	if _, err := bdb.Set[dsc.Relation](ctx, tx, bdb.RelationsObjPath, ds.Relation(updReq).ObjKey(), updReq); err != nil {
		return derr.ErrInvalidRelation.Msg("set")
	}

	if _, err := bdb.Set[dsc.Relation](ctx, tx, bdb.RelationsSubPath, ds.Relation(updReq).SubKey(), updReq); err != nil {
		return derr.ErrInvalidRelation.Msg("set")
	}

	return nil
}

func (s *Writer) relationDeleteHandler(ctx context.Context, tx *bolt.Tx, req *dsc.Relation) error {
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

func updateCounter(c *dsw.ImportCounter, opCode dsw.Opcode, err error) *dsw.ImportCounter {
	c.Recv++

	switch {
	case err != nil:
		c.Error++
	case opCode == dsw.Opcode_OPCODE_SET:
		c.Set++
	case opCode == dsw.Opcode_OPCODE_DELETE:
		c.Delete++
	case opCode == dsw.Opcode_OPCODE_DELETE_WITH_RELATIONS:
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
