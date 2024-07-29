package directory

import (
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/pkg/errors"
)

type ExportCmd struct {
	Directory string `short:"d" required:"" help:"directory to write .json data"`
	directory.DirectoryConfig
}

func (cmd *ExportCmd) Run(c *cc.CommonCtx) error {
	if !c.IsServing(cmd.Host) {
		return errors.Wrap(cc.ErrNotServing, cmd.Host)
	}
	c.Con().Info().Msg(">>> exporting data to %s", cmd.Directory)

	objectsFile := filepath.Join(cmd.Directory, "objects.json")
	relationsFile := filepath.Join(cmd.Directory, "relations.json")

	dirClient, err := directory.NewClient(c, &cmd.DirectoryConfig)
	if err != nil {
		return err
	}

	return dirClient.Export(c.Context, objectsFile, relationsFile)
}
