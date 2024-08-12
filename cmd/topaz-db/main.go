package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/topaz/cmd/topaz-db/cmd"
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
			"directory_svc":   os.Getenv("TOPAZ_DIRECTORY_SVC"),
			"directory_key":   os.Getenv("TOPAZ_DIRECTORY_KEY"),
			"directory_token": "",
			"tenant_id":       os.Getenv("ASERTO_TENANT_ID"),
			"insecure":        strconv.FormatBool(false),
			"no_check":        strconv.FormatBool(false),
		},
	)

	kongCtx.BindTo(ctx, (*context.Context)(nil))

	if err := kongCtx.Run(); err != nil {
		kongCtx.FatalIfErrorf(err)
	}
}
