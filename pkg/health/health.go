package health

import (
	"strings"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/config/handler"

	"github.com/spf13/viper"
)

type Config struct {
	Enabled       bool              `json:"enabled"`
	ListenAddress string            `json:"listen_address"`
	Certificates  *client.TLSConfig `json:"certs,omitempty"`
}

var _ = handler.Config(&Config{})

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault(strings.Join(append(p, "enabled"), "."), false)
	v.SetDefault(strings.Join(append(p, "listen_address"), "."), "0.0.0.0:9494")
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}
