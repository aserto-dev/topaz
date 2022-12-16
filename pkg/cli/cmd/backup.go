package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type BackupCmd struct {
	File string `arg:""  default:"backup.tar.gz" help:"absolute file path to make backup to"`
	clients.Config
}

const defaultFileName = "backup.tar.gz"

func (cmd *BackupCmd) Run(c *cc.CommonCtx) error {
	if running, err := dockerx.IsRunning(dockerx.Topaz); !running || err != nil {
		if err != nil {
			return err
		}
		color.Yellow("!!! topaz is not running")
		return nil
	}

	dirClient, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	if cmd.File == defaultFileName {
		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}
		cmd.File = path.Join(currentDir, defaultFileName)
	}

	fmt.Fprintf(c.UI.Output(), ">>> backup to %s\n", cmd.File)
	return dirClient.Backup(c.Context, cmd.File)
}
