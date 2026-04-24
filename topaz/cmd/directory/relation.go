package directory

import (
	"context"
	"os"

	common "github.com/aserto-dev/topaz/api/directory/v4"
	"github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/api/directory/v4/writer"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"github.com/aserto-dev/topaz/topaz/x"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetRelationCmd struct {
	clients.RequestArgs
	dsc.Config

	req  reader.GetRelationRequest
	resp reader.GetRelationResponse
}

func (cmd *GetRelationCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, reader.Reader_GetRelation_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
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
	clients.RequestArgs
	dsc.Config

	req  writer.SetRelationRequest
	resp writer.SetRelationResponse
}

func (cmd *SetRelationCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, writer.Writer_SetRelation_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
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
			// CreatedAt:       timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Etag:      "",
		},
	}
}

type DeleteRelationCmd struct {
	clients.RequestArgs
	dsc.Config

	req  writer.DeleteRelationRequest
	resp writer.DeleteRelationResponse
}

func (cmd *DeleteRelationCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, writer.Writer_DeleteRelation_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
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
	clients.RequestArgs
	dsc.Config

	req  reader.ListRelationsRequest
	resp reader.ListRelationsResponse
}

func (cmd *ListRelationsCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, reader.Reader_ListRelations_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
}

func (cmd *ListRelationsCmd) template() proto.Message {
	return &reader.ListRelationsRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
		WithObjects:     false,
		Page:            &common.PaginationRequest{Size: x.MaxPaginationSize, Token: ""},
	}
}
