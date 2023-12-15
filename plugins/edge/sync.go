package edge

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/go-edge-ds/pkg/ds"
	"github.com/pkg/errors"

	"github.com/aserto-dev/go-aserto/client"
	topaz "github.com/aserto-dev/topaz/pkg/cc/config"

	cuckoo "github.com/panmari/cuckoofilter"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	syncScheduler  string = "scheduler"
	syncTask       string = "sync-task"
	syncRun        string = "sync-run"
	syncProducer   string = "producer"
	syncSubscriber string = "subscriber"
	status         string = "status"
	started        string = "started"
	finished       string = "finished"
	channelSize    int    = 10000
	localHost      string = "localhost:9292"
	initFilterSize uint   = 1000000
)

type directoryClient struct {
	conn     grpc.ClientConnInterface
	Model    dsm3.ModelClient
	Reader   dsr3.ReaderClient
	Writer   dsw3.WriterClient
	Importer dsi3.ImporterClient
	Exporter dse3.ExporterClient
}

type Sync struct {
	ctx         context.Context
	cfg         *Config
	topazConfig *topaz.Config
	log         *zerolog.Logger
	exportChan  chan *dse3.ExportResponse
	errChan     chan error
	filter      *cuckoo.Filter
	counts      *Counter
}

type Counter struct {
	Received  int32
	Upserts   int32
	Deletes   int32
	Errors    int32
	Manifests int32
	Objects   int32
	Relations int32
}

type Result struct {
	Counts *Counter
	Err    error
}

func NewSyncMgr(ctx context.Context, c *Config, topazConfig *topaz.Config, logger *zerolog.Logger) *Sync {
	return &Sync{
		ctx:         ctx,
		cfg:         c,
		topazConfig: topazConfig,
		log:         logger,
		exportChan:  make(chan *dse3.ExportResponse, channelSize),
		errChan:     make(chan error, 1),
	}
}

func (s *Sync) Run(fs bool) {
	runStartTime := time.Now().UTC()
	s.log.Info().Str(status, started).Bool("full-sync", fs).Msg(syncRun)

	defer func() {
		close(s.errChan)
	}()

	// error spew.
	go func() {
		for e := range s.errChan {
			s.log.Error().Err(e).Msg(syncRun)
		}
	}()

	s.counts = &Counter{}

	if err := s.syncManifest(); err != nil {
		s.log.Error().Str("sync-manifest", "").Err(err).Msg(syncRun)
	}

	g := new(errgroup.Group)

	watermark := s.getWatermark()

	if fs {
		watermark = &timestamppb.Timestamp{}
		s.filter = cuckoo.NewFilter(initFilterSize)
	}

	g.Go(func() error {
		return s.subscriber()
	})

	g.Go(func() error {
		return s.producer(watermark)
	})

	err := g.Wait()
	if err != nil {
		s.log.Error().Err(err).Msg("sync run failed")
	}

	if fs {
		if err := s.diff(); err != nil {
			s.log.Error().Err(err).Msg("failed to diff")
		}
	}

	runEndTime := time.Now().UTC()
	s.log.Info().Str(status, finished).Str("duration", runEndTime.Sub(runStartTime).String()).Msg(syncRun)
}

func (s *Sync) producer(watermark *timestamppb.Timestamp) error {
	defer func() {
		s.log.Debug().Msg("producer closed export channel")
		close(s.exportChan)
	}()

	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	dsc, err := s.getRemoteDirectoryClient(ctx)
	if err != nil {
		s.log.Error().Err(err).Msgf("%s - failed to get directory connection", syncProducer)
		return err
	}
	defer dsc.conn.(*grpc.ClientConn).Close()

	counts := Counter{}
	s.log.Info().Str(status, started).Msg(syncProducer)

	stream, err := dsc.Exporter.Export(ctx, &dse3.ExportRequest{
		Options:   uint32(dse3.Option_OPTION_DATA),
		StartFrom: watermark,
	})
	if err != nil {
		return err
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		atomic.AddInt32(&counts.Received, 1)

		switch m := msg.Msg.(type) {
		case *dse3.ExportResponse_Object:
			atomic.AddInt32(&counts.Objects, 1)
			if fullSync(watermark) {
				s.filter.Insert(getObjectKey(m.Object))
			}
		case *dse3.ExportResponse_Relation:
			atomic.AddInt32(&counts.Relations, 1)
			if fullSync(watermark) {
				s.filter.Insert(getRelationKey(m.Relation))
			}
		default:
			s.log.Debug().Msg("producer unknown message type")
			continue // do not msg to exportChan when unknown.
		}

		s.exportChan <- msg
	}

	s.log.Info().Str(status, finished).Int32("received", counts.Received).Int32("objects", counts.Objects).
		Int32("relations", counts.Relations).Int32("errors", counts.Errors).Msg(syncProducer)

	return nil
}

