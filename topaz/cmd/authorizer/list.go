package authorizer

import (
	"context"
	"os"
	"strings"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/topaz/clients"
	azc "github.com/aserto-dev/topaz/topaz/clients/authorizer"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"github.com/aserto-dev/topaz/topaz/table"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type ListPoliciesCmd struct {
	clients.RequestArgs
	azc.Config

	Raw  bool `name:"raw" help:"return raw request output"`
	req  authorizer.ListPoliciesRequest
	resp authorizer.ListPoliciesResponse
}

func (cmd *ListPoliciesCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, authorizer.Authorizer_ListPolicies_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	if cmd.Raw {
		return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
	}

	tab := table.New(os.Stdout)
	defer tab.Close()

	tab.Header("package path", "id")

	data := [][]any{}

	for _, module := range cmd.resp.GetResult() {
		data = append(data, []any{strings.TrimPrefix(module.GetPackagePath(), "data."), module.GetId()})
	}

	tab.Bulk(data)
	tab.Render()

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
