package authorizer

import (
	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

type DecisionTreeCmd struct {
	Request  string `arg:""  type:"string" name:"request" optional:"" help:"json request or file path to decision tree request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a check permission request template on stdout"`
	clients.AuthorizerConfig
}

func (cmd *DecisionTreeCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printDecisionTreeRequest(c.UI)
	}
	client, err := clients.NewAuthorizerClient(c, &cmd.AuthorizerConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get authorizer client")
	}

	var req authorizer.DecisionTreeRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.DecisionTree(c.Context, &req)
	if err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printDecisionTreeRequest(ui *clui.UI) error {
	req := &authorizer.DecisionTreeRequest{
		PolicyContext: &api.PolicyContext{
			Path:      "",
			Decisions: []string{"allowed"},
		},
		IdentityContext: &api.IdentityContext{
			Identity: "",
			Type:     api.IdentityType_IDENTITY_TYPE_NONE,
		},
		ResourceContext: &structpb.Struct{},
		Options: &authorizer.DecisionTreeOptions{
			PathSeparator: authorizer.PathSeparator_PATH_SEPARATOR_DOT,
		},
		PolicyInstance: &api.PolicyInstance{
			Name:          "",
			InstanceLabel: "",
		},
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}
