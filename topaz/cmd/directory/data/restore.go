package data

import (
	"context"
	"os"
	"path"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
)

type RestoreCmd struct {
	dsc.Config

	File string `arg:""  default:"backup.tar.gz" help:"file path to backup source file"`
}

func (cmd *RestoreCmd) Run(ctx context.Context) error {
	if ok, err := clients.Validate(ctx, &cmd.Config); !ok {
		return err
	}

	dsClient, err := dsc.NewClient(ctx, &cmd.Config)
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

	cc.Con().Info().Msg(">>> restore from %s", cmd.File)

	return dsClient.Restore(ctx, cmd.File)
}
