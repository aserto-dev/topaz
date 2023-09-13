package cmd

import (
	"io"
	"os"
	"path"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/fatih/color"
)

type SaveCmd struct {
	File string `arg:"" type:"path" default:"manifest.yaml" help:"absolute path to manifest file"`
	clients.Config
}

func (cmd *SaveCmd) Run(c *cc.CommonCtx) error {
	if err := CheckRunning(c); err != nil {
		return err
	}

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

	color.Green(">>> save manifest to %s", cmd.File)

	r, err := dirClient.GetManifest(c.Context)
	if err != nil {
		return err
	}

	w, err := os.Create(cmd.File)
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	return nil
}
