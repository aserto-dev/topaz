package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrNotRunning = errors.New("topaz is not running, use 'topaz start' or 'topaz run' to start")
	ErrIsRunning  = errors.New("topaz is already running, use 'topaz stop' to stop")
	ErrNotServing = errors.New("topaz gRPC endpoint not SERVING")
)

type FormatVersion int

const (
	V2 FormatVersion = 2
	V3 FormatVersion = 3

	CLIConfigurationFile = "cli_config.json"
)

type CLI struct {
	Start     StartCmd      `cmd:"" help:"start topaz in daemon mode"`
	Stop      StopCmd       `cmd:"" help:"stop topaz instance"`
	Status    StatusCmd     `cmd:"" help:"status of topaz daemon process"`
	Run       RunCmd        `cmd:"" help:"run topaz in console mode"`
	Manifest  ManifestCmd   `cmd:"" help:"manifest commands"`
	Test      TestCmd       `cmd:"" help:"test assertions commands"`
	Templates TemplateCmd   `cmd:"" help:"template commands"`
	Console   ConsoleCmd    `cmd:"" help:"open console in the browser"`
	Import    ImportCmd     `cmd:"" help:"import directory objects"`
	Export    ExportCmd     `cmd:"" help:"export directory objects"`
	Backup    BackupCmd     `cmd:"" help:"backup directory data"`
	Restore   RestoreCmd    `cmd:"" help:"restore directory data"`
	Install   InstallCmd    `cmd:"" help:"install topaz container"`
	Configure ConfigureCmd  `cmd:"" help:"configure topaz service"`
	List      ListConfigCmd `cmd:"" help:"list available configuration files"`
	Certs     CertsCmd      `cmd:"" help:"cert commands"`
	Update    UpdateCmd     `cmd:"" help:"update topaz container version"`
	Uninstall UninstallCmd  `cmd:"" help:"uninstall topaz container"`
	Load      LoadCmd       `cmd:"" help:"load manifest from file (DEPRECATED)"`
	Save      SaveCmd       `cmd:"" help:"save manifest to file  (DEPRECATED)"`
	Version   VersionCmd    `cmd:"" help:"version information"`
	NoCheck   bool          `name:"no-check" json:"noCheck,omitempty" short:"N" env:"TOPAZ_NO_CHECK" help:"disable local container status check"`
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
