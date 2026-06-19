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

type SubjectSearchCmd struct {
	clients.RequestArgs
	dsc.Config

	req  dsa.SubjectSearchRequest
	resp dsa.SubjectSearchResponse
}

func (cmd *SubjectSearchCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, dsa.Access_SubjectSearch_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
}

func (cmd *SubjectSearchCmd) template() proto.Message {
	return &dsa.SubjectSearchRequest{
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