func (s *Sync) subscriber() error {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	dsc, err := s.getLocalDirectoryClient(ctx)
	if err != nil {
		s.log.Error().Err(err).Msgf("subscriber - failed to get directory connection")
		return err
	}
	defer dsc.conn.(*grpc.ClientConn).Close()

	counts := Counter{}
	s.log.Info().Str(status, started).Msg(syncSubscriber)

	watermark := s.getWatermark()

	writer, err := dsc.Importer.Import(ctx)
	if err != nil {
		s.log.Error().Err(err).Msgf("subscriber - failed to setup import stream")
		return err
	}

	for {
		msg, ok := <-s.exportChan
		if !ok {
			s.log.Debug().Msg("export channel closed")
			break
		}

		atomic.AddInt32(&counts.Received, 1)

		switch m := msg.Msg.(type) {
		case *dse3.ExportResponse_Object:
			if err := writer.Send(&dsi3.ImportRequest{
				OpCode: dsi3.Opcode_OPCODE_SET,
				Msg:    &dsi3.ImportRequest_Object{Object: m.Object},
			}); err == nil {
				watermark = maxTS(watermark, m.Object.GetUpdatedAt())
				atomic.AddInt32(&counts.Objects, 1)
				atomic.AddInt32(&counts.Upserts, 1)
			} else {
				s.log.Error().Err(err).Msgf("failed to set object %v", m.Object)
				s.errChan <- err
				atomic.AddInt32(&counts.Errors, 1)
			}

		case *dse3.ExportResponse_Relation:
			if err := writer.Send(&dsi3.ImportRequest{
				OpCode: dsi3.Opcode_OPCODE_SET,
				Msg:    &dsi3.ImportRequest_Relation{Relation: m.Relation},
			}); err == nil {
				watermark = maxTS(watermark, m.Relation.GetUpdatedAt())
				atomic.AddInt32(&counts.Relations, 1)
				atomic.AddInt32(&counts.Upserts, 1)
			} else {
				s.log.Error().Err(err).Msgf("failed to set object %v", m.Relation)
				s.errChan <- err
				atomic.AddInt32(&counts.Errors, 1)
			}

		default:
			s.log.Debug().Msg("unknown message type")
		}
	}

	if err := writer.CloseSend(); err != nil {
		return err
	}

	if err := s.setWatermark(watermark); err != nil {
		s.log.Error().Err(err).Msg("failed to save watermark")
	}

	s.log.Info().Str(status, finished).Int32("received", counts.Received).Int32("objects", counts.Objects).
		Int32("relations", counts.Relations).Int32("errors", counts.Errors).Msg(syncSubscriber)

	return nil
}

