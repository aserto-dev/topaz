package metrics

import client "github.com/aserto-dev/go-aserto"

type Config struct {
	Enabled       bool              `json:"enabled"`
	ListenAddress string            `json:"listen_address"`
	Certificates  *client.TLSConfig `json:"certs"`
}
