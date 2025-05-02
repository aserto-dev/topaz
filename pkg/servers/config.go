package servers

import (
	"fmt"
	"io"
	"iter"
	"maps"
	"net/http"
	"slices"
	"text/template"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aserto-dev/go-aserto"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/loiter"
)

type (
	ServerName string

	ServiceName string

	Config map[ServerName]*Server

	Server struct {
		GRPC     GRPCServer    `json:"grpc"`
		HTTP     HTTPServer    `json:"http"`
		Services []ServiceName `json:"services"`
	}

	GRPCServer struct {
		ListenAddress     string           `json:"listen_address"`
		FQDN              string           `json:"fqdn"`
		Certs             aserto.TLSConfig `json:"certs"`
		ConnectionTimeout time.Duration    `json:"connection_timeout"` // https://godoc.org/google.golang.org/grpc#ConnectionTimeout
		DisableReflection bool             `json:"disable_reflection"`
	}

	HTTPServer struct {
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

	ServerKind string

	ListenAddress struct {
		Address string
		Kind    ServerKind
	}
)

var (
	_ config.Section = (*Config)(nil)

	ErrPortCollision = errors.Wrap(config.ErrConfig, "service ports must be unique")
	ErrDependency    = errors.Wrap(config.ErrConfig, "undefined depdency")

	Service = struct {
		Access     ServiceName
		Authorizer ServiceName
		Console    ServiceName
		Reader     ServiceName
		Writer     ServiceName
	}{
		Access:     "access",
		Authorizer: "authorizer",
		Console:    "console",
		Reader:     "reader",
		Writer:     "writer",
	}

	DirectoryServices = []ServiceName{Service.Reader, Service.Writer}

	KnownServices = append(DirectoryServices, Service.Access, Service.Authorizer, Service.Console)

	Kind = struct {
		GRPC ServerKind
		HTTP ServerKind
	}{
		GRPC: "grpc",
		HTTP: "http",
	}
)

func (c Config) Defaults() map[string]any {
	return lo.Assign(
		lo.Map(lo.Keys(c), func(name ServerName, _ int) map[string]any {
			return config.PrefixKeys(string(name), c[name].Defaults())
		})...,
	)
}

func (c Config) Validate() error {
	if err := c.validateServers(); err != nil {
		return err
	}

	if err := c.validateListenAddresses(); err != nil {
		return err
	}

	return c.validateDepdencies()
}

func (c Config) EnabledServices() iter.Seq[ServiceName] {
	return loiter.FlatMap(
		maps.Values(c),
		func(svr *Server) iter.Seq[ServiceName] {
			return slices.Values(svr.Services)
		},
	)
}

func (c Config) DirectoryEnabled() bool {
	return loiter.ContainsAny(c.EnabledServices(), DirectoryServices...)
}

func (c Config) ListenAddresses() iter.Seq2[ServerName, ListenAddress] {
	return loiter.ExplodeValues(maps.All(c), func(name ServerName, server *Server) iter.Seq[ListenAddress] {
		return slices.Values([]ListenAddress{
			{server.GRPC.ListenAddress, Kind.GRPC},
			{server.HTTP.ListenAddress, Kind.HTTP},
		})
	})
}

func (c Config) validateServers() error {
	var errs error

	for name, server := range c {
		if err := server.Validate(); err != nil {
			errs = multierror.Append(errs, errors.Wrap(err, string(name)))
		}
	}

	return errs
}

func (c Config) validateListenAddresses() error {
	addrs := make(map[string]string, len(c)*listenersPerService)

	var errs error

	for name, listenAddress := range c.ListenAddresses() {
		addressName := fmt.Sprintf("%s (%s)", name, listenAddress.Kind)

		if existing, ok := addrs[listenAddress.Address]; ok {
			errs = multierror.Append(errs,
				errors.Wrapf(ErrPortCollision, collisionMsg(listenAddress.Address, existing, addressName)),
			)
		}

		addrs[listenAddress.Address] = addressName
	}

	return errs
}

func (c Config) validateDepdencies() error {
	// TODO: Find cycles
	return nil
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

func (c *Server) Defaults() map[string]any {
	return lo.Assign(
		config.PrefixKeys("grpc", c.GRPC.Defaults()),
		config.PrefixKeys("gateway", c.HTTP.Defaults()),
	)
}

func (s *Server) Validate() error {
	var errs error

	for _, service := range s.Services {
		if !slices.Contains(KnownServices, service) {
			errs = multierror.Append(errs, errors.Wrapf(config.ErrConfig, "unknown service %q", service))
		}
	}

	// TODO: validate that a grpc listen address is set if a non-console service is assigned to the server.

	return errs
}

const (
	servicesTemplate string = `
# services configuration
services:
  {{- range $name, $server := . }}
  {{ $name }}:
    {{- with $server.GRPC }}
    grpc:
      listen_address: '{{ .ListenAddress }}'
      fqdn: '{{ .FQDN }}'
      {{- if .Certs }}
      certs:
        tls_key_path: '{{ .Certs.Key }}'
        tls_cert_path: '{{ .Certs.Cert }}'
        tls_ca_cert_path: '{{ .Certs.CA }}'
      {{ end -}}
      connection_timeout: {{ .ConnectionTimeout }}
      disable_reflection: {{ .DisableReflection }}
    {{- end }}

    {{- with $server.HTTP }}
    gateway:
      listen_address: '{{ .ListenAddress }}'
      fqdn: '{{ .FQDN }}'
      {{- if .Certs }}
      certs:
        tls_key_path: '{{ .Certs.Key }}'
        tls_cert_path: '{{ .Certs.Cert }}'
        tls_ca_cert_path: '{{ .Certs.CA }}'
      {{ end -}}
      allowed_origins:
      {{- range .AllowedOrigins }}
        - {{ . -}}
      {{ end }}
      allowed_headers:
      {{- range .AllowedHeaders }}
        - {{ . -}}
      {{ end }}
      allowed_methods:
      {{- range .AllowedMethods }}
        - {{ . -}}
      {{ end }}
      http: {{ .HTTP }}
      read_timeout: {{ .ReadTimeout }}
      read_header_timeout: {{ .ReadHeaderTimeout }}
      write_timeout: {{ .WriteTimeout }}
      idle_timeout: {{ .IdleTimeout }}
    {{- end }}
    includes:
    {{- range $server.Services }}
      - {{ . -}}
    {{ end }}
  {{ end }}
`
)

func (s *GRPCServer) Defaults() map[string]any {
	return map[string]any{
		"listen_address":         "0.0.0:9292",
		"certs.tls_cert_path":    "${TOPAZ_CERTS_DIR}/grpc.crt",
		"certs.tls_key_path":     "${TOPAZ_CERTS_DIR}/grpc.key",
		"certs.tls_ca_cert_path": "${TOPAZ_CERTS_DIR}/grpc-ca.crt",
		"disable_reflection":     false,
	}
}

func (s *GRPCServer) Validate() error {
	return nil
}

func (s *GRPCServer) HasListener() bool {
	return s != nil && s.ListenAddress != ""
}

func (s *GRPCServer) ClientCredentials() (grpc.DialOption, error) {
	if !s.Certs.HasCert() {
		return grpc.WithTransportCredentials(insecure.NewCredentials()), nil
	}

	creds, err := s.Certs.ClientCredentials(true)
	if err != nil {
		return nil, err
	}

	return grpc.WithTransportCredentials(creds), nil
}

func (s *HTTPServer) Defaults() map[string]any {
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

func (s *HTTPServer) Validate() error {
	return nil
}

func (s *HTTPServer) HasListener() bool {
	return s != nil && s.ListenAddress != ""
}

func (s *HTTPServer) Cors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedHeaders: s.AllowedHeaders,
		AllowedOrigins: s.AllowedOrigins,
		AllowedMethods: s.AllowedMethods,
		Debug:          false,
	})
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
