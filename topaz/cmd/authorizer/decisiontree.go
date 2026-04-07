package authorizer

import (
	"context"
	"os"

	"github.com/aserto-dev/topaz/topaz/clients"
	azc "github.com/aserto-dev/topaz/topaz/clients/authorizer"
	"github.com/aserto-dev/topaz/topaz/jsonx"
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

func (cmd *DecisionTreeCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, authorizer.Authorizer_DecisionTree_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
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
		PolicyInstance: &api.PolicyInstance{ //nolint:staticcheck
			Name: "",
		},
	}
}
