package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/topaz/cmd/topaz-db/cmd"
	"github.com/aserto-dev/topaz/pkg/cli/x"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cli := cmd.CLI{}

	kongCtx := kong.Parse(&cli,
		kong.Name("topaz-db"),
		kong.Description("topaz database utility"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppSummary:        false,
			Summary:             false,
			Compact:             true,
			Tree:                false,
			FlagsLast:           true,
			Indenter:            kong.SpaceIndenter,
			NoExpandSubcommands: true,
		}),
		kong.Vars{
			"directory_svc":   os.Getenv(x.EnvTopazDirectorySvc),
			"directory_key":   os.Getenv(x.EnvTopazDirectoryKey),
			"directory_token": "",
			"tenant_id":       os.Getenv(x.EnvAsertoTenantID),
			"insecure":        strconv.FormatBool(false),
			"no_check":        strconv.FormatBool(false),
		},
	)

	kongCtx.BindTo(ctx, (*context.Context)(nil))

	if err := kongCtx.Run(); err != nil {
		kongCtx.FatalIfErrorf(err)
	}
}
