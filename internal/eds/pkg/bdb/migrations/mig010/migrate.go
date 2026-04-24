package mig010

import (
	dsc3 "github.com/aserto-dev/topaz/api/directory/v4"

	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb/migrations/common"
	"github.com/aserto-dev/topaz/internal/eds/pkg/ds"

	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

/*
mig010

update:

In preparation of directory schema v4, move field Object.DisplayName into the property bag field `display_name`.

*/

const (
	Version string = "0.0.10"
)

var fnMap = []func(*zerolog.Logger, *bolt.DB, *bolt.DB) error{
	common.CreateBucket(bdb.SystemPath),
	common.CreateBucket(bdb.ManifestPathV2),

	common.CreateBucket(bdb.ObjectsPath),
	updateObjects(),

	common.CreateBucket(bdb.RelationsObjPath),
	common.CreateBucket(bdb.RelationsSubPath),
}

func Migrate(log *zerolog.Logger, roDB, rwDB *bolt.DB) error {
	log.Info().Str("version", Version).Msg("StartMigration")

	for _, fn := range fnMap {
		if err := fn(log, roDB, rwDB); err != nil {
			return err
		}
	}

	log.Info().Str("version", Version).Msg("FinishedMigration")

	return nil
}

// updateObjects, move field DisplayName into the property bag field `display_name`.
func updateObjects() func(*zerolog.Logger, *bolt.DB, *bolt.DB) error {
	return func(log *zerolog.Logger, roDB *bolt.DB, rwDB *bolt.DB) error {
		log.Info().Str("version", Version).Msg("updateObjects")

		if roDB == nil {
			log.Info().Bool("roDB", roDB == nil).Msg("updateObjects")
			return nil
		}

		if err := roDB.View(func(rtx *bolt.Tx) error {
			wtx, err := rwDB.Begin(true)
			if err != nil {
				return err
			}

			defer func() { _ = wtx.Rollback() }()

			b, err := common.SetBucket(rtx, bdb.ObjectsPath)
			if err != nil {
				return err
			}

			c := b.Cursor()
			for key, value := c.First(); key != nil; key, value = c.Next() {
				obj, err := unmarshal[dsc3.Object](value)
				if err != nil {
					return err
				}

				val, err := marshal(ds.PatchObjectWrite(obj))
				if err != nil {
					return err
				}

				if err := common.SetKey(wtx, bdb.ObjectsPath, key, val); err != nil {
					return err
				}
			}

			return wtx.Commit()
		}); err != nil {
			return err
		}

		return nil
	}
}

type Message[T any] interface {
	proto.Message
	*T
}

var (
	marshalOpts = proto.MarshalOptions{
		AllowPartial:  false,
		Deterministic: false,
		UseCachedSize: false,
	}
	unmarshalOpts = proto.UnmarshalOptions{
		Merge:          false,
		AllowPartial:   false,
		DiscardUnknown: true,
	}
)

func marshal[T any, M Message[T]](t M) ([]byte, error) {
	return marshalOpts.Marshal(t)
}

func unmarshal[T any, M Message[T]](b []byte) (M, error) {
	var t T

	msg := M(&t)
	if err := unmarshalOpts.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return &t, nil
}
