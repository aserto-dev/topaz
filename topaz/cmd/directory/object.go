package directory

import (
	"context"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/aserto-dev/topaz/topaz/fflag"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"github.com/aserto-dev/topaz/topaz/x"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetObjectCmd struct {
	clients.RequestArgs
	dsc.Config

	req  reader.GetObjectRequest
	resp reader.GetObjectResponse
}

func (cmd *GetObjectCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *GetObjectCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, reader.Reader_GetObject_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
}

func (cmd *GetObjectCmd) template() proto.Message {
	return &reader.GetObjectRequest{
		ObjectType:    "",
		ObjectId:      "",
		WithRelations: false,
		Page:          &common.PaginationRequest{Size: x.MaxPaginationSize, Token: ""},
	}
}

type SetObjectCmd struct {
	clients.RequestArgs
	dsc.Config

	req  writer.SetObjectRequest
	resp writer.SetObjectResponse
}

func (cmd *SetObjectCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *SetObjectCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, writer.Writer_SetObject_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
}

func (cmd *SetObjectCmd) template() proto.Message {
	return &writer.SetObjectRequest{
		Object: &common.Object{
			Type:        "",
			Id:          "",
			DisplayName: "",
			Properties:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
			CreatedAt:   &timestamppb.Timestamp{},
			UpdatedAt:   &timestamppb.Timestamp{},
			Etag:        "",
		},
	}
}

type DeleteObjectCmd struct {
	clients.RequestArgs
	dsc.Config

	req  writer.DeleteObjectRequest
	resp writer.DeleteObjectResponse
}

func (cmd *DeleteObjectCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *DeleteObjectCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, writer.Writer_DeleteObject_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
}

func (cmd *DeleteObjectCmd) template() proto.Message {
	return &writer.DeleteObjectRequest{
		ObjectType:    "",
		ObjectId:      "",
		WithRelations: false,
	}
}

type ListObjectsCmd struct {
	clients.RequestArgs
	dsc.Config

	req  reader.GetObjectsRequest
	resp reader.GetObjectsResponse
}

func (cmd *ListObjectsCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *ListObjectsCmd) Run(ctx context.Context) error {
	if cmd.Template {
		return jsonx.OutputJSONPB(os.Stdout, cmd.template())
	}

	if err := cmd.Process(&cmd.req, cmd.template); err != nil {
		return err
	}

	if err := cmd.Invoke(ctx, reader.Reader_GetObjects_FullMethodName, &cmd.req, &cmd.resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(os.Stdout, &cmd.resp)
}

func (cmd *ListObjectsCmd) template() proto.Message {
	return &reader.GetObjectsRequest{
		ObjectType: "",
		Page:       &common.PaginationRequest{Size: x.MaxPaginationSize, Token: ""},
	}
}
