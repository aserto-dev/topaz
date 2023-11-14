package edge

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/go-edge-ds/pkg/bdb"

	"github.com/aserto-dev/go-aserto/client"
	topaz "github.com/aserto-dev/topaz/pkg/cc/config"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	syncRun     string = "run"
	channelSize int    = 2000
	localHost   string = "localhost:9292"
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
	s.log.Info().Time("started", runStartTime).Msg(syncRun)

	defer func() {
		close(s.errChan)
	}()

	// error spew.
	go func() {
		for e := range s.errChan {
			s.log.Error().Err(e).Msg(syncRun)
		}
	}()

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
	s.log.Info().Time("ended", runEndTime).Msg(syncRun)
	s.log.Info().Str("duration", runEndTime.Sub(runStartTime).String()).Msg(syncRun)
}

func (s *Sync) producer() error {
	pluginDirClient, err := s.getPluginDirectoryClient()
	if err != nil {
		s.log.Error().Err(err).Msgf("producer - failed to get directory connection")
		return err
	}

	counts := Counter{}
	s.log.Info().Time("producer-start", time.Now().UTC()).Msg(syncRun)

	watermark, err := s.getWatermark()
	if err != nil {
		return err
	}

	stream, err := pluginDirClient.Exporter.Export(s.ctx, &dse3.ExportRequest{
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
			s.log.Debug().Msg("unknown message type")
		}

		s.exportChan <- msg
	}

	s.log.Debug().Msg("producer closed export channel")
	close(s.exportChan)

	s.log.Info().Int32("received", counts.Received).Msg(syncRun)
	s.log.Info().Int32("objects", counts.Objects).Msg(syncRun)
	s.log.Info().Int32("relations", counts.Relations).Msg(syncRun)
	s.log.Info().Time("producer-end", time.Now().UTC()).Msg(syncRun)

	return nil
}

func (s *Sync) subscriber() error {
	topazDirClient, err := s.getTopazDirectoryClient()
	if err != nil {
		s.log.Error().Err(err).Msgf("subscriber - failed to get directory connection")
		return err
	}

	counts := Counter{}
	s.log.Info().Time("subscriber-start", time.Now().UTC()).Msg(syncRun)

	watermark := &timestamppb.Timestamp{Seconds: 0, Nanos: 0}

	stream, err := topazDirClient.Importer.Import(s.ctx)
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

	s.log.Info().Int32("received", counts.Received).Msg(syncRun)
	s.log.Info().Int32("manifest", counts.Manifests).Msg(syncRun)
	s.log.Info().Int32("objects", counts.Objects).Msg(syncRun)
	s.log.Info().Int32("relations", counts.Relations).Msg(syncRun)

	s.log.Info().Int32("upserts", counts.Upserts).Msg(syncRun)
	s.log.Info().Int32("deletes", counts.Deletes).Msg(syncRun)
	s.log.Info().Int32("errors", counts.Errors).Msg(syncRun)

	s.log.Info().Time("subscriber-end", time.Now().UTC()).Msg(syncRun)

	return nil
}

func (s *Sync) getTopazDirectoryClient() (*directoryClient, error) {
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

func (s *Sync) getPluginDirectoryClient() (*directoryClient, error) {
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

func (s *Sync) getWatermark() (*timestamppb.Timestamp, error) {
	client, err := s.getTopazDirectoryClient()
	if err != nil {
		s.log.Error().Err(err).Msgf("getWatermark - failed to get directory connection")
		return nil, err
	}

	result, err := client.Reader.GetObject(s.ctx,
		&dsr3.GetObjectRequest{
			ObjectType:    "system",
			ObjectId:      "edge_sync",
			WithRelations: false,
			Page:          &dsc3.PaginationRequest{Size: 1},
		},
	)
	switch {
	case bdb.ErrIsNotFound(err):
		return &timestamppb.Timestamp{Seconds: 0, Nanos: 0}, nil
	case err != nil:
		return nil, err
	default:
	}

	seconds := int64(result.Result.Properties.GetFields()["last_updated_seconds"].GetNumberValue())
	nanos := int32(result.Result.Properties.GetFields()["last_updated_nanos"].GetNumberValue())

	return &timestamppb.Timestamp{Seconds: seconds, Nanos: nanos}, nil
}

func (s *Sync) setWatermark(ts *timestamppb.Timestamp) error {
	client, err := s.getTopazDirectoryClient()
	if err != nil {
		s.log.Error().Err(err).Msgf("setWatermark - failed to get directory connection")
		return err
	}

	props := pb.NewStruct()
	props.Fields["last_updated_seconds"] = structpb.NewNumberValue(float64(ts.GetSeconds()))
	props.Fields["last_updated_nanos"] = structpb.NewNumberValue(float64(ts.GetNanos() + 1))

	if _, err := client.Writer.SetObject(s.ctx,
		&dsw3.SetObjectRequest{Object: &dsc3.Object{
			Type:       "system",
			Id:         "edge_sync",
			Properties: props,
		}},
	); err != nil {
		return err
	}

	return nil
}
