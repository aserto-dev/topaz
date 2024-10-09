package templates

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/table"
)

type ListTemplatesCmd struct {
	TemplatesURL string `arg:"" required:"false" default:"https://topaz.sh/assets/templates/templates.json" help:"URL of template catalog"`
}

func (cmd *ListTemplatesCmd) Run(c *cc.CommonCtx) error {
	ctlg, err := getCatalog(cmd.TemplatesURL)
	if err != nil {
		return err
	}

	maxWidth := 0
	for n := range ctlg {
		maxWidth = max(maxWidth, len(n)+1)
	}

	tab := table.New(c.StdOut()).WithColumns(colName, colDescription, colDocumentation)
	tab.WithTableNoAutoWrapText()
	for n, t := range ctlg {
		tab.WithRow(n, t.ShortDescription, t.DocumentationURL)
	}
	tab.Do()

	return nil
}
