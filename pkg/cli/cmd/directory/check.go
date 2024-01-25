package directory

import (
	"encoding/json"
	"os"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

type CheckPermissionCmd struct {
	Request  string `arg:""  type:"existingfile" name:"request" optional:"" help:"file path to check permission request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a check permission request template on stdout"`
	clients.Config
}

func (cmd *CheckPermissionCmd) Run(c *cc.CommonCtx) error {
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

	var req reader.CheckPermissionRequest
	if cmd.Request == "-" {
		decoder := json.NewDecoder(os.Stdin)

		err = decoder.Decode(&req)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal request from stdin")
		}
	} else {
		dat, err := os.ReadFile(cmd.Request)
		if err != nil {
			return errors.Wrapf(err, "opening file [%s]", cmd.Request)
		}

		err = protojson.Unmarshal(dat, &req)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal request from file [%s]", cmd.Request)
		}
	}

	resp, err := client.V3.Reader.CheckPermission(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "check permission call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printCheckPermissionRequest(ui *clui.UI) error {
	req := &reader.CheckPermissionRequest{
		ObjectType:  "",
		ObjectId:    "",
		Permission:  "",
		SubjectType: "",
		SubjectId:   "",
		Trace:       false,
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}

type CheckRelationCmd struct {
	Request  string `arg:""  type:"existingfile" name:"request" optional:"" help:"file path to check relation request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a check relation request template on stdout"`
	clients.Config
}

func (cmd *CheckRelationCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printCheckRelationRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.CheckRelationRequest
	if cmd.Request == "-" {
		decoder := json.NewDecoder(os.Stdin)

		err = decoder.Decode(&req)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal request from stdin")
		}
	} else {
		dat, err := os.ReadFile(cmd.Request)
		if err != nil {
			return errors.Wrapf(err, "opening file [%s]", cmd.Request)
		}

		err = protojson.Unmarshal(dat, &req)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal request from file [%s]", cmd.Request)
		}
	}

	resp, err := client.V3.Reader.CheckRelation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "check relation call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printCheckRelationRequest(ui *clui.UI) error {
	req := &reader.CheckRelationRequest{
		ObjectType:  "",
		ObjectId:    "",
		Relation:    "",
		SubjectType: "",
		SubjectId:   "",
		Trace:       false,
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}
