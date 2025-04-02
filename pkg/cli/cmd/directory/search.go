package directory

import (
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"google.golang.org/protobuf/proto"
)

type SearchCmd struct {
	clients.RequestArgs
	dsc.Config
	req  reader.GetGraphRequest
	resp reader.GetGraphResponse
}

func (cmd *SearchCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.RequestArgs.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Config.Invoke(c.Context, reader.Reader_GetGraph_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
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
