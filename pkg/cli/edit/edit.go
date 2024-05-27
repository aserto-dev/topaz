package edit

import (
	"bytes"
	"os"

	"github.com/aserto-dev/topaz/pkg/cli/editor"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"google.golang.org/protobuf/proto"
)

func Msg(tmpl proto.Message) (string, error) {
	tmp, err := jsonx.MarshalOpts(true).Marshal(tmpl)
	if err != nil {
		return "", err
	}

	e := editor.NewDefaultEditor([]string{"TOPAZ_EDITOR", "EDITOR"})
	name := string(proto.MessageName(tmpl).Name())

	buf, path, err := e.LaunchTempFile("topaz", name, bytes.NewReader(tmp))
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(path) }()

	return string(buf), nil
}
