package directory

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	dse "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) ExportToFile(ctx context.Context, w io.Writer, options uint32) error {
	stream, err := c.Exporter.Export(ctx, &dse.ExportRequest{
		Options:   options,
		StartFrom: &timestamppb.Timestamp{},
	})
	if err != nil {
		return err
	}

	encoder := jsonx.NewEncoder(w)

	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		switch m := msg.GetMsg().(type) {
		case *dse.ExportResponse_Object:
			if err := encoder.Encode(m.Object); err != nil {
				return err
			}

		case *dse.ExportResponse_Relation:
			if err := encoder.Encode(m.Relation); err != nil {
				return err
			}

		default:
			fmt.Fprintf(os.Stderr, "unknown message type\n")
		}
	}

	return nil
}
