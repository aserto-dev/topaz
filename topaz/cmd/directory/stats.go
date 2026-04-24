package directory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	dse3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/azm/stats"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"github.com/aserto-dev/topaz/topaz/table"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

type StatsCmd struct {
	dsc.Config

	Output string `flag:"" short:"o" enum:"table,json" default:"table" help:"output format"`
}

func (cmd *StatsCmd) Run(ctx context.Context) error {
	client, err := dsc.NewClient(ctx, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	stream, err := client.Reader.Export(ctx, &dse3.ExportRequest{
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

		if m, ok := msg.GetMsg().(*dse3.ExportResponse_Stats); ok {
			pbStats = m.Stats
		}
	}

	if cmd.Output == "json" {
		if err := jsonx.OutputJSONPB(os.Stdout, pbStats); err != nil {
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

		statsTable(os.Stderr, &s)
	}

	return nil
}

func statsTable(w io.Writer, s *stats.Stats) {
	tab := table.New(w)
	defer tab.Close()

	tab.Header("ObjType", "Count", "Rel", "Count", "SubType", "Count", "SubRel", "Count")

	data := [][]any{}

	for ot, objType := range s.ObjectTypes {
		for or, objRel := range objType.Relations {
			for st, subType := range objRel.SubjectTypes {
				data = append(data, []any{
					ot.String(), countStr(objType.ObjCount),
					or.String(), countStr(objRel.Count),
					st.String(), countStr(subType.Count),
					"", countStr(0),
				})

				for sr, subRel := range subType.SubjectRelations {
					data = append(data, []any{
						ot.String(), countStr(objType.ObjCount),
						or.String(), countStr(objRel.Count),
						st.String(), countStr(subType.Count),
						sr.String(), countStr(subRel.Count),
					})
				}
			}
		}
	}

	tab.Bulk(data)
	tab.Render()
}

func countStr(c int32) string { return fmt.Sprintf("%8d", c) }
