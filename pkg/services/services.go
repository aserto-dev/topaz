package services

import (
	"time"

	"github.com/aserto-dev/go-aserto"
)

type Config struct {
	Services map[string]Service `json:"services"`
}

type Service struct {
	DependsOn []string       `json:"depends_on"`
	GRPC      GRPCService    `json:"grpc"`
	Gateway   GatewayService `json:"gateway"`
	Includes  []string       `json:"includes"`
}

type GRPCService struct {
	ListenAddress     string           `json:"listen_address"`
	FQDN              string           `json:"fqdn"`
	Certs             aserto.TLSConfig `json:"certs"`
	ConnectionTimeout time.Duration    `json:"connection_timeout"` // https://godoc.org/google.golang.org/grpc#ConnectionTimeout
	DisableReflection bool             `json:"disable_reflection"`
}

type GatewayService struct {
	ListenAddress     string           `json:"listen_address"`
	FQDN              string           `json:"fqdn"`
	Certs             aserto.TLSConfig `json:"certs"`
	AllowedOrigins    []string         `json:"allowed_origins"`
	AllowedHeaders    []string         `json:"allowed_headers"`
	AllowedMethods    []string         `json:"allowed_methods"`
	HTTP              bool             `json:"http"`
	ReadTimeout       time.Duration    `json:"read_timeout"`
	ReadHeaderTimeout time.Duration    `json:"read_header_timeout"`
	WriteTimeout      time.Duration    `json:"write_timeout"`
	IdleTimeout       time.Duration    `json:"idle_timeout"`
}
