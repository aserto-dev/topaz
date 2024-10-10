package configure

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/pkg/errors"
)

type DeleteConfigCmd struct {
	Name      ConfigName `arg:"" required:"" help:"topaz config name"`
	ConfigDir string     `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd *DeleteConfigCmd) Run(c *cc.CommonCtx) error {
	if c.CheckRunStatus(cc.ContainerName(fmt.Sprintf("%s.yaml", cmd.Name)), cc.StatusRunning) {
		return errors.Errorf("configuration %q is running, use 'topaz stop' to stop, before deleting", cmd.Name)
	}

	c.Con().Info().Msg("Removing configuration %q", cmd.Name)

	filename := filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))

	if c.Config.Active.Config == cmd.Name.String() {
		c.Config.Active.Config = ""
		c.Config.Active.ConfigFile = ""
		if err := c.SaveContextConfig(common.CLIConfigurationFile); err != nil {
			return errors.Wrap(err, "failed to update active context")
		}
	}

	return os.Remove(filename)
}
