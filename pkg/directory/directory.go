package directory

import (
	"time"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/config/handler"

	"github.com/spf13/viper"
)

type Config struct {
	Store Store `json:"store"`
}

var _ = handler.Config(&Config{})

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault("plugin", "boltdb")
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

type Store struct {
	Plugin   string                 `json:"plugin"`
	Settings map[string]interface{} `json:"settings"`
}

type LocalStore struct {
	directory.Config
}

const DefaultRequestTimeout = time.Second * 5

func (c *LocalStore) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault("db_path", "${TOPAZ_DB_DIR}/directory.db")
	v.SetDefault("request_timeout", DefaultRequestTimeout.String())
}

func (c *LocalStore) Validate() (bool, error) {
	return true, nil
}

type RemoteStore struct {
	client.Config
}

type PostgresStore struct{}

type NATSKeyValueStore struct{}
