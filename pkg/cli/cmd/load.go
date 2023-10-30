package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/pkg/errors"
)

type LoadCmd struct {
	Path string `arg:"" type:"existingfile" default:"manifest.yaml" help:"absolute path to manifest file"`
	clients.Config
}

// var defaultManifestName = "manifest.yaml"

func (cmd *LoadCmd) Run(c *cc.CommonCtx) error {
	return errors.Errorf("The \"topaz load\" command has been deprecated, see \"topaz manifest set\".")

	// if err := CheckRunning(c); err != nil {
	// 	return err
	// }

	// cmd.Config.SessionID = uuid.NewString()
	// dirClient, err := clients.NewDirectoryClient(c, &cmd.Config)
	// if err != nil {
	// 	return err
	// }

	// if cmd.Path == defaultManifestName {
	// 	currentDir, err := os.Getwd()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	cmd.Path = path.Join(currentDir, defaultManifestName)
	// }

	// r, err := os.Open(cmd.Path)
	// if err != nil {
	// 	return err
	// }

	// color.Green(">>> load manifest from %s", cmd.Path)

	// return dirClient.SetManifest(c.Context, r)
}
