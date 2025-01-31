package directory

import (
	"os"
	"path/filepath"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/pkg/errors"
)

type ImportCmd struct {
	Directory string `short:"d" help:"import from directory"  default:"${cwd}"`
	File      string `short:"f" help:"import from file"`
	Stdin     bool   `flag:"" help:"import from --stdin"`
	Type      string `flag:"" enum:"data,objects,relations" required:"" default:"data" help:"export type (objects|relations|data=objects+relations)"`
	dsc.Config
}

func (cmd *ImportCmd) Run(c *cc.CommonCtx) error {
	if ok, err := clients.Validate(c.Context, &cmd.Config); !ok {
		return err
	}

	dsClient, err := dsc.NewClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	options := typeToOption(cmd.Type)

	var src *os.File
	switch {
	case cmd.Stdin:
		src = os.Stdin
		return dsClient.ImportFromFile(c.Context, src, options)

	case cmd.File != "":
		if fi, err := os.Stat(cmd.File); err != nil || fi.IsDir() {
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return errors.Errorf("--file argument %q is a directory not a file", cmd.File)
			}
		}

		src, err = os.Open(cmd.File)
		if err != nil {
			return err
		}
		return dsClient.ImportFromFile(c.Context, src, options)

	case cmd.Directory != "":
		c.Con().Info().Msg(">>> importing data from %s", cmd.Directory)

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
		return dsClient.ImportFromDirectory(c.Context, files)

	default:
		return errors.Errorf("no valid command line options")
	}
}

func typeToOption(t string) dse3.Option {
	var options dse3.Option
	switch t {
	case "data":
		options = dse3.Option_OPTION_DATA
	case "objects":
		options = dse3.Option_OPTION_DATA_OBJECTS
	case "relations":
		options = dse3.Option_OPTION_DATA_RELATIONS
	default:
		options = dse3.Option_OPTION_DATA
	}
	return options
}
