package configure

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
)

type UseConfigCmd struct {
	Name      ConfigName `arg:"" required:"" help:"topaz config name"`
	ConfigDir string     `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd *UseConfigCmd) Run(c *cc.CommonCtx) error {
	if _, err := os.Stat(filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))); err != nil {
		return err
	}

	topazContainers, err := c.GetRunningContainers()
	if err != nil {
		return err
	}

	if len(topazContainers) > 0 {
		return cc.ErrIsRunning
	}

	c.Config.Active.Config = cmd.Name.String()
	c.Config.Active.ConfigFile = filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))

	c.Con().Info().Msg("Using configuration %q", cmd.Name)

	if err := c.SaveContextConfig(common.CLIConfigurationFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return nil
}
