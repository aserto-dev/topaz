package topaz

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type StatusCmd struct {
	ContainerName string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
}

func (cmd *StatusCmd) Run(c *cc.CommonCtx) error {
	containerName := c.Config.Running.ContainerName
	if cmd.ContainerName != containerName {
		containerName = cmd.ContainerName
	}

	if c.CheckRunStatus(containerName, cc.StatusRunning) {
		c.Con().Info().Msg(">>> topaz %q is running", c.Config.Running.Config)
	} else {
		c.Con().Warn().Msg(">>> topaz is not running")
	}

	return nil
}
