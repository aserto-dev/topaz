package authorizer

import (
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
)

type EvalCmd struct {
	AuthParams     `embed:""`
	Path           string   `name:"path" required:"" help:"policy package to evaluate"`
	Decisions      []string `name:"decisions" required:"" help:"policy decisions to return"`
	clients.Config `envprefix:"TOPAZ_AUTHORIZER_"`
}

func (cmd *EvalCmd) Run(c *cc.CommonCtx) error {
	client, err := clients.NewAuthorizerClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get authorizer client")
	}

	resource, err := cmd.ResourceContext()
	if err != nil {
		return err
	}

	resp, err := client.Is(c.Context, &authorizer.IsRequest{
		PolicyContext: &api.PolicyContext{
			Path:      cmd.Path,
			Decisions: cmd.Decisions,
		},
		IdentityContext: cmd.IdentityContext(),
		ResourceContext: resource,
	})
	if err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}
