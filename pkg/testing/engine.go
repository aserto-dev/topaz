package testing

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/app/topaz"
	"github.com/aserto-dev/topaz/pkg/cc/config"
)

const (
	ten = 10
)

// EngineHarness wraps an Aserto Runtime Engine so we can set it up easily
// and monitor its logs.
type EngineHarness struct {
	Engine      *app.Authorizer
	LogDebugger *LogDebugger

	cleanup func()
	t       *testing.T
}

// Cleanup releases all resources the harness uses and
// shuts down servers and runtimes.
func (h *EngineHarness) Cleanup() {
	assert := require.New(h.t)

	h.Engine.Manager.StopServers(h.Context())

	h.cleanup()

	assert.Eventually(func() bool {
		return !PortOpen("127.0.0.1:8484")
	}, ten*time.Second, ten*time.Millisecond)
	assert.Eventually(func() bool {
		return !PortOpen("127.0.0.1:8383")
	}, ten*time.Second, ten*time.Millisecond)
	assert.Eventually(func() bool {
		return !PortOpen("127.0.0.1:8282")
	}, ten*time.Second, ten*time.Millisecond)
}

func (h *EngineHarness) Runtime() *runtime.Runtime {
	if _, ok := h.Engine.Services["topaz"]; ok {
		result, err := h.Engine.Services["topaz"].(*app.Topaz).Resolver.GetRuntimeResolver().RuntimeFromContext(h.Engine.Context, "", "")
		require.NoError(h.t, err)
		return result
	}
	return nil
}

func (h *EngineHarness) Context() context.Context {
	return context.Background()
}

// SetupOffline sets up an engine that uses a runtime that loads offline bundles,
// from the assets directory.
func SetupOffline(t *testing.T, configOverrides func(*config.Config)) *EngineHarness {
	return setup(t, configOverrides, false)
}

// SetupOnline sets up an engine that uses a runtime that loads online bundles,
// from the online policy registry service.
func SetupOnline(t *testing.T, configOverrides func(*config.Config)) *EngineHarness {
	return setup(t, configOverrides, true)
}

func setup(t *testing.T, configOverrides func(*config.Config), online bool) *EngineHarness {
	assert := require.New(t)

	var err error
	h := &EngineHarness{
		t:           t,
		LogDebugger: NewLogDebugger(t, "topaz"),
	}

	configFile := AssetDefaultConfigLocal()
	if online {
		configFile = AssetDefaultConfigOnline()
	}

	err = os.Setenv("TOPAZ_DIR", "/tmp/topaz/test")
	assert.NoError(err)
	if _, err := os.Stat("/tmp/topaz/test"); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll("/tmp/topaz/test", 0777)
		assert.NoError(err)
		err = os.MkdirAll("/tmp/topaz/test/certs", 0777)
		assert.NoError(err)
	}
	h.Engine, h.cleanup, err = topaz.BuildTestApp(
		h.LogDebugger,
		h.LogDebugger,
		configFile,
		configOverrides,
	)
	assert.NoError(err)
	directory := topaz.DirectoryResolver(h.Engine.Context, h.Engine.Logger, h.Engine.Configuration)
	decisionlog, err := h.Engine.GetDecisionLogger(h.Engine.Configuration.DecisionLogger)
	assert.NoError(err)
	rt, _, err := topaz.NewRuntimeResolver(h.Engine.Context, h.Engine.Logger, h.Engine.Configuration, nil, decisionlog, directory)
	assert.NoError(err)
	err = h.Engine.ConfigServices()
	assert.NoError(err)
	if _, ok := h.Engine.Services["topaz"]; ok {
		h.Engine.Services["topaz"].(*app.Topaz).Resolver.SetRuntimeResolver(rt)
		h.Engine.Services["topaz"].(*app.Topaz).Resolver.SetDirectoryResolver(directory)
	}
	err = h.Engine.Start()
	assert.NoError(err)

	if online {
		for i := 0; i < 2; i++ {
			assert.Eventually(func() bool {
				return h.LogDebugger.Contains("Bundle loaded and activated successfully")
			}, ten*time.Second, ten*time.Millisecond)
		}
	}

	assert.Eventually(func() bool {
		return PortOpen("127.0.0.1:8383")
	}, ten*time.Second, ten*time.Millisecond)

	return h
}
