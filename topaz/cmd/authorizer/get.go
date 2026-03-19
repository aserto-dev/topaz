package authorizer

import (
	"context"
	"os"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/clients"
	azc "github.com/aserto-dev/topaz/topaz/clients/authorizer"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type GetPolicyCmd struct {
	clients.RequestArgs
	azc.Config

	Raw  bool `name:"raw" help:"return raw request output"`
	req  authorizer.GetPolicyRequest
	resp authorizer.GetPolicyResponse
}

func (cmd *GetPolicyCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, authorizer.Authorizer_GetPolicy_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	if cmd.Raw {
		return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
	}

	cc.Out().Msg(cmd.resp.GetResult().GetRaw())

	return nil
}

func (cmd *GetPolicyCmd) template() proto.Message {
	return &authorizer.GetPolicyRequest{
		Id: "",
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{},
		},
		PolicyInstance: &api.PolicyInstance{
			Name:          "",
			InstanceLabel: "",
		},
	}
}
