package main

import (
	"os"

	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/topaz/pkg/app/controller"
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
		app, cleanup, err := topaz.BuildApp(os.Stdout, os.Stderr, configPath, func(cfg *config.Config) {
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
				cleanup()
			}
		}()
		if err != nil {
			return err
		}
		directory := topaz.DirectoryResolver(app.Context, app.Logger, app.Configuration)
		decisionlog, err := app.GetDecisionLogger(app.Configuration.DecisionLogger)
		if err != nil {
			return err
		}

		controllerFactory := controller.NewControllerFactory(app.Logger, app.Configuration.ControllerConfig, client.NewDialOptionsProvider())

		runtime, _, err := topaz.NewRuntimeResolver(app.Context, app.Logger, app.Configuration, controllerFactory, decisionlog, directory)
		if err != nil {
			return err
		}
		app.Resolver.SetRuntimeResolver(runtime)
		app.Resolver.SetDirectoryResolver(directory)

		err = app.Start()
		if err != nil {
			return err
		}

		<-app.Context.Done()

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
