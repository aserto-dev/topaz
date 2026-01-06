package directory

import (
	"io"
	"iter"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/loiter"
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

func (c *NATSKeyValueStore) Paths() iter.Seq2[string, config.AccessMode] {
	return loiter.Seq2[string, config.AccessMode]()
}

func (c *NATSKeyValueStore) Serialize(w io.Writer) error {
	return nil
}
