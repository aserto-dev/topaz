package directory

import (
	"context"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/pkg/errors"
)

type ImportCmd struct {
	dsc.Config

	Directory string `short:"d" required:"" help:"directory containing .json data"  default:"${cwd}"`
}

func (cmd *ImportCmd) Run(ctx context.Context) error {
	if ok, err := clients.Validate(ctx, &cmd.Config); !ok {
		return err
	}

	cc.Con().Info().Msg(">>> importing data from %s", cmd.Directory)

	if fi, err := os.Stat(cmd.Directory); err != nil || !fi.IsDir() {
		if err != nil {
			return err
		}

		if !fi.IsDir() {
			return errors.Errorf("--directory argument %q is not a directory", cmd.Directory)
		}
	}

	files, err := filepath.Glob(filepath.Join(cmd.Directory, "*.json"))
	if err != nil {
		return err
	}

	dsClient, err := dsc.NewClient(ctx, &cmd.Config)
	if err != nil {
		return err
	}

	return dsClient.Import(ctx, files)
}
