package cmd

import "github.com/aserto-dev/topaz/pkg/cli/cc"

type RestartCmd struct {
	StartRunCmd
	Wait bool `optional:"" default:"false" help:"wait for ports to be opened"`
}

func (cmd *RestartCmd) Run(c *cc.CommonCtx) error {
	{
		cmd := &StopCmd{
			ContainerName: cmd.ContainerName,
			Wait:          cmd.Wait,
		}
		if err := cmd.Run(c); err != nil {
			return err
		}
	}
	{
		cmd := &StartCmd{
			StartRunCmd: cmd.StartRunCmd,
			Wait:        cmd.Wait,
		}
		if err := cmd.Run(c); err != nil {
			return err
		}
	}
	return nil
}
