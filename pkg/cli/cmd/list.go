package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type ListConfigCmd struct {
	ConfigDir string `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd ListConfigCmd) Run(c *cc.CommonCtx) error {
	table := c.UI.Normal().WithTable("Available Config Files", "Is Default")
	files, err := os.ReadDir(cmd.ConfigDir)
	if err != nil {
		return err
	}
	for i := range files {
		active := false
		if files[i].Name() == filepath.Base(c.Config.DefaultConfigFile) {
			active = true
		}
		table.WithTableRow(files[i].Name(), fmt.Sprintf("%v", active))
	}
	table.Do()

	return nil
}