package health

import (
	"html/template"
	"io"
	"strings"

	"github.com/Masterminds/sprig/v3"
	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/config/handler"

	"github.com/spf13/viper"
)

type Config struct {
	Enabled       bool             `json:"enabled"`
	ListenAddress string           `json:"listen_address"`
	Certificates  client.TLSConfig `json:"certs,omitempty"`
}

var _ = handler.Config(&Config{})

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault(strings.Join(append(p, "enabled"), "."), false)
	v.SetDefault(strings.Join(append(p, "listen_address"), "."), "0.0.0.0:9494")
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

func (c *Config) Generate(w io.Writer) error {
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
  listen_address: '{{ .ListenAddress}}'
  {{- if .Certificates }}
  certs:
    tls_key_path: '{{ .Certificates.Key }}'
    tls_cert_path: '{{ .Certificates.Cert }}'
    tls_ca_cert_path: '{{ .Certificates.CA }}'
  {{ end }}
`
