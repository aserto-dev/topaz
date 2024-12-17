package directory

import (
	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
)

type Config struct {
	Store Store `json:"store"`
}

type Store struct {
	Plugin   string                 `json:"plugin"`
	Settings map[string]interface{} `json:"settings"`
}

type LocalStore struct {
	directory.Config
}

type RemoteStore struct {
	client.Config
}

type PostgresStore struct{}

type NATSKeyValueStore struct{}
