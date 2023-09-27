package edge

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	dse2 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v2"
	dsi2 "github.com/aserto-dev/go-directory/aserto/directory/importer/v2"
	dsr2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	dsw2 "github.com/aserto-dev/go-directory/aserto/directory/writer/v2"
	"golang.org/x/sync/errgroup"

	dsClient "github.com/aserto-dev/go-aserto/client"
	topaz "github.com/aserto-dev/topaz/pkg/cc/config"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	syncRun     string = "run"
	channelSize int    = 2000
	localHost   string = "localhost:9292"
)

type directoryClient struct {
	conn     grpc.ClientConnInterface
	Reader   dsr2.ReaderClient
	Writer   dsw2.WriterClient
	Importer dsi2.ImporterClient
	Exporter dse2.ExporterClient
}

type Sync struct {
	ctx         context.Context
	cfg         *Config
	topazConfig *topaz.Config
	log         *zerolog.Logger
	exportChan  chan *dse2.ExportResponse
	errChan     chan error
}

type Counter struct {
	Received      int32
	Upserts       int32
	Deletes       int32
	Errors        int32
	ObjectTypes   int32
	RelationTypes int32
	Permissions   int32
	Objects       int32
	Relations     int32
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
		exportChan:  make(chan *dse2.ExportResponse, channelSize),
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

	stream, err := pluginDirClient.Exporter.Export(s.ctx, &dse2.ExportRequest{
		Options:   uint32(dse2.Option_OPTION_ALL),
		StartFrom: &timestamppb.Timestamp{},
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
		case *dse2.ExportResponse_ObjectType:
			atomic.AddInt32(&counts.ObjectTypes, 1)
		case *dse2.ExportResponse_RelationType:
			atomic.AddInt32(&counts.RelationTypes, 1)
		case *dse2.ExportResponse_Permission:
			atomic.AddInt32(&counts.Permissions, 1)
		case *dse2.ExportResponse_Object:
			atomic.AddInt32(&counts.Objects, 1)
		case *dse2.ExportResponse_Relation:
			atomic.AddInt32(&counts.Relations, 1)
		default:
			s.log.Debug().Msg("unknown message type")
		}

		s.exportChan <- msg
	}

	s.log.Debug().Msg("producer closed export channel")
	close(s.exportChan)

	s.log.Info().Int32("received", counts.Received).Msg(syncRun)
	s.log.Info().Int32("object_types", counts.ObjectTypes).Msg(syncRun)
	s.log.Info().Int32("relation_types", counts.RelationTypes).Msg(syncRun)
	s.log.Info().Int32("permissions", counts.Permissions).Msg(syncRun)
	s.log.Info().Int32("objects", counts.Objects).Msg(syncRun)
	s.log.Info().Int32("relations", counts.Relations).Msg(syncRun)
	s.log.Info().Time("producer-end", time.Now().UTC()).Msg(syncRun)

	return nil
}

