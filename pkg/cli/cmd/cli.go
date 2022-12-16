package cmd

import (
	"errors"
	"fmt"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/pkg/version"
)

var ErrNotRunning = errors.New("topaz is not running, use 'topaz start' or 'topaz run' to start")

type CLI struct {
	Backup    BackupCmd    `cmd:"" help:"backup directory data"`
	Configure ConfigureCmd `cmd:"" help:"configure topaz service"`
	Export    ExportCmd    `cmd:"" help:"export directory objects"`
	Install   InstallCmd   `cmd:"" help:"install topaz"`
	Import    ImportCmd    `cmd:"" help:"import directory objects"`
	Load      LoadCmd      `cmd:"" help:"load a manifest file"`
	Restore   RestoreCmd   `cmd:"" help:"restore directory data"`
	Run       RunCmd       `cmd:"" help:""`
	Save      SaveCmd      `cmd:"" help:"save a manifest file"`
	Start     StartCmd     `cmd:"" help:"start topaz instance"`
	Status    StatusCmd    `cmd:"" help:"display topaz instance status"`
	Stop      StopCmd      `cmd:"" help:"stop topaz instance"`
	Update    UpdateCmd    `cmd:"" help:"update topaz container version"`
	Version   VersionCmd   `cmd:"" help:"version information"`
	Uninstall UninstallCmd `cmd:"" help:"uninstall topaz, removes all locally installed artifacts"`
	NoCheck   bool         `name:"no-check" hidden:"" env:"TOPAZ_NO_CHECK" help:"disable running check"`
}

type VersionCmd struct{}

func (cmd *VersionCmd) Run(c *cc.CommonCtx) error {
	fmt.Fprintf(c.UI.Output(), "%s - %s (%s)\n",
		x.AppName,
		version.GetInfo().String(),
		x.AppVersionTag,
	)
	return nil
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
