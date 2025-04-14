package directory

import (
	"io"
	"text/template"

	client "github.com/aserto-dev/go-aserto"
	"github.com/go-viper/mapstructure/v2"
)

const RemoteDirectoryStorePlugin string = "remote_directory"

type RemoteDirectoryStore struct {
	client.Config `json:"config,squash"` //nolint:staticcheck  //squash accepted by mapstructure
}

func (c *RemoteDirectoryStore) Validate() (bool, error) {
	return true, nil
}

func (c *RemoteDirectoryStore) Generate(w io.Writer) error {
	tmpl, err := template.New("STORE").Parse(remoteDirectoryStoreTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

func (c RemoteDirectoryStore) Map() map[string]interface{} {
	var result map[string]interface{}
	if err := mapstructure.Decode(c, &result); err != nil {
		return nil
	}

	return result
}

func RemoteDirectoryStoreFromMap(m map[string]interface{}) *RemoteDirectoryStore {
	var cfg RemoteDirectoryStore
	if err := mapstructure.Decode(m, &cfg); err != nil {
		return nil
	}

	return &cfg
}

func RemoteDirectoryStoreMap(cfg *RemoteDirectoryStore) map[string]interface{} {
	var result map[string]interface{}
	if err := mapstructure.Decode(cfg, &result); err != nil {
		return nil
	}

	return result
}

const remoteDirectoryStoreTemplate = `
        address: '{{ .Address }}'
        tenant_id: '{{ .TenantID }}'
        api_key: '{{ .APIKey }}'
        token: '{{ .Token }}'
        client_cert_path: '{{ .ClientCertPath }}'
        client_key_path: '{{ .ClientKeyPath }}'
        ca_cert_path: '{{ .CACertPath }}'
        insecure: {{ .Insecure }}
        no_tls: {{ .NoTLS }}
        no_proxy: {{ .NoProxy }}
        {{- if .Headers }}
        headers:
          {{- range $name, $value := .Headers }}
          {{ $name }}: {{ $value }}
          {{ end -}}
        {{ end -}}
`
