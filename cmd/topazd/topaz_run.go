package main

import (
	"os"

	"github.com/aserto-dev/aserto-management/controller"
	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/app/topaz"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/plugins/edge"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var (
	flagRunConfigFile        string
	flagRunBundleFiles       []string
	flagRunWatchLocalBundles bool
	flagRunIgnorePaths       []string
	flagRunDebug             bool
	debugService             *debug.Server
	errGroup                 *errgroup.Group
)

var cmdRun = &cobra.Command{
	Use:   "run [args]",
	Short: "Start Topaz authorization service",
	Long:  `Start instance of the Topaz authorization service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.Path(flagRunConfigFile)
		topazApp, cleanup, err := topaz.BuildApp(os.Stdout, os.Stderr, configPath, func(cfg *config.Config) {
			cfg.Command.Mode = config.CommandModeRun

			if len(flagRunBundleFiles) > 0 {
				cfg.OPA.LocalBundles.Paths = append(cfg.OPA.LocalBundles.Paths, flagRunBundleFiles...)
			}

			if len(flagRunIgnorePaths) > 0 {
				cfg.OPA.LocalBundles.Ignore = append(cfg.OPA.LocalBundles.Paths, flagRunIgnorePaths...)
			}

			if flagRunWatchLocalBundles {
				cfg.OPA.LocalBundles.Watch = true
			}

			if flagRunDebug {
				cfg.Debug.Enabled = true
				cfg.Debug.ListenAddress = "localhost:6060"
				cfg.Debug.ShutdownTimeout = 5
			}

		})
		defer func() {
			if cleanup != nil {
				topazApp.Manager.StopServers(topazApp.Context)
				cleanup()
			}
		}()
		if err != nil {
			return err
		}
		err = topazApp.ConfigServices()
		if err != nil {
			return err
		}

		if topazApp.Configuration.Debug.Enabled {
			errGroup = new(errgroup.Group)
			debugService = debug.NewServer(&topazApp.Configuration.Debug, topazApp.Logger, errGroup)
			debugService.Start()
		}

		if _, ok := topazApp.Services["authorizer"]; ok {
			directory := topaz.DirectoryResolver(topazApp.Context, topazApp.Logger, topazApp.Configuration)
			decisionlog, err := topazApp.GetDecisionLogger(topazApp.Configuration.DecisionLogger)
			if err != nil {
				return err
			}

			controllerFactory := controller.NewFactory(
				topazApp.Logger,
				topazApp.Configuration.ControllerConfig,
				app.KeepAliveDialOptionsProvider(),
			)
			runtime, runtimeCleanup, err := topaz.NewRuntimeResolver(
				topazApp.Context,
				topazApp.Logger,
				topazApp.Configuration,
				controllerFactory,
				decisionlog,
				directory,
				topazApp,
			)
			if err != nil {
				return err
			}
			defer runtimeCleanup()
			topazApp.Services["authorizer"].(*app.Authorizer).Resolver.SetRuntimeResolver(runtime)
			topazApp.Services["authorizer"].(*app.Authorizer).Resolver.SetDirectoryResolver(directory)

		}

		err = topazApp.Start()
		if err != nil {
			return err
		}
		if _, edgePluginUsed := topazApp.Configuration.OPA.Config.Plugins[edge.PluginName]; edgePluginUsed{
			topazApp.SetServiceStatus("sync", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		}

		<-topazApp.Context.Done()

		debugService.Stop()

		return nil
	},
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
