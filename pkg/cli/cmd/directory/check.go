package directory

import (
	"github.com/alecthomas/kong"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type CheckCmd struct {
	RequestArgs
	dsc.Config
}

func (cmd *CheckCmd) BeforeReset(ctx *kong.Context) error {
	fflag.UnHideFlags(ctx)
	return nil
}

func (cmd *CheckCmd) Run(c *cc.CommonCtx) error {
	request, err := cmd.RequestArgs.Parse(c, cmd.template)
	if err != nil {
		return err
	}

	client, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return errors.Wrap(err, "failed to get directory client")
	}

	return Invoke[reader.CheckRequest](
		c,
		client.IReader(),
		reader.Reader_Check_FullMethodName,
		request,
	)
}

func (cmd *CheckCmd) template() proto.Message {
	return &reader.CheckRequest{
		ObjectType:  "",
		ObjectId:    "",
		Relation:    "",
		SubjectType: "",
		SubjectId:   "",
		Trace:       false,
	}
}
