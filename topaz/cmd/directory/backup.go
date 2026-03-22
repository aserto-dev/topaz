package directory

import (
	"context"
	"os"
	"path"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
)

type BackupCmd struct {
	dsc.Config

	File string `arg:""  default:"backup.tar.gz" help:"absolute file path to make backup to"`
}

const defaultFileName = "backup.tar.gz"

func (cmd *BackupCmd) Run(ctx context.Context) error {
	if ok, err := clients.Validate(ctx, &cmd.Config); !ok {
		return err
	}

	dsClient, err := dsc.NewClient(ctx, &cmd.Config)
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

	cc.Con().Info().Msg(">>> backup to %s", cmd.File)

	return dsClient.Backup(ctx, cmd.File)
}
