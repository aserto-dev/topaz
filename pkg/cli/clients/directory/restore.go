package directory

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/js"
	"google.golang.org/grpc/codes"

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

	ctr := &Counter{}
	defer ctr.Print(os.Stdout)

	g, iCtx := errgroup.WithContext(context.Background())
	stream, err := c.Importer.Import(iCtx)
	if err != nil {
		return err
	}

	g.Go(c.receiver(stream))

	g.Go(c.restoreHandler(stream, tr, ctr))

	return g.Wait()
}

func (c *Client) receiver(stream dsi3.Importer_ImportClient) func() error {
	return func() error {
		objCounter := &dsi3.ImportCounter{Type: "object"}
		relCounter := &dsi3.ImportCounter{Type: "relation"}

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
				case "object":
					objCounter = m.Counter
				case "relation":
					relCounter = m.Counter
				}

			default:
				msg.Object.Type = "object"
				objCounter = msg.Object
				msg.Relation.Type = "relation"
				relCounter = msg.Relation
			}
		}
	}
}

func printStatus(w io.Writer, status *dsi3.ImportStatus) {
	fmt.Fprintf(w, "%-8s: %s (%d) - %s\n",
		"error",
		codes.Code(status.Code).String(),
		status.Code,
		status.Msg)
}

func printCounter(w io.Writer, ctr *dsi3.ImportCounter) {
	fmt.Fprintf(w, "%-8s: recv:%d set:%d delete:%d error:%d\n",
		ctr.Type,
		ctr.Recv,
		ctr.Set,
		ctr.Delete,
		ctr.Error,
	)
}

func (c *Client) restoreHandler(stream dsi3.Importer_ImportClient, tr *tar.Reader, ctr *Counter) func() error {
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

func (c *Client) loadObjects(stream dsi3.Importer_ImportClient, objects *js.Reader) error {
	defer objects.Close()

	var m dsc3.Object

	for {
		err := objects.Read(&m)
		if err == io.EOF {
			break
		}

		if err != nil {
			if strings.Contains(err.Error(), "unknown field") {
				continue
			}
			return err
		}

		if err := stream.Send(&dsi3.ImportRequest{
			OpCode: dsi3.Opcode_OPCODE_SET,
			Msg: &dsi3.ImportRequest_Object{
				Object: &m,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) loadRelations(stream dsi3.Importer_ImportClient, relations *js.Reader) error {
	defer relations.Close()

	var m dsc3.Relation

	for {
		err := relations.Read(&m)
		if err == io.EOF {
			break
		}
		if err != nil {
			if strings.Contains(err.Error(), "unknown field") {
				continue
			}
			return err
		}

		if err := stream.Send(&dsi3.ImportRequest{
			OpCode: dsi3.Opcode_OPCODE_SET,
			Msg: &dsi3.ImportRequest_Relation{
				Relation: &m,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
