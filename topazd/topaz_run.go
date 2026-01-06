package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/aserto-dev/topaz/pkg/config/v3"
	"github.com/aserto-dev/topaz/topazd/signals"
	"github.com/aserto-dev/topaz/topazd/topaz"
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

func run(_ *cobra.Command, _ []string) error {
	ctx := signals.SetupSignalHandler()

	cfg, err := config.Load(flagRunConfigFile, config.WithOverrides(configOverrides))
	if err != nil {
		return err
	}

	app, err := topaz.NewTopaz(ctx, cfg)
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
