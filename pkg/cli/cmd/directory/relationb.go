package directory

import (
	"encoding/json"
	"os"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetRelationCmd struct {
	Request  string `arg:""  type:"string" name:"request" optional:"" help:"json request or file path to get relation request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a get relation request template on stdout"`
	clients.Config
}

func (cmd *GetRelationCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printGetRelationRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.GetRelationRequest
	if cmd.Request == "-" {
		decoder := json.NewDecoder(os.Stdin)

		err = decoder.Decode(&req)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal request from stdin")
		}
	} else {
		err = protojson.Unmarshal([]byte(cmd.Request), &req)
		if err != nil {
			dat, err := os.ReadFile(cmd.Request)
			if err != nil {
				return errors.Wrapf(err, "opening file [%s]", cmd.Request)
			}

			err = protojson.Unmarshal(dat, &req)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal request from file [%s]", cmd.Request)
			}
		}
	}

	resp, err := client.V3.Reader.GetRelation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "get relation call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printGetRelationRequest(ui *clui.UI) error {
	req := &reader.GetRelationRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
		WithObjects:     true,
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}

type SetRelationCmd struct {
	Request  string `arg:""  type:"existingfile" name:"request" optional:"" help:"file path to set relation request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a set relation request template on stdout"`
	clients.Config
}

func (cmd *SetRelationCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printSetRelationRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req writer.SetRelationRequest
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

	resp, err := client.V3.Writer.SetRelation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "set relation call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printSetRelationRequest(ui *clui.UI) error {
	req := &writer.SetRelationRequest{
		Relation: &common.Relation{
			ObjectType:      "",
			ObjectId:        "",
			Relation:        "",
			SubjectType:     "",
			SubjectId:       "",
			SubjectRelation: "",
			CreatedAt:       timestamppb.Now(),
			UpdatedAt:       timestamppb.Now(),
			Etag:            "",
		},
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}

type DeleteRelationCmd struct {
	Request  string `arg:""  type:"existingfile" name:"request" optional:"" help:"file path to delete relation request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a delete relation request template on stdout"`
	clients.Config
}

func (cmd *DeleteRelationCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printDeleteRelationRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req writer.DeleteRelationRequest
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

	resp, err := client.V3.Writer.DeleteRelation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "delete relation call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printDeleteRelationRequest(ui *clui.UI) error {
	req := &writer.DeleteRelationRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}

type ListRelationsCmd struct {
	Request  string `arg:""  type:"existingfile" name:"request" optional:"" help:"file path to list relations request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a list relations request template on stdout"`
	clients.Config
}

func (cmd *ListRelationsCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printListRelationsRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.GetRelationsRequest
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

	resp, err := client.V3.Reader.GetRelations(c.Context, &req)

	if err != nil {
		return errors.Wrap(err, "get relations call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printListRelationsRequest(ui *clui.UI) error {
	req := &reader.GetRelationsRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
		WithObjects:     true,
		Page:            &common.PaginationRequest{Size: 10, Token: ""},
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}
