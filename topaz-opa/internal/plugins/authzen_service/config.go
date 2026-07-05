package authzen_service

import (
	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/topaz-opa/internal/config"
)

type Config struct {
	Enabled bool    `json:"enabled"`
	Service Service `json:"service"`
}

type Service struct {
	GRPC struct {
		ListenAddress     string           `json:"listen_address"`
		FQDN              string           `json:"fqdn"`
		Certs             client.TLSConfig `json:"certs"`
		ConnectionTimeout config.Duration  `json:"connection_timeout"` // Default 120s https://godoc.org/google.golang.org/grpc#ConnectionTimeout
		EnableReflection  bool             `json:"reflection"`
	} `json:"grpc"`
	Gateway struct {
		Enabled       bool             `json:"enabled"`
		ListenAddress string           `json:"listen_address"`
		FQDN          string           `json:"fqdn"`
		HTTP          bool             `json:"http"`
		Certs         client.TLSConfig `json:"certs"`
		Timeouts      struct {
			ReadTimeout       config.Duration `json:"read_timeout"`
			ReadHeaderTimeout config.Duration `json:"read_header_timeout"`
			WriteTimeout      config.Duration `json:"write_timeout"`
			IdleTimeout       config.Duration `json:"idle_timeout"`
		} `json:"timeouts"`
		CQRS struct {
			AllowedOrigins []string `json:"allowed_origins"`
			AllowedHeaders []string `json:"allowed_headers"`
			AllowedMethods []string `json:"allowed_methods"`
		} `json:"cqrs"`
	} `json:"gateway"`
}
