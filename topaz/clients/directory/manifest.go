package directory

import (
	"bytes"
	"context"
	"io"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
)

func (c *Client) GetManifest(ctx context.Context) (*bytes.Reader, error) {
	manifest, err := c.Reader.GetManifest(ctx, &dsr.GetManifestRequest{})
	if err != nil {
		return nil, err
	}

	manifestReader := bytes.NewReader(manifest.GetManifest().GetContent())

	return manifestReader, nil
}

func (c *Client) SetManifest(ctx context.Context, r io.Reader) error {
	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}

	if _, err := c.Writer.SetManifest(ctx, &dsw.SetManifestRequest{
		Manifest: &dsc.Manifest{
			Content: buf.Bytes(),
		},
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteManifest(ctx context.Context) error {
	_, err := c.Writer.DeleteManifest(ctx, &dsw.DeleteManifestRequest{})
	return err
}
