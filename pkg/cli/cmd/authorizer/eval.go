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

type EvalCmd struct {
	com.RequestArgs
	azc.Config
}

func (cmd *EvalCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *EvalCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := azc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return com.Invoke[authorizer.IsRequest](
		c,
		client.IAuthorizer(),
		authorizer.Authorizer_Is_FullMethodName,
		request,
	)
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
