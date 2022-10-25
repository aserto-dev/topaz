package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type ImportCmd struct {
	Directory string `short:"d" required:"" help:"directory containing .json data"`
	clients.Config
}

func (cmd *ImportCmd) Run(c *cc.CommonCtx) error {
	if running, err := dockerx.IsRunning(dockerx.Topaz); !running || err != nil {
		if err != nil {
			return err
		}
		color.Yellow("!!! topaz is not running")
		return nil
	}

	fmt.Fprintf(c.UI.Err(), ">>> importing data...\n")
	files, err := filepath.Glob(filepath.Join(cmd.Directory, "*.json"))
	if err != nil {
		return err
	}

	dirClient, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	return dirClient.Import(c.Context, files)
}
