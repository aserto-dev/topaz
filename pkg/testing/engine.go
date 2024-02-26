package testing

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
	Engine      *app.Topaz
	LogDebugger *LogDebugger
	cleanup     func()
	t           *testing.T
	topazDir    string
}

// Cleanup releases all resources the harness uses and
// shuts down servers and runtimes.
func (h *EngineHarness) Cleanup() {
	assert := require.New(h.t)

	h.Engine.Manager.StopServers(h.Context())

	h.cleanup()

	assert.Eventually(func() bool {
		return !PortOpen("0.0.0.0:9494")
	}, ten*time.Second, ten*time.Millisecond)
	assert.Eventually(func() bool {
		return !PortOpen("0.0.0.0:8383")
	}, ten*time.Second, ten*time.Millisecond)
	assert.Eventually(func() bool {
		return !PortOpen("0.0.0.0:8282")
	}, ten*time.Second, ten*time.Millisecond)
	assert.Eventually(func() bool {
		return !PortOpen("0.0.0.0:9292")
	}, ten*time.Second, ten*time.Millisecond)
	assert.Eventually(func() bool {
		return !PortOpen("0.0.0.0:9393")
	}, ten*time.Second, ten*time.Millisecond)
	assert.Eventually(func() bool {
		return !PortOpen("0.0.0.0:9494")
	}, ten*time.Second, ten*time.Millisecond)
	assert.Eventually(func() bool {
		return !PortOpen("0.0.0.0:9696")
	}, ten*time.Second, ten*time.Millisecond)

}

func (h *EngineHarness) Context() context.Context {
	return context.Background()
}

func (h *EngineHarness) TopazDir() string {
	return h.topazDir
}

func (h *EngineHarness) TopazCfgDir() string {
	return path.Join(h.topazDir, "cfg")
}

func (h *EngineHarness) TopazCertsDir() string {
	return path.Join(h.topazDir, "certs")
}

func (h *EngineHarness) TopazDataDir() string {
	return path.Join(h.topazDir, "db")
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

	topazDir := path.Join(t.TempDir(), "topaz")
	err := os.MkdirAll(topazDir, 0777)
	assert.NoError(err)
	assert.DirExists(topazDir)

	err = os.Setenv("TOPAZ_DIR", topazDir)
	assert.NoError(err)

	h := &EngineHarness{
		t:           t,
		LogDebugger: NewLogDebugger(t, "topaz"),
		topazDir:    topazDir,
	}

	configFile := AssetDefaultConfigLocal()
	if online {
		configFile = AssetDefaultConfigOnline()
	}

	topazCertsDir := path.Join(topazDir, "certs")
	err = os.MkdirAll(topazCertsDir, 0777)
	assert.NoError(err)
	assert.DirExists(topazCertsDir)

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
	if _, ok := h.Engine.Services["authorizer"]; ok {
		h.Engine.Services["authorizer"].(*app.Authorizer).Resolver.SetRuntimeResolver(rt)
		h.Engine.Services["authorizer"].(*app.Authorizer).Resolver.SetDirectoryResolver(directory)
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
