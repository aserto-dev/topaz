package authorizer

import (
	"cmp"
	"slices"
	"strings"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/edit"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/prompter"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type ListPoliciesCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to list request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a check permission request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit request" hidden:"" type:"fflag.Editor"`
	Raw      bool   `name:"raw" help:"return raw request output"`
	clients.AuthorizerConfig
}

func (cmd *ListPoliciesCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.UI.Output(), cmd.template())
	}

	client, err := clients.NewAuthorizerClient(c, &cmd.AuthorizerConfig)
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

	var req authorizer.ListPoliciesRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.ListPolicies(c.Context, &req)
	if err != nil {
		return err
	}

	slices.SortFunc(resp.Result, func(a, b *api.Module) int {
		return cmp.Compare(a.GetPackagePath()+a.GetId(), b.GetPackagePath()+b.GetId())
	})

	if cmd.Raw {
		return jsonx.OutputJSONPB(c.UI.Output(), resp)
	}

	table := c.UI.Normal().WithTable("package path", "id")
	for _, module := range resp.Result {
		table.WithTableRow(
			strings.TrimPrefix(module.GetPackagePath(), "data."), module.GetId())
	}
	table.Do()

	return nil
}

func (cmd *ListPoliciesCmd) template() proto.Message {
	return &authorizer.ListPoliciesRequest{
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{},
		},
		PolicyInstance: &api.PolicyInstance{
			Name:          "",
			InstanceLabel: "",
		},
	}
}
