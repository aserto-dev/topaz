package directory

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/aserto-dev/azm/stats"
	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/table"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pkg/errors"
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
		if err == io.EOF {
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

		var stats stats.Stats
		if err := json.Unmarshal(buf, &stats); err != nil {
			return err
		}

		statsTable(c.StdErr(), &stats)
	}

	return nil
}

func statsTable(w io.Writer, stats *stats.Stats) {
	tab := table.New(w).WithColumns("obj-type", "obj-type-count", "rel", "rel-count", "sub-type", "sub-type-count", "sub-rel", "sub-rel-count")
	tab.WithTableNoAutoWrapText()

	for ot, objType := range stats.ObjectTypes {
		for or, objRel := range objType.Relations {
			for st, subType := range objRel.SubjectTypes {
				tab.WithRow(
					ot.String(),
					fmt.Sprintf("%8d", objType.ObjCount),
					or.String(),
					fmt.Sprintf("%8d", objRel.Count),
					st.String(),
					fmt.Sprintf("%8d", subType.Count),
					"",
					fmt.Sprintf("%8d", 0))

				for sr, subRel := range subType.SubjectRelations {
					tab.WithRow(
						ot.String(),
						fmt.Sprintf("%8d", objType.ObjCount),
						or.String(),
						fmt.Sprintf("%8d", objRel.Count),
						st.String(),
						fmt.Sprintf("%8d", subType.Count),
						sr.String(),
						fmt.Sprintf("%8d", subRel.Count))
				}
			}
		}
	}

	tab.Do()
}
