package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type StopCmd struct {
	ContainerName string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
	Wait          bool   `optional:"" default:"false" help:"wait for ports to be closed"`
}

func (cmd *StopCmd) Run(c *cc.CommonCtx) error {
	c.Config.NoCheck = false // enforce that Stop does not bypass CheckRunStatus() to short-circuit.
	if c.CheckRunStatus(cmd.ContainerName, cc.StatusNotRunning) {
		return nil
	}

	color.Green(">>> stopping topaz...")
	dc, err := dockerx.New()
	if err != nil {
		return err
	}
	if err := dc.Stop(cmd.ContainerName); err != nil {
		return err
	}

	if cmd.Wait {
		ports, err := config.GetConfig(c.Config.TopazConfigFile).Ports()
		if err != nil {
			return err
		}
		if err := cc.WaitForPorts(ports, cc.PortClosed); err != nil {
			return err
		}
	}

	return nil
}
