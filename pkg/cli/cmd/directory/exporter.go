package directory

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/pkg/errors"
)

type ExportCmd struct {
	Directory string `short:"d" help:"export to directory" default:"${cwd}"`
	File      string `short:"f" help:"export to file"`
	Stdout    bool   `flag:"" help:"export to --stdout"`
	Type      string `flag:"" enum:"data,objects,relations" required:"" default:"data" help:"export type (objects|relations|data=objects+relations)"`
	dsc.Config
}

func (cmd *ExportCmd) Run(c *cc.CommonCtx) error {
	if ok, err := clients.Validate(c.Context, &cmd.Config); !ok {
		return err
	}

	dsClient, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	options := typeToOption(cmd.Type)

	var dest *os.File
	switch {
	case cmd.Stdout:
		dest = os.Stdout
		return dsClient.ExportToFile(c.Context, dest, options)

	case cmd.File != "":
		c.Con().Info().Msg(">>> exporting to %s", cmd.File)

		dest, err = os.Create(cmd.File)
		if err != nil {
			return err
		}
		return dsClient.ExportToFile(c.Context, dest, options)

	case cmd.Directory != "":
		c.Con().Info().Msg(">>> exporting to %s", cmd.Directory)

		if fi, err := os.Stat(cmd.Directory); err != nil || !fi.IsDir() {
			if err != nil {
				return err
			}
			if !fi.IsDir() {
				return errors.Errorf("--directory argument %q is not a directory", cmd.Directory)
			}
		}

		objectsFile := filepath.Join(cmd.Directory, "objects.json")
		relationsFile := filepath.Join(cmd.Directory, "relations.json")
		return dsClient.ExportToDirectory(c.Context, objectsFile, relationsFile)

	default:
		return errors.Errorf("no valid command line options")
	}
}
