package services

import (
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/aserto-dev/go-aserto"

	"github.com/aserto-dev/topaz/pkg/config"
)

type (
	Config map[string]*Service

	Service struct {
		DependsOn []string       `json:"depends_on"`
		GRPC      GRPCService    `json:"grpc"`
		Gateway   GatewayService `json:"gateway"`
		Includes  []string       `json:"includes"`
	}

	GRPCService struct {
		ListenAddress     string           `json:"listen_address"`
		FQDN              string           `json:"fqdn"`
		Certs             aserto.TLSConfig `json:"certs"`
		ConnectionTimeout time.Duration    `json:"connection_timeout"` // https://godoc.org/google.golang.org/grpc#ConnectionTimeout
		DisableReflection bool             `json:"disable_reflection"`
	}

	GatewayService struct {
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
)

var (
	_ config.Section = (*Config)(nil)

	ErrPortCollision = errors.Wrap(config.ErrConfig, "service ports must be unique")
	ErrDependency    = errors.Wrap(config.ErrConfig, "undefined depdency")
)

func (c Config) Defaults() map[string]any {
	return lo.Assign(
		lo.Map(lo.Keys(c), func(name string, _ int) map[string]any {
			return config.PrefixKeys(name, c[name].Defaults())
		})...,
	)
}

func (c Config) Validate() error {
	if err := c.validateServices(); err != nil {
		return err
	}

	if err := c.validateListenAddresses(); err != nil {
		return err
	}

	return c.validateDepdencies()
}

func (c Config) validateServices() error {
	var errs error

	for name, svc := range c {
		if err := svc.Validate(); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, name))
		}
	}

	return errs
}

func (c Config) validateListenAddresses() error {
	addrs := make(map[string]string, len(c)*listenersPerService)

	var errs error

	for name, svc := range c {
		grpcName := name + " (grpc)"

		if existing, ok := addrs[svc.GRPC.ListenAddress]; ok {
			errs = multierror.Append(errs,
				errors.Wrapf(ErrPortCollision, collisionMsg(svc.GRPC.ListenAddress, existing, grpcName)),
			)
		}

		addrs[svc.GRPC.ListenAddress] = grpcName

		httpName := name + " (http)"

		if existing, ok := addrs[svc.Gateway.ListenAddress]; ok {
			errs = multierror.Append(errs,
				errors.Wrapf(ErrPortCollision, collisionMsg(svc.Gateway.ListenAddress, existing, httpName)),
			)
		}

		addrs[svc.Gateway.ListenAddress] = httpName
	}

	return errs
}

func (c Config) validateDepdencies() error {
	var errs error

	for name, svc := range c {
		for _, dep := range svc.DependsOn {
			if _, ok := c[dep]; !ok {
				errs = multierror.Append(errs, errors.Wrapf(ErrDependency, "%s referenced in %s", dep, name))
			}
		}
	}

	return errs
}

func (c Config) Serialize(w io.Writer) error {
	tmpl, err := template.New("SERVICES").Parse(servicesTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

// collisionMsg formats the message for a port collision error.
// It prints the service names in deterministic order for easier testing.
func collisionMsg(addr, svc1, svc2 string) string {
	if svc1 > svc2 {
		svc1, svc2 = svc2, svc1
	}

	return addr + fmt.Sprintf(" [%s, %s]", svc1, svc2)
}

func (c *Service) Defaults() map[string]any {
	return lo.Assign(
		config.PrefixKeys("grpc", c.GRPC.Defaults()),
		config.PrefixKeys("gateway", c.Gateway.Defaults()),
	)
}

func (s *Service) Validate() error {
	return nil
}

const (
	servicesTemplate string = `
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
)

func (s *GRPCService) Defaults() map[string]any {
	return map[string]any{
		"listen_address":         "0.0.0:9292",
		"certs.tls_cert_path":    "${TOPAZ_CERTS_DIR}/grpc.crt",
		"certs.tls_key_path":     "${TOPAZ_CERTS_DIR}/grpc.key",
		"certs.tls_ca_cert_path": "${TOPAZ_CERTS_DIR}/grpc-ca.crt",
		"disable_reflection":     false,
	}
}

func (s *GRPCService) Validate() error {
	return nil
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

func (s *GatewayService) Validate() error {
	return nil
}

const (
	DefaultReadTimeout       = time.Second * 5
	DefaultReadHeaderTimeout = time.Second * 5
	DefaultWriteTimeout      = time.Second * 5
	DefaultIdleTimeout       = time.Second * 30

	listenersPerService = 2 // gRPC and HTTP
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
