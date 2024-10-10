package configure

import (
	"regexp"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/pkg/errors"
)

type ConfigCmd struct {
	Use    UseConfigCmd    `cmd:"" help:"set active configuration"`
	New    NewConfigCmd    `cmd:"" help:"create new configuration"`
	List   ListConfigCmd   `cmd:"" help:"list configurations"`
	Rename RenameConfigCmd `cmd:"" help:"rename configuration"`
	Delete DeleteConfigCmd `cmd:"" help:"delete configuration"`
	Info   InfoConfigCmd   `cmd:"" help:"display configuration information"`
	Edit   EditConfigCmd   `cmd:"" help:"edit config file (defaults to active)" hidden:"" type:"fflag.Editor"`
}

var restrictedNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`)

type ConfigName string

func (c ConfigName) AfterApply(ctx *kong.Context) error {
	if string(c) == "" {
		return errors.Errorf("no configuration name value provided")
	}

	if !restrictedNamePattern.MatchString(string(c)) {
		return errors.Errorf("configuration name is invalid, must match pattern %q", restrictedNamePattern.String())
	}

	return nil
}

func (c ConfigName) String() string {
	return string(c)
}

func (cmd *ConfigCmd) Run(c *cc.CommonCtx) error {
	return nil
}

func (cmd *ConfigCmd) BeforeReset(ctx *kong.Context) error {
	n := ctx.Selected()
	if n != nil {
		fflag.UnHideCmds(ctx)
	}
	return nil
}
