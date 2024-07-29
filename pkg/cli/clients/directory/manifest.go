package directory

import (
	"bytes"
	"context"
	"io"

	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (c *Client) GetManifest(ctx context.Context) (io.Reader, error) {
	stream, err := c.Model.GetManifest(ctx, &dsm3.GetManifestRequest{Empty: &emptypb.Empty{}})
	if err != nil {
		return nil, err
	}

	data := bytes.Buffer{}

	bytesRecv := 0
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		if md, ok := resp.GetMsg().(*dsm3.GetManifestResponse_Metadata); ok {
			_ = md.Metadata
		}

		if body, ok := resp.GetMsg().(*dsm3.GetManifestResponse_Body); ok {
			data.Write(body.Body.Data)
			bytesRecv += len(body.Body.Data)
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
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := stream.Send(&dsm3.SetManifestRequest{
			Msg: &dsm3.SetManifestRequest_Body{
				Body: &dsm3.Body{Data: buf[0:n]},
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
	_, err := c.Model.DeleteManifest(ctx, &dsm3.DeleteManifestRequest{Empty: &emptypb.Empty{}})
	return err
}
