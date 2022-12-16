package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/google/uuid"
)

type ImportCmd struct {
	Directory string `short:"d" required:"" help:"directory containing .json data"`
	clients.Config
}

func (cmd *ImportCmd) Run(c *cc.CommonCtx) error {
	if err := CheckRunning(c); err != nil {
		return err
	}

	fmt.Fprintf(c.UI.Output(), ">>> importing data from %s\n", cmd.Directory)

	files, err := filepath.Glob(filepath.Join(cmd.Directory, "*.json"))
	if err != nil {
		return err
	}

	cmd.Config.SessionID = uuid.NewString()
	dirClient, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	return dirClient.Import(c.Context, files)
}
