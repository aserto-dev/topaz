package v3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	man "github.com/aserto-dev/azm/v3"
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsw "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/go-directory/pkg/validator"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/eds/pkg/ds"

	"github.com/go-http-utils/headers"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Writer struct {
	logger *zerolog.Logger
	store  *bdb.BoltDB
}

var _ dsw.WriterServer = (*Writer)(nil)

func NewWriter(logger *zerolog.Logger, store *bdb.BoltDB) *Writer {
	return &Writer{
		logger: logger,
		store:  store,
	}
}

// SetManifest.
func (s *Writer) SetManifest(ctx context.Context, req *dsw.SetManifestRequest) (*dsw.SetManifestResponse, error) {
	resp := &dsw.SetManifestResponse{}

	if err := validator.SetManifestRequest(req); err != nil {
		return resp, err
	}

	// clone manifest to update etag and updated_at.
	manifest, ok := proto.Clone(req.GetManifest()).(*dsc.Manifest)
	if !ok {
		return resp, status.Error(codes.Internal, "clone manifest failed")
	}

	etag := ds.Manifest(req.GetManifest()).Hash()
	updatedAt := timestamppb.New(time.Now().UTC())

	// update cloned manifest.
	manifest.UpdatedAt = updatedAt
	manifest.Etag = etag

	// models inherit the updated_at and etag values from the originating manifest.
	modModel := &dsc.Model{Content: pb.NewStruct(), UpdatedAt: updatedAt, Etag: etag}
	invModel := &dsc.Model{Content: pb.NewStruct(), UpdatedAt: updatedAt, Etag: etag}

	// load manifest to validate syntax, returns model
	{
		mod, err := man.Load(bytes.NewReader(manifest.GetContent()))
		if err != nil {
			return resp, err
		}

		modBuf, err := json.Marshal(mod)
		if err != nil {
			return resp, err
		}

		if err := modModel.GetContent().UnmarshalJSON(modBuf); err != nil {
			return resp, err
		}

		// invert model
		inv := mod.Invert()

		invBuf, err := json.Marshal(inv)
		if err != nil {
			return resp, err
		}

		if err := invModel.GetContent().UnmarshalJSON(invBuf); err != nil {
			return resp, err
		}
	}

	// ifMatchHeader := metadata.ExtractIncoming(ctx).Get(headers.IfMatch)

	// retrieve current manifest etag
	// if err := s.store.DB().View(func(tx *bolt.Tx) error {
	// 	man, err := bdb.Get[dsc.Manifest](ctx, tx, bdb.ManifestPathV3, bdb.ManifestKeyV3)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	return nil

	// 	// if ifMatchHeader == man.GetEtag() {
	// 	// 	return nil
	// 	// }

	// 	// return status.Error(codes.Aborted, "If-Match header does not match etag")
	// }); err != nil {
	// 	return resp, err
	// }

	err := s.store.DB().Update(func(tx *bolt.Tx) error {
		man, err := bdb.Set(ctx, tx, bdb.ManifestPathV3, bdb.ManifestKeyV3, manifest)
		if err != nil {
			return err
		}

		if _, err := bdb.Set(ctx, tx, bdb.ModelPathV3, bdb.ModelKeyV3, modModel); err != nil {
			return err
		}

		if _, err := bdb.Set(ctx, tx, bdb.ModelPathV3, bdb.InvertedModelKeyV3, invModel); err != nil {
			return err
		}

		resp.Manifest = man

		return nil
	})

	return resp, err
}

