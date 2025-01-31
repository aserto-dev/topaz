package directory

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/js"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func (c *Client) ImportFromDirectory(ctx context.Context, files []string) error {
	g, iCtx := errgroup.WithContext(context.Background())
	stream, err := c.Importer.Import(iCtx)
	if err != nil {
		return err
	}

	g.Go(c.receiver(stream))

	g.Go(c.importHandler(stream, files))

	return g.Wait()
}

func (c *Client) importHandler(stream dsi3.Importer_ImportClient, files []string) func() error {
	return func() error {
		for _, file := range files {
			if err := c.importFile(stream, file); err != nil {
				return err
			}
		}

		if err := stream.CloseSend(); err != nil {
			return err
		}

		return nil
	}
}

func (c *Client) importFile(stream dsi3.Importer_ImportClient, file string) error {
	r, err := os.Open(file)
	if err != nil {
		return errors.Wrapf(err, "failed to open file: [%s]", file)
	}
	defer r.Close()

	reader, err := js.NewReader(r)
	if err != nil || reader == nil {
		fmt.Fprintf(os.Stderr, "Skipping file [%s]: [%s]\n", file, err.Error())
		return nil
	}
	defer reader.Close()

	objectType := reader.GetObjectType()
	switch objectType {
	case ObjectsStr:
		if err := c.loadObjects(stream, reader); err != nil {
			return err
		}

	case RelationsStr:
		if err := c.loadRelations(stream, reader); err != nil {
			return err
		}

	default:
		fmt.Fprintf(os.Stderr, "skipping file [%s] with object type [%s]\n", file, objectType)
	}

	return nil
}

var (
	objPrefix = []byte("{\"object\":")
	relPrefix = []byte("{\"relation\":")
)

var uOpts = protojson.UnmarshalOptions{
	AllowPartial:   true,
	DiscardUnknown: true,
}

func (c *Client) ImportFromFile(ctx context.Context, file *os.File, options dse3.Option) error {
	g, iCtx := errgroup.WithContext(ctx)
	stream, err := c.Importer.Import(iCtx)
	if err != nil {
		return err
	}

	g.Go(func() error {
		objCounter := &dsi3.ImportCounter{Type: objectsCounter}
		relCounter := &dsi3.ImportCounter{Type: relationsCounter}

		defer func() {
			printCounter(os.Stderr, objCounter)
			printCounter(os.Stderr, relCounter)
		}()

		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return err
			}

			switch m := msg.Msg.(type) {
			case *dsi3.ImportResponse_Status:
				printStatus(os.Stderr, m.Status)
			case *dsi3.ImportResponse_Counter:
				switch m.Counter.Type {
				case objectsCounter:
					objCounter = m.Counter
				case relationsCounter:
					relCounter = m.Counter
				}
			}
		}
	})

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		switch {
		case bytes.HasPrefix(scanner.Bytes(), objPrefix):
			setObj := &dsw3.SetObjectRequest{}
			if err := uOpts.Unmarshal(scanner.Bytes(), setObj); err != nil {
				return err
			}

			if err := stream.Send(&dsi3.ImportRequest{
				OpCode: dsi3.Opcode_OPCODE_SET,
				Msg: &dsi3.ImportRequest_Object{
					Object: setObj.Object,
				},
			}); err != nil {
				return err
			}

		case bytes.HasPrefix(scanner.Bytes(), relPrefix):
			setRel := &dsw3.SetRelationRequest{}
			if err := uOpts.Unmarshal(scanner.Bytes(), setRel); err != nil {
				return err
			}

			if err := stream.Send(&dsi3.ImportRequest{
				OpCode: dsi3.Opcode_OPCODE_SET,
				Msg: &dsi3.ImportRequest_Relation{
					Relation: setRel.Relation,
				},
			}); err != nil {
				return err
			}

		default:
			return errors.Errorf("unknown prefix")
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}

	return g.Wait()
}
