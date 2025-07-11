package directory

import (
	"io"
	"text/template"
	"time"

	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/config"
)

type BoltDBStore directory.Config

const BoltDBDefaultRequestTimeout = time.Second * 5

const BoltDBStorePlugin string = "boltdb"

var _ config.Section = (*BoltDBStore)(nil)

func (*BoltDBStore) Defaults() map[string]any {
	return map[string]any{
		"db_path":         "${TOPAZ_DB_DIR}/directory.db",
		"request_timeout": BoltDBDefaultRequestTimeout.String(),
	}
}

func (*BoltDBStore) Validate() error {
	return nil
}

func (c *BoltDBStore) Serialize(w io.Writer) error {
	tmpl, err := template.New("STORE").Parse(boltDBStoreConfigTemplate)
	if err != nil {
		return err
	}

	type params struct {
		*BoltDBStore
		Provider_ string
	}

	p := params{c, BoltDBStorePlugin}

	return tmpl.Execute(w, p)
}

const boltDBStoreConfigTemplate = `
{{- .Provider_ }}:
  db_path: '{{ .DBPath }}'
  request_timeout: {{ .RequestTimeout }}
`
