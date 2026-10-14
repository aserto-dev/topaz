package directory

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) ExportToFile(ctx context.Context, w io.Writer, options uint32) error {
	stream, err := c.Reader.Export(ctx, &dsr.ExportRequest{
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
		case *dsr.ExportResponse_Object:
			if err := encoder.Encode(m.Object); err != nil {
				return err
			}

		case *dsr.ExportResponse_Relation:
			if err := encoder.Encode(m.Relation); err != nil {
				return err
			}

		default:
			fmt.Fprintf(os.Stderr, "unknown message type\n")
		}
	}

	return nil
}
