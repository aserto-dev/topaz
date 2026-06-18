package directory

import (
	"bytes"
	"context"
	"errors"
	"io"

	dsm "github.com/aserto-dev/go-directory/aserto/directory/model/v3"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (c *Client) GetManifest(ctx context.Context) (*bytes.Reader, error) {
	stream, err := c.Model.GetManifest(ctx, &dsm.GetManifestRequest{Empty: &emptypb.Empty{}})
	if err != nil {
		return nil, err
	}

	data := bytes.Buffer{}

	bytesRecv := 0

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		if md, ok := resp.GetMsg().(*dsm.GetManifestResponse_Metadata); ok {
			_ = md.Metadata
		}

		if body, ok := resp.GetMsg().(*dsm.GetManifestResponse_Body); ok {
			data.Write(body.Body.GetData())
			bytesRecv += len(body.Body.GetData())
		}
	}

	return bytes.NewReader(data.Bytes()), nil
}

const blockSize = 1024 * 64

func (c *Client) SetManifest(ctx context.Context, r io.Reader) error {
	stream, err := c.Model.SetManifest(ctx)
	if err != nil {
		return err
	}

	buf := make([]byte, blockSize)

	for {
		n, err := r.Read(buf)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		if err := stream.Send(&dsm.SetManifestRequest{
			Msg: &dsm.SetManifestRequest_Body{
				Body: &dsm.Body{Data: buf[0:n]},
			},
		}); err != nil {
			return err
		}

		if n < blockSize {
			break
		}
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteManifest(ctx context.Context) error {
	_, err := c.Model.DeleteManifest(ctx, &dsm.DeleteManifestRequest{Empty: &emptypb.Empty{}})
	return err
}
