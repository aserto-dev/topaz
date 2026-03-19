package templates

import (
	"context"
	"os"

	"github.com/aserto-dev/topaz/topaz/table"
	"github.com/aserto-dev/topaz/topaz/x"
)

type ListTemplatesCmd struct {
	Legacy       bool   `optional:"" default:"false" help:"use legacy templates"`
	TemplatesURL string `optional:"" default:"${topaz_tmpl_url}" env:"TOPAZ_TMPL_URL" help:"URL of template catalog"`
}

func (cmd *ListTemplatesCmd) Run(ctx context.Context) error {
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

	t := table.New(os.Stdout)

	t.Header(colName, colDescription, colDocumentation)
	t.Bulk(data)
	t.Render()

	return nil
}
