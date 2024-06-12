package testing

import (
	"context"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/app/topaz"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	h.Engine.Manager.StopServers(ctx)
	cancel()
	h.cleanup()

	time.Sleep(ten * time.Second) // wait for graceful shutdown before checking ports
	assert.NoError(h.WaitForPorts(cc.PortClosed))
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

func (h *EngineHarness) WaitForPorts(expectedStatus cc.PortStatus) error {
	portMap := map[int]struct{}{}

	for _, svc := range h.Engine.Configuration.APIConfig.Services {

		grpcAddr, err := net.ResolveTCPAddr("tcp", svc.GRPC.ListenAddress)
		if err == nil && grpcAddr.Port != 0 {
			if _, ok := portMap[grpcAddr.Port]; !ok {
				portMap[grpcAddr.Port] = struct{}{}
			}
		}

		gtwAddr, err := net.ResolveTCPAddr("tcp", svc.Gateway.ListenAddress)
		if err == nil && gtwAddr.Port != 0 {
			if _, ok := portMap[gtwAddr.Port]; !ok {
				portMap[gtwAddr.Port] = struct{}{}
			}
		}
	}

	healthAddr, err := net.ResolveTCPAddr("tcp", h.Engine.Configuration.APIConfig.Health.ListenAddress)
	if err == nil && healthAddr.Port != 0 {
		if _, ok := portMap[healthAddr.Port]; !ok {
			portMap[healthAddr.Port] = struct{}{}
		}
	}

	metricsAddr, err := net.ResolveTCPAddr("tcp", h.Engine.Configuration.APIConfig.Metrics.ListenAddress)
	if err == nil && metricsAddr.Port != 0 {
		if _, ok := portMap[metricsAddr.Port]; !ok {
			portMap[metricsAddr.Port] = struct{}{}
		}
	}

	ports := lo.MapToSlice(portMap, func(port int, _ struct{}) string {
		return strconv.Itoa(port)
	})

	wErr := cc.WaitForPorts(ports, expectedStatus)

	return wErr
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
	err = os.Setenv("TOPAZ_CERTS_DIR", filepath.Join(topazDir, "certs"))
	assert.NoError(err)
	err = os.Setenv("TOPAZ_DB_DIR", filepath.Join(topazDir, "db"))
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
	rt, _, err := topaz.NewRuntimeResolver(h.Engine.Context, h.Engine.Logger, h.Engine.Configuration, nil, decisionlog, directory,h.Engine)
	assert.NoError(err)
	err = h.Engine.ConfigServices()
	assert.NoError(err)
	if _, ok := h.Engine.Services["authorizer"]; ok {
		h.Engine.Services["authorizer"].(*app.Authorizer).Resolver.SetRuntimeResolver(rt)
		h.Engine.Services["authorizer"].(*app.Authorizer).Resolver.SetDirectoryResolver(directory)
	}

	h.Engine.Manager = h.Engine.Manager.WithShutdownTimeout(ten) // set shutdown timeout in engine manager
	err = h.Engine.Manager.StartServers(h.Context())
	assert.NoError(err)

	curLevel := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	t.Cleanup(func() {
		zerolog.SetGlobalLevel(curLevel)
	})

	assert.NoError(h.WaitForPorts(cc.PortOpened))

	if online {
		for i := 0; i < 2; i++ {
			assert.Eventually(func() bool {
				return h.LogDebugger.Contains("Bundle loaded and activated successfully")
			}, ten*time.Second, ten*time.Millisecond)
		}
	}

	return h
}
