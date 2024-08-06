package configure

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/table"
)

type ListConfigCmd struct {
	ConfigDir string `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd ListConfigCmd) Run(c *cc.CommonCtx) error {
	tab := table.New(c.StdOut()).WithColumns("", "Name", "Config File")

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

		tab.WithRow(active, name, files[i].Name())
	}
	tab.Do()

	return nil
}