func (s *Sync) getLocalDirectoryClient(ctx context.Context) (*directoryClient, error) {
	host := localHost
	if s.topazConfig.DirectoryResolver.Address != "" {
		host = s.topazConfig.DirectoryResolver.Address
	}

	caCertPath := ""
	// when reader registered to same port as authorizer.
	if conf, ok := s.topazConfig.Common.APIConfig.Services["authorizer"]; ok {
		if conf.GRPC.ListenAddress == s.topazConfig.DirectoryResolver.Address {
			caCertPath = conf.GRPC.Certs.TLSCACertPath
			host = conf.GRPC.ListenAddress
		}
	}
	// if reader api configured separately.
	if conf, ok := s.topazConfig.Common.APIConfig.Services["writer"]; ok {
		host = conf.GRPC.ListenAddress
		caCertPath = conf.GRPC.Certs.TLSCACertPath
	}

	opts := []client.ConnectionOption{
		client.WithAddr(host),
		client.WithInsecure(s.topazConfig.DirectoryResolver.Insecure),
		client.WithCACertPath(caCertPath),
	}

	if s.topazConfig.DirectoryResolver.APIKey != "" {
		opts = append(opts, client.WithAPIKeyAuth(s.topazConfig.DirectoryResolver.APIKey))
	}

	opts = append(opts, client.WithSessionID(s.cfg.SessionID))

	if s.topazConfig.DirectoryResolver.TenantID != "" {
		opts = append(opts, client.WithTenantID(s.topazConfig.DirectoryResolver.TenantID))
	}

	conn, err := client.NewConnection(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &directoryClient{
		conn:     conn.Conn,
		Model:    dsm3.NewModelClient(conn.Conn),
		Reader:   dsr3.NewReaderClient(conn.Conn),
		Writer:   dsw3.NewWriterClient(conn.Conn),
		Importer: dsi3.NewImporterClient(conn.Conn),
		Exporter: dse3.NewExporterClient(conn.Conn),
	}, nil
}

func (s *Sync) getRemoteDirectoryClient(ctx context.Context) (*directoryClient, error) {
	host := localHost
	if s.cfg.Addr != "" {
		host = s.cfg.Addr
	}

	opts := []client.ConnectionOption{
		client.WithAddr(host),
		client.WithInsecure(s.cfg.Insecure),
	}

	if s.cfg.APIKey != "" {
		opts = append(opts, client.WithAPIKeyAuth(s.cfg.APIKey))
	}

	if s.cfg.SessionID != "" {
		opts = append(opts, client.WithSessionID(s.cfg.SessionID))
	}

	if s.cfg.TenantID != "" {
		opts = append(opts, client.WithTenantID(s.cfg.TenantID))
	}

	conn, err := client.NewConnection(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &directoryClient{
		conn:     conn.Conn,
		Model:    dsm3.NewModelClient(conn.Conn),
		Reader:   dsr3.NewReaderClient(conn.Conn),
		Writer:   dsw3.NewWriterClient(conn.Conn),
		Importer: dsi3.NewImporterClient(conn.Conn),
		Exporter: dse3.NewExporterClient(conn.Conn),
	}, nil
}

func maxTS(lhs, rhs *timestamppb.Timestamp) *timestamppb.Timestamp {
	if lhs.GetSeconds() > rhs.GetSeconds() {
		return lhs
	} else if lhs.GetSeconds() == rhs.GetSeconds() && lhs.GetNanos() > rhs.GetNanos() {
		return lhs
	}
	return rhs
}

func (s *Sync) getWatermark() *timestamppb.Timestamp {
	fi, err := os.Stat(s.syncFilename())
	if err != nil {
		return &timestamppb.Timestamp{Seconds: 0, Nanos: 0}
	}

	return timestamppb.New(fi.ModTime())
}

func (s *Sync) setWatermark(ts *timestamppb.Timestamp) error {
	w, err := os.Create(s.syncFilename())
	if err != nil {
		return err
	}
	_, _ = w.WriteString(ts.AsTime().Format(time.RFC3339Nano) + "\n")
	w.Close()

	wmTime := ts.AsTime().Add(time.Millisecond)

	if err := os.Chtimes(s.syncFilename(), wmTime, wmTime); err != nil {
		return err
	}

	return nil
}

func (s *Sync) syncFilename() string {
	dir, file := filepath.Split(s.topazConfig.Common.Edge.DBPath)
	return filepath.Join(dir, fmt.Sprintf("%s.%s", file, "sync"))
}

func (s *Sync) syncManifest() error {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	ldsc, err := s.getLocalDirectoryClient(ctx)
	if err != nil {
		return err
	}
	defer ldsc.conn.(*grpc.ClientConn).Close()

	rdsc, err := s.getRemoteDirectoryClient(ctx)
	if err != nil {
		return err
	}
	defer rdsc.conn.(*grpc.ClientConn).Close()

	_, rr, err := s.getManifest(rdsc.Model)
	if err != nil {
		return err
	}

	if err := s.setManifest(ldsc.Model, rr); err != nil {
		return err
	}

	return nil
}

func (s *Sync) getManifest(mc dsm3.ModelClient) (*dsm3.Metadata, io.Reader, error) {
	stream, err := mc.GetManifest(s.ctx, &dsm3.GetManifestRequest{Empty: &emptypb.Empty{}})
	if err != nil {
		return nil, nil, err
	}

	data := bytes.Buffer{}
	var metadata *dsm3.Metadata

	bytesRecv := 0
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, nil, err
		}

		if md, ok := resp.GetMsg().(*dsm3.GetManifestResponse_Metadata); ok {
			metadata = md.Metadata
		}

		if body, ok := resp.GetMsg().(*dsm3.GetManifestResponse_Body); ok {
			data.Write(body.Body.Data)
			bytesRecv += len(body.Body.Data)
		}
	}

	return metadata, bytes.NewReader(data.Bytes()), nil
}

