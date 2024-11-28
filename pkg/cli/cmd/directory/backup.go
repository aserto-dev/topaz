package directory

import (
	"os"
	"path"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
)

type BackupCmd struct {
	File string `arg:""  default:"backup.tar.gz" help:"absolute file path to make backup to"`
	dsc.Config
}

const defaultFileName = "backup.tar.gz"

func (cmd *BackupCmd) Run(c *cc.CommonCtx) error {
	if ok, err := clients.Validate(c.Context, &cmd.Config); !ok {
		return err
	}

	dsClient, err := dsc.NewClient(c, &cmd.Config)
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

	c.Con().Info().Msg(">>> backup to %s", cmd.File)

	return dsClient.Backup(c.Context, cmd.File)
}
