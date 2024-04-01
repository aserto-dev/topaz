package directory

import (
	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
)

type SearchCmd struct {
	Request        string `arg:""  type:"string" name:"request" optional:"" help:"json request or file path to get graph request or '-' to read from stdin"`
	Template       bool   `name:"template" help:"prints a get graph request template on stdout"`
	clients.Config `envprefix:"TOPAZ_DIRECTORY_"`
}

func (cmd *SearchCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printGetGraphRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.GetGraphRequest
	err = UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.V3.Reader.GetGraph(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "get graph call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printGetGraphRequest(ui *clui.UI) error {
	req := &reader.GetGraphRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
		Explain:         false,
		Trace:           false,
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}
