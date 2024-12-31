package directory

import (
	"os"

	"github.com/mitchellh/mapstructure"
)

const NATSKeyValueStorePlugin string = "nats_kv"

type NATSKeyValueStore struct{}

func (c *NATSKeyValueStore) Validate() (bool, error) {
	return true, nil
}

func (c *NATSKeyValueStore) Generate(w *os.File) error {
	return nil
}

func NATSKeyValueStoreFromMap(m map[string]interface{}) *NATSKeyValueStore {
	var cfg NATSKeyValueStore
	if err := mapstructure.Decode(m, &cfg); err != nil {
		return nil
	}

	return &cfg
}

func NATSKeyValueStoreMap(cfg *NATSKeyValueStore) map[string]interface{} {
	var result map[string]interface{}
	if err := mapstructure.Decode(cfg, &result); err != nil {
		return nil
	}
	return result
}
