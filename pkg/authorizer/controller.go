package authorizer

import (
	"os"
	"text/template"

	"github.com/aserto-dev/aserto-management/controller"
)

type ControllerConfig struct {
	controller.Config
}

func (c *ControllerConfig) Validate() (bool, error) {
	return true, nil
}

func (c *ControllerConfig) Generate(w *os.File) error {
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
    server:
      address: '{{ .Server.Address }}'
      api_key: '{{ .Server.APIKey }}'
      client_cert_path: '{{ .Server.ClientCertPath }}'
      client_key_path: '{{ .Server.ClientKeyPath }}'
`
