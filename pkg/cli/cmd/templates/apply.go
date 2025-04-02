package templates

import (
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
)

type ApplyTemplateCmd struct {
	Name  string `arg:"" required:"" help:"template name"`
	Force bool   `flag:"" short:"f" default:"false" required:"false" help:"skip confirmation prompt"`
	dsc.Config
}
