package authorizer

import (
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
)

type ListPoliciesCmd struct {
	PolicyName    string `name:"policy-name" default:"" required:"false" help:"policy name"`
	InstanceLabel string `name:"instance-label" default:"" required:"false" help:"policy's instance label"`
	clients.AuthorizerConfig
}

func (cmd *ListPoliciesCmd) Run(c *cc.CommonCtx) error {
	client, err := clients.NewAuthorizerClient(c, &cmd.AuthorizerConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get authorizer client")
	}

	resp, err := client.ListPolicies(c.Context, &authorizer.ListPoliciesRequest{
		PolicyInstance: &api.PolicyInstance{Name: cmd.PolicyName, InstanceLabel: cmd.InstanceLabel},
	})
	if err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}
