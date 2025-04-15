package directory

import (
	"bytes"
	"io"
	"reflect"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/aserto-dev/topaz/pkg/config/handler"
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
	handler.PluginConfig `json:"plugin_config,squash"` //nolint:staticcheck  //squash accepted by mapstructure

	Bolt     BoltDBStore          `json:"boltdb,omitempty"`
	Remote   RemoteDirectoryStore `json:"remote_directory,omitempty"`
	Postgres PostgresStore        `json:"postgres,omitempty"`
	NatsKV   NATSKeyValueStore    `json:"nats_kv,omitempty"`
}

var _ handler.Config = (*Config)(nil)

func (c *Config) Defaults() map[string]any {
	return lo.Assign(
		map[string]any{
			"read_timeout":  defaultReadTimeout,
			"write_timeout": defaultWriteTimeout,
			"store.plugin":  defaultPlugin,
		},
		handler.PrefixKeys("store.boltdb", c.Store.Bolt.Defaults()),
		handler.PrefixKeys("store.remote_directory", c.Store.Remote.Defaults()),
	)
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

	var buf bytes.Buffer
	if err := c.generatePlugins(&buf); err != nil {
		return err
	}

	_, err = w.Write([]byte(handler.Indent(buf.String(), pluginIndentLevel)))

	return err
}

func (c *Config) generatePlugins(w io.Writer) error {
	if err := writeIfNotEmpty(w, &c.Store.Bolt); err != nil {
		return err
	}

	if err := writeIfNotEmpty(w, &c.Store.Remote); err != nil {
		return err
	}

	if err := writeIfNotEmpty(w, &c.Store.Postgres); err != nil {
		return err
	}

	if err := writeIfNotEmpty(w, &c.Store.NatsKV); err != nil {
		return err
	}

	return nil
}

type config[T any] interface {
	handler.Config
	*T
}

func writeIfNotEmpty[T any, P config[T]](w io.Writer, t *T) error {
	if nilOrEmpty(t) {
		return nil
	}

	return P(t).Generate(w)
}

func nilOrEmpty[T any](t *T) bool {
	if t == nil {
		return true
	}

	var zero T

	return reflect.DeepEqual(zero, *t)
}

const directoryTemplate = `
# directory configuration.
directory:
  read_timeout: {{ .ReadTimeout }}
  write_timeout: {{ .WriteTimeout }}
  # directory store configuration.
  store:
    plugin: {{ .Store.Plugin }}
`
