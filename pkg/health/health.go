package health

import (
	"html/template"
	"io"

	"github.com/Masterminds/sprig/v3"
	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/config"
)

type Config struct {
	Enabled       bool             `json:"enabled"`
	ListenAddress string           `json:"listen_address"`
	Certificates  client.TLSConfig `json:"certs,omitempty"`
}

var _ config.Section = (*Config)(nil)

func (c *Config) Defaults() map[string]any {
	return map[string]any{
		"enabled":        false,
		"listen_address": "0.0.0.0:9494",
	}
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Serialize(w io.Writer) error {
	tmpl := template.New("HEALTH")

	var funcMap template.FuncMap = map[string]interface{}{}
	tmpl = tmpl.Funcs(sprig.TxtFuncMap()).Funcs(funcMap)

	tmpl, err := tmpl.Parse(healthTemplate)
	if err != nil {
		return err
	}

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
{{- with .Certificates }}
  certs:
    tls_key_path: '{{ .Key }}'
    tls_cert_path: '{{ .Cert }}'
    tls_ca_cert_path: '{{ .CA }}'
{{- end }}
`
