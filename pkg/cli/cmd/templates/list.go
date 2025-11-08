package templates

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/table"
)

type ListTemplatesCmd struct {
	TemplatesURL string `optional:"" default:"${topaz_tmpl_url}" env:"TOPAZ_TMPL_URL" help:"URL of template catalog"`
}

func (cmd *ListTemplatesCmd) Run(c *cc.CommonCtx) error {
	ctlg, err := getCatalog(cmd.TemplatesURL)
	if err != nil {
		return err
	}

	data := [][]any{}
	for n, t := range ctlg {
		data = append(data, []any{n, t.ShortDescription, t.DocumentationURL})
	}

	t := table.New(c.StdOut())

	t.Header(colName, colDescription, colDocumentation)
	t.Bulk(data)
	t.Render()

	return nil
}
