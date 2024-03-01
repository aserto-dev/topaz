package cmd

import (
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type StartCmd struct {
	StartRunCmd
	Wait bool `optional:"" default:"false" help:"wait for ports to be opened"`
}

func (cmd *StartCmd) Run(c *cc.CommonCtx) error {

	if err := cmd.run(c, modeDaemon); err != nil {
		return err
	}

	if cmd.Wait {
		cfg, err := config.LoadConfiguration(filepath.Join(cc.GetTopazCfgDir(), "config.yaml"))
		if err != nil {
			return err
		}

		ports, err := cfg.GetPorts()
		if err != nil {
			return err
		}

		if err := cc.WaitForPorts(ports, cc.PortOpened); err != nil {
			return err
		}
	}

	return nil
}
