package authorizer

import (
	"context"
	"errors"
	"time"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	rt "github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins/az"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins/ds"
	ctrl "github.com/aserto-dev/topaz/topazd/authorizer/controller"
	"github.com/aserto-dev/topaz/topazd/authorizer/decisionlog"
	decisionlog_plugin "github.com/aserto-dev/topaz/topazd/authorizer/plugins/decisionlog"
	"github.com/aserto-dev/topaz/topazd/authorizer/plugins/edge"
	"github.com/aserto-dev/topaz/topazd/authorizer/resolvers"
	"github.com/authzen/access.go/api/access/v1"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
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
	dsConn *grpc.ClientConn,
	edgeFactory *edge.PluginFactory,
) (*RuntimeResolver, error) {
	logger := zerolog.Ctx(ctx)

	dsReader := dsr3.NewReaderClient(dsConn)
	acClient := access.NewAccessClient(dsConn)

	runtime, err := rt.New(ctx, (*rt.Config)(&cfg.OPA),
		// directory get functions
		rt.WithBuiltin1(ds.RegisterIdentity(logger, builtins.DSIdentity, dsReader)),
		rt.WithBuiltin1(ds.RegisterUser(logger, builtins.DSUser, dsReader)),
		rt.WithBuiltin1(ds.RegisterObject(logger, builtins.DSObject, dsReader)),
		rt.WithBuiltin1(ds.RegisterRelation(logger, builtins.DSRelation, dsReader)),
		rt.WithBuiltin1(ds.RegisterRelations(logger, builtins.DSRelations, dsReader)),
		rt.WithBuiltin1(ds.RegisterGraph(logger, builtins.DSGraph, dsReader)),

		// authorization check functions
		rt.WithBuiltin1(ds.RegisterCheck(logger, builtins.DSCheck, dsReader)),
		rt.WithBuiltin1(ds.RegisterChecks(logger, builtins.DSChecks, dsReader)),
		rt.WithBuiltin1(ds.RegisterCheckPermission(logger, builtins.DSCheckPermission, dsReader)),
		rt.WithBuiltin1(ds.RegisterCheckRelation(logger, builtins.DSCheckRelation, dsReader)),

		// authZen built-ins
		rt.WithBuiltin1(az.RegisterEvaluation(logger, builtins.AZEvaluation, acClient)),
		rt.WithBuiltin1(az.RegisterEvaluations(logger, builtins.AZEvaluations, acClient)),
		rt.WithBuiltin1(az.RegisterSubjectSearch(logger, builtins.AZSubjectSearch, acClient)),
		rt.WithBuiltin1(az.RegisterResourceSearch(logger, builtins.AZResourceSearch, acClient)),
		rt.WithBuiltin1(az.RegisterActionSearch(logger, builtins.AZActionSearch, acClient)),

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
