package topaz

import (
	"context"
)

type RestartCmd struct {
	StartRunCmd

	Wait bool `optional:"" default:"false" help:"wait for ports to be opened"`
}

func (cmd *RestartCmd) Run(ctx context.Context) error {
	if err := (&StopCmd{
		ContainerName: cmd.ContainerName,
		Wait:          true, // force wait, to prevent errors running start.
	}).Run(ctx); err != nil {
		return err
	}

	if err := (&StartCmd{
		StartRunCmd: cmd.StartRunCmd,
		Wait:        cmd.Wait,
	}).Run(ctx); err != nil {
		return err
	}

	return nil
}
