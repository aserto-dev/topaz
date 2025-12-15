package datasync

import (
	"context"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/go-directory/pkg/validator"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/ds"

	bolt "go.etcd.io/bbolt"
)

func (s *Sync) objectSetHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Object) error {
	s.logger.Debug().Interface("object", req).Msg("ImportObject")

	if req == nil {
		return derr.ErrInvalidObject.Msg("nil")
	}

	if err := validator.Object(req); err != nil {
		return err
	}

	obj := ds.Object(req)
	if err := obj.Validate(s.store.MC()); err != nil {
		// The object violates the model.
		return err
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

func (s *Sync) objectDeleteHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Object) error {
	s.logger.Debug().Interface("object", req).Msg("ImportObject")

	if req == nil {
		return derr.ErrInvalidObject.Msg("nil")
	}

	if err := validator.Object(req); err != nil {
		return err
	}

	obj := ds.Object(req)
	if err := obj.Validate(s.store.MC()); err != nil {
		return err
	}

	if err := bdb.Delete(ctx, tx, bdb.ObjectsPath, obj.Key()); err != nil {
		return derr.ErrInvalidObject.Msg("delete")
	}

	return nil
}

func (s *Sync) relationSetHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Relation) error {
	s.logger.Debug().Interface("relation", req).Msg("ImportRelation")

	if req == nil {
		return derr.ErrInvalidRelation.Msg("nil")
	}

	if err := validator.Relation(req); err != nil {
		return err
	}

	rel := ds.Relation(req)
	if err := rel.Validate(s.store.MC()); err != nil {
		return err
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

	if _, err := bdb.Set[dsc3.Relation](ctx, tx, bdb.RelationsObjPath, rel.ObjKey(), updReq); err != nil {
		return derr.ErrInvalidRelation.Msg("set")
	}

	if _, err := bdb.Set[dsc3.Relation](ctx, tx, bdb.RelationsSubPath, rel.SubKey(), updReq); err != nil {
		return derr.ErrInvalidRelation.Msg("set")
	}

	return nil
}

func (s *Sync) relationDeleteHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Relation) error {
	s.logger.Debug().Interface("relation", req).Msg("ImportRelation")

	if req == nil {
		return derr.ErrInvalidRelation.Msg("nil")
	}

	if err := validator.Relation(req); err != nil {
		return err
	}

	rel := ds.Relation(req)
	if err := rel.Validate(s.store.MC()); err != nil {
		return err
	}

	if err := bdb.Delete(ctx, tx, bdb.RelationsObjPath, rel.ObjKey()); err != nil {
		return derr.ErrInvalidRelation.Msg("delete")
	}

	if err := bdb.Delete(ctx, tx, bdb.RelationsSubPath, rel.SubKey()); err != nil {
		return derr.ErrInvalidRelation.Msg("delete")
	}

	return nil
}
