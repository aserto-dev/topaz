package health

import (
	"html/template"
	"io"
	"time"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/servers"
)

type Config struct {
	servers.GRPCServer

	Enabled bool `json:"enabled"`
}

var _ config.Section = (*Config)(nil)

//nolint:mnd  // this is where default values are defined.
func (c *Config) Defaults() map[string]any {
	return map[string]any{
		"enabled":            false,
		"listen_address":     "0.0.0.0:9494",
		"connection_timeout": 5 * time.Second,
	}
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Serialize(w io.Writer) error {
	tmpl := template.Must(template.New("HEALTH").Parse(healthTemplate))
	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const healthTemplate = `
# health service settings.
health:
  enabled: {{ .Enabled }}
  listen_address: '{{ .ListenAddress }}'
{{- with .TryCerts }}
  certs:
    tls_key_path: '{{ .Key }}'
    tls_cert_path: '{{ .Cert }}'
    tls_ca_cert_path: '{{ .CA }}'
{{- end }}
`
