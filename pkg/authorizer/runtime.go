package authorizer

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	rt "github.com/aserto-dev/runtime"

	"github.com/aserto-dev/topaz/builtins/edge/ds"
	ctrl "github.com/aserto-dev/topaz/controller"
	"github.com/aserto-dev/topaz/decisionlog"
	decisionlog_plugin "github.com/aserto-dev/topaz/plugins/decisionlog"
	"github.com/aserto-dev/topaz/plugins/edge"
	"github.com/aserto-dev/topaz/resolvers"
)

var _ resolvers.RuntimeResolver = (*RuntimeResolver)(nil)

type RuntimeResolver struct {
	runtime    *rt.Runtime
	controller *ctrl.Controller
	pluginWait time.Duration
}

func NewRuntimeResolver(
	ctx context.Context,
	cfg *Config,
	decisionLogger decisionlog.DecisionLogger,
	dsReader dsr3.ReaderClient,
	edgeFactory *edge.PluginFactory,
) (*RuntimeResolver, error) {
	logger := zerolog.Ctx(ctx)

	runtime, err := rt.New(ctx, (*rt.Config)(&cfg.OPA),
		// directory get functions
		rt.WithBuiltin1(ds.RegisterIdentity(logger, "ds.identity", dsReader)),
		rt.WithBuiltin1(ds.RegisterUser(logger, "ds.user", dsReader)),
		rt.WithBuiltin1(ds.RegisterObject(logger, "ds.object", dsReader)),
		rt.WithBuiltin1(ds.RegisterRelation(logger, "ds.relation", dsReader)),
		rt.WithBuiltin1(ds.RegisterRelations(logger, "ds.relations", dsReader)),
		rt.WithBuiltin1(ds.RegisterGraph(logger, "ds.graph", dsReader)),

		// authorization check functions
		rt.WithBuiltin1(ds.RegisterCheck(logger, "ds.check", dsReader)),
		rt.WithBuiltin1(ds.RegisterChecks(logger, "ds.checks", dsReader)),

		// plugins
		rt.WithPlugin(decisionlog_plugin.PluginName, decisionlog_plugin.NewFactory(decisionLogger)),
		rt.WithPlugin(edge.PluginName, edgeFactory),
	)
	if err != nil {
		return nil, err
	}

	controller, err := newController(cfg, logger, runtime)
	if err != nil {
		return nil, err
	}

	return &RuntimeResolver{
		runtime:    runtime,
		controller: controller,
		pluginWait: time.Duration(cfg.OPA.MaxPluginWaitTimeSeconds) * time.Second,
	}, err
}

func (r *RuntimeResolver) Start(ctx context.Context) error {
	r.controller.Start(ctx)

	if err := r.runtime.Start(ctx); err != nil {
		return err
	}

	if err := r.runtime.WaitForPlugins(ctx, r.pluginWait); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return aerr.ErrRuntimeLoading.Err(err).Msg("timeout while waiting for plugins to start")
		}

		return aerr.ErrBadRuntime.Err(err)
	}

	return nil
}

func (r *RuntimeResolver) Stop(ctx context.Context) error {
	r.runtime.Stop(ctx)

	return r.controller.Stop()
}

func (r *RuntimeResolver) RuntimeFromContext(ctx context.Context, policyName string) (*rt.Runtime, error) {
	return r.GetRuntime(ctx, "", policyName)
}

func (r *RuntimeResolver) GetRuntime(ctx context.Context, opaInstanceID, policyName string) (*rt.Runtime, error) {
	return r.runtime, nil
}
