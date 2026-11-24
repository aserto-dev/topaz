package topaz_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aserto-dev/topaz/topaz-opa/internal/config"
	"github.com/aserto-dev/topaz/topaz-opa/internal/errs"
	"github.com/aserto-dev/topaz/topaz-opa/internal/plugins/noop"

	"github.com/aserto-dev/topaz/topaz-opa/internal/plugins/topaz"
	opaconfig "github.com/open-policy-agent/opa/v1/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pluginConfig parses an OPA config file and returns the raw "topaz" plugin
// section, mirroring how OPA hands plugin config to a plugin factory.
func pluginConfig(t *testing.T, file string) ([]byte, bool) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", file))
	require.NoError(t, err)

	cfg, err := opaconfig.ParseConfig(raw, "test")
	require.NoError(t, err)

	section, ok := cfg.Plugins[topaz.PluginName]

	return section, ok
}

// reset clears the package level config so tests do not leak state.
func reset(t *testing.T) {
	t.Helper()

	topaz.ResetConfig()
	t.Cleanup(func() { topaz.ResetConfig() })
}

func requireDisabled(t *testing.T) {
	t.Helper()

	_, err := topaz.GetDirectoryClient()()
	require.ErrorIs(t, err, errs.ErrTopazPluginDisabled)

	_, err = topaz.GetAccessClient()()
	require.ErrorIs(t, err, errs.ErrTopazPluginDisabled)

	_, err = topaz.GetDirectoryConn()()
	require.ErrorIs(t, err, errs.ErrTopazPluginDisabled)
}

// No plugins section: OPA never calls the factory, so the plugin must
// behave as disabled without any configuration having been stored.
func TestConfigNoPlugins(t *testing.T) {
	reset(t)

	_, found := pluginConfig(t, "config-no-plugins.yaml")
	require.False(t, found)

	cfg := topaz.GetConfig()
	assert.False(t, cfg.Enabled)
	assert.Equal(t, config.Duration(10*time.Second), cfg.RequestTimeout)

	requireDisabled(t)
}

// topaz.Plugin explicitly disabled with a partial config.
func TestConfigTopazPluginDisabled(t *testing.T) {
	reset(t)

	section, found := pluginConfig(t, "config-topaz-plugin-disabled.yaml")
	require.True(t, found)

	f := &topaz.PluginFactory{}

	parsed, err := f.Validate(nil, section)
	require.NoError(t, err)

	p := f.New(nil, parsed)
	assert.IsType(t, &noop.Noop{}, p)

	cfg := topaz.GetConfig()
	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.Connection.Address)

	requireDisabled(t)
}

// topaz.Plugin fully configured and enabled.
func TestConfigTopazPluginEnabled(t *testing.T) {
	reset(t)

	section, found := pluginConfig(t, "config-topaz-plugin-enabled.yaml")
	require.True(t, found)

	f := &topaz.PluginFactory{}

	parsed, err := f.Validate(nil, section)
	require.NoError(t, err)

	p := f.New(nil, parsed)
	assert.IsType(t, &topaz.Plugin{}, p)

	cfg := topaz.GetConfig()
	assert.True(t, cfg.Enabled)
	assert.Equal(t, ":9292", cfg.Connection.Address)
	assert.True(t, cfg.Connection.Insecure)
	assert.False(t, cfg.Connection.NoTLS)
	assert.Equal(t, config.Duration(5*time.Second), cfg.RequestTimeout)

	// gRPC connections are lazy, so creating clients must succeed
	// without a running Topaz server.
	dsr, err := topaz.GetDirectoryClient()()
	require.NoError(t, err)
	assert.NotNil(t, dsr)

	acr, err := topaz.GetAccessClient()()
	require.NoError(t, err)
	assert.NotNil(t, acr)
}
