package bdb

import (
	"context"
	"encoding/json"

	"github.com/aserto-dev/azm/model"
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
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
		if ok, _ := BucketExists(tx, ModelPathV3); !ok {
			return nil
		}

		m, err := Get[dsc.Model](ctx, tx, ModelPathV3, ModelKeyV3)
		switch {
		case status.Code(err) == codes.NotFound:
			return nil
		case err != nil:
			return err
		}

		var mod model.Model

		buf, err := m.GetContent().MarshalJSON()
		if err != nil {
			return err
		}

		if err := json.Unmarshal(buf, &mod); err != nil {
			return err
		}

		if err := s.mc.UpdateModel(&mod); err != nil {
			return err
		}

		return nil
	})

	return err
}
