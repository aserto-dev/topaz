package configure

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/table"
)

type ListConfigCmd struct {
	ConfigDir string `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd ListConfigCmd) Run(c *cc.CommonCtx) error {
	tab := table.New(c.StdOut())
	defer tab.Close()

	tab.Header("", "Name", "Config File")

	files, err := os.ReadDir(cmd.ConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	data := [][]any{}

	for i := range files {
		name := strings.Split(files[i].Name(), ".")[0]
		active := ""

		if files[i].Name() == filepath.Base(c.Config.Active.ConfigFile) {
			active = "*"
		}

		data = append(data, []any{active, name, files[i].Name()})
	}

	tab.Bulk(data)
	tab.Render()

	return nil
}
