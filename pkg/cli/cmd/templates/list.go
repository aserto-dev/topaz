package templates

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/table"
)

type ListTemplatesCmd struct {
	CatalogArgs
}

func (cmd *ListTemplatesCmd) Run(c *cc.CommonCtx) error {
	catalogURL, err := cmd.CatalogArgs.URL()
	if err != nil {
		return err
	}

	catalog, err := getCatalog(catalogURL)
	if err != nil {
		return err
	}

	maxWidth := 0
	for n := range catalog {
		maxWidth = max(maxWidth, len(n)+1)
	}

	tab := table.New(c.StdOut()).WithColumns(colName, colDescription, colDocumentation)
	tab.WithTableNoAutoWrapText()
	for n, t := range catalog {
		tab.WithRow(n, t.ShortDescription, t.DocumentationURL)
	}
	tab.Do()

	return nil
}
