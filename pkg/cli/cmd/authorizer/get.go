package authorizer

import (
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	com "github.com/aserto-dev/topaz/pkg/cli/cmd/common"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type GetPolicyCmd struct {
	com.RequestArgs
	Raw bool `name:"raw" help:"return raw request output"`
	azc.Config
}

func (cmd *GetPolicyCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := azc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return com.Invoke[authorizer.GetPolicyRequest](
		c,
		client.IAuthorizer(),
		authorizer.Authorizer_GetPolicy_FullMethodName,
		request,
	)
}

func (cmd *GetPolicyCmd) template() proto.Message {
	return &authorizer.GetPolicyRequest{
		Id: "",
		FieldMask: &fieldmaskpb.FieldMask{
			Paths: []string{},
		},
		PolicyInstance: &api.PolicyInstance{
			Name:          "",
			InstanceLabel: "",
		},
	}
}
