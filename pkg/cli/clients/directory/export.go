package directory

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/js"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) ExportToDirectory(ctx context.Context, objectsFile, relationsFile string) error {
	stream, err := c.Exporter.Export(ctx, &dse3.ExportRequest{
		Options:   uint32(dse3.Option_OPTION_DATA),
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

		switch m := msg.Msg.(type) {
		case *dse3.ExportResponse_Object:
			err = objects.Write(m.Object)
			objectsCounter.Incr().Print(os.Stdout)

		case *dse3.ExportResponse_Relation:
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

var mOpts = protojson.MarshalOptions{
	Multiline:         false,
	Indent:            "",
	UseProtoNames:     true,
	UseEnumNumbers:    false,
	AllowPartial:      true,
	EmitUnpopulated:   false,
	EmitDefaultValues: false,
}

func (c *Client) ExportToFile(ctx context.Context, file *os.File, options dse3.Option) error {
	stream, err := c.Exporter.Export(ctx, &dse3.ExportRequest{
		Options:   uint32(options),
		StartFrom: &timestamppb.Timestamp{},
	})
	if err != nil {
		return err
	}

	defer func() { _ = file.Sync() }()

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

		switch m := msg.Msg.(type) {
		case *dse3.ExportResponse_Object:
			mm := &dsw3.SetObjectRequest{Object: m.Object}
			mm.Object.CreatedAt = nil
			mm.Object.UpdatedAt = nil
			mm.Object.Etag = ""

			buf, err := mOpts.Marshal(mm)
			if err != nil {
				return err
			}
			_, _ = file.Write(buf)
			_, _ = file.WriteString("\n")

			objectsCounter.Incr()

		case *dse3.ExportResponse_Relation:
			mm := &dsw3.SetRelationRequest{Relation: m.Relation}
			mm.Relation.CreatedAt = nil
			mm.Relation.UpdatedAt = nil
			mm.Relation.Etag = ""

			buf, err := mOpts.Marshal(mm)
			if err != nil {
				return err
			}
			_, _ = file.Write(buf)
			_, _ = file.WriteString("\n")

			relationsCounter.Incr()

		default:
			fmt.Fprintf(os.Stderr, "unknown message type\n")
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "err: %v\n", err)
		}
	}

	ctr.Print(os.Stderr)

	return nil
}
