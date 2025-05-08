package directory_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/directory"
)

func TestMarshaling(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cfg    string
		verify func(*testing.T, *directory.Store)
	}{
		{"boltdb", boltConfig, func(t *testing.T, s *directory.Store) {
			assert.Equal(t, directory.BoltDBStorePlugin, s.Provider)
			require.NotNil(t, s.Bolt)
			assert.Equal(t, "/path/to/bolt.db", s.Bolt.DBPath)
		}},
		{"remote_directory", remoteConfig, func(t *testing.T, s *directory.Store) {
			assert.Equal(t, directory.RemoteDirectoryStorePlugin, s.Provider)
			require.NotNil(t, s.Remote)
			assert.Equal(t, "localhost:9292", s.Remote.Address)
			assert.Equal(t, "tenant-id", s.Remote.TenantID)
			assert.Equal(t, "api-key", s.Remote.APIKey)
			assert.Equal(t, "token", s.Remote.Token)
			assert.Empty(t, s.Remote.ClientCertPath)
			assert.Empty(t, s.Remote.ClientKeyPath)
			assert.Equal(t, "ca-cert-path", s.Remote.CACertPath)
			assert.False(t, s.Remote.Insecure)
			assert.True(t, s.Remote.NoTLS)
			assert.False(t, s.Remote.NoProxy)
			assert.Contains(t, s.Remote.Headers, "x-foo")
			assert.Equal(t, "foo-value", s.Remote.Headers["x-foo"])
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.TrimN(tc.cfg)

			v := config.NewViper()
			v.ReadConfig(
				strings.NewReader(cfg),
			)

			var c directory.Config
			err := v.Unmarshal(&c)
			require.NoError(t, err)

			tc.verify(t, &c.Store)

			var out bytes.Buffer

			require.NoError(t,
				c.Serialize(&out),
			)

			assert.Equal(t, config.TrimN(preamble)+config.Indent(cfg, 2), out.String())
		})
	}
}

func TestDefaults(t *testing.T) {
	var c directory.Config

	v := config.NewViper()

	v.SetDefaults(&c)
	v.ReadConfig(
		strings.NewReader(""),
	)

	err := v.Unmarshal(&c)
	require.NoError(t, err)

	assert.Equal(t, "5s", c.ReadTimeout.String())
	assert.Equal(t, directory.BoltDBStorePlugin, c.Store.Provider)
	require.NotNil(t, c.Store.Bolt)
	assert.Equal(t, "${TOPAZ_DB_DIR}/directory.db", c.Store.Bolt.DBPath)
}

func TestEnvVars(t *testing.T) {
	t.Setenv("TOPAZ_TEST_READ_TIMEOUT", "2s")
	t.Setenv("TOPAZ_TEST_STORE_BOLTDB_DB_PATH", "/bolt/db/path")

	var c directory.Config

	v := config.NewViper()
	v.SetDefaults(&c)
	v.SetEnvPrefix("TOPAZ_TEST")
	v.AutomaticEnv()
	v.ReadConfig(
		strings.NewReader(boltConfig),
	)

	err := v.Unmarshal(&c)
	require.NoError(t, err)

	assert.Equal(t, "2s", c.ReadTimeout.String())
	assert.Equal(t, directory.BoltDBStorePlugin, c.Store.Provider)
	require.NotNil(t, c.Store.Bolt)
	assert.Equal(t, "/bolt/db/path", c.Store.Bolt.DBPath)
}

const (
	preamble = `
# directory configuration.
directory:
`

	boltConfig = `
read_timeout: 1s
write_timeout: 1s
# directory store configuration.
store:
  provider: boltdb
  boltdb:
    db_path: '/path/to/bolt.db'
    request_timeout: 5s
`

	remoteConfig = `
read_timeout: 1s
write_timeout: 1s
# directory store configuration.
store:
  provider: remote_directory
  remote_directory:
    address: 'localhost:9292'
    tenant_id: 'tenant-id'
    api_key: 'api-key'
    token: 'token'
    ca_cert_path: 'ca-cert-path'
    insecure: false
    no_tls: true
    no_proxy: false
    headers:
      x-foo: foo-value
`
)
