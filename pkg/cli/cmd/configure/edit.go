package configure

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/editor"
)

type EditConfigCmd struct {
	Name      ConfigName `arg:"" required:"" default:"${active_config}" help:"topaz config name"`
	ConfigDir string     `flag:"" default:"${topaz_cfg_dir}" help:"path to config folder"`
}

func (cmd *EditConfigCmd) Run(c *cc.CommonCtx) error {
	cfg := filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))
	if cmd.Name == "defaults" {
		cfg = filepath.Join(cc.GetTopazDir(), common.CLIConfigurationFile)
	}

	if _, err := os.Stat(cfg); err != nil {
		return err
	}

	e := editor.NewDefaultEditor([]string{"TOPAZ_EDITOR", "EDITOR"})

	if err := e.Launch(cfg); err != nil {
		return err
	}

	return nil
}
