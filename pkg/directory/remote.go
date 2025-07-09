package directory

import (
	"io"
	"text/template"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/config"
	"google.golang.org/grpc"
)

const RemoteDirectoryStorePlugin string = "remote_directory"

type RemoteDirectoryStore client.Config

var _ config.Section = (*RemoteDirectoryStore)(nil)

func (*RemoteDirectoryStore) Defaults() map[string]any {
	return map[string]any{}
}

func (*RemoteDirectoryStore) Validate() error {
	return nil
}

func (c *RemoteDirectoryStore) Serialize(w io.Writer) error {
	tmpl, err := template.New("STORE").
		Funcs(config.TemplateFuncs()).
		Parse(remoteDirectoryStoreConfigTemplate)
	if err != nil {
		return err
	}

	type params struct {
		*RemoteDirectoryStore
		Provider_ string
	}

	p := params{c, RemoteDirectoryStorePlugin}

	return tmpl.Execute(w, p)
}

func (c *RemoteDirectoryStore) Connect() (*grpc.ClientConn, error) {
	return (*client.Config)(c).Connect()
}

const remoteDirectoryStoreConfigTemplate = `
{{- .Provider_ }}:
  address: '{{ .Address }}'

  {{- with .TenantID }}
  tenant_id: '{{ . }}'
  {{- end }}

  {{- with .APIKey }}
  api_key: '{{ . }}'
  {{- end }}

  {{- with .Token }}
  token: '{{ . }}'
  {{- end }}

  {{- with .CACertPath }}
  ca_cert_path: '{{ . }}'
  {{- end }}

  {{- with .Insecure }}
  insecure: {{ . }}
  {{- end }}

  {{- with .NoTLS }}
  no_tls: {{ . }}
  {{- end }}

  {{- with .NoProxy }}
  no_proxy: {{ . }}
  {{- end }}

  {{- with .Headers }}
  headers:
    {{- . | toYAML | nindent 4 }}
  {{- end }}
`
