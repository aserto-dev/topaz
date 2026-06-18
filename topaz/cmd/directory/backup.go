package directory

import (
	"context"
	"os"
	"path"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
)

const defBackupFileName = "backup.tar.gz"

type BackupCmd struct {
	dsc.Config

	File string `arg:"" default:"backup.tar.gz" help:"path to target backup file"`
}

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

		cmd.File = path.Join(currentDir, defBackupFileName)
	}

	w, err := os.Create(cmd.File)
	if err != nil {
		return err
	}
	defer w.Close()

	cc.Con().Info().Msg(">>> backup to %q", cmd.File)

	return dsClient.BackupToFile(ctx, w)
}
