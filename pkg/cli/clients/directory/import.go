package directory

import (
	"context"
	"fmt"
	"os"

	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/js"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func (c *Client) Import(ctx context.Context, files []string) error {
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
