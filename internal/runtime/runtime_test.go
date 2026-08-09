package runtime_test

import (
	"context"
	"os"
	"testing"
	"time"

	runtime "github.com/aserto-dev/topaz/internal/runtime"
	"github.com/aserto-dev/topaz/internal/runtime/testutil"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins/ds"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/download"
	bp "github.com/open-policy-agent/opa/v1/plugins/bundle"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestEmptyRuntime(t *testing.T) {
	assert := require.New(t)
	ctx, cancel := context.WithTimeout(zerolog.New(os.Stderr).WithContext(t.Context()), 60*time.Second)
	t.Cleanup(cancel)

	r, err := runtime.New(ctx, &runtime.Config{})
	assert.NoError(err)

	// Act
	s := r.Status()

	// Assert
	assert.True(s.Ready)
}

func TestLocalBundle(t *testing.T) {
	assert := require.New(t)
	ctx, cancel := context.WithTimeout(zerolog.New(os.Stderr).WithContext(t.Context()), 60*time.Second)
	t.Cleanup(cancel)

	r, err := runtime.New(ctx, &runtime.Config{
		InstanceID:                    "-",
		PluginsErrorLimit:             5,
		GracefulShutdownPeriodSeconds: 2,
		MaxPluginWaitTimeSeconds:      30,
		Flags:                         runtime.Flags{EnableStatusPlugin: true},
		LocalBundles: runtime.LocalBundlesConfig{
			Paths: []string{testutil.AssetSimpleBundle()},
		},
	})
	assert.NoError(err)

	s := r.Status()

	// Assert
	assert.True(s.Ready)
	assert.Empty(s.Errors)
	assert.Len(s.Bundles, 1)
}

func TestFailingLocalBundle(t *testing.T) {
	assert := require.New(t)
	ctx, cancel := context.WithTimeout(zerolog.New(os.Stderr).WithContext(t.Context()), 60*time.Second)
	t.Cleanup(cancel)

	// Act
	_, err := runtime.New(ctx, &runtime.Config{
		InstanceID:                    "-",
		PluginsErrorLimit:             5,
		GracefulShutdownPeriodSeconds: 2,
		MaxPluginWaitTimeSeconds:      30,
		Flags:                         runtime.Flags{EnableStatusPlugin: true},
		LocalBundles: runtime.LocalBundlesConfig{
			Paths: []string{testutil.AssetBuiltinsBundle()},
		},
	})

	// Assert
	assert.Error(err)
}

func TestRemoteBundleV0(t *testing.T) {
	assert := require.New(t)
	ctx, cancel := context.WithTimeout(zerolog.New(os.Stderr).WithContext(t.Context()), 60*time.Second)
	t.Cleanup(cancel)

	r, err := runtime.New(ctx, &runtime.Config{
		InstanceID:                    "-",
		PluginsErrorLimit:             5,
		GracefulShutdownPeriodSeconds: 2,
		MaxPluginWaitTimeSeconds:      30,
		Flags:                         runtime.Flags{EnableStatusPlugin: true},
		Config: runtime.OPAConfig{
			Services: map[string]any{
				"acmecorp": map[string]any{
					"url":                             "https://ghcr.io",
					"response_header_timeout_seconds": 5,
					"type":                            "oci",
				},
			},
			Bundles: map[string]*bp.Source{
				"testBundleV0": {
					Service:  "acmecorp",
					Resource: "ghcr.io/aserto-policies/policy-peoplefinder-rbac:2",
				},
			},
		},
	},
		runtime.WithRegoVersion(ast.RegoV0),
	)

	assert.NoError(err)

	// Act
	assert.NoError(
		r.Start(ctx),
	)
	t.Cleanup(func() { r.Stop(ctx) })

	assert.NoError(
		r.WaitForPlugins(ctx, time.Second*5),
	)

	s := r.Status()

	// Assert
	assert.True(s.Ready)
	assert.Empty(s.Errors)
	assert.Len(s.Bundles, 1)
}

func TestRemoteBundleV1(t *testing.T) {
	assert := require.New(t)
	ctx, cancel := context.WithTimeout(zerolog.New(os.Stderr).WithContext(t.Context()), 60*time.Second)
	t.Cleanup(cancel)

	opts := []runtime.Option{
		runtime.WithBuiltin1(ds.RegisterCheck(&zerolog.Logger{}, builtins.DSCheck, nil)),
		runtime.WithRegoVersion(ast.RegoV1),
	}

	//nolint:gosec // G101: Potential hardcoded credentials.
	r, err := runtime.New(ctx, &runtime.Config{
		InstanceID:                    "-",
		PluginsErrorLimit:             5,
		GracefulShutdownPeriodSeconds: 2,
		MaxPluginWaitTimeSeconds:      30,
		Flags:                         runtime.Flags{EnableStatusPlugin: true},
		Config: runtime.OPAConfig{
			Services: map[string]any{
				"ghcr": map[string]any{
					"url":                             "https://ghcr.io",
					"response_header_timeout_seconds": 5,
					"type":                            "oci",
					"credentials": map[string]any{
						"bearer": map[string]any{
							"scheme": "Bearer",
							"token":  "${GIT_TOKEN}",
						},
					},
				},
			},
			Bundles: map[string]*bp.Source{
				"testBundleV1": {
					Service:  "ghcr",
					Resource: "ghcr.io/aserto-policies/policy-rebac:latest",
					Persist:  false,
					Config: download.Config{
						Polling: download.PollingConfig{
							MinDelaySeconds: new(int64(60)),
							MaxDelaySeconds: new(int64(120)),
						},
					},
				},
			},
		},
	},
		opts...,
	)

	assert.NoError(err)

	// Act
	assert.NoError(
		r.Start(ctx),
	)
	t.Cleanup(func() { r.Stop(ctx) })

	assert.NoError(
		r.WaitForPlugins(ctx, time.Second*5),
	)

	s := r.Status()

	// Assert
	assert.True(s.Ready)
	assert.Empty(s.Errors)
	assert.Len(s.Bundles, 1)
}
