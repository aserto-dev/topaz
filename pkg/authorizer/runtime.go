package authorizer

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-grpc/aserto/api/v2"
	rt "github.com/aserto-dev/runtime"

	"github.com/aserto-dev/topaz/builtins/edge/ds"
	ctrl "github.com/aserto-dev/topaz/controller"
	"github.com/aserto-dev/topaz/decisionlog"
	"github.com/aserto-dev/topaz/pkg/app/management"
	"github.com/aserto-dev/topaz/pkg/cli/x"
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

	runtime, err := rt.NewRuntime(ctx, logger, (*rt.Config)(&cfg.OPA),
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

func newController(cfg *Config, logger *zerolog.Logger, runtime *rt.Runtime) (*ctrl.Controller, error) {
	if cfg.OPA.Config.Discovery == nil {
		return &ctrl.Controller{}, nil
	}

	host, err := hostname()
	if err != nil {
		return nil, err
	}

	if cfg.OPA.Config.Discovery.Resource == nil {
		return nil, aerr.ErrBadRuntime.Msg("discovery resource must be provided")
	}

	details := strings.Split(*cfg.OPA.Config.Discovery.Resource, "/")

	if cfg.Controller.Server.TenantID == "" {
		cfg.Controller.Server.TenantID = cfg.OPA.InstanceID // get the tenant id from the opa instance id config.
	}

	if len(details) < 1 {
		return nil, aerr.ErrBadRuntime.Msg("provided discovery resource not formatted correctly")
	}

	return ctrl.NewController(
		logger,
		details[0],
		host,
		(*ctrl.Config)(&cfg.Controller),
		func(cmdCtx context.Context, cmd *api.Command) error {
			return management.HandleCommand(cmdCtx, cmd, runtime)
		},
	)
}

func hostname() (string, error) {
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
