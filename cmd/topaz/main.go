package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/samber/lo"
)

const (
	CLIConfigurationFile = "cli_config.json"
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
		kong.Configuration(kong.JSON, cliConfig),
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

	if err := kongCtx.Run(ctx); err != nil {
		kongCtx.FatalIfErrorf(err)
	}

	// only if not version command used.
	if !lo.Contains(kongCtx.Args, "version") {
		// only save on config change.
		if lo.Contains(kongCtx.Args, "configure") || lo.Contains(kongCtx.Args, "-c") || lo.Contains(kongCtx.Args, "--config") {
			if err := ctx.SaveContextConfig(cmd.CLIConfigurationFile); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}
	}
}
