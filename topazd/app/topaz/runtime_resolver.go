package topaz

import (
	"context"
	"errors"
	"time"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	runtime "github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins/az"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins/ds"
	decisionlog_plugin "github.com/aserto-dev/topaz/topazd/authorizer/plugins/decisionlog"
	"github.com/aserto-dev/topaz/topazd/authorizer/plugins/edge"
	"github.com/aserto-dev/topaz/topazd/authorizer/resolvers"
	"github.com/authzen/access.go/api/access/v1"
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
	dsClient := reader.NewReaderClient(dsConn)
	acClient := access.NewAccessClient(dsConn)

	sidecarRuntime, err := runtime.New(ctx, &cfg.OPA,

		// directory get functions
		runtime.WithBuiltin1(ds.RegisterIdentity(logger, builtins.DSIdentity, dsClient)),
		runtime.WithBuiltin1(ds.RegisterUser(logger, builtins.DSUser, dsClient)),
		runtime.WithBuiltin1(ds.RegisterObject(logger, builtins.DSObject, dsClient)),
		runtime.WithBuiltin1(ds.RegisterRelation(logger, builtins.DSRelation, dsClient)),
		runtime.WithBuiltin1(ds.RegisterRelations(logger, builtins.DSRelations, dsClient)),
		runtime.WithBuiltin1(ds.RegisterGraph(logger, builtins.DSGraph, dsClient)),

		// authorization check functions
		runtime.WithBuiltin1(ds.RegisterCheck(logger, builtins.DSCheck, dsClient)),
		runtime.WithBuiltin1(ds.RegisterChecks(logger, builtins.DSChecks, dsClient)),

		// authZen built-ins
		runtime.WithBuiltin1(az.RegisterEvaluation(logger, builtins.AZEvaluation, acClient)),
		runtime.WithBuiltin1(az.RegisterEvaluations(logger, builtins.AZEvaluations, acClient)),
		runtime.WithBuiltin1(az.RegisterSubjectSearch(logger, builtins.AZSubjectSearch, acClient)),
		runtime.WithBuiltin1(az.RegisterResourceSearch(logger, builtins.AZResourceSearch, acClient)),
		runtime.WithBuiltin1(az.RegisterActionSearch(logger, builtins.AZActionSearch, acClient)),

		// plugins
		runtime.WithPlugin(decisionlog_plugin.PluginName, decisionlog_plugin.NewFactory()),
		runtime.WithPlugin(edge.PluginName, edge.NewPluginFactory(ctx, cfg, logger)),

		runtime.WithRegoVersion(ast.RegoV0),
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

func (r *RuntimeResolver) RuntimeFromContext(ctx context.Context, policyName string) (*runtime.Runtime, error) {
	return r.GetRuntime(ctx, policyName)
}

func (r *RuntimeResolver) GetRuntime(ctx context.Context, policyName string) (*runtime.Runtime, error) {
	return r.runtime, nil
}
