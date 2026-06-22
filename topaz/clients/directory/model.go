package directory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	dsm "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (c *Client) GetModel(ctx context.Context) (*bytes.Reader, error) {
	md := metadata.Pairs(
		"Aserto-Manifest-Request", "model-only",
	)

	headerCtx := metadata.NewOutgoingContext(ctx, md)

	stream, err := c.Model.GetManifest(headerCtx, &dsm.GetManifestRequest{Empty: &emptypb.Empty{}})
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
			_ = body.Body.GetData()
		}

		if model, ok := resp.GetMsg().(*dsm.GetManifestResponse_Model); ok {
			buf, err := json.MarshalIndent(model.Model.AsMap(), "", "  ")
			if err != nil {
				return nil, err
			}

			data.Write(buf)
			bytesRecv += len(buf)
		}
	}

	return bytes.NewReader(data.Bytes()), nil
}
