package topaz

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/aserto-dev/aserto-management/controller"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/go-grpc/aserto/api/v2"
	runtime "github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/builtins/edge/ds"
	decisionlog "github.com/aserto-dev/topaz/decision_log"
	"github.com/aserto-dev/topaz/pkg/app/management"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	decisionlog_plugin "github.com/aserto-dev/topaz/plugins/decision_log"
	"github.com/aserto-dev/topaz/plugins/edge"
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
	ctrlf *controller.Factory,
	decisionLogger decisionlog.DecisionLogger,
	directoryResolver resolvers.DirectoryResolver,
) (resolvers.RuntimeResolver, func(), error) {
	sidecarRuntime, cleanupRuntime, err := runtime.NewRuntime(ctx, logger, &cfg.OPA,
		// directory get functions
		runtime.WithBuiltin1(ds.RegisterIdentity(logger, "ds.identity", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterUser(logger, "ds.user", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterObject(logger, "ds.object", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterRelation(logger, "ds.relation", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterRelations(logger, "ds.relations", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterGraph(logger, "ds.graph", directoryResolver)),

		// authorization check functions
		runtime.WithBuiltin1(ds.RegisterCheck(logger, "ds.check", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterChecks(logger, "ds.checks", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterCheckRelation(logger, "ds.check_relation", directoryResolver)),
		runtime.WithBuiltin1(ds.RegisterCheckPermission(logger, "ds.check_permission", directoryResolver)),

		// plugins
		runtime.WithPlugin(decisionlog_plugin.PluginName, decisionlog_plugin.NewFactory(decisionLogger)),
		runtime.WithPlugin(edge.PluginName, edge.NewPluginFactory(ctx, cfg, logger)),
	)
	if err != nil {
		return nil, cleanupRuntime, err
	}
	cleanup := func() {
		if cleanupRuntime != nil {
			cleanupRuntime()
		}
	}

	if cfg.OPA.Config.Discovery != nil && ctrlf != nil {
		host := os.Getenv("ASERTO_HOSTNAME")
		if host == "" {
			if host, err = os.Hostname(); err != nil {
				host = os.Getenv("HOSTNAME")
			}
		}
		details := strings.Split(*cfg.OPA.Config.Discovery.Resource, "/")
		cleanupController, err := ctrlf.OnRuntimeStarted(ctx, cfg.OPA.InstanceID, "", details[0],

			details[1], host, func(cmdCtx context.Context, cmd *api.Command) error {
				return management.HandleCommand(cmdCtx, cmd, sidecarRuntime)
			})

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

	err = sidecarRuntime.Start(ctx)
	if err != nil {
		return nil, cleanup, err
	}

	err = sidecarRuntime.WaitForPlugins(ctx, time.Duration(cfg.OPA.MaxPluginWaitTimeSeconds)*time.Second)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, cleanup, aerr.ErrRuntimeLoading.Err(err).Msg("timeout while waiting for runtime to load")
		}
		return nil, cleanup, aerr.ErrBadRuntime.Err(err)
	}

	return &RuntimeResolver{
		runtime: sidecarRuntime,
	}, cleanup, err
}

func (r *RuntimeResolver) RuntimeFromContext(ctx context.Context, policyName string) (*runtime.Runtime, error) {
	return r.GetRuntime(ctx, "", policyName)
}

func (r *RuntimeResolver) GetRuntime(ctx context.Context, opaInstanceID, policyName string) (*runtime.Runtime, error) {
	return r.runtime, nil
}

func (r *RuntimeResolver) PeekRuntime(ctx context.Context, opaInstanceID, policyName string) (*runtime.Runtime, error) {
	return r.runtime, nil
}

func (r *RuntimeResolver) ReloadRuntime(ctx context.Context, opaInstanceID, policyName string) error {
	return nil
}

func (r *RuntimeResolver) ListRuntimes(ctx context.Context) (map[string]*runtime.Runtime, error) {
	if r.runtime == nil {
		return nil, nil
	}

	return map[string]*runtime.Runtime{r.runtime.Config.InstanceID: r.runtime}, nil
}

func (r *RuntimeResolver) UnloadRuntime(ctx context.Context, opaInstanceID, policyName string) {
}
