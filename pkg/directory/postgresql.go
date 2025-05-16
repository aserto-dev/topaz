package directory

import (
	"io"

	"github.com/aserto-dev/topaz/pkg/config"
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

func (c *PostgresStore) Serialize(w io.Writer) error {
	return nil
}
