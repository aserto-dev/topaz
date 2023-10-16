package cmd

import (
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/cli/browser"
)

type ConsoleCmd struct {
	ConsoleAdress string `arg:""  default:"https://localhost:8080/ui/directory" help:"gateway address of the console service"`
}

func (cmd *ConsoleCmd) Run(c *cc.CommonCtx) error {
	if !strings.HasSuffix(cmd.ConsoleAdress, "/ui/directory") {
		cmd.ConsoleAdress += "/ui/directory"
	}
	if !strings.HasPrefix(cmd.ConsoleAdress, "https://") {
		cmd.ConsoleAdress = "https://" + cmd.ConsoleAdress
	}
	return browser.OpenURL(cmd.ConsoleAdress)
}
