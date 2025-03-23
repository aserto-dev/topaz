package access

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/edit"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/pb"
	"github.com/aserto-dev/topaz/pkg/cli/prompter"
	dsa1 "github.com/authzen/access.go/api/access/v1"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type ResourceSearchCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to get relation request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a get relation request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit request" hidden:"" type:"fflag.Editor"`
	dsc.Config
}

func (cmd *ResourceSearchCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" && cmd.Editor && fflag.Enabled(fflag.Editor) {
		req, err := edit.Msg(cmd.template())
		if err != nil {
			return err
		}
		cmd.Request = req
	}

	if cmd.Request == "" && fflag.Enabled(fflag.Prompter) {
		p := prompter.New(cmd.template())
		if err := p.Show(); err != nil {
			return err
		}
		cmd.Request = jsonx.MaskedMarshalOpts().Format(p.Req())
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req dsa1.ResourceSearchRequest
	err = pb.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.Access.ResourceSearch(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "resource search call failed")
	}

	return jsonx.OutputJSONPB(c.StdOut(), resp)
}

func (cmd *ResourceSearchCmd) template() proto.Message {
	return &dsa1.ResourceSearchRequest{
		Subject: &dsa1.Subject{
			Type:       "",
			Id:         "",
			Properties: &structpb.Struct{},
		},
		Action: &dsa1.Action{
			Name:       "",
			Properties: &structpb.Struct{},
		},
		Resource: &dsa1.Resource{
			Type:       "",
			Id:         "",
			Properties: &structpb.Struct{},
		},
		Context: &structpb.Struct{},
		Page: &dsa1.Page{
			NextToken: "",
		},
	}
}
