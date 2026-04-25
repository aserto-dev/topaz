package authorizer

import (
	"context"
	"os"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/topaz/clients"
	azc "github.com/aserto-dev/topaz/topaz/clients/authorizer"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type QueryCmd struct {
	clients.RequestArgs
	azc.Config

	req  authorizer.QueryRequest
	resp authorizer.QueryResponse
}

func (cmd *QueryCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, authorizer.Authorizer_Query_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
}

func (cmd *QueryCmd) template() proto.Message {
	return &authorizer.QueryRequest{
		Query: "",
		Input: "",
		Options: &authorizer.QueryOptions{
			Metrics:      false,
			Instrument:   false,
			Trace:        authorizer.TraceLevel_TRACE_LEVEL_OFF,
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
	}
}
