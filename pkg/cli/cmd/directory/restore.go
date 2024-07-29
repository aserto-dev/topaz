package directory

import (
	"os"
	"path"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/pkg/errors"
)

type RestoreCmd struct {
	File string `arg:""  default:"backup.tar.gz" help:"absolute file path to local backup tarball"`
	directory.DirectoryConfig
}

func (cmd *RestoreCmd) Run(c *cc.CommonCtx) error {
	if !c.IsServing(cmd.Host) {
		return errors.Wrap(cc.ErrNotServing, cmd.Host)
	}

	dirClient, err := directory.NewClient(c, &cmd.DirectoryConfig)
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

	c.Con().Info().Msg(">>> restore from %s", cmd.File)

	return dirClient.Restore(c.Context, cmd.File)
}
