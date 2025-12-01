package topaz

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/go-grpc/aserto/api/v2"
	runtime "github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/builtins"
	"github.com/aserto-dev/topaz/builtins/az"
	"github.com/aserto-dev/topaz/builtins/ds"
	"github.com/aserto-dev/topaz/controller"
	"github.com/aserto-dev/topaz/decisionlog"
	"github.com/aserto-dev/topaz/pkg/app/management"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	decisionlog_plugin "github.com/aserto-dev/topaz/plugins/decisionlog"
	"github.com/aserto-dev/topaz/plugins/edge"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"
)

var _ resolvers.RuntimeResolver = (*RuntimeResolver)(nil)

type RuntimeResolver struct {
	runtime *runtime.Runtime
}

//nolint:funlen,nestif
func NewRuntimeResolver(
	ctx context.Context,
	logger *zerolog.Logger,
	cfg *config.Config,
	decisionLogger decisionlog.DecisionLogger,
	directoryResolver resolvers.DirectoryResolver,
) (resolvers.RuntimeResolver, func(), error) {
	sidecarRuntime, err := runtime.New(ctx, &cfg.OPA,
		// directory get functions
		runtime.WithBuiltin1(ds.RegisterIdentity(logger, builtins.DSIdentity, directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterUser(logger, builtins.DSUser, directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterObject(logger, builtins.DSObject, directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterRelation(logger, builtins.DSRelation, directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterRelations(logger, builtins.DSRelations, directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterGraph(logger, builtins.DSGraph, directoryResolver)),

		// authorization check functions
		runtime.WithBuiltin1(ds.RegisterCheck(logger, builtins.DSCheck, directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterChecks(logger, builtins.DSChecks, directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterCheckRelation(logger, builtins.DSCheckRelation, directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterCheckPermission(logger, builtins.DSCheckPermission, directoryResolver)),

		// authZen built-ins
		runtime.WithBuiltin1(az.RegisterEvaluation(logger, builtins.AZEvaluation, directoryResolver)),
		runtime.WithBuiltin1(az.RegisterEvaluations(logger, builtins.AZEvaluations, directoryResolver)),
		runtime.WithBuiltin1(az.RegisterSubjectSearch(logger, builtins.AZSubjectSearch, directoryResolver)),
		runtime.WithBuiltin1(az.RegisterResourceSearch(logger, builtins.AZResourceSearch, directoryResolver)),
		runtime.WithBuiltin1(az.RegisterActionSearch(logger, builtins.AZActionSearch, directoryResolver)),

		// plugins
		runtime.WithPlugin(decisionlog_plugin.PluginName, decisionlog_plugin.NewFactory(decisionLogger)),
		runtime.WithPlugin(edge.PluginName, edge.NewPluginFactory(ctx, cfg, logger)),
	)
	if err != nil {
		return nil, func() {}, err
	}

	cleanupRuntime := func() {
		sidecarRuntime.Stop(ctx)
	}

	cleanup := func() {
		if cleanupRuntime != nil {
			cleanupRuntime()
		}
	}

	if cfg.OPA.Config.Discovery != nil {
		host, err := discoveryHostname()
		if err != nil {
			return nil, func() {}, err
		}

		if cfg.OPA.Config.Discovery.Resource == nil {
			return nil, func() {}, aerr.ErrBadRuntime.Msg("discovery resource must be provided")
		}

		details := strings.Split(*cfg.OPA.Config.Discovery.Resource, "/")

		if cfg.ControllerConfig.Server.TenantID == "" {
			cfg.ControllerConfig.Server.TenantID = cfg.OPA.InstanceID // get the tenant id from the opa instance id config.
		}

		if len(details) < 1 {
			return nil, func() {}, aerr.ErrBadRuntime.Msg("provided discovery resource not formatted correctly")
		}

		ctrl, err := controller.NewController(logger, details[0], host, &cfg.ControllerConfig, func(cmdCtx context.Context, cmd *api.Command) error {
			return management.HandleCommand(cmdCtx, cmd, sidecarRuntime)
		})
		if err != nil {
			return nil, func() {}, err
		}

		cleanupController := ctrl.Start(ctx)

		cleanup = func() {
			if cleanupController != nil {
				cleanupController()
			}

			if cleanupRuntime != nil {
				cleanupRuntime()
			}
		}
		if err != nil {
			return nil, cleanup, err
		}
	}

	if err := sidecarRuntime.Start(ctx); err != nil {
		return nil, cleanup, err
	}

	if err := sidecarRuntime.WaitForPlugins(ctx, time.Duration(cfg.OPA.MaxPluginWaitTimeSeconds)*time.Second); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, cleanup, aerr.ErrRuntimeLoading.Err(err).Msg("timeout while waiting for runtime to load")
		}

		return nil, cleanup, aerr.ErrBadRuntime.Err(err)
	}

	return &RuntimeResolver{
		runtime: sidecarRuntime,
	}, cleanup, err
}

func discoveryHostname() (string, error) {
	if host := os.Getenv(x.EnvAsertoHostName); host != "" {
		return host, nil
	}

	if host, err := os.Hostname(); err == nil && host != "" {
		return host, nil
	}

	if host := os.Getenv(x.EnvHostName); host != "" {
		return host, nil
	}

	return "", aerr.ErrBadRuntime.Msg("discovery hostname not set")
}

func (r *RuntimeResolver) RuntimeFromContext(ctx context.Context, policyName string) (*runtime.Runtime, error) {
	return r.GetRuntime(ctx, "", policyName)
}

func (r *RuntimeResolver) GetRuntime(ctx context.Context, opaInstanceID, policyName string) (*runtime.Runtime, error) {
	return r.runtime, nil
}
