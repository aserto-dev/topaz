package v3

import (
	"context"
	"errors"
	"fmt"
	"io"

	aerr "github.com/aserto-dev/errors"
	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/go-directory/pkg/validator"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/ds"

	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc/status"
)

type Importer struct {
	dsi3.UnimplementedImporterServer

	logger *zerolog.Logger
	store  *bdb.BoltDB
}

const (
	object   string = "object"
	relation string = "relation"
)

type counters map[string]*dsi3.ImportCounter

func NewImporter(logger *zerolog.Logger, store *bdb.BoltDB) *Importer {
	return &Importer{
		logger: logger,
		store:  store,
	}
}

func (s *Importer) Import(stream dsi3.Importer_ImportServer) error {
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
					_ = stream.Send(&dsi3.ImportResponse{Msg: &dsi3.ImportResponse_Counter{Counter: c}})
				}

				// backwards compatible response.
				return stream.Send(&dsi3.ImportResponse{
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
					status := &dsi3.ImportStatus{
						Code: uint32(stat.Code()),
						Msg:  stat.Message(),
						Req:  req,
					}

					if err := stream.Send(&dsi3.ImportResponse{Msg: &dsi3.ImportResponse_Status{Status: status}}); err != nil {
						s.logger.Err(err).Msg("failed to send import status")
					}
				}
			}
		}
	})

	return importErr
}

func (s *Importer) handleImportRequest(ctx context.Context, tx *bolt.Tx, req *dsi3.ImportRequest, ctr counters) error {
	switch m := req.GetMsg().(type) {
	case *dsi3.ImportRequest_Object:
		if req.GetOpCode() == dsi3.Opcode_OPCODE_SET {
			err := s.objectSetHandler(ctx, tx, m.Object)
			ctr[object] = updateCounter(ctr[object], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsi3.Opcode_OPCODE_DELETE {
			err := s.objectDeleteHandler(ctx, tx, m.Object)
			ctr[object] = updateCounter(ctr[object], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsi3.Opcode_OPCODE_DELETE_WITH_RELATIONS {
			err := s.objectDeleteWithRelationsHandler(ctx, tx, m.Object)
			ctr[object] = updateCounter(ctr[object], req.GetOpCode(), err)

			return err
		}

		return derr.ErrUnknownOpCode.Msgf("%s - %d", req.GetOpCode().String(), int32(req.GetOpCode()))

	case *dsi3.ImportRequest_Relation:
		if req.GetOpCode() == dsi3.Opcode_OPCODE_SET {
			err := s.relationSetHandler(ctx, tx, m.Relation)
			ctr[relation] = updateCounter(ctr[relation], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsi3.Opcode_OPCODE_DELETE {
			err := s.relationDeleteHandler(ctx, tx, m.Relation)
			ctr[relation] = updateCounter(ctr[relation], req.GetOpCode(), err)

			return err
		}

		if req.GetOpCode() == dsi3.Opcode_OPCODE_DELETE_WITH_RELATIONS {
			return derr.ErrInvalidOpCode.Msgf("%s for type relation", req.GetOpCode().String())
		}

		return derr.ErrUnknownOpCode.Msgf("%s - %d", req.GetOpCode().String(), int32(req.GetOpCode()))

	default:
		return derr.ErrUnknown.Msgf("import request")
	}
}

func (s *Importer) objectSetHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Object) error {
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

func (s *Importer) objectDeleteHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Object) error {
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

func (s *Importer) objectDeleteWithRelationsHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Object) error {
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

func (*Importer) deleteObjectRelations(ctx context.Context, tx *bolt.Tx, path bdb.Path, obj *dsc3.Object) error {
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

func (s *Importer) relationSetHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Relation) error {
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

func (s *Importer) relationDeleteHandler(ctx context.Context, tx *bolt.Tx, req *dsc3.Relation) error {
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

func updateCounter(c *dsi3.ImportCounter, opCode dsi3.Opcode, err error) *dsi3.ImportCounter {
	c.Recv++

	switch {
	case err != nil:
		c.Error++
	case opCode == dsi3.Opcode_OPCODE_SET:
		c.Set++
	case opCode == dsi3.Opcode_OPCODE_DELETE:
		c.Delete++
	case opCode == dsi3.Opcode_OPCODE_DELETE_WITH_RELATIONS:
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
