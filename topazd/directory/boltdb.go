package directory

import (
	"io"
	"iter"
	"text/template"
	"time"

	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/loiter"
	"github.com/samber/lo"
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

func (c *BoltDBStore) Paths() iter.Seq2[string, config.AccessMode] {
	return loiter.Seq2(lo.T2(c.DBPath, config.ReadWrite))
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
