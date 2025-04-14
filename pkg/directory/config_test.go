package directory_test

import (
	"strings"
	"testing"
	"time"

	"github.com/aserto-dev/topaz/pkg/config/handler"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMarshaling(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cfg    string
		verify func(*testing.T, *directory.Store)
	}{
		{"boltdb", boltConfig, func(t *testing.T, s *directory.Store) {
			assert.Equal(t, directory.BoltDBStorePlugin, s.Plugin)
			require.NotNil(t, s.Bolt)
			assert.Equal(t, "/path/to/bolt.db", s.Bolt.DBPath)
			assert.Equal(t, 5*time.Second, s.Bolt.RequestTimeout)
		}},
		{"remote_directory", remoteConfig, func(t *testing.T, s *directory.Store) {
			assert.Equal(t, directory.RemoteDirectoryStorePlugin, s.Plugin)
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
			v := viper.NewWithOptions(viper.WithDecodeHook(handler.PluginDecodeHook()))
			v.SetConfigType("yaml")
			v.ReadConfig(
				strings.NewReader(tc.cfg),
			)

			var s directory.Store
			err := v.Unmarshal(&s, func(dc *mapstructure.DecoderConfig) { dc.TagName = "json" })
			require.NoError(t, err)

			tc.verify(t, &s)

			settings := v.AllSettings()

			out, err := yaml.Marshal(settings)
			require.NoError(t, err)

			assert.Equal(t, tc.cfg, "\n"+string(out))
		})
	}
}

const (
	boltConfig = `
plugin: boltdb
settings:
    db_path: /path/to/bolt.db
    request_timeout: 5s
`

	remoteConfig = `
plugin: remote_directory
settings:
    address: localhost:9292
    api_key: api-key
    ca_cert_path: ca-cert-path
    headers:
        x-foo: foo-value
    insecure: false
    no_tls: true
    tenant_id: tenant-id
    token: token
`
)
