package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/Masterminds/semver/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/pkg/errors"

	ver "github.com/aserto-dev/topaz/pkg/version"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
)

const docLink = "https://www.topaz.sh/docs/command-line-interface/topaz-cli/configuration"

const (
	rcOK  int = 0
	rcErr int = 1
)

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}

	os.Exit(run())
}

func run() (exitCode int) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fflag.Init()

	cli := cmd.CLI{}

	cliConfigFile := filepath.Join(cc.GetTopazDir(), common.CLIConfigurationFile)

	oldDBPath := filepath.Join(cc.GetTopazDir(), "db")
	warn, err := checkDBFiles(oldDBPath)
	if err != nil {
		return exitErr(err)
	}

	c, err := cc.NewCommonContext(ctx, cli.NoCheck, cliConfigFile)
	if err != nil {
		return exitErr(err)
	}

	err = checkVersion(c)
	if err != nil {
		return exitErr(err)
	}

	if warn && len(os.Args) == 1 {
		c.Con().Warn().Msg("Detected directory db files in the old data location %q", oldDBPath)
		c.Con().Msg("Check the documentation on how to update your configuration:\n%s", docLink)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return exitErr(errors.Wrap(err, "failed to determine current working directory"))
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
			"topaz_tmpl_dir":     cc.GetTopazTemplateDir(),
			"container_registry": cc.ContainerRegistry(),
			"container_image":    cc.ContainerImage(),
			"container_tag":      cc.ContainerTag(),
			"container_platform": cc.ContainerPlatform(),
			"container_name":     cc.ContainerName(c.Config.Active.ConfigFile),
			"directory_svc":      cc.DirectorySvc(),
			"directory_key":      cc.DirectoryKey(),
			"directory_token":    cc.DirectoryToken(),
			"authorizer_svc":     cc.AuthorizerSvc(),
			"authorizer_key":     cc.AuthorizerKey(),
			"authorizer_token":   cc.AuthorizerToken(),
			"tenant_id":          cc.TenantID(),
			"insecure":           strconv.FormatBool(cc.Insecure()),
			"plaintext":          strconv.FormatBool(cc.Plaintext()),
			"no_check":           strconv.FormatBool(cc.NoCheck()),
			"no_color":           strconv.FormatBool(cc.NoColor()),
			"active_config":      c.Config.Active.Config,
			"cwd":                cwd,
			"timeout":            cc.Timeout().String(),
		},
	)
	zerolog.SetGlobalLevel(logLevel(cli.LogLevel))

	if cli.NoColor {
		os.Setenv("TOPAZ_NO_COLOR", "TRUE")
	}

	if err := cc.EnsureDirs(); err != nil {
		return exitErr(err)
	}

	if err := kongCtx.Run(c); err != nil {
		return exitErr(err)
	}

	return rcOK
}

func exitErr(err error) int {
	fmt.Fprintln(os.Stderr, err.Error())
	return rcErr
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

// check set version in defaults and suggest update if needed.
func checkVersion(c *cc.CommonCtx) error {
	if cc.ContainerTag() == "latest" {
		return nil
	}

	buildVer, err := semver.NewVersion(ver.GetInfo().Version)
	if err != nil {
		return err
	}
	if buildVer.Prerelease() != "" {
		return nil
	}

	tagVer, err := semver.NewVersion(cc.ContainerTag())
	if err != nil {
		return err
	}

	if buildVer.Major() == tagVer.Major() && buildVer.Minor() == tagVer.Minor() && buildVer.Patch() == tagVer.Patch() {
		return nil
	}

	c.Con().Warn().Msg("The default container tag configuration setting (%s), is different from the current topaz version (%v).",
		c.Config.Defaults.ContainerTag,
		ver.GetInfo().Version,
	)
	if !common.PromptYesNo("Do you want to update the configuration setting?", false) {
		return nil
	}

	c.Config.Defaults.ContainerTag = ver.GetInfo().Version
	return c.SaveContextConfig(common.CLIConfigurationFile)
}
