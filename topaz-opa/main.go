package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/aserto-dev/topaz/topaz-opa/internal/plugin"
	"github.com/spf13/cobra"

	"github.com/open-policy-agent/opa/cmd"
)

const brand string = "topaz-opa"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// initialize builtins & plugins
	plugin.Init()

	rootCommand := &cobra.Command{
		Use:   brand,
		Short: "OPA with Topaz Directory support",
		Long:  "Extended Open Policy Agent (OPA) with Topaz support.",
	}

	if err := cmd.Command(rootCommand, brand).ExecuteContext(ctx); err != nil {
		panic(err)
	}
}
