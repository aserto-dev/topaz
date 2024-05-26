package directory

import (
	"bytes"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/editor"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/prompter"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetObjectCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to get object request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a get object request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit config file" hidden:"" type:"fflag.Editor"`
	clients.DirectoryConfig
}

func (cmd *GetObjectCmd) BeforeReset(ctx *kong.Context) error {
	n := ctx.Selected()
	if n != nil {
		fflag.UnHideFlags(ctx)
	}
	return nil
}

func (cmd *GetObjectCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return cmd.print(c.UI.Output())
	}

	if cmd.Request == "" && cmd.Editor && fflag.Enabled(fflag.Editor) {
		req, err := cmd.edit(cmd.template())
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

	var req reader.GetObjectRequest
	if err := clients.UnmarshalRequest(cmd.Request, &req); err != nil {
		return err
	}

	client, err := clients.NewDirectoryClient(c, &cmd.DirectoryConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
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

func (cmd *GetObjectCmd) edit(tmpl proto.Message) (string, error) {
	tmp, err := jsonx.MarshalOpts(true).Marshal(tmpl)
	if err != nil {
		return "", err
	}

	e := editor.NewDefaultEditor([]string{"TOPAZ_EDITOR"})
	name := string(proto.MessageName(cmd.template()).Name())

	buf, path, err := e.LaunchTempFile("topaz", name, bytes.NewReader(tmp))
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(path) }()

	return string(buf), nil
}

// func (cmd *GetObjectCmd) prompt(c *cc.CommonCtx) (string, error) {
// 	request, err := input.Prompt(cmd.template())
// 	if err != nil {
// 		return "", err
// 	}

// 	return request, nil
// }

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
