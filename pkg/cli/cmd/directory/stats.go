package directory

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/aserto-dev/azm/model"
	"github.com/aserto-dev/azm/stats"
	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/table"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

type StatsCmd struct {
	Output string `flag:"" short:"o" enum:"table,json" default:"table" help:"output format"`
	dsc.Config
}

func (cmd *StatsCmd) Run(c *cc.CommonCtx) error {
	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	stream, err := client.Exporter.Export(c.Context, &dse3.ExportRequest{
		Options: uint32(dse3.Option_OPTION_DATA + dse3.Option_OPTION_STATS),
	})
	if err != nil {
		return err
	}

	var pbStats *structpb.Struct

	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if m, ok := msg.Msg.(*dse3.ExportResponse_Stats); ok {
			pbStats = m.Stats
		}
	}

	if cmd.Output == "json" {
		if err := jsonx.OutputJSONPB(c.StdOut(), pbStats); err != nil {
			return err
		}
		return nil
	}

	if cmd.Output == "table" {
		buf, err := pbStats.MarshalJSON()
		if err != nil {
			return err
		}

		var s stats.Stats
		if err := json.Unmarshal(buf, &s); err != nil {
			return err
		}

		statsTable(c.StdErr(), &s)
	}

	return nil
}

func statsTable(w io.Writer, s *stats.Stats) {
	tab := table.New(w).WithColumns("obj-type", "obj-type-count", "rel", "rel-count", "sub-type", "sub-type-count", "sub-rel", "sub-rel-count")
	tab.WithTableNoAutoWrapText()

	for ot, objType := range s.ObjectTypes {
		for or, objRel := range objType.Relations {
			for st, subType := range objRel.SubjectTypes {
				tabRow(tab, ot, objType.ObjCount, or, objRel.Count, st, subType.Count, "", 0)

				for sr, subRel := range subType.SubjectRelations {
					tabRow(tab, ot, objType.ObjCount, or, objRel.Count, st, subType.Count, sr, subRel.Count)
				}
			}
		}
	}

	tab.Do()
}

func countStr(c int32) string { return fmt.Sprintf("%8d", c) }

func tabRow(tab *table.TableWriter, objType model.ObjectName, objTypeCount int32, objRel model.RelationName, objRelCount int32, subType model.ObjectName, subTypeCount int32, subRel model.RelationName, subRelCount int32) {
	tab.WithRow(
		objType.String(),
		countStr(objTypeCount),
		objRel.String(),
		countStr(objRelCount),
		subType.String(),
		countStr(subTypeCount),
		subRel.String(),
		countStr(subRelCount),
	)
}
