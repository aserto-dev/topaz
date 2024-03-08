package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd"
	"github.com/aserto-dev/topaz/pkg/cli/x"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
)

func main() {

	cli := cmd.CLI{}

	cliConfigFile := filepath.Join(cc.GetTopazDir(), cmd.CLIConfigurationFile)

	ctx, err := cc.NewCommonContext(cli.NoCheck, cliConfigFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	kongCtx := kong.Parse(&cli,
		kong.Name(x.AppName),
		kong.Description(x.AppDescription),
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
			"topaz_dir":          cc.GetTopazDir(),
			"topaz_certs_dir":    cc.GetTopazCertsDir(),
			"topaz_cfg_dir":      cc.GetTopazCfgDir(),
			"topaz_db_dir":       cc.GetTopazDataDir(),
			"container_registry": cc.ContainerRegistry(),
			"container_image":    cc.ContainerImage(),
			"container_tag":      cc.ContainerTag(),
			"container_platform": cc.ContainerPlatform(),
			"container_name":     cc.ContainerName(ctx.Config.TopazConfigFile),
		},
	)
	zerolog.SetGlobalLevel(logLevel(cli.LogLevel))

	if err := kongCtx.Run(ctx); err != nil {
		kongCtx.FatalIfErrorf(err)
	}

	// only save on config change.
	if _, ok := ctx.Context.Value(cmd.Save).(bool); ok {
		if err := ctx.SaveContextConfig(cmd.CLIConfigurationFile); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
}

func logLevel(level int) zerolog.Level {
	switch level {
	case 0:
		return zerolog.Disabled
	case 1:
		return zerolog.InfoLevel
	case 2:
		return zerolog.WarnLevel
	case 3:
		return zerolog.ErrorLevel
	case 4:
		return zerolog.DebugLevel
	case 5:
		return zerolog.TraceLevel
	default:
		return zerolog.Disabled
	}
}
