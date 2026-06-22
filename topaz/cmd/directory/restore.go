package directory

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

	File string `arg:"" default:"backup.tar.gz" help:"path to source backup file"`
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

	r, err := os.Open(cmd.File)
	if err != nil {
		return err
	}
	defer r.Close()

	cc.Con().Info().Msg(">>> restore from %q", cmd.File)

	return dsClient.RestoreFromFile(ctx, r)
}
