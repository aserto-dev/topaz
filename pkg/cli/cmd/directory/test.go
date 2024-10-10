package directory

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
)

type TestCmd struct {
	Exec     TestExecCmd     `cmd:"" help:"execute assertions"`
	Template TestTemplateCmd `cmd:"" help:"output assertions template"`
}

type TestExecCmd struct {
	common.TestExecCmd
	dsc.Config
}

func (cmd *TestExecCmd) Run(c *cc.CommonCtx) error {
	files := []string{}
	for _, file := range cmd.Files {
		if expanded, err := filepath.Glob(file); err == nil {
			files = append(files, expanded...)
		}
	}

	runner, err := common.NewDirectoryTestRunner(
		c,
		&common.TestExecCmd{
			Files:   files,
			Stdin:   cmd.Stdin,
			Summary: cmd.Summary,
			Format:  cmd.Format,
			Desc:    cmd.Desc,
		},
		&cmd.Config,
	)
	if err != nil {
		return err
	}

	return runner.Run(c)
}

type TestTemplateCmd struct {
	Pretty bool `flag:"" default:"false" help:"pretty print JSON"`
}

const assertionsTemplateV3 string = `{
  "assertions": [
  	{"check": {"object_type": "", "object_id": "", "relation": "", "subject_type": "", "subject_id": ""}, "expected": true, "description": ""}
  ]
}`

func (cmd *TestTemplateCmd) Run(c *cc.CommonCtx) error {
	if !cmd.Pretty {
		c.Out().Msg(assertionsTemplateV3)
		return nil
	}

	r := strings.NewReader(assertionsTemplateV3)
	dec := json.NewDecoder(r)

	var template interface{}
	if err := dec.Decode(&template); err != nil {
		return err
	}

	enc := json.NewEncoder(c.StdOut())
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(template); err != nil {
		return err
	}

	return nil
}
