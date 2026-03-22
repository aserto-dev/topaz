package configure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/cmd/common"
)

type UseConfigCmd struct {
	Name      ConfigName `arg:"" required:"" help:"topaz config name"`
	ConfigDir string     `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd *UseConfigCmd) Run(ctx context.Context) error {
	if _, err := os.Stat(filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))); err != nil {
		return err
	}

	cfg := cc.GetConfig()

	topazContainers, err := cfg.GetRunningContainers()
	if err != nil {
		return err
	}

	if len(topazContainers) > 0 {
		return cc.ErrIsRunning
	}

	cfg.Active.Config = cmd.Name.String()
	cfg.Active.ConfigFile = filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))

	cc.Con().Info().Msg("Using configuration %q", cmd.Name)

	if err := cfg.SaveContextConfig(common.CLIConfigurationFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return nil
}
