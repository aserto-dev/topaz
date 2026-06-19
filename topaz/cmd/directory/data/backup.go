package data

import (
	"context"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
)

type BackupCmd struct {
	dsc.Config

	File string `arg:""  default:"backup.tar.gz" help:"file path to target backup file"`
}

const defBackupFileName = "backup.tar.gz"

func (cmd *BackupCmd) Run(ctx context.Context) error {
	if ok, err := clients.Validate(ctx, &cmd.Config); !ok {
		return err
	}

	dsClient, err := dsc.NewClient(ctx, &cmd.Config)
	if err != nil {
		return err
	}

	if cmd.File == defBackupFileName {
		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}

		cmd.File = filepath.Join(currentDir, defBackupFileName)
	}

	cc.Con().Info().Msg(">>> backup to %s", cmd.File)

	return dsClient.Backup(ctx, cmd.File)
}
