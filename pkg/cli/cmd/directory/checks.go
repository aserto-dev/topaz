package directory

import (
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/edit"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/pb"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//nolint:lll
type ChecksCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request, file path to check permission request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a check permission request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit request" hidden:"" type:"fflag.Editor"`
	dsc.Config
}

func (cmd *ChecksCmd) Run(c *cc.CommonCtx) error {
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
		return status.Error(codes.Unavailable, "prompter is unavailable for directory checks command")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.ChecksRequest
	err = pb.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.Reader.Checks(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "checks permission call failed")
	}

	return jsonx.OutputJSONPB(c.StdOut(), resp)
}

func (cmd *ChecksCmd) template() *reader.ChecksRequest {
	return &reader.ChecksRequest{
		Default: &reader.CheckRequest{
			ObjectType:  "",
			ObjectId:    "",
			Relation:    "",
			SubjectType: "",
			SubjectId:   "",
			Trace:       false,
		},
		Checks: []*reader.CheckRequest{
			{
				ObjectType:  "",
				ObjectId:    "",
				Relation:    "",
				SubjectType: "",
				SubjectId:   "",
				Trace:       false,
			},
		},
	}
}
