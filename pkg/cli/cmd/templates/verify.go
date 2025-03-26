package templates

import (
	"bytes"
	"encoding/json"
	"strconv"

	v3 "github.com/aserto-dev/azm/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/table"
	"github.com/rs/zerolog"
)

type VerifyTemplateCmd struct {
	TemplatesURL string `optional:"" default:"${topaz_tmpl_url}" env:"TOPAZ_TMPL_URL" help:"URL of template catalog"`
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
			v := validateManifest(absURL)
			errStr := ""
			if v.err != nil {
				errStr = v.err.Error()
			}
			tab.WithRow(tmplName, absURL, strconv.FormatBool(v.exists), strconv.FormatBool(v.parsed), errStr)
		}
		{
			assets := []string{}
			assets = append(assets, tmpl.Assets.Assertions...)
			assets = append(assets, tmpl.Assets.IdentityData...)
			assets = append(assets, tmpl.Assets.DomainData...)

			for _, assetURL := range assets {
				absURL := tmpl.AbsURL(assetURL)
				v := validateJSON(absURL)
				errStr := ""
				if v.err != nil {
					errStr = v.err.Error()
				}
				tab.WithRow(tmplName, absURL, strconv.FormatBool(v.exists), strconv.FormatBool(v.parsed), errStr)
			}
		}
	}

	tab.Do()

	return nil
}

type valid struct {
	exists bool
	parsed bool
	err    error
}

func validateManifest(absURL string) valid {
	b, err := getBytes(absURL)
	if err != nil {
		return valid{exists: false, parsed: false, err: err}
	}

	if _, err := v3.Load(bytes.NewReader(b)); err != nil {
		return valid{exists: true, parsed: false, err: err}
	}

	return valid{exists: true, parsed: true, err: nil}
}

func validateJSON(absURL string) valid {
	b, err := getBytes(absURL)
	if err != nil {
		return valid{exists: false, parsed: false, err: err}
	}

	var v map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&v); err != nil {
		return valid{exists: true, parsed: false, err: err}
	}

	return valid{exists: true, parsed: true, err: nil}
}
