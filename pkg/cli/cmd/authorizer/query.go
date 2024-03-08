package authorizer

import (
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
)

type ExecQueryCmd struct {
	AuthParams     `embed:""`
	Statement      string `arg:"stmt" name:"stmt" required:"" help:"query statement"`
	Path           string `name:"path" help:"policy package to evaluate"`
	Input          string `name:"input" help:"query input context"`
	clients.Config `envprefix:"TOPAZ_AUTHORIZER_"`
}

func (cmd *ExecQueryCmd) Run(c *cc.CommonCtx) error {
	client, err := clients.NewAuthorizerClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get authorizer client")
	}

	resource, err := cmd.ResourceContext()
	if err != nil {
		return err
	}

	resp, err := client.Query(c.Context, &authorizer.QueryRequest{
		Query:           cmd.Statement,
		Input:           cmd.Input,
		IdentityContext: cmd.IdentityContext(),
		PolicyContext: &api.PolicyContext{
			Path: cmd.Path,
		},
		ResourceContext: resource,
		Options: &authorizer.QueryOptions{
			Metrics:      false,
			Instrument:   false,
			Trace:        authorizer.TraceLevel_TRACE_LEVEL_OFF,
			TraceSummary: false,
		},
	})
	if err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}
