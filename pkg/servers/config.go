package servers

import (
	"fmt"
	"io"
	"iter"
	"maps"
	"slices"
	"text/template"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/loiter"
)

type (
	ServerName  string
	ServiceName string
	ServerKind  string

	Config map[ServerName]*Server

	Server struct {
		GRPC     GRPCServer    `json:"grpc"`
		HTTP     HTTPServer    `json:"http"`
		Services []ServiceName `json:"services"`
	}

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
