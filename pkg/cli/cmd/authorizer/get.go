package authorizer

import (
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type GetPolicyCmd struct {
	clients.RequestArgs
	Raw bool `name:"raw" help:"return raw request output"`
	azc.Config
	req  authorizer.GetPolicyRequest
	resp authorizer.GetPolicyResponse
}

func (cmd *GetPolicyCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.RequestArgs.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Config.Invoke(c.Context, authorizer.Authorizer_GetPolicy_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	if cmd.Raw {
		return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
	}

	c.Out().Msg(cmd.resp.GetResult().GetRaw())

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
