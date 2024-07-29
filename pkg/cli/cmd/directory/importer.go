package directory

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/pkg/errors"
)

type ImportCmd struct {
	Directory string `short:"d" required:"" help:"directory containing .json data"`
	dsc.DirectoryConfig
}

func (cmd *ImportCmd) Run(c *cc.CommonCtx) error {
	if !c.IsServing(cmd.Host) {
		return errors.Wrap(cc.ErrNotServing, cmd.Host)
	}
	c.Con().Info().Msg(">>> importing data from %s", cmd.Directory)

	if fi, err := os.Stat(cmd.Directory); err != nil || !fi.IsDir() {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			return fmt.Errorf("--directory argument %q is not a directory", cmd.Directory)
		}
	}

	files, err := filepath.Glob(filepath.Join(cmd.Directory, "*.json"))
	if err != nil {
		return err
	}

	dsClient, err := dsc.NewClient(c, &cmd.DirectoryConfig)
	if err != nil {
		return err
	}

	return dsClient.Import(c.Context, files)
}
