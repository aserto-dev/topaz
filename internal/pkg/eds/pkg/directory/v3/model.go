package v3

import (
	"bytes"
	"context"
	"hash/fnv"
	"io"
	"strconv"

	azmModel "github.com/aserto-dev/azm/model"
	manifest "github.com/aserto-dev/azm/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/go-directory/pkg/gateway/model/v3"
	mnfst "github.com/aserto-dev/go-directory/pkg/manifest"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/go-directory/pkg/validator"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/ds"

	"github.com/go-http-utils/headers"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Model struct {
	dsm3.UnimplementedModelServer

	logger *zerolog.Logger
	store  *bdb.BoltDB
}

// NOTES:
//
// store layout: _manifest/{name}/{version}/[metadata|manifest|model]
// The current single manifest implementation uses a constant name and version, defined in pkg/bdb/path.go.
//
// examples:
// _manifest/default/0.0.1/metadata		-- contains the model.Metadata message
// _manifest/default/0.0.1/manifest		-- contains the manifest raw byte stream
// _manifest/default/0.0.1/model		-- contains the serialized model representation of the manifest byte stream

func NewModel(logger *zerolog.Logger, store *bdb.BoltDB) *Model {
	return &Model{
		logger: logger,
		store:  store,
	}
}

var _ = dsm3.ModelServer(&Model{})

func (s *Model) GetManifest(req *dsm3.GetManifestRequest, stream dsm3.Model_GetManifestServer) error {
	if err := validator.GetManifestRequest(req); err != nil {
		return err
	}

	md := &dsm3.Metadata{UpdatedAt: timestamppb.Now(), Etag: ""}

	modelErr := s.store.DB().View(func(tx *bolt.Tx) error {
		manifest, err := ds.Manifest(md).Get(stream.Context(), tx)

		switch {
		case status.Code(err) == codes.NotFound:
			if manifest == nil {
				manifest = ds.Manifest(&dsm3.Metadata{})
			}
		case err != nil:
			return errors.Errorf("failed to get manifest")
		}

		if err := stream.Send(&dsm3.GetManifestResponse{
			Msg: &dsm3.GetManifestResponse_Metadata{
				Metadata: manifest.Metadata,
			},
		}); err != nil {
			return err
		}

		// optimistic concurrency check
		inMD, _ := metadata.FromIncomingContext(stream.Context())
		if lo.Contains(inMD.Get(headers.IfNoneMatch), manifest.Metadata.GetEtag()) {
			return nil
		}

		amr := mnfst.IncomingManifestRequest(stream.Context())
		if amr.WithBody() {
			body := &dsm3.Body{}

			for curByte := 0; curByte < len(manifest.Body.GetData()); curByte += model.MaxChunkSizeBytes {
				if curByte+model.MaxChunkSizeBytes > len(manifest.Body.GetData()) {
					body.Data = manifest.Body.GetData()[curByte:len(manifest.Body.GetData())]
				} else {
					body.Data = manifest.Body.GetData()[curByte : curByte+model.MaxChunkSizeBytes]
				}

				if err := stream.Send(&dsm3.GetManifestResponse{
					Msg: &dsm3.GetManifestResponse_Body{
						Body: body,
					},
				}); err != nil {
					return err
				}
			}
		}

		if amr.WithModel() {
			return s.getModel(stream, tx, md)
		}

		return nil
	})

	return modelErr
}

func (s *Model) SetManifest(stream dsm3.Model_SetManifestServer) error {
	logger := s.logger.With().Str("method", "SetManifest").Logger()
	logger.Trace().Send()

	// optimistic concurrency check
	etag := metautils.ExtractIncoming(stream.Context()).Get(headers.IfMatch)
	if etag != "" && etag != s.store.MC().Metadata().ETag {
		return derr.ErrHashMismatch
	}

	h := fnv.New64a()
	h.Reset()

	data := bytes.NewBuffer([]byte{})

	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return errors.Wrap(err, "failed to receive manifest")
		}

		if body, ok := msg.GetMsg().(*dsm3.SetManifestRequest_Body); ok {
			if err := validator.Body(body.Body); err != nil {
				return err
			}

			data.Write(body.Body.GetData())

			_, _ = h.Write(data.Bytes())
		}
	}

	if err := stream.SendAndClose(&dsm3.SetManifestResponse{
		Result: &emptypb.Empty{},
	}); err != nil {
		return err
	}

	md := &dsm3.Metadata{
		UpdatedAt: timestamppb.Now(),
		Etag:      strconv.FormatUint(h.Sum64(), 10),
	}

	if err := validator.Metadata(md); err != nil {
		return err
	}

	m, err := manifest.Load(bytes.NewReader(data.Bytes()))
	if err != nil {
		return derr.ErrInvalidArgument.Msg(err.Error())
	}

	if err := s.store.DB().Update(func(tx *bolt.Tx) error {
		return s.setManifest(stream, tx, m, md, data)
	}); err != nil {
		return err
	}

	logger.Info().Msg("manifest updated")

	return s.store.MC().UpdateModel(m)
}

