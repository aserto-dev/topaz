package main

import (
	"os"

	"github.com/aserto-dev/aserto-management/controller"
	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/app/topaz"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/spf13/cobra"
)

var (
	flagRunConfigFile        string
	flagRunBundleFiles       []string
	flagRunWatchLocalBundles bool
	flagRunIgnorePaths       []string
)

var cmdRun = &cobra.Command{
	Use:   "run [args]",
	Short: "Start Topaz authorization service",
	Long:  `Start instance of the Topaz authorization service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.Path(flagRunConfigFile)
		authorizer, cleanup, err := topaz.BuildApp(os.Stdout, os.Stderr, configPath, func(cfg *config.Config) {
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
		})
		defer func() {
			if cleanup != nil {
				authorizer.Manager.StopServers(authorizer.Context)
				cleanup()
			}
		}()
		if err != nil {
			return err
		}
		err = authorizer.ConfigServices()
		if err != nil {
			return err
		}
		if _, ok := authorizer.Services["authorizer"]; ok {
			directory := topaz.DirectoryResolver(authorizer.Context, authorizer.Logger, authorizer.Configuration)
			decisionlog, err := authorizer.GetDecisionLogger(authorizer.Configuration.DecisionLogger)
			if err != nil {
				return err
			}

			controllerFactory := controller.NewFactory(authorizer.Logger, authorizer.Configuration.ControllerConfig, client.NewDialOptionsProvider())

			runtime, runtimeCleanup, err := topaz.NewRuntimeResolver(authorizer.Context, authorizer.Logger, authorizer.Configuration, controllerFactory, decisionlog, directory)
			if err != nil {
				return err
			}

			defer runtimeCleanup()

			authorizer.Services["authorizer"].(*app.Topaz).Resolver.SetRuntimeResolver(runtime)
			authorizer.Services["authorizer"].(*app.Topaz).Resolver.SetDirectoryResolver(directory)
		}

		err = authorizer.Start()
		if err != nil {
			return err
		}

		<-authorizer.Context.Done()

		return nil
	},
}

// nolint: gochecknoinits
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

	rootCmd.AddCommand(cmdRun)
}
