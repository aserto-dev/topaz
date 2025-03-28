package directory

import (
	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetRelationCmd struct {
	clients.RequestArgs
	dsc.Config
	req  reader.GetRelationRequest
	resp reader.GetRelationResponse
}

func (cmd *GetRelationCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.RequestArgs.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Config.Invoke(c.Context, reader.Reader_GetRelation_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
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

func (cmd *SetRelationCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.RequestArgs.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Config.Invoke(c.Context, writer.Writer_SetRelation_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
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
	clients.RequestArgs
	dsc.Config
	req  writer.DeleteRelationRequest
	resp writer.DeleteRelationResponse
}

func (cmd *DeleteRelationCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.RequestArgs.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Config.Invoke(c.Context, writer.Writer_DeleteRelation_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
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
	req  reader.GetRelationsRequest
	resp reader.GetRelationsResponse
}

func (cmd *ListRelationsCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(c.StdOut(), cmd.template())
	}

	if err := cmd.RequestArgs.Process(c, &cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Config.Invoke(c.Context, reader.Reader_GetRelations_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), &cmd.resp)
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
		Page:            &common.PaginationRequest{Size: x.MaxPaginationSize, Token: ""},
	}
}
