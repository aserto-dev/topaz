package runtime

import (
	"context"
	"sync"
	"time"

	"github.com/aserto-dev/topaz/internal/runtime/logger"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/hooks"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/sdk"
	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
	"github.com/open-policy-agent/opa/v1/topdown/cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const defaultInstanceID string = "topazd"

type Runtime struct {
	Logger           *zerolog.Logger
	Config           *Config
	InterQueryCache  cache.InterQueryCache
	opaInstance      *sdk.OPA
	storage          storage.Store
	pluginsManager   *plugins.Manager
	regoVersion      ast.RegoVersion
	plugins          map[string]plugins.Factory
	builtins1        map[*rego.Function]rego.Builtin1
	builtins2        map[*rego.Function]rego.Builtin2
	builtins3        map[*rego.Function]rego.Builtin3
	builtins4        map[*rego.Function]rego.Builtin4
	builtinsDyn      map[*rego.Function]rego.BuiltinDyn
	builtins         []func(*rego.Rego)
	compilerBuiltins map[string]*ast.Builtin
	imports          []string
}

type BundleState struct {
	ID             string
	Revision       string
	LastDownload   time.Time
	LastActivation time.Time
	Errors         []error
}

type State struct {
	Ready   bool
	Errors  []error
	Bundles []BundleState
}

var builtinsLock sync.Mutex

func New(ctx context.Context, cfg *Config, opts ...Option) (*Runtime, error) {
	newLogger := zerolog.Ctx(ctx).With().Str("component", "runtime").Str("instance-id", cfg.InstanceID).Logger()

	if cfg.InstanceID == "" {
		cfg.InstanceID = defaultInstanceID
	}

	rt := &Runtime{
		Logger:           &newLogger,
		Config:           cfg,
		regoVersion:      DefaultRegoVersion.ToAstRegoVersion(),
		plugins:          map[string]plugins.Factory{},
		builtins1:        map[*rego.Function]rego.Builtin1{},
		builtins2:        map[*rego.Function]rego.Builtin2{},
		builtins3:        map[*rego.Function]rego.Builtin3{},
		builtins4:        map[*rego.Function]rego.Builtin4{},
		builtinsDyn:      map[*rego.Function]rego.BuiltinDyn{},
		builtins:         []func(*rego.Rego){},
		compilerBuiltins: map[string]*ast.Builtin{},
		imports:          []string{},
	}

	for _, opt := range opts {
		opt(rt)
	}

	if rt.storage == nil {
		rt.storage = inmem.New()
	}

	rt.registerBuiltins()

	return rt, nil
}

func (r *Runtime) Start(ctx context.Context) error {
	readyChannel := make(chan struct{})

	loadedBundles, err := r.loadPaths([]string{})
	if err != nil {
		return errors.Wrap(err, "local bundle load error")
	}

	opaConfig, err := r.Config.opaConfigReader()
	if err != nil {
		return err
	}

	sdkOpts := sdk.Options{
		ID:            r.Config.InstanceID,
		Config:        opaConfig,
		Logger:        logger.NewOpaLogger(r.Logger),
		ConsoleLogger: logger.NewOpaLogger(r.Logger),
		Plugins:       r.plugins,
		Store:         r.storage,
		Hooks:         hooks.Hooks{},
		V0Compatible:  false,
		V1Compatible:  false,
		RegoVersion:   r.regoVersion,
		ManagerOpts: []func(manager *plugins.Manager){
			plugins.InitBundles(loadedBundles),
			plugins.Info(ast.NewTerm(r.info())),
			plugins.MaxErrors(r.Config.PluginsErrorLimit),
			plugins.WithParserOptions(ast.ParserOptions{RegoVersion: r.regoVersion}),
			plugins.GracefulShutdownPeriod(r.Config.GracefulShutdownPeriodSeconds),
			plugins.Logger(logger.NewOpaLogger(r.Logger)),
			func(manager *plugins.Manager) {
				// copy the plugin manager reference to the runtime struct, for later usage.
				r.pluginsManager = manager
			},
		},
		Ready: readyChannel,
	}

	opaInstance, err := sdk.New(ctx, sdkOpts)
	if err != nil {
		return err
	}

	r.opaInstance = opaInstance

	select {
	case <-readyChannel:
	case <-ctx.Done():
		r.Logger.Error().Err(ctx.Err()).Msg("creating opa instance")
		return ctx.Err()
	}

	r.InterQueryCache = cache.NewInterQueryCache(r.pluginsManager.InterQueryBuiltinCacheConfig())

	return nil
}

func (r *Runtime) Stop(ctx context.Context) {
	if r.opaInstance != nil {
		r.opaInstance.Stop(ctx)
	}
}

func (r *Runtime) GetPluginsManager() *plugins.Manager {
	return r.pluginsManager
}

func (r *Runtime) registerBuiltins() {
	// We shouldn't register global builtins, these should be per runtime.
	// In order for that to work, the plugin manager has to allow us to tell the compiler
	// of our builtins.
	builtinsLock.Lock()

	defer builtinsLock.Unlock()

	for decl, impl := range r.builtins1 {
		r.Logger.Info().Str("name", decl.Name).Msg("registering builtin1")
		rego.RegisterBuiltin1(decl, impl)
	}

	for decl, impl := range r.builtins2 {
		r.Logger.Info().Str("name", decl.Name).Msg("registering builtin2")
		rego.RegisterBuiltin2(decl, impl)
	}

	for decl, impl := range r.builtins3 {
		r.Logger.Info().Str("name", decl.Name).Msg("registering builtin3")
		rego.RegisterBuiltin3(decl, impl)
	}

	for decl, impl := range r.builtins4 {
		r.Logger.Info().Str("name", decl.Name).Msg("registering builtin4")
		rego.RegisterBuiltin4(decl, impl)
	}

	for decl, impl := range r.builtinsDyn {
		r.Logger.Info().Str("name", decl.Name).Msg("registering builtinDyn")
		rego.RegisterBuiltinDyn(decl, impl)
	}
}
