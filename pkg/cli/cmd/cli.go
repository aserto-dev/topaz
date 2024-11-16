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

var Save SaveContext

type CLI struct {
	Start      topaz.StartCmd           `cmd:"" help:"start topaz instance (daemon mode)"`
	Stop       topaz.StopCmd            `cmd:"" help:"stop topaz instance"`
	Restart    topaz.RestartCmd         `cmd:"" help:"restart topaz instance"`
	Status     topaz.StatusCmd          `cmd:"" help:"status of topaz daemon process"`
	Config     configure.ConfigCmd      `cmd:"" help:"configure topaz instance"`
	Run        topaz.RunCmd             `cmd:"" help:"start topaz instance (console mode)"`
	Templates  templates.TemplateCmd    `cmd:"" help:"template commands"`
	Console    topaz.ConsoleCmd         `cmd:"" help:"open topaz console in the browser"`
	Directory  directory.DirectoryCmd   `cmd:"" aliases:"ds" help:"directory commands"`
	Authorizer authorizer.AuthorizerCmd `cmd:"" aliases:"az" help:"authorizer commands"`
	Certs      certs.CertsCmd           `cmd:"" help:"certificate management"`
	Install    topaz.InstallCmd         `cmd:"" help:"install topaz container"`
	Uninstall  topaz.UninstallCmd       `cmd:"" help:"uninstall topaz container"`
	Update     topaz.UpdateCmd          `cmd:"" help:"update topaz container version"`
	Version    VersionCmd               `cmd:"" help:"version information"`
	NoCheck    bool                     `flag:"" name:"no-check" json:"noCheck,omitempty" short:"N" env:"TOPAZ_NO_CHECK" help:"disable local container status check"`
	NoColor    bool                     `flag:"" name:"no-color" json:"no_color,omitempty" env:"TOPAZ_NO_COLOR" help:"disable colored terminal output"`
	LogLevel   int                      `flag:"" name:"verbosity" short:"v" type:"counter" default:"0" help:"log level"`
	Import     ImportCmd                `cmd:"" help:"'topaz import' was moved to 'topaz directory import'" hidden:""`
	Export     ExportCmd                `cmd:"" help:"'topaz export' was moved to 'topaz directory export'" hidden:""`
	Backup     BackupCmd                `cmd:"" help:"'topaz backup' was moved to 'topaz directory backup'" hidden:""`
	Restore    RestoreCmd               `cmd:"" help:"'topaz restore' was moved to 'topaz directory restore'" hidden:""`
	Configure  ConfigureCmd             `cmd:"" help:"'topaz configure' was moved to 'topaz config new'" hidden:""`
	Test       TestCmd                  `cmd:"" help:"'topaz test' was moved to 'topaz directory test'" hidden:""`
	Manifest   ManifestCmd              `cmd:"" help:"'topaz manifest ...' was moved to 'topaz directory get|set|delete manifest'" hidden:""`
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

type ManifestCmd struct {
	Get    GetManifestCmd    `cmd:"" help:"'topaz manifest get' was moved to 'topaz directory get manifest'" hidden:""`
	Set    SetManifestCmd    `cmd:"" help:"'topaz manifest set' was moved to 'topaz directory set manifest'" hidden:""`
	Delete DeleteManifestCmd `cmd:"" help:"'topaz manifest delete' was moved to 'topaz directory delete manifest'" hidden:""`
}

func (cmd *ManifestCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz manifest ...", "topaz directory get|set|delete manifest")
}

type GetManifestCmd struct{}

func (cmd *GetManifestCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz manifest get", "topaz directory get manifest")
}

type SetManifestCmd struct{}

func (cmd *SetManifestCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz manifest set", "topaz directory set manifest")
}

type DeleteManifestCmd struct{}

func (cmd *DeleteManifestCmd) Run(c *cc.CommonCtx) error {
	return movedErr("topaz manifest delete", "topaz directory delete manifest")
}

func movedErr(oldCmd, newCmd string) error {
	return errors.Errorf("the command %q was moved to %q.", oldCmd, newCmd)
}