func (s *Model) DeleteManifest(ctx context.Context, req *dsm3.DeleteManifestRequest) (*dsm3.DeleteManifestResponse, error) {
	resp := &dsm3.DeleteManifestResponse{}
	if err := validator.DeleteManifestRequest(req); err != nil {
		return resp, err
	}

	h := fnv.New64a()
	h.Reset()

	data := bytes.NewBuffer([]byte{})
	_, _ = h.Write(data.Bytes())

	md := &dsm3.Metadata{
		UpdatedAt: timestamppb.Now(),
		Etag:      strconv.FormatUint(h.Sum64(), 10),
	}

	m, err := manifest.Load(bytes.NewReader(data.Bytes()))
	if err != nil {
		return resp, derr.ErrInvalidArgument.Msg(err.Error())
	}

	if err := s.store.DB().Update(func(tx *bolt.Tx) error {
		// optimistic concurrency check
		ifMatchHeader := metautils.ExtractIncoming(ctx).Get(headers.IfMatch)
		if ifMatchHeader != "" {
			dbMd := &dsm3.Metadata{UpdatedAt: timestamppb.Now(), Etag: ""}
			manifest, err := ds.Manifest(dbMd).Get(ctx, tx)
			if err != nil {
				return nil //nolint:nilerr // early return when manifest does not exists, delete should not fail.
			}

			if ifMatchHeader != manifest.Metadata.GetEtag() {
				return derr.ErrHashMismatch
			}
		}

		if err := ds.Manifest(&dsm3.Metadata{}).Delete(ctx, tx); err != nil {
			return derr.ErrUnknown.Msgf("failed to delete manifest: %s", err.Error())
		}

		if err := ds.Manifest(md).Set(ctx, tx, data); err != nil {
			return derr.ErrUnknown.Msgf("failed to set manifest: %s", err.Error())
		}

		if err := ds.Manifest(md).SetModel(ctx, tx, m); err != nil {
			return derr.ErrUnknown.Msgf("failed to set model: %s", err.Error())
		}

		return nil
	}); err != nil {
		return resp, err
	}

	return &dsm3.DeleteManifestResponse{Result: &emptypb.Empty{}}, nil
}

func (*Model) getModel(stream dsm3.Model_GetManifestServer, tx *bolt.Tx, md *dsm3.Metadata) error {
	model, err := ds.Manifest(md).GetModel(stream.Context(), tx)

	switch {
	case status.Code(err) == codes.NotFound:
		return derr.ErrNotFound.Msg("model")
	case err != nil:
		return errors.Errorf("failed to get model")
	}

	m := pb.NewStruct()

	r, err := model.Reader()
	if err != nil {
		return err
	}

	if err := pb.BufToProto(r, m); err != nil {
		return err
	}

	if err := stream.Send(&dsm3.GetManifestResponse{
		Msg: &dsm3.GetManifestResponse_Model{
			Model: m,
		},
	}); err != nil {
		return err
	}

	return nil
}

func (s *Model) setManifest(stream dsm3.Model_SetManifestServer, tx *bolt.Tx, m *azmModel.Model, md *dsm3.Metadata, data *bytes.Buffer) error {
	stats, err := ds.CalculateStats(stream.Context(), tx)
	if err != nil {
		return derr.ErrUnknown.Msgf("failed to calculate stats: %s", err.Error())
	}

	if err := s.store.MC().CanUpdate(m, stats); err != nil {
		return err
	}

	if err := ds.Manifest(md).Set(stream.Context(), tx, data); err != nil {
		return derr.ErrUnknown.Msgf("failed to set manifest: %s", err.Error())
	}

	if err := ds.Manifest(md).SetModel(stream.Context(), tx, m); err != nil {
		return derr.ErrUnknown.Msgf("failed to set model: %s", err.Error())
	}

	return nil
}
