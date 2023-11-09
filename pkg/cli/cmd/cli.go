package cmd

import (
	"errors"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
)

var ErrNotRunning = errors.New("topaz is not running, use 'topaz start' or 'topaz run' to start")

type FormatVersion int

const (
	V2 FormatVersion = 2
	V3 FormatVersion = 3
)

type CLI struct {
	Start     StartCmd     `cmd:"" help:"start topaz in daemon mode"`
	Stop      StopCmd      `cmd:"" help:"stop topaz instance"`
	Status    StatusCmd    `cmd:"" help:"status of topaz daemon process"`
	Run       RunCmd       `cmd:"" help:"run topaz in console mode"`
	Manifest  ManifestCmd  `cmd:"" help:"manifest commands"`
	Load      LoadCmd      `cmd:"" help:"load manifest from file (DEPRECATED)"`
	Save      SaveCmd      `cmd:"" help:"save manifest to file  (DEPRECATED)"`
	Import    ImportCmd    `cmd:"" help:"import directory objects"`
	Export    ExportCmd    `cmd:"" help:"export directory objects"`
	Backup    BackupCmd    `cmd:"" help:"backup directory data"`
	Restore   RestoreCmd   `cmd:"" help:"restore directory data"`
	Test      TestCmd      `cmd:"" help:"execute assertions"`
	Install   InstallCmd   `cmd:"" help:"install topaz container"`
	Configure ConfigureCmd `cmd:"" help:"configure topaz service"`
	Update    UpdateCmd    `cmd:"" help:"update topaz container version"`
	Uninstall UninstallCmd `cmd:"" help:"uninstall topaz container"`
	Version   VersionCmd   `cmd:"" help:"version information"`
	Console   ConsoleCmd   `cmd:"" help:"opens the console in the browser"`
	NoCheck   bool         `name:"no-check" short:"N" env:"TOPAZ_NO_CHECK" help:"disable local container status check"`
}

func CheckRunning(c *cc.CommonCtx) error {
	if c.NoCheck {
		return nil
	}

	if running, err := dockerx.IsRunning(dockerx.Topaz); !running || err != nil {
		if !running {
			return ErrNotRunning
		}
		if err != nil {
			return err
		}
	}

	return nil
}
