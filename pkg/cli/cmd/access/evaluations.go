package access

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/edit"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/pb"
	dsa1 "github.com/authzen/access.go/api/access/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type EvaluationsCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to get relation request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a get relation request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit request" hidden:"" type:"fflag.Editor"`
	dsc.Config
}

func (cmd *EvaluationsCmd) Run(c *cc.CommonCtx) error {
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
		return status.Error(codes.Unavailable, "prompter is unavailable for access evaluations command")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req dsa1.EvaluationRequest
	if err := pb.UnmarshalRequest(cmd.Request, &req); err != nil {
		return err
	}

	resp, err := client.Access.Evaluation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "evaluations call failed")
	}

	return jsonx.OutputJSONPB(c.StdOut(), resp)
}

func (cmd *EvaluationsCmd) template() *dsa1.EvaluationsRequest {
	return &dsa1.EvaluationsRequest{
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
		Context:     &structpb.Struct{},
		Evaluations: []*dsa1.EvaluationRequest{},
		Options:     &structpb.Struct{},
	}
}
