package main

import (
	"context"
	"time"

	"github.com/aserto-dev/topaz/pkg/cc/signals"
	"github.com/aserto-dev/topaz/pkg/topaz"
	"github.com/aserto-dev/topaz/pkg/topaz/config"
	"github.com/spf13/cobra"
)

var (
	flagRunConfigFile        string
	flagRunBundleFiles       []string
	flagRunWatchLocalBundles bool
	flagRunIgnorePaths       []string
	flagRunDebug             bool

	gracefulShutdownPeriod = 5 * time.Second
)

var cmdRun = &cobra.Command{
	Use:   "run [args]",
	Short: "Start Topaz authorization service",
	Long:  `Start instance of the Topaz authorization service.`,
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	ctx := signals.SetupSignalHandler()

	app, err := topaz.NewTopaz(ctx, flagRunConfigFile, configOverrides)
	if err != nil {
		return err
	}

	// Start topaz.
	ctx, err = app.Start(ctx)
	if err != nil {
		return err
	}

	// Wait for shutdown signal or for any of the servers to stop unexpectedly.
	<-ctx.Done()

	return stopAndWait(app)
}

func stopAndWait(app *topaz.Topaz) error {
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownPeriod)
	defer cancel()

	return app.Stop(ctx)
}

func configOverrides(cfg *config.Config) {
	if len(flagRunBundleFiles) > 0 {
		cfg.Authorizer.OPA.LocalBundles.Paths = append(cfg.Authorizer.OPA.LocalBundles.Paths, flagRunBundleFiles...)
	}

	if len(flagRunIgnorePaths) > 0 {
		cfg.Authorizer.OPA.LocalBundles.Ignore = append(cfg.Authorizer.OPA.LocalBundles.Ignore, flagRunIgnorePaths...)
	}

	if flagRunWatchLocalBundles {
		cfg.Authorizer.OPA.LocalBundles.Watch = true
	}

	if flagRunDebug {
		cfg.Debug.Enabled = true
	}
}
