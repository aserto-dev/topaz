package authorizer

import (
	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

type QueryCmd struct {
	Request  string `arg:""  type:"string" name:"request" optional:"" help:"json request or file path to check permission request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a check permission request template on stdout"`
	clients.AuthorizerConfig
}

func (cmd *QueryCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printQueryRequest(c.UI)
	}

	client, err := clients.NewAuthorizerClient(c, &cmd.AuthorizerConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get authorizer client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req authorizer.QueryRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.Query(c.Context, &req)
	if err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printQueryRequest(ui *clui.UI) error {
	req := &authorizer.QueryRequest{
		Query: "",
		Input: "",
		Options: &authorizer.QueryOptions{
			Metrics:      false,
			Instrument:   false,
			Trace:        authorizer.TraceLevel_TRACE_LEVEL_FAILS,
			TraceSummary: false,
		},
		PolicyContext: &api.PolicyContext{
			Path:      "",
			Decisions: []string{"allowed"},
		},
		IdentityContext: &api.IdentityContext{
			Identity: "",
			Type:     api.IdentityType_IDENTITY_TYPE_NONE,
		},
		ResourceContext: &structpb.Struct{},
		PolicyInstance: &api.PolicyInstance{
			Name:          "",
			InstanceLabel: "",
		},
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}
