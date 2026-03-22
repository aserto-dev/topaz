package configure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/cmd/common"
	"github.com/pkg/errors"
)

type RenameConfigCmd struct {
	Name      ConfigName `arg:"" required:"" help:"topaz config name"`
	NewName   ConfigName `arg:"" required:"" help:"topaz new config name"`
	ConfigDir string     `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd *RenameConfigCmd) Run(ctx context.Context) error {
	cfg := cc.GetConfig()

	if cfg.CheckRunStatus(cc.ContainerName(fmt.Sprintf("%s.yaml", cmd.Name)), cc.StatusRunning) {
		return errors.Errorf("configuration %q is running, use 'topaz stop' to stop, before renaming", cmd.Name)
	}

	cc.Con().Info().Msg("Renaming configuration %q to %q", cmd.Name, cmd.NewName)

	oldFile := filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))
	newFile := filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.NewName))

	if cfg.Active.Config == cmd.Name.String() {
		cfg.Active.Config = cmd.NewName.String()
		cfg.Active.ConfigFile = newFile

		if err := cfg.SaveContextConfig(common.CLIConfigurationFile); err != nil {
			return errors.Wrap(err, "failed to update active context")
		}
	}

	return os.Rename(oldFile, newFile)
}
