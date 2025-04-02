package templates

type DownloadTemplateCmd struct {
	Name         string `arg:"" required:"" help:"template name"`
	Force        bool   `flag:"" short:"f" default:"false" required:"false" help:"skip confirmation prompt"`
	TemplatesURL string `optional:"" default:"${topaz_tmpl_url}" env:"TOPAZ_TMPL_URL" help:"URL of template catalog"`
	ConfigName   string `optional:"" help:"set config name"`
}
