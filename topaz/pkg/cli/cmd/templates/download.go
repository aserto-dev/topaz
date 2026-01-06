package templates

import (
	"path"

	"github.com/aserto-dev/topaz/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/x"
	"github.com/pkg/errors"
)

type DownloadTemplateCmd struct {
	Name         string `arg:"" required:"" help:"template name"`
	Force        bool   `flag:"" short:"f" default:"false" required:"false" help:"skip confirmation prompt"`
	Legacy       bool   `optional:"" default:"false" help:"use legacy templates"`
	TemplatesURL string `optional:"" default:"${topaz_tmpl_url}" env:"TOPAZ_TMPL_URL" help:"URL of template catalog"`
	ConfigName   string `optional:"" help:"set config name"`
}

func (cmd *DownloadTemplateCmd) Run(c *cc.CommonCtx) error {
	if cmd.ConfigName == "" {
		cmd.ConfigName = cmd.Name
	}

	if cmd.ConfigName != "" && !common.RestrictedNamePattern.MatchString(cmd.ConfigName) {
		return errors.Errorf("%s name must match pattern %s", cmd.Name, common.RestrictedNamePattern.String())
	}

	topazTemplateDir := cc.GetTopazTemplateDir()

	if cmd.Legacy {
		cmd.TemplatesURL = x.TopazTmplV32URL
	}

	catalog, err := getCatalog(cmd.TemplatesURL)
	if err != nil {
		return err
	}

	if _, ok := catalog[cmd.Name]; !ok {
		return errors.Errorf("template %s not found", cmd.Name)
	}

	tmpl, err := getTemplate(cmd.Name, cmd.TemplatesURL)
	if err != nil {
		return err
	}

	// manifest
	{
		manifestDir := path.Join(topazTemplateDir, cmd.ConfigName, "model")

		s, err := download(tmpl.AbsURL(tmpl.Assets.Manifest), manifestDir)
		if err != nil {
			return err
		}

		c.Con().Msg(s)
	}

	// data
	{
		dataDir := path.Join(topazTemplateDir, cmd.ConfigName, "data")

		// identity data
		{
			for _, url := range tmpl.Assets.IdentityData {
				s, err := download(tmpl.AbsURL(url), dataDir)
				if err != nil {
					return err
				}

				c.Con().Msg(s)
			}
		}

		// domain data
		{
			for _, url := range tmpl.Assets.DomainData {
				s, err := download(tmpl.AbsURL(url), dataDir)
				if err != nil {
					return err
				}

				c.Con().Msg(s)
			}
		}
	}

	// test data
	{
		testDir := path.Join(topazTemplateDir, cmd.ConfigName, "assertions")

		for _, url := range tmpl.Assets.Assertions {
			s, err := download(tmpl.AbsURL(url), testDir)
			if err != nil {
				return err
			}

			c.Con().Msg(s)
		}
	}

	return nil
}
