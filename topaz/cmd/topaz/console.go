package topaz

import (
	"context"
	"strings"

	"github.com/cli/browser"
)

type ConsoleCmd struct {
	ConsoleAddress string `arg:""  default:"https://localhost:8080/ui/directory" help:"gateway address of the console service"`
}

func (cmd *ConsoleCmd) Run(ctx context.Context) error {
	if !strings.HasSuffix(cmd.ConsoleAddress, "/ui/directory") {
		cmd.ConsoleAddress += "/ui/directory"
	}

	if !strings.HasPrefix(cmd.ConsoleAddress, "https://") {
		cmd.ConsoleAddress = "https://" + cmd.ConsoleAddress
	}

	return browser.OpenURL(cmd.ConsoleAddress)
}
