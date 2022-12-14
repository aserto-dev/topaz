package topaz

import (
	"context"
	"errors"
	"time"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	runtime "github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/builtins/edge/ds"
	decisionlog "github.com/aserto-dev/topaz/decision_log"
	decisionlog_plugin "github.com/aserto-dev/topaz/decision_log/plugin"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"
)

var _ resolvers.RuntimeResolver = (*RuntimeResolver)(nil)

type RuntimeResolver struct {
	runtime *runtime.Runtime
}

func NewRuntimeResolver(
	ctx context.Context,
	logger *zerolog.Logger,
	cfg *config.Config,
	decisionLogger decisionlog.DecisionLogger,
	directoryResolver resolvers.DirectoryResolver) (resolvers.RuntimeResolver, func(), error) {

	sidecarRuntime, cleanupRuntime, err := runtime.NewRuntime(ctx, logger, &cfg.OPA,
		// directory get functions
		runtime.WithBuiltin1(ds.RegisterIdentity(logger, "ds.identity", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterUser(logger, "ds.user", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterObject(logger, "ds.object", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterRelation(logger, "ds.relation", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterGraph(logger, "ds.graph", directoryResolver)),

		// authorization check functions
		runtime.WithBuiltin1(ds.RegisterCheckRelation(logger, "ds.check_relation", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterCheckPermission(logger, "ds.check_permission", directoryResolver)),

		// plugins
		runtime.WithPlugin(decisionlog_plugin.PluginName, decisionlog_plugin.NewFactory(decisionLogger)),
	)
	if err != nil {
		return nil, cleanupRuntime, err
	}

	err = sidecarRuntime.Start(ctx)
	if err != nil {
		return nil, cleanupRuntime, err
	}

	err = sidecarRuntime.WaitForPlugins(ctx, time.Duration(cfg.OPA.MaxPluginWaitTimeSeconds)*time.Second)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, cleanupRuntime, aerr.ErrRuntimeLoading.Err(err).Msg("timeout while waiting for runtime to load")
		}
		return nil, cleanupRuntime, aerr.ErrBadRuntime.Err(err)
	}

	return &RuntimeResolver{
		runtime: sidecarRuntime,
	}, cleanupRuntime, err
}

func (r *RuntimeResolver) RuntimeFromContext(ctx context.Context, policyName, instanceLabel string) (*runtime.Runtime, error) {
	return r.GetRuntime(ctx, "", policyName, instanceLabel)
}

func (r *RuntimeResolver) GetRuntime(ctx context.Context, opaInstanceID, policyName, instanceLabel string) (*runtime.Runtime, error) {
	return r.runtime, nil
}

func (r *RuntimeResolver) PeekRuntime(ctx context.Context, opaInstanceID, policyName, instanceLabel string) (*runtime.Runtime, error) {
	return r.runtime, nil
}

func (r *RuntimeResolver) ReloadRuntime(ctx context.Context, opaInstanceID, policyName, instanceLabel string) error {
	return nil
}

func (r *RuntimeResolver) ListRuntimes(ctx context.Context) (map[string]*runtime.Runtime, error) {
	if r.runtime == nil {
		return nil, nil
	}

	return map[string]*runtime.Runtime{r.runtime.Config.InstanceID: r.runtime}, nil
}

func (r *RuntimeResolver) UnloadRuntime(ctx context.Context, opaInstanceID, policyName, instanceLabel string) {
}
