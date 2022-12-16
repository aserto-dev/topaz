package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/google/uuid"
)

type ExportCmd struct {
	Directory string `short:"d" required:"" help:"directory to write .json data"`
	clients.Config
}

func (cmd *ExportCmd) Run(c *cc.CommonCtx) error {
	if err := CheckRunning(c); err != nil {
		return err
	}

	fmt.Fprintf(c.UI.Output(), ">>> exporting data to %s\n", cmd.Directory)

	objectsFile := filepath.Join(cmd.Directory, "objects.json")
	relationsFile := filepath.Join(cmd.Directory, "relations.json")

	cmd.Config.SessionID = uuid.NewString()
	dirClient, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	return dirClient.Export(c.Context, objectsFile, relationsFile)
}
