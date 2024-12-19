package services

import (
	"net/http"
	"strings"
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/config/handler"

	"github.com/go-http-utils/headers"
	"github.com/spf13/viper"
)

type Config map[string]*Service

var _ = handler.Config(&Config{})

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	for name, service := range *c {
		service.SetDefaults(v, strings.Join(append(p, name), "."))
	}
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

type Service struct {
	DependsOn []string       `json:"depends_on"`
	GRPC      GRPCService    `json:"grpc"`
	Gateway   GatewayService `json:"gateway"`
	Includes  []string       `json:"includes"`
}

func (s *Service) SetDefaults(v *viper.Viper, p ...string) {
	s.GRPC.SetDefaults(v, strings.Join(append(p, "grpc"), "."))
	s.Gateway.SetDefaults(v, strings.Join(append(p, "gateway"), "."))
}

func (s *Service) Validate() (bool, error) {
	return true, nil
}

type GRPCService struct {
	ListenAddress     string           `json:"listen_address"`
	FQDN              string           `json:"fqdn"`
	Certs             aserto.TLSConfig `json:"certs"`
	ConnectionTimeout time.Duration    `json:"connection_timeout"` // https://godoc.org/google.golang.org/grpc#ConnectionTimeout
	DisableReflection bool             `json:"disable_reflection"`
}

func (s *GRPCService) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault(strings.Join(append(p, "listen_address"), "."), "0.0.0.0:9292")
	v.SetDefault(strings.Join(append(p, "certs", "tls_cert_path"), "."), "${TOPAZ_CERTS_DIR}/grpc.crt")
	v.SetDefault(strings.Join(append(p, "certs", "tls_key_path"), "."), "${TOPAZ_CERTS_DIR}/grpc.key")
	v.SetDefault(strings.Join(append(p, "certs", "tls_ca_cert_path"), "."), "${TOPAZ_CERTS_DIR}/grpc-ca.crt")
	v.SetDefault(strings.Join(append(p, "disable_reflection"), "."), false)
}

func (s *GRPCService) Validate() (bool, error) {
	return true, nil
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

func (s *GatewayService) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault(strings.Join(append(p, "listen_address"), "."), "0.0.0.0:9393")
	v.SetDefault(strings.Join(append(p, "certs", "tls_cert_path"), "."), "${TOPAZ_CERTS_DIR}/gateway.crt")
	v.SetDefault(strings.Join(append(p, "certs", "tls_key_path"), "."), "${TOPAZ_CERTS_DIR}/gateway.key")
	v.SetDefault(strings.Join(append(p, "certs", "tls_ca_cert_path"), "."), "${TOPAZ_CERTS_DIR}/gateway-ca.crt")
	v.SetDefault(strings.Join(append(p, "allowed_origins"), "."), DefaultAllowedOrigins(s.HTTP))
	v.SetDefault(strings.Join(append(p, "allowed_headers"), "."), DefaultAllowedHeaders())
	v.SetDefault(strings.Join(append(p, "allowed_methods"), "."), DefaultAllowedMethods())
	v.SetDefault(strings.Join(append(p, "http"), "."), false)
	v.SetDefault(strings.Join(append(p, "read_timeout"), "."), DefaultReadTimeout.String())
	v.SetDefault(strings.Join(append(p, "read_header_timeout"), "."), DefaultReadHeaderTimeout.String())
	v.SetDefault(strings.Join(append(p, "write_timeout"), "."), DefaultWriteTimeout.String())
	v.SetDefault(strings.Join(append(p, "idle_timeout"), "."), DefaultIdleTimeout.String())
}

func (s *GatewayService) Validate() (bool, error) {
	return true, nil
}

const (
	DefaultReadTimeout       = time.Second * 5
	DefaultReadHeaderTimeout = time.Second * 5
	DefaultWriteTimeout      = time.Second * 5
	DefaultIdleTimeout       = time.Second * 30
)

func DefaultAllowedOrigins(useHTTP bool) []string {
	if useHTTP {
		return []string{
			"http://localhost",
			"http://localhost:*",
			"http://127.0.0.1",
			"http://127.0.0.1:*",
			"http://0.0.0.0",
			"http://0.0.0.0:*",
		}
	}
	return []string{
		"https://localhost",
		"https://localhost:*",
		"https://127.0.0.1",
		"https://127.0.0.1:*",
		"https://0.0.0.0",
		"https://0.0.0.0:*",
	}
}

func DefaultAllowedHeaders() []string {
	return []string{
		headers.Authorization,
		headers.ContentType,
		headers.IfMatch,
		headers.IfNoneMatch,
		"Depth",
	}
}

func DefaultAllowedMethods() []string {
	return []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodHead,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		"PROPFIND",
		"MKCOL",
		"COPY",
		"MOVE",
	}
}
