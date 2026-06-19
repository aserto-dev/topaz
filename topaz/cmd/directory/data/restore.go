package data

import (
	"context"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
)

type RestoreCmd struct {
	dsc.Config

	File string `arg:""  default:"backup.tar.gz" help:"file path to source backup file"`
}

func (cmd *RestoreCmd) Run(ctx context.Context) error {
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

	cc.Con().Info().Msg(">>> restore from %s", cmd.File)

	return dsClient.Restore(ctx, cmd.File)
}
