package directory

import (
	"io"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/aserto-dev/topaz/pkg/config/handler"
)

const (
	defaultReadTimeout         = 5 * time.Second
	defaultWriteTimeout        = 5 * time.Second
	defaultPlugin       string = BoltDBStorePlugin
)

type Config struct {
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	Store        Store         `json:"store"`
}

var _ = handler.Config(&Config{})

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault("read_timeout", defaultReadTimeout)
	v.SetDefault("write_timeout", defaultWriteTimeout)
	v.SetDefault("store.plugin", defaultPlugin)
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

func (c *Config) Generate(w io.Writer) error {
	tmpl, err := template.New("DIRECTORY").Parse(directoryTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	switch c.Store.Plugin {
	case BoltDBStorePlugin:
		return c.Store.Bolt.Generate(w)
	case RemoteDirectoryStorePlugin:
		return c.Store.Remote.Generate(w)
	case PostgresStorePlugin:
		return c.Store.Postgres.Generate(w)
	case NATSKeyValueStorePlugin:
		return c.Store.NatsKV.Generate(w)
	default:
		return errors.Errorf("unknown store plugin %q", c.Store.PluginConfig)
	}
}

const directoryTemplate = `
# directory configuration.
directory:
  read_timeout: {{ .ReadTimeout }}
  write_timeout: {{ .WriteTimeout }}
  # directory store configuration.
  store:
    plugin: {{ .Store.Plugin }}
    settings:
`

type Store struct {
	handler.PluginConfig `json:"plugin_config,squash"` //nolint:staticcheck  //squash accepted by mapstructure

	Bolt     *BoltDBStore          `json:"boltdb,omitempty"`
	Remote   *RemoteDirectoryStore `json:"remote_directory,omitempty"`
	Postgres *PostgresStore        `json:"postgres,omitempty"`
	NatsKV   *NATSKeyValueStore    `json:"nats_kv,omitempty"`
}
