package configure

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type ListConfigCmd struct {
	ConfigDir string `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd ListConfigCmd) Run(c *cc.CommonCtx) error {
	table := c.UI.Normal().WithTable("", "Name", "Config File")

	files, err := os.ReadDir(cmd.ConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for i := range files {
		name := strings.Split(files[i].Name(), ".")[0]
		active := ""
		if files[i].Name() == filepath.Base(c.Config.Active.ConfigFile) {
			active = "*"
		}

		table.WithTableRow(active, name, files[i].Name())
	}
	table.Do()

	return nil
}
