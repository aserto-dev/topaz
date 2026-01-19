package templates

import (
	"bytes"
	"encoding/json"
	"strconv"

	v3 "github.com/aserto-dev/azm/v3"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/table"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/x"
	"github.com/rs/zerolog"
)

type VerifyTemplateCmd struct {
	Name         string `arg:"" optional:"" help:"template name"`
	Legacy       bool   `optional:"" default:"false" help:"use legacy templates"`
	TemplatesURL string `optional:"" default:"${topaz_tmpl_url}" env:"TOPAZ_TMPL_URL" help:"URL of template catalog"`
}

func (cmd *VerifyTemplateCmd) Run(c *cc.CommonCtx) error {
	if cmd.Legacy {
		cmd.TemplatesURL = x.TopazTmplV32URL
	}

	catalog, err := getCatalog(cmd.TemplatesURL)
	if err != nil {
		return err
	}

	// limit the amount of noise from the azm parser.
	zerolog.SetGlobalLevel(zerolog.Disabled)

	tab := table.New(c.StdOut())
	defer tab.Close()

	tab.Header("template", "asset", "exists", "parsed", "error")

	data := [][]any{}

	for tmplName := range catalog {
		if cmd.Name != "" && tmplName != cmd.Name {
			continue
		}

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

			data = append(data, []any{tmplName, absURL, strconv.FormatBool(v.exists), strconv.FormatBool(v.parsed), errStr})
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

				data = append(data, []any{tmplName, absURL, strconv.FormatBool(v.exists), strconv.FormatBool(v.parsed), errStr})
			}
		}
	}

	tab.Bulk(data)
	tab.Render()

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

	var v map[string]any
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&v); err != nil {
		return valid{exists: true, parsed: false, err: err}
	}

	return valid{exists: true, parsed: true, err: nil}
}
