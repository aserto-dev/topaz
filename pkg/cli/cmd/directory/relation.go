package directory

import (
	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/edit"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/pb"
	"github.com/aserto-dev/topaz/pkg/cli/prompter"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetRelationCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to get relation request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a get relation request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit request" hidden:"" type:"fflag.Editor"`
	dsc.Config
}

func (cmd *GetRelationCmd) Run(c *cc.CommonCtx) error {
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
		p := prompter.New(cmd.template())
		if err := p.Show(); err != nil {
			return err
		}
		cmd.Request = jsonx.MaskedMarshalOpts().Format(p.Req())
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.GetRelationRequest
	err = pb.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.Reader.GetRelation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "get relation call failed")
	}

	return jsonx.OutputJSONPB(c.StdOut(), resp)
}

func (cmd *GetRelationCmd) template() proto.Message {
	return &reader.GetRelationRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
		WithObjects:     false,
	}
}

type SetRelationCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"file path to set relation request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a set relation request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit request" hidden:"" type:"fflag.Editor"`
	dsc.Config
}

func (cmd *SetRelationCmd) Run(c *cc.CommonCtx) error {
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
		p := prompter.New(cmd.template())
		if err := p.Show(); err != nil {
			return err
		}
		cmd.Request = jsonx.MaskedMarshalOpts().Format(p.Req())
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req writer.SetRelationRequest
	err = pb.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.Writer.SetRelation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "set relation call failed")
	}

	return jsonx.OutputJSONPB(c.StdOut(), resp)
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

type DeleteRelationCmd struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"file path to delete relation request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a delete relation request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit request" hidden:"" type:"fflag.Editor"`
	dsc.Config
}

func (cmd *DeleteRelationCmd) Run(c *cc.CommonCtx) error {
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
		p := prompter.New(cmd.template())
		if err := p.Show(); err != nil {
			return err
		}
		cmd.Request = jsonx.MaskedMarshalOpts().Format(p.Req())
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req writer.DeleteRelationRequest
	err = pb.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.Writer.DeleteRelation(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "delete relation call failed")
	}

	return jsonx.OutputJSONPB(c.StdOut(), resp)
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

type ListRelationsCmd struct {
	Request  string `arg:"" type:"s" name:"request" optional:"" help:"file path to list relations request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a list relations request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit request" hidden:"" type:"fflag.Editor"`
	dsc.Config
}

func (cmd *ListRelationsCmd) Run(c *cc.CommonCtx) error {
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
		p := prompter.New(cmd.template())
		if err := p.Show(); err != nil {
			return err
		}
		cmd.Request = jsonx.MaskedMarshalOpts().Format(p.Req())
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	var req reader.GetRelationsRequest
	err = pb.UnmarshalRequest(cmd.Request, &req)
	if err != nil {
		return err
	}

	resp, err := client.Reader.GetRelations(c.Context, &req)
	if err != nil {
		return errors.Wrap(err, "get relations call failed")
	}

	return jsonx.OutputJSONPB(c.StdOut(), resp)
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
