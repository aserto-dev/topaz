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
	"github.com/aserto-dev/topaz/builtins/edge/ds"
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

//nolint:funlen
func NewRuntimeResolver(
	ctx context.Context,
	logger *zerolog.Logger,
	cfg *config.Config,
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

	if cfg.OPA.Config.Discovery != nil {
		host := os.Getenv(x.EnvAsertoHostName)
		if host == "" {
			if host, err = os.Hostname(); err != nil {
				host = os.Getenv(x.EnvHostName)
				if host == "" {
					panic("hostname not set")
				}
			}
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
