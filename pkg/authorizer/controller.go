package authorizer

import (
	"io"
	"text/template"

	"github.com/aserto-dev/topaz/controller"
	"github.com/aserto-dev/topaz/pkg/config"
)

type ControllerConfig controller.Config

var _ config.Section = (*ControllerConfig)(nil)

func (c *ControllerConfig) Defaults() map[string]any {
	return map[string]any{}
}

func (c *ControllerConfig) Validate() error {
	return nil
}

func (c *ControllerConfig) Serialize(w io.Writer) error {
	tmpl, err := template.New("CONTROLLER").Parse(controllerTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const controllerTemplate = `
# control plane configuration
controller:
  enabled: {{ .Enabled }}
  {{- if .Enabled }}
  server:
    address: '{{ .Server.Address }}'
    api_key: '{{ .Server.APIKey }}'
    client_cert_path: '{{ .Server.ClientCertPath }}'
    client_key_path: '{{ .Server.ClientKeyPath }}'
  {{ end }}
`
