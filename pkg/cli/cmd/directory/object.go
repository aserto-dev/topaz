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
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetObjectCmd struct {
	Request  string `arg:""  type:"string" name:"request" optional:"" help:"json request or file path to get object request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a get object request template on stdout"`
	clients.Config
}

func (cmd *GetObjectCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printGetObjectRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.GetObjectRequest
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

	resp, err := client.V3.Reader.GetObject(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "get object call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printGetObjectRequest(ui *clui.UI) error {
	req := &reader.GetObjectRequest{
		ObjectType:    "",
		ObjectId:      "",
		WithRelations: true,
		Page:          &common.PaginationRequest{Size: 10, Token: ""},
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}

type SetObjectCmd struct {
	Request  string `arg:""  type:"existingfile" name:"request" optional:"" help:"file path to set object request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a set object request template on stdout"`
	clients.Config
}

func (cmd *SetObjectCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printSetObjectRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req writer.SetObjectRequest
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

	resp, err := client.V3.Writer.SetObject(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "failed to set object")
	}
	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printSetObjectRequest(ui *clui.UI) error {
	properties := map[string]interface{}{"property1": 123, "property2": ""}
	props, _ := structpb.NewStruct(properties)
	req := &writer.SetObjectRequest{
		Object: &common.Object{
			Type:        "",
			Id:          "",
			DisplayName: "",
			Properties:  props,
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
			Etag:        "",
		},
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}

type DeleteObjectCmd struct {
	Request  string `arg:""  type:"existingfile" name:"request" optional:"" help:"file path to delete object request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a delete object request template on stdout"`
	clients.Config
}

func (cmd *DeleteObjectCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printDeleteObjectRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req writer.DeleteObjectRequest
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

	resp, err := client.V3.Writer.DeleteObject(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "delete object call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printDeleteObjectRequest(ui *clui.UI) error {
	req := &writer.DeleteObjectRequest{
		ObjectType:    "",
		ObjectId:      "",
		WithRelations: true,
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}

type ListObjectsCmd struct {
	Request  string `arg:""  type:"existingfile" name:"request" optional:"" help:"file path to list objects request or '-' to read from stdin"`
	Template bool   `name:"template" help:"prints a list objects request template on stdout"`
	clients.Config
}

func (cmd *ListObjectsCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return printListObjectsRequest(c.UI)
	}

	client, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.GetObjectsRequest
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

	resp, err := client.V3.Reader.GetObjects(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "get objects call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func printListObjectsRequest(ui *clui.UI) error {
	req := &reader.GetObjectsRequest{
		ObjectType: "",
		Page:       &common.PaginationRequest{Size: 10, Token: ""},
	}
	return jsonx.OutputJSONPB(ui.Output(), req)
}
