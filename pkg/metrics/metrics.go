package metrics

import (
	"strings"

	client "github.com/aserto-dev/go-aserto"
	"github.com/spf13/viper"
)

type Config struct {
	Enabled       bool              `json:"enabled"`
	ListenAddress string            `json:"listen_address"`
	Certificates  *client.TLSConfig `json:"certs,omitempty"`
}

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault(strings.Join(append(p, "enabled"), "."), false)
	v.SetDefault(strings.Join(append(p, "listen_address"), "."), "0.0.0.0:9696")
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}
