package clients

import (
	"bufio"
	"io"
	"os"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/edit"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/prompter"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type RequestArgs struct {
	Request  string `arg:"" type:"string" name:"request" optional:"" help:"read the JSON request from the stringified argument, file path, or stdin '-'"`
	Template bool   `name:"template" short:"t" help:"print request template to stdout"`
	Editor   bool   `name:"edit" short:"e" help:"open request in editor" hidden:"" type:"fflag.Editor"`
}

func (cmd *RequestArgs) Process(c *cc.CommonCtx, req proto.Message, tmpl func() proto.Message) error {
	if cmd.Request == "" && cmd.Editor && fflag.Enabled(fflag.Editor) {
		req, err := edit.Msg(tmpl())
		if err != nil {
			return err
		}

		cmd.Request = req
	}

	if cmd.Request == "" && fflag.Enabled(fflag.Prompter) {
		p := prompter.New(tmpl())
		if err := p.Show(); err != nil {
			return err
		}

		cmd.Request = jsonx.MaskedMarshalOpts().Format(p.Req())
	}

	if cmd.Request == "" {
		return errors.New("request argument is required")
	}

	if err := unmarshalRequest(cmd.Request, req); err != nil {
		return err
	}

	return nil
}

func unmarshalRequest(src string, msg proto.Message) error {
	if src == "-" {
		reader := bufio.NewReader(os.Stdin)

		bytes, err := io.ReadAll(reader)
		if err != nil {
			return errors.Wrap(err, "failed to read from stdin")
		}

		return protojson.Unmarshal(bytes, msg)
	}

	if fi, err := os.Stat(src); err == nil && !fi.IsDir() {
		bytes, err := os.ReadFile(src)
		if err != nil {
			return errors.Wrapf(err, "opening file [%s]", src)
		}

		return protojson.Unmarshal(bytes, msg)
	}

	return protojson.Unmarshal([]byte(src), msg)
}
