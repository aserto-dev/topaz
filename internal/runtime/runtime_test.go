//nolint:funlen,goconst,dupl
package runtime_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aserto-dev/topaz/internal/builtins"
	"github.com/aserto-dev/topaz/internal/builtins/az"
	"github.com/aserto-dev/topaz/internal/builtins/ds"
	runtime "github.com/aserto-dev/topaz/internal/runtime"
	"github.com/aserto-dev/topaz/internal/runtime/testutil"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/authorizer/plugins/edge"
	"github.com/aserto-dev/topaz/topazd/authorizer/plugins/topaz_file_decision_logger"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/download"
	"github.com/open-policy-agent/opa/v1/plugins/bundle"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

const defaultTestContextTimeout = 60 * time.Second

func testContextTimeout(t *testing.T) time.Duration {
	t.Helper()

	str := os.Getenv("TEST_CONTEXT_TIMEOUT")
	if str == "" {
		return defaultTestContextTimeout
	}

	t.Logf("env TEST_CONTEXT_TIME=%s", str)

	parsed, err := time.ParseDuration(str)
	if err != nil {
		t.Logf("parsing TEST_CONTEXT_TIME failed %q", err.Error())
		return defaultTestContextTimeout
	}

	return parsed
}

func TestEmptyRuntime(t *testing.T) {
	assert := require.New(t)

	ctx, cancel := context.WithTimeout(t.Context(), testContextTimeout(t))
	t.Cleanup(cancel)

	r, err := runtime.New(ctx, &runtime.Config{})
	assert.NoError(err)
	assert.NotNil(r)

	assert.NoError(
		r.Start(ctx),
	)
	t.Cleanup(func() { r.Stop(ctx) })

	assert.NoError(
		r.CheckPluginsStatus(),
	)

	b, err := r.GetBundles(ctx)
	assert.NoError(err)
	assert.Empty(b)
}

func TestLocalBundle(t *testing.T) {
	assert := require.New(t)

	ctx, cancel := context.WithTimeout(t.Context(), testContextTimeout(t))
	t.Cleanup(cancel)

	r, err := runtime.New(ctx, &runtime.Config{
		LocalBundles: runtime.LocalBundlesConfig{
			Paths: []string{testutil.AssetSimpleBundle()},
		},
	})
	assert.NoError(err)
	assert.NotNil(r)

	assert.NoError(
		r.Start(ctx),
	)
	t.Cleanup(func() { r.Stop(ctx) })

	assert.NoError(
		r.CheckPluginsStatus(),
	)

	b, err := r.GetBundles(ctx)
	assert.NoError(err)
	assert.Len(b, 1)
}

func TestFailingLocalBundle(t *testing.T) {
	assert := require.New(t)

	ctx, cancel := context.WithTimeout(t.Context(), testContextTimeout(t))
	t.Cleanup(cancel)

	r, err := runtime.New(ctx, &runtime.Config{
		LocalBundles: runtime.LocalBundlesConfig{
			Paths: []string{testutil.AssetBuiltinsBundle()},
		},
	})
	assert.NoError(err)
	assert.NotNil(r)

	assert.Error(
		r.Start(ctx),
	)
	t.Cleanup(func() { r.Stop(ctx) })

	assert.Error(
		r.CheckPluginsStatus(),
	)

	b, err := r.GetBundles(ctx)
	assert.Error(err)
	assert.Empty(b)
}

func TestRemoteBundleV0(t *testing.T) {
	tok := os.Getenv("GH_TOKEN")
	if tok == "" {
		t.Skipf("GH_TOKEN NOT SET")
	}

	assert := require.New(t)

	var (
		logger *zerolog.Logger
		cfg    *config.Config
		dsConn *grpc.ClientConn
	)

	dsClient := builtins.GetDirectoryClient(dsConn)
	acClient := builtins.GetAccessClient(dsConn)

	ctx, cancel := context.WithTimeout(t.Context(), testContextTimeout(t))
	t.Cleanup(cancel)

	r, err := runtime.New(
		t.Context(), &runtime.Config{
			Config: runtime.OPAConfig{
				Services: map[string]any{
					"ghcr": map[string]any{
						"url":  "https://ghcr.io",
						"type": "oci",
						"credentials": map[string]any{
							"bearer": map[string]any{
								"scheme": "Bearer",
								"token":  tok,
							},
						},
						"response_header_timeout_seconds": 5,
					},
				},
				Bundles: map[string]*bundle.Source{
					"testbundle": {
						Service:  "ghcr",
						Resource: "ghcr.io/aserto-policies/policy-peoplefinder-rbac:2",
						Persist:  false,
						Config: download.Config{
							Polling: testPollingConfig(),
						},
					},
				},
			},
		},
		runtime.WithBuiltin1(ds.RegisterIdentity(builtins.DSIdentity, dsClient)),
		runtime.WithBuiltin1(ds.RegisterUser(builtins.DSUser, dsClient)),
		runtime.WithBuiltin1(ds.RegisterObject(builtins.DSObject, dsClient)),
		runtime.WithBuiltin1(ds.RegisterRelation(builtins.DSRelation, dsClient)),
		runtime.WithBuiltin1(ds.RegisterRelations(builtins.DSRelations, dsClient)),
		runtime.WithBuiltin1(ds.RegisterGraph(builtins.DSGraph, dsClient)),
		runtime.WithBuiltin1(ds.RegisterCheck(builtins.DSCheck, dsClient)),
		runtime.WithBuiltin1(ds.RegisterChecks(builtins.DSChecks, dsClient)),
		runtime.WithBuiltin1(az.RegisterEvaluation(builtins.AZEvaluation, acClient)),
		runtime.WithBuiltin1(az.RegisterEvaluations(builtins.AZEvaluations, acClient)),
		runtime.WithBuiltin1(az.RegisterSubjectSearch(builtins.AZSubjectSearch, acClient)),
		runtime.WithBuiltin1(az.RegisterResourceSearch(builtins.AZResourceSearch, acClient)),
		runtime.WithBuiltin1(az.RegisterActionSearch(builtins.AZActionSearch, acClient)),
		runtime.WithPlugin(topaz_file_decision_logger.PluginName, topaz_file_decision_logger.NewFactory(ctx)),
		runtime.WithPlugin(edge.PluginName, edge.NewPluginFactory(ctx, cfg, logger)),
		runtime.WithRegoVersion(ast.RegoV0),
	)

	assert.NoError(err)
	assert.NotNil(r)

	assert.NoError(
		r.Start(ctx),
	)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(t.Context(), testContextTimeout(t))
		r.Stop(cleanupCtx)
		cleanupCancel()
	})

	assert.Equal(ast.RegoV0, r.GetPluginsManager().ParserOptions().RegoVersion)

	assert.NoError(
		r.CheckPluginsStatus(),
	)

	b, err := r.GetBundles(ctx)
	assert.NoError(err)
	assert.Len(b, 1)
}

