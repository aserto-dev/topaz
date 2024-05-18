package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd"
	"github.com/aserto-dev/topaz/pkg/cli/x"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
)

const (
	docLink = "https://www.topaz.sh/docs/command-line-interface/topaz-cli/configuration"
)

func main() {

	cli := cmd.CLI{}

	cliConfigFile := filepath.Join(cc.GetTopazDir(), cmd.CLIConfigurationFile)

	oldDBPath := filepath.Join(cc.GetTopazDir(), "db")
	warn, err := checkDBFiles(oldDBPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	ctx, err := cc.NewCommonContext(cli.NoCheck, cliConfigFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if warn && len(os.Args) == 1 {
		ctx.UI.Exclamation().Msgf("Detected directory db files in the old data location %q\nCheck the documentation on how to update your configuration:\n%s\n",
			oldDBPath, docLink)
	}

	kongCtx := kong.Parse(&cli,
		kong.Name(x.AppName),
		kong.Description(x.AppDescription),
		kong.UsageOnError(),
		kong.Exit(exit),
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
			"topaz_tmpl_dir":     cc.GetTopazTemplateDir(),
			"container_registry": cc.ContainerRegistry(),
			"container_image":    cc.ContainerImage(),
			"container_tag":      cc.ContainerTag(),
			"container_platform": cc.ContainerPlatform(),
			"container_name":     cc.ContainerName(ctx.Config.Active.ConfigFile),
			"directory_svc":      cc.DirectorySvc(),
			"directory_key":      cc.DirectoryKey(),
			"directory_token":    cc.DirectoryToken(),
			"authorizer_svc":     cc.AuthorizerSvc(),
			"authorizer_key":     cc.AuthorizerKey(),
			"authorizer_token":   cc.AuthorizerToken(),
			"tenant_id":          cc.TenantID(),
			"insecure":           strconv.FormatBool(cc.Insecure()),
			"no_check":           strconv.FormatBool(cc.NoCheck()),
		},
	)
	zerolog.SetGlobalLevel(logLevel(cli.LogLevel))

	if err := cc.EnsureDirs(); err != nil {
		kongCtx.FatalIfErrorf(err)
	}

	if err := kongCtx.Run(ctx); err != nil {
		kongCtx.FatalIfErrorf(err)
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

func checkDBFiles(topazDBDir string) (bool, error) {
	if _, err := os.Stat(topazDBDir); os.IsNotExist(err) {
		return false, nil
	}
	if topazDBDir == cc.GetTopazDataDir() {
		return false, nil
	}

	files, err := os.ReadDir(topazDBDir)
	if err != nil {
		return false, err
	}

	return len(files) > 0, nil
}

// set status code to 0 when executing with no arguments, help only output.
func exit(rc int) {
	if len(os.Args) == 1 {
		os.Exit(0)
	}
	os.Exit(rc)
}
