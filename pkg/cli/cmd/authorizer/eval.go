package authorizer

import (
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type EvalCmd struct {
	clients.RequestArgs
	azc.Config
	req  authorizer.IsRequest
	resp authorizer.IsResponse
}

func (cmd *EvalCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(c.Context, authorizer.Authorizer_Is_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
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
			Name: "",
		},
		ResourceContext: &structpb.Struct{},
	}
}