// DeleteManifest.
func (s *Writer) DeleteManifest(ctx context.Context, req *dsw.DeleteManifestRequest) (*dsw.DeleteManifestResponse, error) {
	resp := &dsw.DeleteManifestResponse{}

	// ifMatchHeader := metadata.ExtractIncoming(ctx).Get(headers.IfMatch)
	// if ifMatchHeader == "" {
	// 	return resp, status.Error(codes.Aborted, "no If-Match header with matching etag provided")
	// }

	// retrieve current manifest etag
	// if err := s.store.DB().View(func(tx *bolt.Tx) error {
	// 	man, err := bdb.Get[dsc.Manifest](ctx, tx, bdb.ManifestPathV3, bdb.ManifestKeyV3)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if ifMatchHeader == man.GetEtag() {
	// 		return nil
	// 	}

	// 	return status.Error(codes.Aborted, "If-Match header does not match etag")
	// }); err != nil {
	// 	return resp, err
	// }

	// delete & rebuild store.
	if err := s.store.DB().Update(func(tx *bolt.Tx) error {
		if err := bdb.DeleteBucket(tx, bdb.ManifestPathV3); err != nil {
			return err
		}

		if _, err := bdb.CreateBucket(tx, bdb.ManifestPathV3); err != nil {
			return err
		}

		if err := bdb.SetKey(tx, bdb.ManifestPathV3, bdb.ManifestKeyV3, []byte{}); err != nil {
			return err
		}

		if err := bdb.DeleteBucket(tx, bdb.ModelPathV3); err != nil {
			return err
		}

		if _, err := bdb.CreateBucket(tx, bdb.ModelPathV3); err != nil {
			return err
		}

		if err := bdb.SetKey(tx, bdb.ModelPathV3, bdb.ModelKeyV3, []byte{}); err != nil {
			return err
		}

		if err := bdb.DeleteBucket(tx, bdb.ObjectsPath); err != nil {
			return err
		}

		if _, err := bdb.CreateBucket(tx, bdb.ObjectsPath); err != nil {
			return err
		}

		if err := bdb.DeleteBucket(tx, bdb.RelationsObjPath); err != nil {
			return err
		}

		if _, err := bdb.CreateBucket(tx, bdb.RelationsObjPath); err != nil {
			return err
		}

		if err := bdb.DeleteBucket(tx, bdb.RelationsSubPath); err != nil {
			return err
		}

		if _, err := bdb.CreateBucket(tx, bdb.RelationsSubPath); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return resp, err
	}

	return resp, nil
}

// SetObject.
func (s *Writer) SetObject(ctx context.Context, req *dsw.SetObjectRequest) (*dsw.SetObjectResponse, error) {
	resp := &dsw.SetObjectResponse{}

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
		ifMatchHeader := metadata.ExtractIncoming(ctx).Get(headers.IfMatch)
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

func (s *Writer) DeleteObject(ctx context.Context, req *dsw.DeleteObjectRequest) (*dsw.DeleteObjectResponse, error) {
	resp := &dsw.DeleteObjectResponse{}

	if err := validator.DeleteObjectRequest(req); err != nil {
		return resp, err
	}

	objIdent := ds.ObjectIdentifier(&dsc.ObjectIdentifier{ObjectType: req.GetObjectType(), ObjectId: req.GetObjectId()})

	if err := objIdent.Validate(s.store.MC()); err != nil {
		return resp, err
	}

	err := s.store.DB().Update(func(tx *bolt.Tx) error {
		objIdent := ds.ObjectIdentifier(&dsc.ObjectIdentifier{ObjectType: req.GetObjectType(), ObjectId: req.GetObjectId()})

		// optimistic concurrency check
		ifMatchHeader := metadata.ExtractIncoming(ctx).Get(headers.IfMatch)
		if ifMatchHeader != "" {
			obj := &dsc.Object{ObjectType: req.GetObjectType(), ObjectId: req.GetObjectId()}

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

		return nil
	})

	return resp, err
}

// SetRelation.
func (s *Writer) SetRelation(ctx context.Context, req *dsw.SetRelationRequest) (*dsw.SetRelationResponse, error) {
	resp := &dsw.SetRelationResponse{}

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
		ifMatchHeader := metadata.ExtractIncoming(ctx).Get(headers.IfMatch)
		// if the updReq.Etag == "" this means the this is an insert
		if ifMatchHeader != "" && updRel.GetEtag() != "" && ifMatchHeader != updRel.GetEtag() {
			return derr.ErrHashMismatch.Msgf(
				"for relation with objectType [%s], objectId [%s], relation [%s], subjectType [%s], SubjectId [%s]",
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

func (s *Writer) DeleteRelation(ctx context.Context, req *dsw.DeleteRelationRequest) (*dsw.DeleteRelationResponse, error) {
	resp := &dsw.DeleteRelationResponse{}

	if err := validator.DeleteRelationRequest(req); err != nil {
		return resp, err
	}

	rel := &dsc.Relation{
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
		ifMatchHeader := metadata.ExtractIncoming(ctx).Get(headers.IfMatch)
		if ifMatchHeader != "" {
			updRel, err := ds.UpdateMetadataRelation(ctx, tx, bdb.RelationsObjPath, rid.ObjKey(), rel)
			if err != nil {
				return err
			}

			if ifMatchHeader != updRel.GetEtag() {
				return derr.ErrHashMismatch.Msgf(
					"for relation with objectType [%s], objectId [%s], relation [%s], subjectType [%s], SubjectId [%s]",
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

		return nil
	})

	return resp, err
}

func (s *Writer) Import(stream grpc.BidiStreamingServer[dsw.ImportRequest, dsw.ImportResponse]) error {
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
					_ = stream.Send(&dsw.ImportResponse{Msg: &dsw.ImportResponse_Counter{Counter: c}})
				}

				return stream.Send(&dsw.ImportResponse{})
			}

			if err != nil {
				s.logger.Trace().Str("err", err.Error()).Msg("cannot receive req")
				continue
			}

			if err := s.handleImportRequest(ctx, tx, req, ctr); err != nil {
				if stat, ok := status.FromError(err); ok {
					status := &dsw.ImportStatus{
						Code: uint32(stat.Code()),
						Msg:  stat.Message(),
						Req:  req,
					}

					if err := stream.Send(&dsw.ImportResponse{Msg: &dsw.ImportResponse_Status{Status: status}}); err != nil {
						s.logger.Err(err).Msg("failed to send import status")
					}
				}
			}
		}
	})

	return importErr
}

// Batch, allows sending multiple writer requests as a single request, executed within a single transaction scope.
func (*Writer) Batch(ctx context.Context, req *dsw.BatchRequest) (*dsw.BatchResponse, error) {
	return nil, status.Error(codes.Unimplemented, "writer.Batch not implemented")
}

func (*Writer) deleteRelations(ctx context.Context, path bdb.Path, tx *bolt.Tx, oid *dsc.ObjectIdentifier) error {
	objIdent := ds.ObjectIdentifier(oid)

	iter, err := bdb.NewScanIterator[dsc.Relation](
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
