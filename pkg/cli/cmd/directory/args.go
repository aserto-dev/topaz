package directory

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/edit"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/prompter"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type RequestArgs struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"json request or file path to get object request or '-' to read from stdin"`
	Template bool   `name:"template" short:"t" help:"prints a get object request template on stdout"`
	Editor   bool   `name:"edit" short:"e" help:"edit request" hidden:"" type:"fflag.Editor"`
}

func (cmd *RequestArgs) Parse(c *cc.CommonCtx, tmpl func() proto.Message) (string, error) {
	// template
	if cmd.Template {
		return "", jsonx.OutputJSONPB(c.StdOut(), tmpl())
	}

	// editor
	if cmd.Request == "" && cmd.Editor && fflag.Enabled(fflag.Editor) {
		req, err := edit.Msg(tmpl())
		if err != nil {
			return "", err
		}
		return req, nil
	}

	// prompter
	if cmd.Request == "" && fflag.Enabled(fflag.Prompter) {
		p := prompter.New(tmpl())
		if err := p.Show(); err != nil {
			return "", err
		}
		return jsonx.MaskedMarshalOpts().Format(p.Req()), nil
	}

	// string argument
	if cmd.Request == "" {
		return "", errors.New("request argument is required")
	}
	return "", nil
}
