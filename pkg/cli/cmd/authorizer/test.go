package authorizer

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
)

type TestCmd struct {
	Exec     TestExecCmd     `cmd:"" help:"execute assertions"`
	Template TestTemplateCmd `cmd:"" help:"output assertions template"`
}

type TestExecCmd struct {
	common.TestExecCmd
	azc.Config
}

func (cmd *TestExecCmd) Run(c *cc.CommonCtx) error {
	files := []string{}
	for _, file := range cmd.Files {
		if expanded, err := filepath.Glob(file); err == nil {
			files = append(files, expanded...)
		}
	}

	runner, err := common.NewAuthorizerTestRunner(
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

const assertionsTemplate string = `{
  "assertions": [
	{"check_decision": {"identity_context": {"identity": "", "type": ""}, "resource_context": {}, "policy_context": {"path": "", "decisions": [""]}}, "expected":true, "description": ""},
  ]
}`

func (cmd *TestTemplateCmd) Run(c *cc.CommonCtx) error {
	if !cmd.Pretty {
		fmt.Fprintln(c.StdOut(), assertionsTemplate)
		return nil
	}

	r := strings.NewReader(assertionsTemplate)

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
