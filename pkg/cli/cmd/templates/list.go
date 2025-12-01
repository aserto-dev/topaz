package templates

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/table"
	"github.com/aserto-dev/topaz/pkg/cli/x"
)

type ListTemplatesCmd struct {
	Legacy       bool   `optional:"" default:"false" help:"use legacy templates"`
	TemplatesURL string `optional:"" default:"${topaz_tmpl_url}" env:"TOPAZ_TMPL_URL" help:"URL of template catalog"`
}

func (cmd *ListTemplatesCmd) Run(c *cc.CommonCtx) error {
	if cmd.Legacy {
		cmd.TemplatesURL = x.TopazTmplV32URL
	}

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
