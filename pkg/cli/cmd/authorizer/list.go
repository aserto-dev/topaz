package authorizer

import (
	"strings"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/table"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type ListPoliciesCmd struct {
	clients.RequestArgs
	Raw bool `name:"raw" help:"return raw request output"`
	azc.Config
	req  authorizer.ListPoliciesRequest
	resp authorizer.ListPoliciesResponse
}

func (cmd *ListPoliciesCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.RequestArgs.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Config.Invoke(c.Context, authorizer.Authorizer_Is_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	if cmd.Raw {
		return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
	}

	tab := table.New(c.StdOut()).WithColumns("package path", "id")
	for _, module := range cmd.resp.GetResult() {
		tab.WithRow(strings.TrimPrefix(module.GetPackagePath(), "data."), module.GetId())
	}

	tab.Do()

	return nil
}

func (cmd *ListPoliciesCmd) template() proto.Message {
	return &authorizer.ListPoliciesRequest{
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{},
		},
		PolicyInstance: &api.PolicyInstance{
			Name: "",
		},
	}
}
