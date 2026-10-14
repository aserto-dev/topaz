package v3

import (
	"encoding/json"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dse "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/eds/pkg/ds"

	bolt "go.etcd.io/bbolt"
)

func exportObjects(tx *bolt.Tx, stream dsr.Reader_ExportServer) error {
	iter, err := bdb.NewScanIterator[dsc.Object](stream.Context(), tx, bdb.ObjectsPath)
	if err != nil {
		return err
	}

	for iter.Next() {
		if err := stream.Send(&dsr.ExportResponse{Msg: &dsr.ExportResponse_Object{Object: iter.Value()}}); err != nil {
			return err
		}
	}

	return nil
}

func exportRelations(tx *bolt.Tx, stream dsr.Reader_ExportServer) error {
	iter, err := bdb.NewScanIterator[dsc.Relation](stream.Context(), tx, bdb.RelationsObjPath)
	if err != nil {
		return err
	}

	for iter.Next() {
		if err := stream.Send(&dsr.ExportResponse{Msg: &dsr.ExportResponse_Relation{Relation: iter.Value()}}); err != nil {
			return err
		}
	}

	return nil
}

func exportStats(tx *bolt.Tx, stream dsr.Reader_ExportServer, opts uint32) error {
	stats := ds.NewStats()

	// object stats.
	if opts&uint32(dse.Option_OPTION_DATA_OBJECTS) != 0 {
		if err := stats.CountObjects(stream.Context(), tx); err != nil {
			return err
		}
	}

	// relation stats.
	if opts&uint32(dse.Option_OPTION_DATA_RELATIONS) != 0 {
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

	if err := stream.Send(&dsr.ExportResponse{Msg: &dsr.ExportResponse_Stats{Stats: resp}}); err != nil {
		return err
	}

	return nil
}
