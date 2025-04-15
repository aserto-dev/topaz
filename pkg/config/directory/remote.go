package directory

import (
	"io"
	"strings"
	"text/template"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/config/handler"
)

const RemoteDirectoryStorePlugin string = "remote_directory"

type RemoteDirectoryStore client.Config

var _ handler.Config = (*RemoteDirectoryStore)(nil)

func (c *RemoteDirectoryStore) Defaults() map[string]any {
	return map[string]any{}
}

func (c *RemoteDirectoryStore) Validate() (bool, error) {
	return true, nil
}

func (c *RemoteDirectoryStore) Generate(w io.Writer) error {
	tmpl, err := template.New("STORE").Parse(strings.TrimLeft(remoteDirectoryStoreTemplate, "\n"))
	if err != nil {
		return err
	}

	type params struct {
		*RemoteDirectoryStore
		Plugin_ string
	}

	p := params{c, RemoteDirectoryStorePlugin}
	if err := tmpl.Execute(w, p); err != nil {
		return err
	}

	return nil
}

var remoteDirectoryStoreTemplate = `
{{ .Plugin_ }}:
  address: '{{ .Address }}'
  tenant_id: '{{ .TenantID }}'
  api_key: '{{ .APIKey }}'
  token: '{{ .Token }}'
  ca_cert_path: '{{ .CACertPath }}'
  insecure: {{ .Insecure }}
  no_tls: {{ .NoTLS }}
  no_proxy: {{ .NoProxy }}
  {{- if .Headers }}
  headers:
    {{- range $name, $value := .Headers }}
    {{ $name }}: {{ $value }}
    {{- end }}
  {{- end }}
`