const blockSize = 1024 * 64

func (s *Sync) setManifest(mc dsm3.ModelClient, r io.Reader) error {
	stream, err := mc.SetManifest(s.ctx)
	if err != nil {
		return err
	}

	buf := make([]byte, blockSize)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := stream.Send(&dsm3.SetManifestRequest{
			Msg: &dsm3.SetManifestRequest_Body{
				Body: &dsm3.Body{Data: buf[0:n]},
			},
		}); err != nil {
			return err
		}

		if n < blockSize {
			break
		}
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		return err
	}

	return nil
}

func getObjectKey(obj *dsc3.Object) []byte {
	return []byte(fmt.Sprintf("%s:%s", obj.Type, obj.Id))
}

func getRelationKey(rel *dsc3.Relation) []byte {
	return []byte(fmt.Sprintf("%s:%s#%s@%s:%s%s",
		rel.ObjectType, rel.ObjectId, rel.Relation, rel.SubjectType, rel.SubjectId,
		ds.Iff(rel.SubjectRelation == "", "", fmt.Sprintf("#%s", rel.SubjectRelation))))
}

func fullSync(watermark *timestamppb.Timestamp) bool {
	return (watermark.Seconds == 0 && watermark.Nanos == 0)
}

func (s *Sync) diff() error {
	if s.filter == nil {
		return errors.New("filter not initialized")
	}

	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	dsc, err := s.getLocalDirectoryClient(ctx)
	if err != nil {
		s.log.Error().Err(err).Msgf("diff - failed to create directory client")
		return err
	}
	defer dsc.conn.(*grpc.ClientConn).Close()

	writer, err := dsc.Importer.Import(ctx)
	if err != nil {
		s.log.Error().Err(err).Msgf("subscriber - failed to setup import stream")
		return err
	}

	objDeleted := 0
	relDeleted := 0

	{
		reader, err := dsc.Exporter.Export(ctx, &dse3.ExportRequest{StartFrom: &timestamppb.Timestamp{}, Options: uint32(dse3.Option_OPTION_DATA_OBJECTS)})
		if err != nil {
			return err
		}

		for {
			msg, err := reader.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			obj := msg.Msg.(*dse3.ExportResponse_Object).Object
			if !s.filter.Lookup(getObjectKey(obj)) {
				s.log.Trace().Str("key", string(getObjectKey(obj))).Msg("delete")

				if err := writer.Send(&dsi3.ImportRequest{
					OpCode: dsi3.Opcode_OPCODE_DELETE,
					Msg:    &dsi3.ImportRequest_Object{Object: obj},
				}); err == nil {
					atomic.AddInt32(&s.counts.Deletes, 1)
					objDeleted++
				} else {
					s.log.Error().Err(err).Msgf("failed to delete object %v", obj)
					s.errChan <- err
					atomic.AddInt32(&s.counts.Errors, 1)
				}
			}
		}
	}

	{
		reader, err := dsc.Exporter.Export(ctx, &dse3.ExportRequest{StartFrom: &timestamppb.Timestamp{}, Options: uint32(dse3.Option_OPTION_DATA_RELATIONS)})
		if err != nil {
			return err
		}
		for {
			msg, err := reader.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			rel := msg.Msg.(*dse3.ExportResponse_Relation).Relation
			if !s.filter.Lookup(getRelationKey(rel)) {
				s.log.Trace().Str("key", string(getRelationKey(rel))).Msg("delete")

				if err := writer.Send(&dsi3.ImportRequest{
					OpCode: dsi3.Opcode_OPCODE_DELETE,
					Msg:    &dsi3.ImportRequest_Relation{Relation: rel},
				}); err == nil {
					atomic.AddInt32(&s.counts.Deletes, 1)
					relDeleted++
				} else {
					s.log.Error().Err(err).Msgf("failed to delete relation %v", rel)
					s.errChan <- err
					atomic.AddInt32(&s.counts.Errors, 1)
				}
			}
		}
	}

	if err := writer.CloseSend(); err != nil {
		return err
	}

	s.filter = nil

	s.log.Info().Int("obj", objDeleted).Int("rel", relDeleted).Msg("diff")

	return nil
}
