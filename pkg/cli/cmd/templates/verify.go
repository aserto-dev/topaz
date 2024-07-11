package templates

import (
	"bytes"
	"encoding/json"
	"fmt"

	v3 "github.com/aserto-dev/azm/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/table"
	"github.com/rs/zerolog"
)

type VerifyTemplateCmd struct {
	TemplatesURL string `arg:"" required:"false" default:"https://topaz.sh/assets/templates/templates.json" help:"URL of template catalog"`
}

func (cmd *VerifyTemplateCmd) Run(c *cc.CommonCtx) error {
	ctlg, err := getCatalog(cmd.TemplatesURL)
	if err != nil {
		return err
	}

	// limit the amount of noise from the azm parser.
	zerolog.SetGlobalLevel(zerolog.Disabled)

	tab := table.New(c.StdOut()).WithColumns("template", "asset", "exists", "parsed", "error")
	tab.WithTableNoAutoWrapText()

	for tmplName := range ctlg {

		tmpl, err := getTemplate(tmplName, cmd.TemplatesURL)
		if err != nil {
			return err
		}
		{
			absURL := tmpl.AbsURL(tmpl.Assets.Manifest)
			exists, parsed, err := validateManifest(absURL)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			tab.WithRow(tmplName, absURL, fmt.Sprintf("%t", exists), fmt.Sprintf("%t", parsed), errStr)
		}
		{
			assets := []string{}
			assets = append(assets, tmpl.Assets.Assertions...)
			assets = append(assets, tmpl.Assets.IdentityData...)
			assets = append(assets, tmpl.Assets.DomainData...)

			for _, assetURL := range assets {
				absURL := tmpl.AbsURL(assetURL)
				exists, parsed, err := validateJSON(absURL)
				errStr := ""
				if err != nil {
					errStr = err.Error()
				}
				tab.WithRow(tmplName, absURL, fmt.Sprintf("%t", exists), fmt.Sprintf("%t", parsed), errStr)
			}
		}
	}

	tab.Do()

	return nil
}

func validateManifest(absURL string) (exists, parsed bool, err error) {
	b, err := getBytes(absURL)
	if err != nil {
		return false, false, err
	}

	if _, err := v3.Load(bytes.NewReader(b)); err != nil {
		return true, false, err
	}

	return true, true, nil
}

func validateJSON(absURL string) (exists, parsed bool, err error) {
	b, err := getBytes(absURL)
	if err != nil {
		return false, false, err
	}

	var v map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&v); err != nil {
		return true, false, err
	}

	return true, true, nil
}
