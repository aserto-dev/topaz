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
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetObjectCmd struct {
	RequestArgs
	dsc.Config
}

func (cmd *GetObjectCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *GetObjectCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return Invoke[reader.GetObjectRequest](
		c,
		client.IReader(),
		reader.Reader_GetObject_FullMethodName,
		request,
	)
}

func (cmd *GetObjectCmd) template() proto.Message {
	return &reader.GetObjectRequest{
		ObjectType:    "",
		ObjectId:      "",
		WithRelations: false,
		Page:          &common.PaginationRequest{Size: 100, Token: ""},
	}
}

type SetObjectCmd struct {
	RequestArgs
	dsc.Config
}

func (cmd *SetObjectCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *SetObjectCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return Invoke[writer.SetObjectRequest](
		c,
		client.IWriter(),
		writer.Writer_SetObject_FullMethodName,
		request,
	)
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
	RequestArgs
	dsc.Config
}

func (cmd *DeleteObjectCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *DeleteObjectCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return Invoke[writer.DeleteObjectRequest](
		c,
		client.IWriter(),
		writer.Writer_DeleteObject_FullMethodName,
		request,
	)
}

func (cmd *DeleteObjectCmd) template() proto.Message {
	return &writer.DeleteObjectRequest{
		ObjectType:    "",
		ObjectId:      "",
		WithRelations: false,
	}
}

type ListObjectsCmd struct {
	RequestArgs
	dsc.Config
}

func (cmd *ListObjectsCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *ListObjectsCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return Invoke[reader.GetObjectsRequest](
		c,
		client.IReader(),
		reader.Reader_GetObjects_FullMethodName,
		request,
	)
}

func (cmd *ListObjectsCmd) template() proto.Message {
	return &reader.GetObjectsRequest{
		ObjectType: "",
		Page:       &common.PaginationRequest{Size: 100, Token: ""},
	}
}
