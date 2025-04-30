package directory

import (
	"io"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/aserto-dev/topaz/pkg/config"
)

const (
	defaultReadTimeout         = 5 * time.Second
	defaultWriteTimeout        = 5 * time.Second
	defaultPlugin       string = BoltDBStorePlugin

	pluginIndentLevel = 4
)

type Config struct {
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	Store        Store         `json:"store"`
}

type Store struct {
	Provider string `json:"provider"`

	Bolt     BoltDBStore          `json:"boltdb,omitempty"`
	Remote   RemoteDirectoryStore `json:"remote_directory,omitempty"`
	Postgres PostgresStore        `json:"postgres,omitempty"`
	NatsKV   NATSKeyValueStore    `json:"nats_kv,omitempty"`
}

var _ config.Section = (*Config)(nil)

func (c *Config) Defaults() map[string]any {
	return lo.Assign(
		map[string]any{
			"read_timeout":   defaultReadTimeout,
			"write_timeout":  defaultWriteTimeout,
			"store.provider": defaultPlugin,
		},
		config.PrefixKeys("store.boltdb", c.Store.Bolt.Defaults()),
		config.PrefixKeys("store.remote_directory", c.Store.Remote.Defaults()),
	)
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Serialize(w io.Writer) error {
	tmpl, err := template.New("DIRECTORY").Parse(config.TrimN(configTemplate))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return c.generatePlugins(config.IndentWriter(w, pluginIndentLevel))
}

func (c *Config) IsRemote() bool {
	return c.Store.Provider == RemoteDirectoryStorePlugin
}

func (c *Config) generatePlugins(w io.Writer) error {
	if err := config.WriteNonEmpty(w, &c.Store.Bolt); err != nil {
		return err
	}

	if err := config.WriteNonEmpty(w, &c.Store.Remote); err != nil {
		return err
	}

	if err := config.WriteNonEmpty(w, &c.Store.Postgres); err != nil {
		return err
	}

	if err := config.WriteNonEmpty(w, &c.Store.NatsKV); err != nil {
		return err
	}

	return nil
}

const configTemplate = `
# directory configuration.
directory:
  read_timeout: {{ .ReadTimeout }}
  write_timeout: {{ .WriteTimeout }}
  # directory store configuration.
  store:
    provider: {{ .Store.Provider }}
`
