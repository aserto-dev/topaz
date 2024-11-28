package directory

import (
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
)

type ExportCmd struct {
	Directory string `short:"d" required:"" help:"directory to write .json data" default:"${cwd}"`
	dsc.Config
}

func (cmd *ExportCmd) Run(c *cc.CommonCtx) error {
	if ok, err := clients.Validate(c.Context, &cmd.Config); !ok {
		return err
	}

	c.Con().Info().Msg(">>> exporting data to %s", cmd.Directory)

	objectsFile := filepath.Join(cmd.Directory, "objects.json")
	relationsFile := filepath.Join(cmd.Directory, "relations.json")

	dsClient, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	return dsClient.Export(c.Context, objectsFile, relationsFile)
}
