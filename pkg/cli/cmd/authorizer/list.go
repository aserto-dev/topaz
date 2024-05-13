package authorizer

import (
	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type ListPoliciesCmd struct {
	Request  string `arg:""  type:"string" name:"request" optional:"" help:"json request or file path to list request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a check permission request template on stdout"`
	clients.AuthorizerConfig
}

func (cmd *ListPoliciesCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printListRequest(c.UI)
	}

	client, err := clients.NewAuthorizerClient(c, &cmd.AuthorizerConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get authorizer client")
	}
	var req authorizer.ListPoliciesRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.ListPolicies(c.Context, &req)
	if err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printListRequest(ui *clui.UI) error {
	req := &authorizer.ListPoliciesRequest{
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{},
		},
		PolicyInstance: &api.PolicyInstance{
			Name:          "",
			InstanceLabel: "",
		},
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}
