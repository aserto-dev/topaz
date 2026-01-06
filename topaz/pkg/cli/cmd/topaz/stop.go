package topaz

import (
	"fmt"
	"os"

	"github.com/aserto-dev/topaz/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/cmd/common"
	clicfg "github.com/aserto-dev/topaz/topaz/pkg/cli/config"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/dockerx"
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

	c.Config.Defaults.NoCheck = false // enforce that Stop does not bypass CheckRunStatus() to short-circuit.
	if c.CheckRunStatus(cmd.ContainerName, cc.StatusNotRunning) {
		return nil
	}

	c.Con().Info().Msg(">>> stopping topaz %q...", c.Config.Running.Config)

	if err := dc.Stop(cmd.ContainerName); err != nil {
		return err
	}

	if cmd.Wait {
		cfg, err := clicfg.Load(c.Config.Running.ConfigFile)
		if err != nil {
			return err
		}

		if err := cc.WaitForPorts(cfg.Ports(), cc.PortClosed); err != nil {
			return err
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
