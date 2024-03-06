package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cmd/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/directory"
)

type SaveContext bool

const (
	CLIConfigurationFile = "cli_config.json"
)

var (
	Save SaveContext
)

type CLI struct {
	Authorizer authorizer.AuthorizerCmd `cmd:"" help:"authorizer commands"`
	Directory  directory.DirectoryCmd   `cmd:"" help:"directory commands"`
	Start      StartCmd                 `cmd:"" help:"start topaz in daemon mode"`
	Stop       StopCmd                  `cmd:"" help:"stop topaz instance"`
	Status     StatusCmd                `cmd:"" help:"status of topaz daemon process"`
	Run        RunCmd                   `cmd:"" help:"run topaz in console mode"`
	Manifest   ManifestCmd              `cmd:"" help:"manifest commands"`
	Test       TestCmd                  `cmd:"" help:"test assertions commands"`
	Templates  TemplateCmd              `cmd:"" help:"template commands"`
	Console    ConsoleCmd               `cmd:"" help:"open console in the browser"`
	Backup     BackupCmd                `cmd:"" help:"backup directory data"`
	Restore    RestoreCmd               `cmd:"" help:"restore directory data"`
	Install    InstallCmd               `cmd:"" help:"install topaz container"`
	Configure  ConfigureCmd             `cmd:"" help:"configure topaz service"`
	List       ListConfigCmd            `cmd:"" help:"list available configuration files"`
	Certs      CertsCmd                 `cmd:"" help:"cert commands"`
	Update     UpdateCmd                `cmd:"" help:"update topaz container version"`
	Uninstall  UninstallCmd             `cmd:"" help:"uninstall topaz container"`
	Version    VersionCmd               `cmd:"" help:"version information"`
	NoCheck    bool                     `name:"no-check" json:"noCheck,omitempty" short:"N" env:"TOPAZ_NO_CHECK" help:"disable local container status check"`
	LogLevel   int                      `name:"log" short:"L" type:"counter" default:"0" help:"log level"`
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
