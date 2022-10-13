package cmd

import (
	"fmt"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/pkg/version"
)

type CLI struct {
	Backup    BackupCmd    `cmd:"" help:"backup directory data"`
	Configure ConfigureCmd `cmd:"" help:"configure topaz service"`
	Export    ExportCmd    `cmd:"" help:"export directory objects"`
	Import    ImportCmd    `cmd:"" help:"import directory objects"`
	Load      LoadCmd      `cmd:"" help:"load a manifest file"`
	Restore   RestoreCmd   `cmd:"" help:"restore directory data"`
	Run       RunCmd       `cmd:"" help:""`
	Save      SaveCmd      `cmd:"" help:"save a manifest file"`
	Start     StartCmd     `cmd:"" help:"start topaz instance"`
	Status    StatusCmd    `cmd:"" help:"display topaz instance status"`
	Stop      StopCmd      `cmd:"" help:"stop topaz instance"`
	Version   VersionCmd   `cmd:"" help:"version information"`
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
