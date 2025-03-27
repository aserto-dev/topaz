package authorizer

import (
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/edit"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/pb"
	"github.com/aserto-dev/topaz/pkg/cli/prompter"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

type EvalCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to eval policy request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a check permission request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit request" hidden:"" type:"fflag.Editor"`
	azc.Config
}

func (cmd *EvalCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	azClient, err := azc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get authorizer client")
	}

	if cmd.Request == "" && cmd.Editor && fflag.Enabled(fflag.Editor) {
		req, err := edit.Msg(cmd.template())
		if err != nil {
			return err
		}

		cmd.Request = req
	}

	if cmd.Request == "" && fflag.Enabled(fflag.Prompter) {
		p := prompter.New(cmd.template())
		if err := p.Show(); err != nil {
			return err
		}

		cmd.Request = jsonx.MaskedMarshalOpts().Format(p.Req())
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req authorizer.IsRequest
	if err := pb.UnmarshalRequest(cmd.Request, &req); err != nil {
		return err
	}

	resp, err := azClient.Authorizer.Is(c.Context, &req)
	if err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), resp)
}

func (cmd *EvalCmd) template() *authorizer.IsRequest {
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
			Name: "",
		},
		ResourceContext: &structpb.Struct{},
	}
}
