package authorizer

import (
	"io"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type GetPolicyCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to get policy request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a check permission request template on stdout"`
	clients.AuthorizerConfig
}

func (cmd *GetPolicyCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return cmd.print(c.UI.Output())
	}

	client, err := clients.NewAuthorizerClient(c, &cmd.AuthorizerConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get authorizer client")
	}
	var req authorizer.GetPolicyRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.GetPolicy(c.Context, &req)
	if err != nil {
		return err
	}
	return jsonx.OutputJSONPB(c.UI.Output(), resp)
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

func (cmd *GetPolicyCmd) print(w io.Writer) error {
	return jsonx.OutputJSONPB(w, cmd.template())
}
