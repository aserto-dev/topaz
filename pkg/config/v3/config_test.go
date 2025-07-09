package config_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/pkg/config/v3"

	assrt "github.com/stretchr/testify/assert"
	rqur "github.com/stretchr/testify/require"
)

const (
	dbDir    = "/db"
	certsDir = "/certs"
	topazDor = "/topaz"
)

func TestLoadConfigV3(t *testing.T) {
	setTopazEnv(t)

	require := rqur.New(t)
	assert := assrt.New(t)

	r, err := os.Open("schema/config.yaml")
	require.NoError(err)
	t.Cleanup(func() { _ = r.Close() })

	cfg3, err := config.NewConfig(r)
	require.NoError(err)

	// Log config if test fails
	logConfig(t, cfg3)

	assert.Equal(config.Version, cfg3.Version)
	assert.NotEmpty(cfg3.Logging)

	// Authentication
	assert.False(cfg3.Authentication.Enabled)
	assert.Equal("local", cfg3.Authentication.Provider)
	assert.Len(cfg3.Authentication.Local.Keys, 2)
	assert.False(cfg3.Authentication.Local.Options.Default.AllowAnonymous)
	assert.Len(cfg3.Authentication.Local.Options.Overrides, 1)
	assert.Len(cfg3.Authentication.Local.Options.Overrides[0].Paths, 3)
	assert.True(cfg3.Authentication.Local.Options.Overrides[0].Override.AllowAnonymous)

	// Debug
	assert.False(cfg3.Debug.Enabled)
	assert.Equal("0.0.0.0:6666", cfg3.Debug.ListenAddress)

	// Health
	assert.True(cfg3.Health.Enabled)
	assert.Equal("0.0.0.0:9494", cfg3.Health.ListenAddress)

	// Metrics
	assert.True(cfg3.Metrics.Enabled)
	assert.Equal("0.0.0.0:9696", cfg3.Metrics.ListenAddress)

	// Servers
	assert.Len(cfg3.Servers, 3)
}

func setTopazEnv(t *testing.T) {
	t.Helper()

	t.Setenv(x.EnvTopazDBDir, dbDir)
	t.Setenv(x.EnvTopazCertsDir, certsDir)
	t.Setenv(x.EnvTopazDir, topazDor)
}

func logConfig(t *testing.T, cfg3 *config.Config) {
	t.Helper()

	b := bytes.Buffer{}

	rqur.NoError(t,
		cfg3.Serialize(&b),
	)

	t.Logf("Config:\n%s", b.String())
}
