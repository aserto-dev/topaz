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
	"google.golang.org/protobuf/types/known/structpb"
)

type EvalCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to eval policy request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a check permission request template on stdout"`
	clients.AuthorizerConfig
}

func (cmd *EvalCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return cmd.print(c.UI.Output())
	}
	client, err := clients.NewAuthorizerClient(c, &cmd.AuthorizerConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get authorizer client")
	}

	var req authorizer.IsRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.Is(c.Context, &req)
	if err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func (cmd *EvalCmd) template() proto.Message {
	return &authorizer.IsRequest{
		PolicyContext: &api.PolicyContext{
			Path:      "",
			Decisions: []string{"allowed"},
		},
		IdentityContext: &api.IdentityContext{
			Identity: "",
			Type:     api.IdentityType_IDENTITY_TYPE_NONE,
		},
		PolicyInstance: &api.PolicyInstance{
			Name:          "",
			InstanceLabel: "",
		},
		ResourceContext: &structpb.Struct{},
	}
}

func (cmd *EvalCmd) print(w io.Writer) error {
	return jsonx.OutputJSONPB(w, cmd.template())
}
