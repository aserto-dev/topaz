package directory

import (
	"io"

	"github.com/aserto-dev/topaz/pkg/config"
)

const NATSKeyValueStorePlugin string = "nats_kv"

type NATSKeyValueStore struct{}

var _ config.Section = (*NATSKeyValueStore)(nil)

func (c *NATSKeyValueStore) Defaults() map[string]any {
	return map[string]any{}
}

func (c *NATSKeyValueStore) Validate() error {
	return nil
}

func (c *NATSKeyValueStore) Serialize(w io.Writer) error {
	return nil
}
