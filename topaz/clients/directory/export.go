package directory

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/topaz/js"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) Export(ctx context.Context, objectsFile, relationsFile string) error {
	stream, err := c.Reader.Export(ctx, &dsr.ExportRequest{
		Options:   uint32(dsr.Option_OPTION_DATA),
		StartFrom: &timestamppb.Timestamp{},
	})
	if err != nil {
		return err
	}

	objects, err := js.NewWriter(objectsFile, ObjectsStr)
	if err != nil {
		return err
	}
	defer objects.Close()

	relations, err := js.NewWriter(relationsFile, RelationsStr)
	if err != nil {
		return err
	}
	defer relations.Close()

	ctr := &Counter{}
	objectsCounter := ctr.Objects()
	relationsCounter := ctr.Relations()

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
			err = objects.Write(m.Object)

			objectsCounter.Incr().Print(os.Stdout)

		case *dsr.ExportResponse_Relation:
			err = relations.Write(m.Relation)

			relationsCounter.Incr().Print(os.Stdout)

		default:
			fmt.Fprintf(os.Stderr, "unknown message type\n")
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "err: %v\n", err)
		}
	}

	ctr.Print(os.Stdout)

	return nil
}
