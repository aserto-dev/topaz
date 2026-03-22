package topaz

import (
	"context"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topaz/cc"
)

type StartCmd struct {
	StartRunCmd

	Wait bool `optional:"" default:"false" help:"wait for ports to be opened"`
}

func (cmd *StartCmd) Run(ctx context.Context) error {
	cfg := cc.GetConfig()

	if err := cmd.run(cfg, modeDaemon); err != nil {
		return err
	}

	if cmd.Wait {
		ports, err := config.GetConfig(cfg.Active.ConfigFile).Ports()
		if err != nil {
			return err
		}

		if err := cc.WaitForPorts(ports, cc.PortOpened); err != nil {
			return err
		}
	}

	return nil
}
