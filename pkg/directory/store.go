package directory

import (
	"iter"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/loiter"
)

type Store struct {
	Provider string `json:"provider"`

	Bolt     BoltDBStore          `json:"boltdb"`
	Remote   RemoteDirectoryStore `json:"remote_directory"`
	Postgres PostgresStore        `json:"postgres"`
	NatsKV   NATSKeyValueStore    `json:"nats_kv"`
}

func (s *Store) Paths() iter.Seq2[string, config.AccessMode] {
	switch s.Provider {
	case BoltDBStorePlugin:
		return s.Bolt.Paths()
	case RemoteDirectoryStorePlugin:
		return s.Remote.Paths()
	case PostgresStorePlugin:
		return s.Postgres.Paths()
	case NATSKeyValueStorePlugin:
		return s.NatsKV.Paths()
	default:
		return loiter.Seq2[string, config.AccessMode]()
	}
}
