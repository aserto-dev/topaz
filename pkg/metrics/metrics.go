package metrics

import (
	"io"
	"text/template"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/config/handler"
)

type Config struct {
	Enabled       bool             `json:"enabled"`
	ListenAddress string           `json:"listen_address"`
	Certificates  client.TLSConfig `json:"certs,omitempty"`
}

var _ handler.Config = (*Config)(nil)

func (c *Config) Defaults() map[string]any {
	return map[string]any{
		"enabled":        false,
		"listen_address": "0.0.0.0:9696",
	}
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

func (c *Config) Generate(w io.Writer) error {
	tmpl, err := template.New("METRICS").Parse(metricsTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const metricsTemplate = `
# metric service settings.
metrics:
  enabled: {{ .Enabled }}
  listen_address: '{{ .ListenAddress}}'
  {{- if .Certificates }}
  certs:
    tls_key_path: '{{ .Certificates.Key }}'
    tls_cert_path: '{{ .Certificates.Cert }}'
    tls_ca_cert_path: '{{ .Certificates.CA }}'
  {{ end }}
`
