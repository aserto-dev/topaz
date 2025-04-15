package directory

import (
	"io"
	"text/template"
	"time"

	"github.com/aserto-dev/go-edge-ds/pkg/directory"
)

type BoltDBStore directory.Config

const BoltDBDefaultRequestTimeout = time.Second * 5

const BoltDBStorePlugin string = "boltdb"

func (c *BoltDBStore) Defaults() map[string]any {
	return map[string]any{
		"db_path":         "${TOPAZ_DB_DIR}/directory.db",
		"request_timeout": BoltDBDefaultRequestTimeout.String(),
	}
}

func (c *BoltDBStore) Validate() (bool, error) {
	return true, nil
}

func (c *BoltDBStore) Generate(w io.Writer) error {
	tmpl, err := template.New("STORE").Parse(boltDBStoreConfigTemplate)
	if err != nil {
		return err
	}

	type params struct {
		*BoltDBStore
		Plugin_ string
	}

	p := params{c, BoltDBStorePlugin}
	if err := tmpl.Execute(w, p); err != nil {
		return err
	}

	return nil
}

const boltDBStoreConfigTemplate = `
{{ .Plugin_ }}:
  db_path: '{{ .DBPath }}'
  request_timeout: {{ .RequestTimeout }}
`