func (s *Sync) subscriber() error {
	topazDirclient, err := s.getTopazDirectoryClient()
	if err != nil {
		s.log.Error().Err(err).Msgf("subscriber - failed to get directory connection")
		return err
	}

	counts := Counter{}
	s.log.Info().Time("subscriber-start", time.Now().UTC()).Msg(syncRun)

	for {
		msg, ok := <-s.exportChan
		if !ok {
			s.log.Debug().Msg("export channel closed")
			break
		}
		atomic.AddInt32(&counts.Received, 1)

		switch m := msg.Msg.(type) {
		case *dse2.ExportResponse_ObjectType:
			_, err := topazDirclient.Writer.SetObjectType(s.ctx, &dsw2.SetObjectTypeRequest{ObjectType: m.ObjectType})
			if err != nil {
				s.log.Error().Err(err).Msgf("failed to set object type %v", m.ObjectType)
				s.errChan <- err
				atomic.AddInt32(&counts.Errors, 1)
			}
			atomic.AddInt32(&counts.ObjectTypes, 1)

		case *dse2.ExportResponse_RelationType:
			_, err := topazDirclient.Writer.SetRelationType(s.ctx, &dsw2.SetRelationTypeRequest{RelationType: m.RelationType})
			if err != nil {
				s.log.Error().Err(err).Msgf("failed to set relation type %v", m.RelationType)
				s.errChan <- err
				atomic.AddInt32(&counts.Errors, 1)
			}
			atomic.AddInt32(&counts.RelationTypes, 1)

		case *dse2.ExportResponse_Permission:
			_, err := topazDirclient.Writer.SetPermission(s.ctx, &dsw2.SetPermissionRequest{Permission: m.Permission})
			if err != nil {
				s.log.Error().Err(err).Msgf("failed to set permission %v", m.Permission)
				s.errChan <- err
				atomic.AddInt32(&counts.Errors, 1)
			}
			atomic.AddInt32(&counts.Permissions, 1)

		case *dse2.ExportResponse_Object:
			_, err := topazDirclient.Writer.SetObject(s.ctx, &dsw2.SetObjectRequest{Object: m.Object})
			if err != nil {
				s.log.Error().Err(err).Msgf("failed to set object %v", m.Object)
				s.errChan <- err
				atomic.AddInt32(&counts.Errors, 1)
			}
			atomic.AddInt32(&counts.Objects, 1)

		case *dse2.ExportResponse_Relation:
			_, err := topazDirclient.Writer.SetRelation(s.ctx, &dsw2.SetRelationRequest{Relation: m.Relation})
			if err != nil {
				s.log.Error().Err(err).Msgf("failed to set relation %v", m.Relation)
				s.errChan <- err
				atomic.AddInt32(&counts.Errors, 1)
			}
			atomic.AddInt32(&counts.Relations, 1)

		default:
			s.log.Debug().Msg("unknown message type")
		}
	}

	s.log.Info().Int32("received", counts.Received).Msg(syncRun)
	s.log.Info().Int32("object_types", counts.ObjectTypes).Msg(syncRun)
	s.log.Info().Int32("relation_types", counts.RelationTypes).Msg(syncRun)
	s.log.Info().Int32("permissions", counts.Permissions).Msg(syncRun)
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
	if conf, ok := s.topazConfig.Common.Services["authorizer"]; ok {
		if conf.GRPC.ListenAddress == s.topazConfig.DirectoryResolver.Address {
			caCertPath = conf.GRPC.Certs.TLSCACertPath
			host = conf.GRPC.ListenAddress
		}
	}
	// if reader api configured separately.
	if conf, ok := s.topazConfig.Common.Services["writer"]; ok {
		host = conf.GRPC.ListenAddress
		caCertPath = conf.GRPC.Certs.TLSCACertPath
	}

	opts := []dsClient.ConnectionOption{
		dsClient.WithAddr(host),
		dsClient.WithInsecure(s.topazConfig.DirectoryResolver.Insecure),
		dsClient.WithCACertPath(caCertPath),
	}

	if s.topazConfig.DirectoryResolver.APIKey != "" {
		opts = append(opts, dsClient.WithAPIKeyAuth(s.topazConfig.DirectoryResolver.APIKey))
	}

	opts = append(opts, dsClient.WithSessionID(s.cfg.SessionID))

	if s.topazConfig.DirectoryResolver.TenantID != "" {
		opts = append(opts, dsClient.WithTenantID(s.topazConfig.DirectoryResolver.TenantID))
	}

	conn, err := dsClient.NewConnection(s.ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &directoryClient{
		conn:     conn.Conn,
		Reader:   dsr2.NewReaderClient(conn.Conn),
		Writer:   dsw2.NewWriterClient(conn.Conn),
		Importer: dsi2.NewImporterClient(conn.Conn),
		Exporter: dse2.NewExporterClient(conn.Conn),
	}, nil
}

func (s *Sync) getPluginDirectoryClient() (*directoryClient, error) {
	host := localHost
	if s.cfg.Addr != "" {
		host = s.cfg.Addr
	}

	opts := []dsClient.ConnectionOption{
		dsClient.WithAddr(host),
		dsClient.WithInsecure(s.cfg.Insecure),
	}

	if s.cfg.APIKey != "" {
		opts = append(opts, dsClient.WithAPIKeyAuth(s.cfg.APIKey))
	}

	if s.cfg.SessionID != "" {
		opts = append(opts, dsClient.WithSessionID(s.cfg.SessionID))
	}

	if s.cfg.TenantID != "" {
		opts = append(opts, dsClient.WithTenantID(s.cfg.TenantID))
	}

	conn, err := dsClient.NewConnection(s.ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &directoryClient{
		conn:     conn.Conn,
		Reader:   dsr2.NewReaderClient(conn.Conn),
		Writer:   dsw2.NewWriterClient(conn.Conn),
		Importer: dsi2.NewImporterClient(conn.Conn),
		Exporter: dse2.NewExporterClient(conn.Conn),
	}, nil
}
