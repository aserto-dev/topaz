package directory

import (
	"context"
	"os"

	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"google.golang.org/protobuf/proto"
)

type SearchCmd struct {
	clients.RequestArgs
	dsc.Config

	req  reader.GetGraphRequest
	resp reader.GetGraphResponse
}

func (cmd *SearchCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, reader.Reader_GetGraph_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
}

func (cmd *SearchCmd) template() proto.Message {
	return &reader.GetGraphRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
		Explain:         false,
		Trace:           false,
	}
}
