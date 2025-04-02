package access

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	dsa1 "github.com/authzen/access.go/api/access/v1"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type ActionSearchCmd struct {
	clients.RequestArgs
	dsc.Config
	req  dsa1.ActionSearchRequest
	resp dsa1.ActionSearchResponse
}

func (cmd *ActionSearchCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.RequestArgs.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Config.Invoke(c.Context, dsa1.Access_ActionSearch_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
}

func (cmd *ActionSearchCmd) template() proto.Message {
	return &dsa1.ActionSearchRequest{
		Subject: &dsa1.Subject{
			Type:       "",
			Id:         "",
			Properties: &structpb.Struct{},
		},
		Action: &dsa1.Action{
			Name:       "",
			Properties: &structpb.Struct{},
		},
		Resource: &dsa1.Resource{
			Type:       "",
			Id:         "",
			Properties: &structpb.Struct{},
		},
		Context: &structpb.Struct{},
		Page: &dsa1.Page{
			NextToken: "",
		},
	}
}
