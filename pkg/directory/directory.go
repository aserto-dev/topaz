package directory

import (
	"os"
	"text/template"
	"time"

	"github.com/aserto-dev/topaz/pkg/config/handler"
	"github.com/pkg/errors"

	"github.com/spf13/viper"
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
	v.SetDefault("plugin", defaultPlugin)
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

func (c *Config) Generate(w *os.File) error {
	tmpl, err := template.New("DIRECTORY").Parse(directoryTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	switch c.Store.Plugin {
	case BoltDBStorePlugin:
		cfg := BoltDBStoreFromMap(c.Store.Settings)
		return cfg.Generate(w)
	case RemoteDirectoryStorePlugin:
		cfg := RemoteDirectoryStoreFromMap(c.Store.Settings)
		return cfg.Generate(w)
	case PostgresStorePlugin:
		cfg := PostgresStoreFromMap(c.Store.Settings)
		return cfg.Generate(w)
	case NATSKeyValueStorePlugin:
		cfg := NATSKeyValueStoreFromMap(c.Store.Settings)
		return cfg.Generate(w)
	default:
		return errors.Errorf("unknown store plugin %q", c.Store.Plugin)
	}
}

const directoryTemplate = `
# directory configuration.
directory:
  read_timeout: {{ .ReadTimeout }}
  write_timeout: {{ .WriteTimeout }}
`

type Store struct {
	Plugin   string                 `json:"plugin"`
	Settings map[string]interface{} `json:"settings"`
}
