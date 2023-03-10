package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
)

type StartCmd struct {
	*dockerx.Container `embed:""`
}

func (cmd *StartCmd) Run(c *cc.CommonCtx) error {
	if err := createMountDirs(); err != nil {
		return err
	}
	return cmd.Start(dockerx.Deamon)
}

func createMountDirs() error {
	if _, err := CreateCertsDir(); err != nil {
		return err
	}

	if _, err := CreateDataDir(); err != nil {
		return err
	}

	return nil
}
