package directory

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsi "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	"github.com/aserto-dev/topaz/topaz/js"

	"golang.org/x/sync/errgroup"
)

func (c *Client) Restore(ctx context.Context, file string) error {
	tf, err := os.Open(file)
	if err != nil {
		return err
	}
	defer tf.Close()

	gz, err := gzip.NewReader(tf)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	g, iCtx := errgroup.WithContext(context.Background())

	stream, err := c.Importer.Import(iCtx)
	if err != nil {
		return err
	}

	g.Go(c.receiver(stream))

	g.Go(c.restoreHandler(stream, tr))

	return g.Wait()
}

func (c *Client) receiver(stream dsi.Importer_ImportClient) func() error {
	return func() error {
		objCounter := &dsi.ImportCounter{Type: objectsCounter}
		relCounter := &dsi.ImportCounter{Type: relationsCounter}

		defer func() {
			printCounter(os.Stderr, objCounter)
			printCounter(os.Stderr, relCounter)
		}()

		for {
			msg, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}

			if err != nil {
				return err
			}

			switch m := msg.GetMsg().(type) {
			case *dsi.ImportResponse_Status:
				printStatus(os.Stderr, m.Status)

			case *dsi.ImportResponse_Counter:
				switch m.Counter.GetType() {
				case objectsCounter:
					objCounter = m.Counter
				case relationsCounter:
					relCounter = m.Counter
				}

			// handle obsolete message usage as the default.
			//nolint:staticcheck // SA1019
			default:
				msg.Object.Type = objectsCounter
				objCounter = msg.GetObject()
				msg.Relation.Type = relationsCounter
				relCounter = msg.GetRelation()
			}
		}
	}
}

func (c *Client) restoreHandler(stream dsi.Importer_ImportClient, tr *tar.Reader) func() error {
	return func() error {
		for {
			header, err := tr.Next()
			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				return err
			}

			if header == nil || header.Typeflag != tar.TypeReg {
				continue
			}

			r, err := js.NewReader(tr)
			if err != nil {
				return err
			}

			name := path.Clean(header.Name)
			switch name {
			case ObjectsFileName:
				if err := c.loadObjects(stream, r); err != nil {
					return err
				}

			case RelationsFileName:
				if err := c.loadRelations(stream, r); err != nil {
					return err
				}
			}
		}

		if err := stream.CloseSend(); err != nil {
			return err
		}

		return nil
	}
}

func (c *Client) loadObjects(stream dsi.Importer_ImportClient, objects *js.Reader) error {
	defer objects.Close()

	var m dsc.Object

	for {
		err := objects.Read(&m)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			if strings.Contains(err.Error(), "unknown field") {
				continue
			}

			return err
		}

		if err := stream.Send(&dsi.ImportRequest{
			OpCode: dsi.Opcode_OPCODE_SET,
			Msg: &dsi.ImportRequest_Object{
				Object: &m,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) loadRelations(stream dsi.Importer_ImportClient, relations *js.Reader) error {
	defer relations.Close()

	var m dsc.Relation

	for {
		err := relations.Read(&m)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			if strings.Contains(err.Error(), "unknown field") {
				continue
			}

			return err
		}

		if err := stream.Send(&dsi.ImportRequest{
			OpCode: dsi.Opcode_OPCODE_SET,
			Msg: &dsi.ImportRequest_Relation{
				Relation: &m,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
