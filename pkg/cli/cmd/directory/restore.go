package directory

import (
	"os"
	"path"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

type RestoreCmd struct {
	File string `arg:""  default:"backup.tar.gz" help:"absolute file path to local backup tarball"`
	clients.DirectoryConfig
}

func (cmd *RestoreCmd) Run(c *cc.CommonCtx) error {
	if !c.IsServing(cmd.Host) {
		return errors.Wrap(cc.ErrNotServing, cmd.Host)
	}

	dirClient, err := clients.NewDirectoryClient(c, &cmd.DirectoryConfig)
	if err != nil {
		return err
	}

	if cmd.File == "backup.tar.gz" {
		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}
		cmd.File = path.Join(currentDir, "backup.tar.gz")
	}

	color.Green(">>> restore from %s", cmd.File)

	return dirClient.V3.Restore(c.Context, cmd.File)
}
