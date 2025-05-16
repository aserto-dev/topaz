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
		Model      ServiceName
		Importer   ServiceName
		Exporter   ServiceName
	}{
		Access:     "access",
		Authorizer: "authorizer",
		Console:    "console",
		Reader:     "reader",
		Writer:     "writer",
		Model:      "model",
		Importer:   "importer",
		Exporter:   "exporter",
	}

	DirectoryServices = []ServiceName{Service.Reader, Service.Writer, Service.Access, Service.Model, Service.Importer, Service.Exporter}

	KnownServices = append(DirectoryServices, Service.Authorizer, Service.Console)

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

	return c.validateListenAddresses()
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

func (c Config) FindService(name ServiceName) (*Server, bool) {
	return loiter.Find(maps.Values(c), func(s *Server) bool {
		return slices.Contains(s.Services, name)
	})
}

func (c Config) ServiceEnabled(name ServiceName) bool {
	_, found := c.FindService(name)
	return found
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

func (c Config) Serialize(w io.Writer) error {
	tmpl, err := template.New("SERVERS").
		Funcs(config.TemplateFuncs()).
		Parse(servicesTemplate)
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
		config.PrefixKeys("http", c.HTTP.Defaults()),
	)
}

func (s *Server) Validate() error {
	var (
		errs      error
		needsGRPC bool
	)

	for _, service := range s.Services {
		if !slices.Contains(KnownServices, service) {
			errs = multierror.Append(errs, errors.Wrapf(config.ErrConfig, "unknown service %q", service))
			continue
		}

		if service != Service.Console {
			// All services except the console require grpc configuration.
			needsGRPC = true
		} else if !s.HTTP.HasListener() {
			errs = multierror.Append(errs, errors.Wrap(config.ErrConfig, "http listen_address is required for console service"))
		}
	}

	if needsGRPC && !s.GRPC.HasListener() {
		errs = multierror.Append(errs, errors.Wrap(config.ErrConfig, "grpc listen_address is required"))
	}

	return errs
}

// TryGRPC returns the server's gRPC configuration if it isn't empty or nil otherwise.
// It is used by the serialization template.
func (s *Server) TryGRPC() *GRPCServer {
	zero := GRPCServer{}
	if s.GRPC == zero {
		return nil
	}

	return &s.GRPC
}

// TryHTTP returns the server's http configuration if it isn't empty or nil otherwise.
// It is used by the serialization template.
func (s *Server) TryHTTP() *HTTPServer {
	if s.HTTP.IsEmpty() {
		return nil
	}

	return &s.HTTP
}

const (
	servicesTemplate string = `
#  grpc and http server configuration
servers:
  {{- range $name, $server := . }}
  {{ $name }}:

    {{- with $server.Services }}
    services:
      {{- . | toYaml | nindent 6 }}
    {{- end }}

  {{- with $server.TryGRPC }}
    grpc:
      listen_address: '{{ .ListenAddress }}'

      {{- with .Certs }}
      certs:
        tls_key_path: '{{ .Key }}'
        tls_cert_path: '{{ .Cert }}'
        tls_ca_cert_path: '{{ .CA }}'
      {{ end -}}

      connection_timeout: {{ .ConnectionTimeout }}

    {{- with .NoReflection }}
      no_reflection: {{ . }}
    {{- end }}
  {{- end }}

  {{- with $server.TryHTTP }}
    http:
      listen_address: '{{ .ListenAddress }}'
      fqdn: '{{ .FQDN }}'

      {{- with .Certs }}
      certs:
        tls_key_path: '{{ .Key }}'
        tls_cert_path: '{{ .Cert }}'
        tls_ca_cert_path: '{{ .CA }}'
      {{- end }}

      {{- with .AllowedOrigins }}
      allowed_origins:
        {{- . | toYaml | nindent 8 }}
      {{- end }}

      {{- with .AllowedHeaders }}
      allowed_headers:
        {{- . | toYaml | nindent 8 }}
      {{- end }}

      {{- with .AllowedMethods }}
      allowed_methods:
        {{- . | toYaml | nindent 8 }}
      {{ end -}}

      read_timeout: {{ .ReadTimeout }}
      read_header_timeout: {{ .ReadHeaderTimeout }}
      write_timeout: {{ .WriteTimeout }}
      idle_timeout: {{ .IdleTimeout }}
    {{- end }}
  {{ end }}
`
)
