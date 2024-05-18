package directory

import (
	"io"

	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetObjectCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to get object request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a get object request template on stdout"`
	clients.DirectoryConfig
}

func (cmd *GetObjectCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return cmd.print(c.UI.Output())
	}

	client, err := clients.NewDirectoryClient(c, &cmd.DirectoryConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.GetObjectRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.V3.Reader.GetObject(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "get object call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func (cmd *GetObjectCmd) template() proto.Message {
	return &reader.GetObjectRequest{
		ObjectType:    "",
		ObjectId:      "",
		WithRelations: false,
		Page:          &common.PaginationRequest{Size: 100, Token: ""},
	}
}

func (cmd *GetObjectCmd) print(w io.Writer) error {
	return jsonx.OutputJSONPB(w, cmd.template())
}

type SetObjectCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"file path to set object request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a set object request template on stdout"`
	clients.DirectoryConfig
}

func (cmd *SetObjectCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return cmd.print(c.UI.Output())
	}

	client, err := clients.NewDirectoryClient(c, &cmd.DirectoryConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req writer.SetObjectRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.V3.Writer.SetObject(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "failed to set object")
	}
	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func (cmd *SetObjectCmd) template() proto.Message {
	properties := map[string]interface{}{}
	props, _ := structpb.NewStruct(properties)
	return &writer.SetObjectRequest{
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
}

func (cmd *SetObjectCmd) print(w io.Writer) error {
	return jsonx.OutputJSONPB(w, cmd.template())
}

type DeleteObjectCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"file path to delete object request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a delete object request template on stdout"`
	clients.DirectoryConfig
}

func (cmd *DeleteObjectCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return cmd.print(c.UI.Output())
	}

	client, err := clients.NewDirectoryClient(c, &cmd.DirectoryConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req writer.DeleteObjectRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.V3.Writer.DeleteObject(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "delete object call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func (cmd *DeleteObjectCmd) template() proto.Message {
	return &writer.DeleteObjectRequest{
		ObjectType:    "",
		ObjectId:      "",
		WithRelations: false,
	}
}

func (cmd *DeleteObjectCmd) print(w io.Writer) error {
	return jsonx.OutputJSONPB(w, cmd.template())
}

type ListObjectsCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"file path to list objects request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a list objects request template on stdout"`
	clients.DirectoryConfig
}

func (cmd *ListObjectsCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return cmd.print(c.UI.Output())
	}

	client, err := clients.NewDirectoryClient(c, &cmd.DirectoryConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.GetObjectsRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.V3.Reader.GetObjects(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "get objects call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func (cmd *ListObjectsCmd) template() proto.Message {
	return &reader.GetObjectsRequest{
		ObjectType: "",
		Page:       &common.PaginationRequest{Size: 100, Token: ""},
	}
}

func (cmd *ListObjectsCmd) print(w io.Writer) error {
	return jsonx.OutputJSONPB(w, cmd.template())
}
