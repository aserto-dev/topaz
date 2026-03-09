package main

import (
	"os"

	runtime "github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/app"
	"github.com/aserto-dev/topaz/topazd/app/directory"
	"github.com/aserto-dev/topaz/topazd/app/topaz"
	"github.com/aserto-dev/topaz/topazd/debug"
	zerologdecisionlog "github.com/dagdynamik/topaz-opa-envoy-log-plugin"
	envoy_plugin "github.com/open-policy-agent/opa-envoy-plugin/plugin"
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
	Short: "Start Topaz authorization service with Envoy ext_authz support",
	Long:  `Start instance of the Topaz authorization service with the OPA Envoy ext_authz gRPC plugin.`,
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

		runtime, runtimeCleanup, err := topaz.NewRuntimeResolver(
			topazApp.Context,
			topazApp.Logger,
			topazApp.Configuration,
			dirResolver.GetConn(),
			decisionlog,
			// envoy ext_authz plugin
			runtime.WithPlugin(envoy_plugin.PluginName, &envoy_plugin.Factory{}),
			// zerolog decision log plugin for structured JSON output
			runtime.WithPlugin(zerologdecisionlog.PluginName, zerologdecisionlog.NewFactory(topazApp.Logger)),
		)
		if err != nil {
			return err
		}

		defer runtimeCleanup()

		if authorizer, ok := topazApp.Services["authorizer"].(*app.Authorizer); ok {
			authorizer.Resolver.SetRuntimeResolver(runtime)
			authorizer.Resolver.SetDirectoryResolver(dirResolver)
		}
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

	cfg.DebugService.Enabled = flagRunDebug
}
