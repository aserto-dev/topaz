package topaz

import (
	"context"

	"github.com/aserto-dev/topaz/topaz/cc"
)

type StatusCmd struct {
	ContainerName string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
}

func (cmd *StatusCmd) Run(ctx context.Context) error {
	cfg := cc.GetConfig()

	containerName := cfg.Running.ContainerName
	if cmd.ContainerName != containerName {
		containerName = cmd.ContainerName
	}

	if cfg.CheckRunStatus(containerName, cc.StatusRunning) {
		cc.Con().Info().Msg(">>> topaz %q is running", cfg.Running.Config)
	} else {
		cc.Con().Warn().Msg(">>> topaz is not running")
	}

	return nil
}
