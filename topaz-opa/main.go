package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/aserto-dev/topaz/topaz-opa/internal/plugin"

	"github.com/open-policy-agent/opa/cmd"
)

const brand string = "topaz-opa"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// initialize Topaz builtins & plugin
	plugin.Init()

	if err := cmd.Command(nil, brand).ExecuteContext(ctx); err != nil {
		panic(err)
	}
}
