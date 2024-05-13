package configure

import (
	"fmt"
	"regexp"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type ConfigCmd struct {
	Use    UseConfigCmd    `cmd:"" help:"set active configuration"`
	New    NewConfigCmd    `cmd:"" help:"create new configuration"`
	List   ListConfigCmd   `cmd:"" help:"list configurations"`
	Rename RenameConfigCmd `cmd:"" help:"rename configuration"`
	Delete DeleteConfigCmd `cmd:"" help:"delete configuration"`
	Info   InfoConfigCmd   `cmd:"" help:"display configuration information"`
}

var restrictedNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`)

type ConfigName string

func (c ConfigName) AfterApply(ctx *kong.Context) error {
	if string(c) == "" {
		return fmt.Errorf("no configuration name value provided")
	}

	if !restrictedNamePattern.MatchString(string(c)) {
		return fmt.Errorf("configuration name is invalid, must match pattern %q", restrictedNamePattern.String())
	}

	return nil
}

func (c ConfigName) String() string {
	return string(c)
}

func (cmd *ConfigCmd) Run(c *cc.CommonCtx) error {
	return nil
}
