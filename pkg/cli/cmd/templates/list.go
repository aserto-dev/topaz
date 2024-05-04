package templates

import "github.com/aserto-dev/topaz/pkg/cli/cc"

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

	table := c.UI.Normal().WithTable(colName, colDescription, colDocumentation)
	table.WithTableNoAutoWrapText()
	for n, t := range ctlg {
		table.WithTableRow(n, t.ShortDescription, t.DocumentationURL)
	}
	table.Do()

	return nil
}

func max(rhs, lhs int) int {
	if rhs < lhs {
		return lhs
	}
	return rhs
}
