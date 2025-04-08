package directory

import (
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"google.golang.org/protobuf/proto"
)

type ChecksCmd struct {
	clients.RequestArgs
	dsc.Config
	req  reader.ChecksRequest
	resp reader.ChecksResponse
}

func (cmd *ChecksCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(c.Context, reader.Reader_Checks_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
}

func (cmd *ChecksCmd) template() proto.Message {
	return &reader.ChecksRequest{
		Default: &reader.CheckRequest{
			ObjectType:  "",
			ObjectId:    "",
			Relation:    "",
			SubjectType: "",
			SubjectId:   "",
			Trace:       false,
		},
		Checks: []*reader.CheckRequest{
			{
				ObjectType:  "",
				ObjectId:    "",
				Relation:    "",
				SubjectType: "",
				SubjectId:   "",
				Trace:       false,
			},
		},
	}
}
