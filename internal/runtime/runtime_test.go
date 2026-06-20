package runtime_test

import (
	"testing"
	"time"

	runtime "github.com/aserto-dev/topaz/internal/runtime"
	"github.com/aserto-dev/topaz/internal/runtime/testutil"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/plugins/bundle"
	"github.com/stretchr/testify/require"
)

func TestEmptyRuntime(t *testing.T) {
	// Arrange
	assert := require.New(t)
	ctx := t.Context()

	r, err := runtime.New(ctx, &runtime.Config{})
	assert.NoError(err)

	// Act
	s := r.Status()

	// Assert
	assert.True(s.Ready)
}

func TestLocalBundle(t *testing.T) {
	// Arrange
	assert := require.New(t)
	ctx := t.Context()

	r, err := runtime.New(ctx, &runtime.Config{
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
	// Arrange
	assert := require.New(t)

	// Act
	_, err := runtime.New(t.Context(), &runtime.Config{
		LocalBundles: runtime.LocalBundlesConfig{
			Paths: []string{testutil.AssetBuiltinsBundle()},
		},
	})

	// Assert
	assert.Error(err)
}

func TestRemoteBundleV0(t *testing.T) {
	// Arrange
	assert := require.New(t)
	ctx := t.Context()

	r, err := runtime.New(ctx, &runtime.Config{
		Config: runtime.OPAConfig{
			Services: map[string]any{
				"acmecorp": map[string]any{
					"url":                             "https://ghcr.io",
					"response_header_timeout_seconds": 5,
					"type":                            "oci",
				},
			},
			Bundles: map[string]*bundle.Source{
				"testbundle": {
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
