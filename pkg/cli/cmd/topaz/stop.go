package topaz

import (
	"fmt"
	"os"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
)

type StopCmd struct {
	ContainerName string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
	Wait          bool   `optional:"" default:"false" help:"wait for ports to be closed"`
}

func (cmd *StopCmd) Run(c *cc.CommonCtx) error {
	dc, err := dockerx.New()
	if err != nil {
		return err
	}
	if strings.ContainsAny(cmd.ContainerName, "*") {
		topazContainers, err := c.GetRunningContainers()
		if err != nil {
			return err
		}

		c.Con().Info().Msg(">>> stopping topaz...")
		for _, container := range topazContainers {
			c.Con().Info().Msg(">>> stopping topaz %q...", c.Config.Running.Config)

			if err := dc.Stop(container.Names[0]); err != nil {
				return err
			}

			if cmd.Wait {
				var ports []string
				for _, port := range container.Ports {
					ports = append(ports, fmt.Sprintf("%d", port.PublicPort))
				}
				if err := cc.WaitForPorts(ports, cc.PortClosed); err != nil {
					return err
				}
			}

		}
	} else {
		c.Config.Defaults.NoCheck = false // enforce that Stop does not bypass CheckRunStatus() to short-circuit.
		if c.CheckRunStatus(cmd.ContainerName, cc.StatusNotRunning) {
			return nil
		}

		c.Con().Info().Msg(">>> stopping topaz %q...", c.Config.Running.Config)

		if err := dc.Stop(cmd.ContainerName); err != nil {
			return err
		}

		if cmd.Wait {
			ports, err := config.GetConfig(c.Config.Running.ConfigFile).Ports()
			if err != nil {
				return err
			}
			if err := cc.WaitForPorts(ports, cc.PortClosed); err != nil {
				return err
			}
		}
	}

	// empty running config
	c.Config.Running = cc.RunningConfig{}

	if err := c.SaveContextConfig(common.CLIConfigurationFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return nil
}
