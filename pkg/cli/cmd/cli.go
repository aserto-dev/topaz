package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/directory"
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
	Run        RunCmd                   `cmd:"" help:"run topaz in console mode"`
	Start      StartCmd                 `cmd:"" help:"start topaz in daemon mode"`
	Stop       StopCmd                  `cmd:"" help:"stop topaz instance"`
	Restart    RestartCmd               `cmd:"" help:"restart topaz instance"`
	Status     StatusCmd                `cmd:"" help:"status of topaz daemon process"`
	Manifest   ManifestCmd              `cmd:"" help:"manifest commands"`
	Templates  TemplateCmd              `cmd:"" help:"template commands"`
	Console    ConsoleCmd               `cmd:"" help:"open console in the browser"`
	Directory  directory.DirectoryCmd   `cmd:"" aliases:"ds" help:"directory commands"`
	Authorizer authorizer.AuthorizerCmd `cmd:"" aliases:"az" help:"authorizer commands"`
	Config     ConfigCmd                `cmd:"" help:"configure topaz service"`
	Certs      CertsCmd                 `cmd:"" help:"cert commands"`
	Install    InstallCmd               `cmd:"" help:"install topaz container"`
	Uninstall  UninstallCmd             `cmd:"" help:"uninstall topaz container"`
	Update     UpdateCmd                `cmd:"" help:"update topaz container version"`
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

func PromptYesNo(label string, def bool) bool {
	choices := "Y/n"
	if !def {
		choices = "y/N"
	}

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return def
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
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
