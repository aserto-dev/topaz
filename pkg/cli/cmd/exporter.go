package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
	"github.com/google/uuid"
)

type ExportCmd struct {
	Directory string `short:"d" required:"" help:"directory to write .json data"`
	clients.Config
}

func (cmd *ExportCmd) Run(c *cc.CommonCtx) error {
	if running, err := dockerx.IsRunning(dockerx.Topaz); !running || err != nil {
		if err != nil {
			return err
		}
		color.Yellow("!!! topaz is not running")
		return nil
	}

	fmt.Fprintf(c.UI.Err(), ">>> exporting data...\n")
	objectsFile := filepath.Join(cmd.Directory, "objects.json")
	relationsFile := filepath.Join(cmd.Directory, "relations.json")

	cmd.Config.SessionID = uuid.NewString()
	dirClient, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	return dirClient.Export(c.Context, objectsFile, relationsFile)
}
