package authorizer

import (
	"github.com/aserto-dev/topaz/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/clients"
	azc "github.com/aserto-dev/topaz/topaz/pkg/cli/clients/authorizer"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/jsonx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

type DecisionTreeCmd struct {
	clients.RequestArgs
	azc.Config

	req  authorizer.DecisionTreeRequest
	resp authorizer.DecisionTreeResponse
}

func (cmd *DecisionTreeCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(c.Context, authorizer.Authorizer_DecisionTree_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
}

func (cmd *DecisionTreeCmd) template() proto.Message {
	return &authorizer.DecisionTreeRequest{
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
			Name: "",
		},
	}
}
