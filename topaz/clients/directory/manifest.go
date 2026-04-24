package directory

import (
	"bytes"
	"context"
	"io"

	dsc3 "github.com/aserto-dev/topaz/api/directory/v4"
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
	dsw "github.com/aserto-dev/topaz/api/directory/v4/writer"
)

func (c *Client) GetManifest(ctx context.Context) (*bytes.Reader, error) {
	resp, err := c.Reader.GetManifest(ctx, &dsr.GetManifestRequest{})
	if err != nil {
		return nil, err
	}

	// stream, err := c.Model.GetManifest(ctx, &dsm3.GetManifestRequest{Empty: &emptypb.Empty{}})
	// if err != nil {
	// 	return nil, err
	// }
	// resp.Manifest.Content
	// data := bytes.Buffer{}

	// bytesRecv := 0

	// for {
	// 	resp, err := stream.Recv()
	// 	if errors.Is(err, io.EOF) {
	// 		break
	// 	}

	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	if md, ok := resp.GetMsg().(*dsm3.GetManifestResponse_Metadata); ok {
	// 		_ = md.Metadata
	// 	}

	// 	if body, ok := resp.GetMsg().(*dsm3.GetManifestResponse_Body); ok {
	// 		data.Write(body.Body.GetData())
	// 		bytesRecv += len(body.Body.GetData())
	// 	}
	// }

	return bytes.NewReader(resp.GetManifest().GetContent()), nil
}

const blockSize = 1024 * 64

func (c *Client) SetManifest(ctx context.Context, r io.Reader) error {
	resp, err := c.Writer.SetManifest(ctx, &dsw.SetManifestRequest{
		Manifest: &dsc3.Manifest{
			Content: []byte{},
		},
	})
	if err != nil {
		return err
	}

	_ = resp
	// stream, err := c.Model.SetManifest(ctx)
	// if err != nil {
	// 	return err
	// }

	// buf := make([]byte, blockSize)

	// for {
	// 	n, err := r.Read(buf)
	// 	if errors.Is(err, io.EOF) {
	// 		break
	// 	}

	// 	if err != nil {
	// 		return err
	// 	}

	// 	if err := stream.Send(&dsm3.SetManifestRequest{
	// 		Msg: &dsm3.SetManifestRequest_Body{
	// 			Body: &dsm3.Body{Data: buf[0:n]},
	// 		},
	// 	}); err != nil {
	// 		return err
	// 	}

	// 	if n < blockSize {
	// 		break
	// 	}
	// }

	// if _, err := stream.CloseAndRecv(); err != nil {
	// 	return err
	// }

	return nil
}

func (c *Client) DeleteManifest(ctx context.Context) error {
	_, err := c.Writer.DeleteManifest(ctx, &dsw.DeleteManifestRequest{})
	return err
}