func TestRemoteBundleV1(t *testing.T) {
	tok := os.Getenv("GH_TOKEN")
	if tok == "" {
		t.Skipf("GH_TOKEN NOT SET")
	}

	assert := require.New(t)

	var (
		logger *zerolog.Logger
		cfg    *config.Config
		dsConn *grpc.ClientConn
	)

	dsClient := builtins.GetDirectoryClient(dsConn)
	acClient := builtins.GetAccessClient(dsConn)

	ctx, cancel := context.WithTimeout(t.Context(), testContextTimeout(t))
	t.Cleanup(cancel)

	r, err := runtime.New(
		t.Context(), &runtime.Config{
			Config: runtime.OPAConfig{
				Services: map[string]any{
					"ghcr": map[string]any{
						"url":  "https://ghcr.io",
						"type": "oci",
						"credentials": map[string]any{
							"bearer": map[string]any{
								"scheme": "Bearer",
								"token":  tok,
							},
						},
						"response_header_timeout_seconds": 5,
					},
				},
				Bundles: map[string]*bundle.Source{
					"testbundle": {
						Service:  "ghcr",
						Resource: "ghcr.io/aserto-policies/policy-rebac:latest",
						Persist:  false,
						Config: download.Config{
							Polling: testPollingConfig(),
						},
					},
				},
			},
		},
		runtime.WithBuiltin1(ds.RegisterIdentity(builtins.DSIdentity, dsClient)),
		runtime.WithBuiltin1(ds.RegisterUser(builtins.DSUser, dsClient)),
		runtime.WithBuiltin1(ds.RegisterObject(builtins.DSObject, dsClient)),
		runtime.WithBuiltin1(ds.RegisterRelation(builtins.DSRelation, dsClient)),
		runtime.WithBuiltin1(ds.RegisterRelations(builtins.DSRelations, dsClient)),
		runtime.WithBuiltin1(ds.RegisterGraph(builtins.DSGraph, dsClient)),
		runtime.WithBuiltin1(ds.RegisterCheck(builtins.DSCheck, dsClient)),
		runtime.WithBuiltin1(ds.RegisterChecks(builtins.DSChecks, dsClient)),
		runtime.WithBuiltin1(az.RegisterEvaluation(builtins.AZEvaluation, acClient)),
		runtime.WithBuiltin1(az.RegisterEvaluations(builtins.AZEvaluations, acClient)),
		runtime.WithBuiltin1(az.RegisterSubjectSearch(builtins.AZSubjectSearch, acClient)),
		runtime.WithBuiltin1(az.RegisterResourceSearch(builtins.AZResourceSearch, acClient)),
		runtime.WithBuiltin1(az.RegisterActionSearch(builtins.AZActionSearch, acClient)),
		runtime.WithPlugin(topaz_file_decision_logger.PluginName, topaz_file_decision_logger.NewFactory(ctx)),
		runtime.WithPlugin(edge.PluginName, edge.NewPluginFactory(ctx, cfg, logger)),
		runtime.WithRegoVersion(ast.RegoV1),
	)

	assert.NoError(err)
	assert.NotNil(r)

	assert.NoError(
		r.Start(ctx),
	)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(t.Context(), testContextTimeout(t))
		r.Stop(cleanupCtx)
		cleanupCancel()
	})

	assert.Equal(ast.RegoV1, r.GetPluginsManager().ParserOptions().RegoVersion)

	assert.NoError(
		r.CheckPluginsStatus(),
	)

	b, err := r.GetBundles(ctx)
	assert.NoError(err)
	assert.Len(b, 1)
}

func testPollingConfig() download.PollingConfig {
	return download.PollingConfig{
		MinDelaySeconds:           func() *int64 { v := int64(60); return &v }(),
		MaxDelaySeconds:           func() *int64 { v := int64(120); return &v }(),
		LongPollingTimeoutSeconds: func() *int64 { v := int64(360); return &v }(),
	}
}
