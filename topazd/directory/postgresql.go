package directory

import (
	"io"
	"iter"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/loiter"
)

const PostgresStorePlugin string = "postgres"

type PostgresStore struct{}

var _ config.Section = (*PostgresStore)(nil)

func (c *PostgresStore) Defaults() map[string]any {
	return map[string]any{}
}

func (c *PostgresStore) Validate() error {
	return nil
}

func (c *PostgresStore) Paths() iter.Seq2[string, config.AccessMode] {
	return loiter.Seq2[string, config.AccessMode]()
}

func (c *PostgresStore) Serialize(w io.Writer) error {
	return nil
}
