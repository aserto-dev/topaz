package directory

import (
	"github.com/alecthomas/kong"
	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetRelationCmd struct {
	RequestArgs
	dsc.Config
}

func (cmd *GetRelationCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *GetRelationCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return Invoke[reader.GetRelationRequest](
		c,
		client.IReader(),
		reader.Reader_GetRelation_FullMethodName,
		request,
	)
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
	RequestArgs
	dsc.Config
}

func (cmd *SetRelationCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *SetRelationCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return Invoke[writer.SetRelationRequest](
		c,
		client.IWriter(),
		writer.Writer_SetRelation_FullMethodName,
		request,
	)
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
	RequestArgs
	dsc.Config
}

func (cmd *DeleteRelationCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *DeleteRelationCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return Invoke[writer.DeleteRelationRequest](
		c,
		client.IWriter(),
		writer.Writer_DeleteRelation_FullMethodName,
		request,
	)
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
	RequestArgs
	dsc.Config
}

func (cmd *ListRelationsCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *ListRelationsCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return Invoke[reader.GetRelationsRequest](
		c,
		client.IReader(),
		reader.Reader_GetRelations_FullMethodName,
		request,
	)
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
