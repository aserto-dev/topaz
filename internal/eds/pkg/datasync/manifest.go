//nolint:gocritic,unused,dupword
package datasync

import (
	"context"

	"github.com/aserto-dev/azm/model"
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Sync) syncManifest(ctx context.Context, conn *grpc.ClientConn) error {
	return status.Error(codes.Unimplemented, "code not updated")

	// runStartTime := time.Now().UTC()

	// s.logger.Info().Str(syncStatus, syncStarted).Str("mode", Manifest.String()).Msg(syncManifest)

	// ctx, cancel := context.WithCancel(ctx)
	// defer cancel()

	// rc := dsr.NewReaderClient(conn)
	// resp, err := rc.GetManifest(ctx, &dsr.GetManifestRequest{Empty: &emptypb.Empty{}})
	// // remoteMD, remoteBuf, err := s.getManifest(ctx, dsr.NewReaderClient(conn))
	// if err != nil {
	// 	return err
	// }

	// remoteManifest := resp.GetManifest()

	// localManifest, err := func() (*dsc.Manifest, error) {
	// 	var (
	// 		localMD     *dsm.Metadata
	// 		localReader io.Reader
	// 	)

	// 	err := s.store.DB().View(func(tx *bolt.Tx) error {
	// 		md := &dsm.Metadata{UpdatedAt: timestamppb.Now(), Etag: ""}
	// 		manifest, err := ds.Manifest(md).Get(ctx, tx)

	// 		switch {
	// 		case status.Code(err) == codes.NotFound:
	// 			if manifest == nil {
	// 				manifest = ds.Manifest(&dsm.Metadata{})
	// 			}
	// 		case err != nil:
	// 			return errors.Errorf("failed to get manifest")
	// 		}

	// 		localMD = manifest.Metadata
	// 		localReader = bytes.NewReader(manifest.Body.GetData())

	// 		return nil
	// 	})

	// 	return localMD, localReader, err
	// }()
	// if err != nil {
	// 	return err
	// }

	// s.logger.Debug().
	// 	Str("local.etag", localMD.GetEtag()).Str("remote.etag", remoteMD.GetEtag()).
	// 	Bool("identical", localMD.GetEtag() == remoteMD.GetEtag()).Msg(syncManifest)

	// if localMD.GetEtag() == remoteMD.GetEtag() {
	// 	return nil
	// }

	// m, err := s.setManifest(ctx, remoteBuf)
	// if err != nil {
	// 	return err
	// }

	// runEndTime := time.Now().UTC()

	// s.logger.Info().
	// 	Str(syncStatus, syncFinished).Str("mode", Manifest.String()).
	// 	Str("duration", runEndTime.Sub(runStartTime).String()).Msg(syncManifest)

	// return s.store.MC().UpdateModel(m)
}

func (s *Sync) setManifest(ctx context.Context, remoteBuf []byte) (*model.Model, error) {
	return nil, status.Error(codes.Unimplemented, "code not updated")

	// // calc new ETag from remote manifest.
	// h := fnv.New64a()
	// h.Reset()
	// _, _ = h.Write(remoteBuf)

	// md := &dsm.Metadata{
	// 	UpdatedAt: timestamppb.Now(),
	// 	Etag:      strconv.FormatUint(h.Sum64(), 10),
	// }

	// if err := validator.Metadata(md); err != nil {
	// 	return nil, err
	// }

	// m, err := manifest.Load(bytes.NewReader(remoteBuf))
	// if err != nil {
	// 	return nil, derr.ErrInvalidArgument.Msg(err.Error())
	// }

	// if err := s.store.DB().Update(func(tx *bolt.Tx) error {
	// 	stats, err := ds.CalculateStats(ctx, tx)
	// 	if err != nil {
	// 		return derr.ErrUnknown.Msgf("failed to calculate stats: %s", err.Error())
	// 	}

	// 	if err := s.store.MC().CanUpdate(m, stats); err != nil {
	// 		return err
	// 	}

	// 	if err := ds.Manifest(md).Set(ctx, tx, bytes.NewBuffer(remoteBuf)); err != nil {
	// 		return derr.ErrUnknown.Msgf("failed to set manifest: %s", err.Error())
	// 	}

	// 	if err := ds.Manifest(md).SetModel(ctx, tx, m); err != nil {
	// 		return derr.ErrUnknown.Msgf("failed to set model: %s", err.Error())
	// 	}

	// 	return nil
	// }); err != nil {
	// 	return nil, err
	// }

	// return m, nil
}

func (s *Sync) getManifest(ctx context.Context, rc dsr.ReaderClient) (*dsc.Manifest, error) {
	return nil, status.Error(codes.Unimplemented, "code not updated")

	// manifest, err := rc.GetManifest(ctx, &dsr.GetManifestRequest{Empty: &emptypb.Empty{}})
	// if err != nil {
	// 	return nil, nil, err
	// }

	// data := bytes.Buffer{}
	// metadata := &dsm.Metadata{}
	// bytesRecv := 0

	// for {
	// 	resp, err := stream.Recv()
	// 	if errors.Is(err, io.EOF) {
	// 		break
	// 	}

	// 	if err != nil {
	// 		return nil, nil, err
	// 	}

	// 	if md, ok := resp.GetMsg().(*dsm.GetManifestResponse_Metadata); ok {
	// 		metadata = md.Metadata
	// 	}

	// 	if body, ok := resp.GetMsg().(*dsm.GetManifestResponse_Body); ok {
	// 		data.Write(body.Body.GetData())
	// 		bytesRecv += len(body.Body.GetData())
	// 	}
	// }

	// return metadata, data.Bytes(), nil
}
