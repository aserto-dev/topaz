package directory

import (
	"bytes"
	"context"

	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
)

func (c *Client) GetModel(ctx context.Context) (*bytes.Reader, error) {
	model, err := c.Reader.GetModel(ctx, &dsr.GetModelRequest{})
	if err != nil {
		return nil, err
	}

	buf, err := model.GetModel().GetContent().MarshalJSON()
	if err != nil {
		return nil, err
	}

	modelReader := bytes.NewReader(buf)

	return modelReader, nil
}
