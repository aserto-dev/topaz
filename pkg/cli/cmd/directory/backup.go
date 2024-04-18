package directory

import (
	"os"
	"path"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

type BackupCmd struct {
	File   string        `arg:""  default:"backup.tar.gz" help:"absolute file path to make backup to"`
	Format FormatVersion `flag:"" short:"f" enum:"3,2" name:"format" default:"3" help:"format of json data"`
	clients.Config
}

const defaultFileName = "backup.tar.gz"

func (cmd *BackupCmd) Run(c *cc.CommonCtx) error {
	if !c.IsServing(cmd.Host) {
		return errors.Wrap(cc.ErrNotServing, cmd.Host)
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

	color.Green(">>> backup to %s", cmd.File)

	if cmd.Format == V2 {
		return dirClient.V2.Backup(c.Context, cmd.File)
	}
	return dirClient.V3.Backup(c.Context, cmd.File)
}
