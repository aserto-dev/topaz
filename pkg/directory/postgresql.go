package directory

import (
	"io"

	"github.com/go-viper/mapstructure/v2"
)

const PostgresStorePlugin string = "postgres"

type PostgresStore struct{}

func (c *PostgresStore) Validate() (bool, error) {
	return true, nil
}

func (c *PostgresStore) Generate(w io.Writer) error {
	return nil
}

func PostgresStoreFromMap(m map[string]interface{}) *PostgresStore {
	var cfg PostgresStore
	if err := mapstructure.Decode(m, &cfg); err != nil {
		return nil
	}

	return &cfg
}

func PostgresStoreMap(cfg *PostgresStore) map[string]interface{} {
	var result map[string]interface{}
	if err := mapstructure.Decode(cfg, &result); err != nil {
		return nil
	}

	return result
}
