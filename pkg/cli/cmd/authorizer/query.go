package authorizer

import (
	"github.com/alecthomas/kong"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	com "github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type QueryCmd struct {
	com.RequestArgs
	azc.Config
}

func (cmd *QueryCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *QueryCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := azc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return com.Invoke[authorizer.QueryRequest](
		c,
		client.IAuthorizer(),
		authorizer.Authorizer_Query_FullMethodName,
		request,
	)
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
		PolicyInstance: &api.PolicyInstance{
			Name: "",
		},
	}
}
