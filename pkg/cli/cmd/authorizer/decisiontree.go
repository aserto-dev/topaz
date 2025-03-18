package authorizer

import (
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	com "github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type DecisionTreeCmd struct {
	com.RequestArgs
	azc.Config
}

func (cmd *DecisionTreeCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *DecisionTreeCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := azc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return com.Invoke[authorizer.DecisionTreeRequest](
		c,
		client.IAuthorizer(),
		authorizer.Authorizer_DecisionTree_FullMethodName,
		request,
	)
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
