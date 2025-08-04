package topaz

import (
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
		if err := cc.WaitForPorts(cmd.cfg.Ports(), cc.PortOpened); err != nil {
			return err
		}
	}

	return nil
}
