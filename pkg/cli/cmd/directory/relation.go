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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetRelationCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to get relation request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a get relation request template on stdout"`
	clients.DirectoryConfig
}

func (cmd *GetRelationCmd) Run(c *cc.CommonCtx) error {
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

	var req reader.GetRelationRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.V3.Reader.GetRelation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "get relation call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func (cmd *GetRelationCmd) template() proto.Message {
	return &reader.GetRelationRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
		WithObjects:     true,
	}
}

func (cmd *GetRelationCmd) print(w io.Writer) error {
	return jsonx.OutputJSONPB(w, cmd.template())
}

type SetRelationCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"file path to set relation request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a set relation request template on stdout"`
	clients.DirectoryConfig
}

func (cmd *SetRelationCmd) Run(c *cc.CommonCtx) error {
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

	var req writer.SetRelationRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.V3.Writer.SetRelation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "set relation call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func (cmd *SetRelationCmd) template() proto.Message {
	return &writer.SetRelationRequest{
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
}

func (cmd *SetRelationCmd) print(w io.Writer) error {
	return jsonx.OutputJSONPB(w, cmd.template())
}

type DeleteRelationCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"file path to delete relation request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a delete relation request template on stdout"`
	clients.DirectoryConfig
}

func (cmd *DeleteRelationCmd) Run(c *cc.CommonCtx) error {
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

	var req writer.DeleteRelationRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.V3.Writer.DeleteRelation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "delete relation call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func (cmd *DeleteRelationCmd) template() proto.Message {
	return &writer.DeleteRelationRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
	}
}

func (cmd *DeleteRelationCmd) print(w io.Writer) error {
	return jsonx.OutputJSONPB(w, cmd.template())
}

type ListRelationsCmd struct {
	Request  string `arg:"" type:"s" name:"request" optional:"" help:"file path to list relations request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a list relations request template on stdout"`
	clients.DirectoryConfig
}

func (cmd *ListRelationsCmd) Run(c *cc.CommonCtx) error {
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

	var req reader.GetRelationsRequest
	err = clients.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.V3.Reader.GetRelations(c.Context, &req)

	if err != nil {
		return errors.Wrap(err, "get relations call failed")
	}

	return jsonx.OutputJSONPB(c.UI.Output(), resp)
}

func (cmd *ListRelationsCmd) template() proto.Message {
	return &reader.GetRelationsRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
		WithObjects:     false,
		Page:            &common.PaginationRequest{Size: 100, Token: ""},
	}
}

func (cmd *ListRelationsCmd) print(w io.Writer) error {
	return jsonx.OutputJSONPB(w, cmd.template())
}
