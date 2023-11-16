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

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"

	"github.com/aserto-dev/go-aserto/client"
	topaz "github.com/aserto-dev/topaz/pkg/cc/config"

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
	channelSize    int    = 2000
	localHost      string = "localhost:9292"
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

func NewSyncMgr(c *Config, topazConfig *topaz.Config, logger *zerolog.Logger) *Sync {
	return &Sync{
		ctx:         context.Background(),
		cfg:         c,
		topazConfig: topazConfig,
		log:         logger,
		exportChan:  make(chan *dse3.ExportResponse, channelSize),
		errChan:     make(chan error, 1),
	}
}

func (s *Sync) Run() {
	runStartTime := time.Now().UTC()
	s.log.Info().Str(status, started).Msg(syncRun)

	defer func() {
		close(s.errChan)
	}()

	// error spew.
	go func() {
		for e := range s.errChan {
			s.log.Error().Err(e).Msg(syncRun)
		}
	}()

	if err := s.syncManifest(); err != nil {
		s.log.Error().Str("sync-manifest", "").Err(err).Msg(syncRun)
	}

	g := new(errgroup.Group)

	g.Go(func() error {
		return s.subscriber()
	})

	g.Go(func() error {
		return s.producer()
	})

	err := g.Wait()
	if err != nil {
		s.log.Error().Err(err).Msg("sync run failed")
	}

	runEndTime := time.Now().UTC()
	s.log.Info().Str(status, finished).Str("duration", runEndTime.Sub(runStartTime).String()).Msg(syncRun)
}

func (s *Sync) producer() error {
	dsc, err := s.getRemoteDirectoryClient()
	if err != nil {
		s.log.Error().Err(err).Msgf("%s - failed to get directory connection", syncProducer)
		return err
	}

	counts := Counter{}
	s.log.Info().Str(status, started).Msg(syncProducer)

	watermark := s.getWatermark()

	stream, err := dsc.Exporter.Export(s.ctx, &dse3.ExportRequest{
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

		switch msg.Msg.(type) {
		case *dse3.ExportResponse_Object:
			atomic.AddInt32(&counts.Objects, 1)
		case *dse3.ExportResponse_Relation:
			atomic.AddInt32(&counts.Relations, 1)
		default:
			s.log.Debug().Msg("producer unknown message type")
		}

		s.exportChan <- msg
	}

	s.log.Debug().Msg("producer closed export channel")
	close(s.exportChan)

	s.log.Info().Str(status, finished).Int32("received", counts.Received).Int32("objects", counts.Objects).
		Int32("relations", counts.Relations).Int32("errors", counts.Errors).Msg(syncProducer)

	return nil
}

func (s *Sync) subscriber() error {
	dsc, err := s.getLocalDirectoryClient()
	if err != nil {
		s.log.Error().Err(err).Msgf("subscriber - failed to get directory connection")
		return err
	}

	counts := Counter{}
	s.log.Info().Str(status, started).Msg(syncSubscriber)

	watermark := s.getWatermark()

	stream, err := dsc.Importer.Import(s.ctx)
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
			if err := stream.Send(&dsi3.ImportRequest{
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
			if err := stream.Send(&dsi3.ImportRequest{
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

	if err := stream.CloseSend(); err != nil {
		return err
	}

	if err := s.setWatermark(watermark); err != nil {
		s.log.Error().Err(err).Msg("failed to save watermark")
	}

	s.log.Info().Str(status, finished).Int32("received", counts.Received).Int32("objects", counts.Objects).
		Int32("relations", counts.Relations).Int32("errors", counts.Errors).Msg(syncSubscriber)

	return nil
}

func (s *Sync) getLocalDirectoryClient() (*directoryClient, error) {
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

	conn, err := client.NewConnection(s.ctx, opts...)
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

func (s *Sync) getRemoteDirectoryClient() (*directoryClient, error) {
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

	conn, err := client.NewConnection(s.ctx, opts...)
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
	ldsc, err := s.getLocalDirectoryClient()
	if err != nil {
		return err
	}

	rdsc, err := s.getRemoteDirectoryClient()
	if err != nil {
		return err
	}

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
