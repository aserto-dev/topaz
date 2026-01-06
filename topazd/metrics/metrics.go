package metrics

import (
	"io"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config"
)

type Config struct {
	config.Optional
	config.Listener
}

var _ config.Section = (*Config)(nil)

func (c *Config) Defaults() map[string]any {
	return map[string]any{
		"enabled":        false,
		"listen_address": "0.0.0.0:9696",
	}
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Serialize(w io.Writer) error {
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
  {{- with .TryCerts }}
  certs:
    tls_key_path: '{{ .Key }}'
    tls_cert_path: '{{ .Cert }}'
    tls_ca_cert_path: '{{ .CA }}'
  {{ end }}
`
