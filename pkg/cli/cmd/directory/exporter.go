package directory

import (
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

type ExportCmd struct {
	Directory      string        `short:"d" required:"" help:"directory to write .json data"`
	Format         FormatVersion `flag:"" short:"f" enum:"3,2" name:"format" default:"3" help:"format of json data"`
	clients.Config `envprefix:"TOPAZ_DIRECTORY_"`
}

func (cmd *ExportCmd) Run(c *cc.CommonCtx) error {
	if !c.IsServing(cmd.Host) {
		return errors.Wrap(cc.ErrNotServing, cmd.Host)
	}
	color.Green(">>> exporting data to %s", cmd.Directory)

	objectsFile := filepath.Join(cmd.Directory, "objects.json")
	relationsFile := filepath.Join(cmd.Directory, "relations.json")

	dirClient, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	if cmd.Format == V2 {
		return dirClient.V2.Export(c.Context, objectsFile, relationsFile)
	}
	return dirClient.V3.Export(c.Context, objectsFile, relationsFile)
}
