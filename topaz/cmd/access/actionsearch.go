package access

import (
	"context"
	"os"

	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	dsa "github.com/authzen/access.go/api/access/v1"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type ActionSearchCmd struct {
	clients.RequestArgs
	dsc.Config

	req  dsa.ActionSearchRequest
	resp dsa.ActionSearchResponse
}

func (cmd *ActionSearchCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, dsa.Access_ActionSearch_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
}

func (cmd *ActionSearchCmd) template() proto.Message {
	return &dsa.ActionSearchRequest{
		Subject: &dsa.Subject{
			Type:       "",
			Id:         "",
			Properties: &structpb.Struct{},
		},
		Action: &dsa.Action{
			Name:       "",
			Properties: &structpb.Struct{},
		},
		Resource: &dsa.Resource{
			Type:       "",
			Id:         "",
			Properties: &structpb.Struct{},
		},
		Context: &structpb.Struct{},
		Page: &dsa.PaginationRequest{
			Token:      nil,
			Limit:      nil,
			Properties: nil,
		},
	}
}
