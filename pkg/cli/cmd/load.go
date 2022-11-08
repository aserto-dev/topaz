package cmd

import (
	"os"
	"path"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
	"github.com/google/uuid"
)

type LoadCmd struct {
	File string `arg:""  default:"manifest.yaml" help:"absolute path to manifest file"`
	clients.Config
}

var defaultManifestName = "manifest.yaml"

func (cmd *LoadCmd) Run(c *cc.CommonCtx) error {
	if running, err := dockerx.IsRunning(dockerx.Topaz); !running || err != nil {
		if err != nil {
			return err
		}
		color.Yellow("!!! topaz is not running")
		return nil
	}

	cmd.Config.SessionID = uuid.NewString()
	dirClient, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	if cmd.File == defaultManifestName {
		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}
		cmd.File = path.Join(currentDir, defaultManifestName)
	}

	return dirClient.Load(c.Context, cmd.File)
}
