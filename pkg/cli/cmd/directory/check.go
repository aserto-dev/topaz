package directory

import (
	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
)

type CheckCmd struct {
	Request        string `arg:""  type:"string" name:"request" optional:"" help:"json request or file path to check permission request or '-' to read from stdin"`
	Template       bool   `name:"template" help:"prints a check permission request template on stdout"`
	clients.Config `envprefix:"TOPAZ_DIRECTORY_"`
}

func (cmd *CheckCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printCheckPermissionRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.CheckRequest
	err = UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.V3.Reader.Check(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "check permission call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printCheckPermissionRequest(ui *clui.UI) error {
	req := &reader.CheckRequest{
		ObjectType:  "",
		ObjectId:    "",
		Relation:    "",
		SubjectType: "",
		SubjectId:   "",
		Trace:       false,
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}
