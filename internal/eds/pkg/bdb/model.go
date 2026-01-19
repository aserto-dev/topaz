package bdb

import (
	"context"

	"github.com/aserto-dev/azm/model"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoadModel, reads the serialized model from the store
// and swaps the model instance in the cache.Cache using
// cache.UpdateModel.
func (s *BoltDB) LoadModel() error {
	ctx := context.Background()

	err := s.db.View(func(tx *bolt.Tx) error {
		if ok, _ := BucketExists(tx, ManifestPath); !ok {
			return nil
		}

		mod, err := GetAny[model.Model](ctx, tx, ManifestPath, ModelKey)

		switch {
		case status.Code(err) == codes.NotFound:
			return nil
		case err != nil:
			return err
		}

		if err := s.mc.UpdateModel(mod); err != nil {
			return err
		}

		return nil
	})

	return err
}
