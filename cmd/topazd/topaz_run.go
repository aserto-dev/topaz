package main

import (
	"os"

	"github.com/aserto-dev/aserto-management/controller"
	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/app/directory"
	"github.com/aserto-dev/topaz/pkg/app/topaz"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/spf13/cobra"
)

var (
	flagRunConfigFile        string
	flagRunBundleFiles       []string
	flagRunWatchLocalBundles bool
	flagRunIgnorePaths       []string
	flagRunDebug             bool
	debugService             *debug.Server
)

var cmdRun = &cobra.Command{
	Use:   "run [args]",
	Short: "Start Topaz authorization service",
	Long:  `Start instance of the Topaz authorization service.`,
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	configPath := config.Path(flagRunConfigFile)
	topazApp, cleanup, err := topaz.BuildApp(os.Stdout, os.Stderr, configPath, configOverrides)
	if err != nil {
		return err
	}
	defer topazApp.Manager.StopServers(topazApp.Context)
	defer cleanup()

	if err := topazApp.ConfigServices(); err != nil {
		return err
	}

	if topazApp.Configuration.DebugService.Enabled {
		debugService = debug.NewServer(&topazApp.Configuration.DebugService, topazApp.Logger)
		debugService.Start()
		defer debugService.Stop()
	}

	if _, ok := topazApp.Services["authorizer"]; ok {
		dirResolver, err := directory.NewResolver(topazApp.Logger, &topazApp.Configuration.DirectoryResolver)
		if err != nil {
			return err
		}
		defer dirResolver.Close()

		decisionlog, err := topazApp.GetDecisionLogger(topazApp.Configuration.DecisionLogger)
		if err != nil {
			return err
		}
		defer decisionlog.Shutdown()

		controllerFactory := controller.NewFactory(
			topazApp.Logger,
			topazApp.Configuration.ControllerConfig,
			app.KeepAliveDialOption(),
		)

		runtime, runtimeCleanup, err := topaz.NewRuntimeResolver(
			topazApp.Context,
			topazApp.Logger,
			topazApp.Configuration,
			controllerFactory,
			decisionlog,
			dirResolver,
		)
		if err != nil {
			return err
		}

		defer runtimeCleanup()

		topazApp.Services["authorizer"].(*app.Authorizer).Resolver.SetRuntimeResolver(runtime)
		topazApp.Services["authorizer"].(*app.Authorizer).Resolver.SetDirectoryResolver(dirResolver)
	}

	err = topazApp.Start()
	if err != nil {
		return err
	}

	<-topazApp.Context.Done()

	return nil
}

func configOverrides(cfg *config.Config) {
	cfg.Command.Mode = config.CommandModeRun

	if len(flagRunBundleFiles) > 0 {
		cfg.OPA.LocalBundles.Paths = append(cfg.OPA.LocalBundles.Paths, flagRunBundleFiles...)
	}

	if len(flagRunIgnorePaths) > 0 {
		cfg.OPA.LocalBundles.Ignore = append(cfg.OPA.LocalBundles.Ignore, flagRunIgnorePaths...)
	}

	if flagRunWatchLocalBundles {
		cfg.OPA.LocalBundles.Watch = true
	}

	cfg.Common.DebugService.Enabled = flagRunDebug
}

// nolint: gochecknoinits, errcheck
func init() {
	cmdRun.Flags().StringVarP(
		&flagRunConfigFile,
		"config-file", "c", "",
		"set path of configuration file")
	cmdRun.Flags().StringSliceVarP(
		&flagRunBundleFiles,
		"bundle", "b", []string{},
		"load paths as bundle files or root directories (can be specified more than once)")
	cmdRun.Flags().BoolVarP(
		&flagRunWatchLocalBundles,
		"watch", "w", false,
		"if set, local changes to bundle paths trigger a reload")
	cmdRun.Flags().StringSliceVarP(
		&flagRunIgnorePaths,
		"ignore", "", []string{},
		"set file and directory names to ignore during loading local bundles (e.g., '.*' excludes hidden files)")
	cmdRun.Flags().BoolVarP(
		&flagRunDebug,
		"debug", "", false,
		"start debug service")
	rootCmd.AddCommand(cmdRun)
	cmdRun.MarkFlagRequired("config-file")
}
