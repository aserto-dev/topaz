package v3

import (
	"encoding/json"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/ds"

	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
)

type Exporter struct {
	dse3.UnimplementedExporterServer

	logger *zerolog.Logger
	store  *bdb.BoltDB
}

func NewExporter(logger *zerolog.Logger, store *bdb.BoltDB) *Exporter {
	return &Exporter{
		logger: logger,
		store:  store,
	}
}

func (s *Exporter) Export(req *dse3.ExportRequest, stream dse3.Exporter_ExportServer) error {
	logger := s.logger.With().Str("method", "Export").Interface("req", req).Logger()

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		// stats mode, short circuits when enabled
		if req.GetOptions()&uint32(dse3.Option_OPTION_STATS) != 0 {
			if err := exportStats(tx, stream, req.GetOptions()); err != nil {
				logger.Error().Err(err).Msg("export_stats")
				return err
			}

			return nil
		}

		if req.GetOptions()&uint32(dse3.Option_OPTION_DATA_OBJECTS) != 0 {
			if err := exportObjects(tx, stream); err != nil {
				logger.Error().Err(err).Msg("export_objects")
				return err
			}
		}

		if req.GetOptions()&uint32(dse3.Option_OPTION_DATA_RELATIONS) != 0 {
			if err := exportRelations(tx, stream); err != nil {
				logger.Error().Err(err).Msg("export_relations")
				return err
			}
		}

		return nil
	})

	return err
}

func exportObjects(tx *bolt.Tx, stream dse3.Exporter_ExportServer) error {
	iter, err := bdb.NewScanIterator[dsc3.Object](stream.Context(), tx, bdb.ObjectsPath)
	if err != nil {
		return err
	}

	for iter.Next() {
		if err := stream.Send(&dse3.ExportResponse{Msg: &dse3.ExportResponse_Object{Object: iter.Value()}}); err != nil {
			return err
		}
	}

	return nil
}

func exportRelations(tx *bolt.Tx, stream dse3.Exporter_ExportServer) error {
	iter, err := bdb.NewScanIterator[dsc3.Relation](stream.Context(), tx, bdb.RelationsObjPath)
	if err != nil {
		return err
	}

	for iter.Next() {
		if err := stream.Send(&dse3.ExportResponse{Msg: &dse3.ExportResponse_Relation{Relation: iter.Value()}}); err != nil {
			return err
		}
	}

	return nil
}

func exportStats(tx *bolt.Tx, stream dse3.Exporter_ExportServer, opts uint32) error {
	stats := ds.NewStats()

	// object stats.
	if opts&uint32(dse3.Option_OPTION_DATA_OBJECTS) != 0 {
		if err := stats.CountObjects(stream.Context(), tx); err != nil {
			return err
		}
	}

	// relation stats.
	if opts&uint32(dse3.Option_OPTION_DATA_RELATIONS) != 0 {
		if err := stats.CountRelations(stream.Context(), tx); err != nil {
			return err
		}
	}

	buf, err := json.Marshal(stats.Stats)
	if err != nil {
		return err
	}

	resp := pb.NewStruct()
	if err := resp.UnmarshalJSON(buf); err != nil {
		return err
	}

	if err := stream.Send(&dse3.ExportResponse{Msg: &dse3.ExportResponse_Stats{Stats: resp}}); err != nil {
		return err
	}

	return nil
}
