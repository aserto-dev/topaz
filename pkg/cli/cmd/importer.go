package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/fatih/color"
	"github.com/google/uuid"
)

type ImportCmd struct {
	Directory string        `short:"d" required:"" help:"directory containing .json data"`
	Format    FormatVersion `flag:"" short:"f" enum:"3,2" name:"format" default:"3" help:"format of json data"`
	clients.Config
}

func (cmd *ImportCmd) Run(c *cc.CommonCtx) error {
	if err := CheckRunning(c); err != nil {
		return err
	}

	color.Green(">>> importing data from %s", cmd.Directory)

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

	cmd.Config.SessionID = uuid.NewString()
	dirClient, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	if cmd.Format == V2 {
		return dirClient.V2.Import(c.Context, files)
	}
	return dirClient.V3.Import(c.Context, files)
}
