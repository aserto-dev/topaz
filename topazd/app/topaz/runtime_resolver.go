package topaz

import (
	"context"

	"github.com/aserto-dev/topaz/internal/builtins"
	"github.com/aserto-dev/topaz/internal/builtins/az"
	"github.com/aserto-dev/topaz/internal/builtins/ds"
	"github.com/aserto-dev/topaz/internal/runtime"
	"github.com/aserto-dev/topaz/pkg/authorizer/aerr"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/authorizer/plugins/edge"
	"github.com/aserto-dev/topaz/topazd/authorizer/plugins/topaz_file_decision_logger"
	"github.com/aserto-dev/topaz/topazd/authorizer/resolvers"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var _ resolvers.RuntimeResolver = (*RuntimeResolver)(nil)

type RuntimeResolver struct {
	runtime *runtime.Runtime
}

func NewRuntimeResolver(
	ctx context.Context,
	logger *zerolog.Logger,
	cfg *config.Config,
	dsConn *grpc.ClientConn,
) (resolvers.RuntimeResolver, func(), error) {
	dsClient := builtins.GetDirectoryClient(dsConn)
	acClient := builtins.GetAccessClient(dsConn)

	rt, err := runtime.New(
		ctx, &cfg.OPA,

		// directory get functions
		runtime.WithBuiltin1(ds.RegisterIdentity(builtins.DSIdentity, dsClient)),
		runtime.WithBuiltin1(ds.RegisterUser(builtins.DSUser, dsClient)),
		runtime.WithBuiltin1(ds.RegisterObject(builtins.DSObject, dsClient)),
		runtime.WithBuiltin1(ds.RegisterRelation(builtins.DSRelation, dsClient)),
		runtime.WithBuiltin1(ds.RegisterRelations(builtins.DSRelations, dsClient)),
		runtime.WithBuiltin1(ds.RegisterGraph(builtins.DSGraph, dsClient)),

		// authorization check functions
		runtime.WithBuiltin1(ds.RegisterCheck(builtins.DSCheck, dsClient)),
		runtime.WithBuiltin1(ds.RegisterChecks(builtins.DSChecks, dsClient)),

		// authZen built-ins
		runtime.WithBuiltin1(az.RegisterEvaluation(builtins.AZEvaluation, acClient)),
		runtime.WithBuiltin1(az.RegisterEvaluations(builtins.AZEvaluations, acClient)),
		runtime.WithBuiltin1(az.RegisterSubjectSearch(builtins.AZSubjectSearch, acClient)),
		runtime.WithBuiltin1(az.RegisterResourceSearch(builtins.AZResourceSearch, acClient)),
		runtime.WithBuiltin1(az.RegisterActionSearch(builtins.AZActionSearch, acClient)),

		// plugins
		runtime.WithPlugin(topaz_file_decision_logger.PluginName, topaz_file_decision_logger.NewFactory(logger.WithContext(ctx))),
		runtime.WithPlugin(edge.PluginName, edge.NewPluginFactory(ctx, cfg, logger)),

		runtime.WithRegoVersion(ast.RegoV0),
	)
	if err != nil {
		return nil, func() {}, err
	}

	cleanupRuntime := func() {
		rt.Stop(ctx)
	}

	cleanup := func() {
		if cleanupRuntime != nil {
			cleanupRuntime()
		}
	}

	if err := rt.Start(ctx); err != nil {
		return nil, cleanup, err
	}

	if err := rt.CheckPluginsStatus(); err != nil {
		return nil, cleanup, aerr.ErrBadRuntime.Err(err)
	}

	return &RuntimeResolver{
		runtime: rt,
	}, cleanup, err
}

func (r *RuntimeResolver) GetRuntime(ctx context.Context) (*runtime.Runtime, error) {
	return r.runtime, nil
}
