package directory

import (
	"context"
	"os"

	"github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"google.golang.org/protobuf/proto"
)

type ChecksCmd struct {
	clients.RequestArgs
	dsc.Config

	req  reader.ChecksRequest
	resp reader.ChecksResponse
}

func (cmd *ChecksCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, reader.Reader_Checks_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
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
