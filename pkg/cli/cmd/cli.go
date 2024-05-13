package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/certs"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/configure"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/templates"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/topaz"
	"github.com/pkg/errors"
)

type SaveContext bool

const (
	CLIConfigurationFile = "topaz.json"
)

var (
	Save SaveContext
)

type CLI struct {
	Run        topaz.RunCmd             `cmd:"" help:"run topaz in console mode"`
	Start      topaz.StartCmd           `cmd:"" help:"start topaz in daemon mode"`
	Stop       topaz.StopCmd            `cmd:"" help:"stop topaz instance"`
	Restart    topaz.RestartCmd         `cmd:"" help:"restart topaz instance"`
	Status     topaz.StatusCmd          `cmd:"" help:"status of topaz daemon process"`
	Manifest   directory.ManifestCmd    `cmd:"" help:"manifest commands"`
	Templates  templates.TemplateCmd    `cmd:"" help:"template commands"`
	Console    topaz.ConsoleCmd         `cmd:"" help:"open console in the browser"`
	Directory  directory.DirectoryCmd   `cmd:"" aliases:"ds" help:"directory commands"`
	Authorizer authorizer.AuthorizerCmd `cmd:"" aliases:"az" help:"authorizer commands"`
	Config     configure.ConfigCmd      `cmd:"" help:"configure topaz service"`
	Certs      certs.CertsCmd           `cmd:"" help:"cert commands"`
	Install    topaz.InstallCmd         `cmd:"" help:"install topaz container"`
	Uninstall  topaz.UninstallCmd       `cmd:"" help:"uninstall topaz container"`
	Update     topaz.UpdateCmd          `cmd:"" help:"update topaz container version"`
	Version    VersionCmd               `cmd:"" help:"version information"`
	NoCheck    bool                     `name:"no-check" json:"noCheck,omitempty" short:"N" env:"TOPAZ_NO_CHECK" help:"disable local container status check"`
	LogLevel   int                      `name:"log" short:"L" type:"counter" default:"0" help:"log level"`
	Import     ImportCmd                `cmd:"" help:"'topaz import' was moved to 'topaz directory import'" hidden:""`
	Export     ExportCmd                `cmd:"" help:"'topaz export' was moved to 'topaz directory export'" hidden:""`
	Backup     BackupCmd                `cmd:"" help:"'topaz backup' was moved to 'topaz directory backup'" hidden:""`
	Restore    RestoreCmd               `cmd:"" help:"'topaz restore' was moved to 'topaz directory restore'" hidden:""`
	Configure  ConfigureCmd             `cmd:"" help:"'topaz configure' was moved to 'topaz config new'" hidden:""`
	Test       TestCmd                  `cmd:"" help:"'topaz test' was moved to 'topaz directory test'" hidden:""`
}

type ImportCmd struct{}

func (cmd *ImportCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz import", "topaz directory import")
}

type ExportCmd struct{}

func (cmd *ExportCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz export", "topaz directory export")
}

type BackupCmd struct{}

func (cmd *BackupCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz backup", "topaz directory backup")
}

type RestoreCmd struct{}

func (cmd *RestoreCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz restore", "topaz directory restore")
}

type ConfigureCmd struct{}

func (cmd *ConfigureCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz configure", "topaz config new")
}

type TestCmd struct {
	Exec     TestExecCmd     `cmd:"" help:"'topaz test exec' was moved to 'topaz directory test exec'" hidden:""`
	Template TestTemplateCmd `cmd:"" help:"'topaz test template' was moved to 'topaz directory test template'" hidden:""`
}

func (cmd *TestCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz test", "topaz directory test")
}

type TestExecCmd struct{}

func (cmd *TestExecCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz test exec", "topaz directory test exec")
}

type TestTemplateCmd struct{}

func (cmd *TestTemplateCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz test template", "topaz directory test template")
}

func movedErr(oldCmd, newCmd string) error {
	return errors.Errorf("The command %q was moved to %q.", oldCmd, newCmd)
}
