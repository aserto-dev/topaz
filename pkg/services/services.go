package services

import (
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/samber/lo"

	"github.com/go-http-utils/headers"
)

type Config map[string]*Service

var _ config.Section = (*Config)(nil)

func (c Config) Defaults() map[string]any {
	return lo.Assign(
		lo.Map(lo.Keys(c), func(name string, _ int) map[string]any {
			return config.PrefixKeys(name, c[name].Defaults())
		})...,
	)
}

func (c Config) Validate() (bool, error) {
	return true, nil
}

func (c Config) Generate(w io.Writer) error {
	tmpl, err := template.New("SERVICES").Parse(servicesTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

type Service struct {
	DependsOn []string       `json:"depends_on"`
	GRPC      GRPCService    `json:"grpc"`
	Gateway   GatewayService `json:"gateway"`
	Includes  []string       `json:"includes"`
}

func (c *Service) Defaults() map[string]any {
	return lo.Assign(
		config.PrefixKeys("grpc", c.GRPC.Defaults()),
		config.PrefixKeys("gateway", c.Gateway.Defaults()),
	)
}

func (s *Service) Validate() (bool, error) {
	return true, nil
}

const servicesTemplate string = `
# services configuration
services:
  {{- range $name, $service := . }}
  {{ $name }}:
    grpc:
      listen_address: '{{ $service.GRPC.ListenAddress }}'
      fqdn: '{{ $service.GRPC.FQDN }}'
      {{- if $service.GRPC.Certs }}
      certs:
        tls_key_path: '{{ $service.GRPC.Certs.Key }}'
        tls_cert_path: '{{ $service.GRPC.Certs.Cert }}'
        tls_ca_cert_path: '{{ $service.GRPC.Certs.CA }}'
      {{ end -}}
      connection_timeout: {{ $service.GRPC.ConnectionTimeout }}
      disable_reflection: {{ $service.GRPC.DisableReflection }}

    gateway:
      listen_address: '{{ $service.Gateway.ListenAddress }}'
      fqdn: '{{ $service.Gateway.FQDN }}'
      {{- if $service.Gateway.Certs }}
      certs:
        tls_key_path: '{{ $service.Gateway.Certs.Key }}'
        tls_cert_path: '{{ $service.Gateway.Certs.Cert }}'
        tls_ca_cert_path: '{{ $service.Gateway.Certs.CA }}'
      {{ end -}}
      allowed_origins:
      {{- range $service.Gateway.AllowedOrigins }}
        - {{ . -}}
      {{ end }}
      allowed_headers:
      {{- range $service.Gateway.AllowedHeaders }}
        - {{ . -}}
      {{ end }}
      allowed_methods:
      {{- range $service.Gateway.AllowedMethods }}
        - {{ . -}}
      {{ end }}
      http: {{ $service.Gateway.HTTP }}
      read_timeout: {{ $service.Gateway.ReadTimeout }}
      read_header_timeout: {{ $service.Gateway.ReadHeaderTimeout }}
      write_timeout: {{ $service.Gateway.WriteTimeout }}
      idle_timeout: {{ $service.Gateway.IdleTimeout }}
    includes:
    {{- range $service.Includes }}
      - {{ . -}}
    {{ end }}
  {{ end }}
`

type GRPCService struct {
	ListenAddress     string           `json:"listen_address"`
	FQDN              string           `json:"fqdn"`
	Certs             aserto.TLSConfig `json:"certs"`
	ConnectionTimeout time.Duration    `json:"connection_timeout"` // https://godoc.org/google.golang.org/grpc#ConnectionTimeout
	DisableReflection bool             `json:"disable_reflection"`
}

func (s *GRPCService) Defaults() map[string]any {
	return map[string]any{
		"listen_address":         "0.0.0:9292",
		"certs.tls_cert_path":    "${TOPAZ_CERTS_DIR}/grpc.crt",
		"certs.tls_key_path":     "${TOPAZ_CERTS_DIR}/grpc.key",
		"certs.tls_ca_cert_path": "${TOPAZ_CERTS_DIR}/grpc-ca.crt",
		"disable_reflection":     false,
	}
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

func (s *GatewayService) Defaults() map[string]any {
	return map[string]any{
		"listen_address":         "0.0.0:9393",
		"certs.tls_cert_path":    "${TOPAZ_CERTS_DIR}/gateway.crt",
		"certs.tls_key_path":     "${TOPAZ_CERTS_DIR}/gateway.key",
		"certs.tls_ca_cert_path": "${TOPAZ_CERTS_DIR}/gateway-ca.crt",
		"allowed_origins":        DefaultAllowedOrigins(s.HTTP),
		"allowed_headers":        DefaultAllowedHeaders(),
		"allowed_methods":        DefaultAllowedMethods(),
		"http":                   false,
		"read_timeout":           DefaultReadTimeout.String(),
		"read_header_timeout":    DefaultReadHeaderTimeout.String(),
		"write_timeout":          DefaultWriteTimeout.String(),
		"idle_timeout":           DefaultIdleTimeout.String(),
	}
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
