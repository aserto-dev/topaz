package mig007

import (
	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"

	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb/migrations/common"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/ds"

	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// mig007
//
// change encoding from [protojson.Marshal|Unmarshal] to [proto.Marshal|Unmarshal].
// requires re-encoding each entry with the exception of manifest.model, which is using json.Marshal.
const (
	Version string = "0.0.7"
)

var fnMap = []func(*zerolog.Logger, *bolt.DB, *bolt.DB) error{
	common.DeleteBucket(bdb.SystemPath),
	common.CreateBucket(bdb.SystemPath),

	common.DeleteBucket(bdb.ManifestPathV1),
	common.CreateBucket(bdb.ManifestPathV1),
	updateManifest(bdb.ManifestPathV1),

	common.DeleteBucket(bdb.ObjectsPath),
	common.CreateBucket(bdb.ObjectsPath),
	updateEncodingObjects(),

	common.DeleteBucket(bdb.RelationsObjPath),
	common.CreateBucket(bdb.RelationsObjPath),
	common.DeleteBucket(bdb.RelationsSubPath),
	common.CreateBucket(bdb.RelationsSubPath),
	updateEncodingRelations(),
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

func updateManifest(path bdb.Path) func(*zerolog.Logger, *bolt.DB, *bolt.DB) error {
	return func(log *zerolog.Logger, roDB *bolt.DB, rwDB *bolt.DB) error {
		log.Info().Str("version", Version).Msg("updateManifest")

		if roDB == nil {
			log.Info().Bool("roDB", roDB == nil).Msg("updateManifest")
			return nil
		}

		if err := roDB.View(func(rtx *bolt.Tx) error {
			wtx, err := rwDB.Begin(true)
			if err != nil {
				return err
			}
			defer func() { _ = wtx.Rollback() }()

			b, err := common.SetBucket(rtx, path)
			if err != nil {
				return err
			}

			// re-encode body value.
			if err := encodeBody(b, wtx, path); err != nil {
				return err
			}

			// re-encode metadata value.
			if err := encodeMetaData(b, wtx, path); err != nil {
				return err
			}

			// copy model value as-is.
			{
				modelValue := b.Get(bdb.ModelKey)

				if err := common.SetKey(wtx, path, bdb.ModelKey, modelValue); err != nil {
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

func encodeBody(b *bolt.Bucket, wtx *bolt.Tx, path bdb.Path) error {
	bodyValue := b.Get(bdb.BodyKey)

	body, err := unmarshal[dsm3.Body](bodyValue)
	if err != nil {
		return err
	}

	bodyBuf, err := marshal(body)
	if err != nil {
		return err
	}

	return common.SetKey(wtx, path, bdb.BodyKey, bodyBuf)
}

func encodeMetaData(b *bolt.Bucket, wtx *bolt.Tx, path bdb.Path) error {
	metadataValue := b.Get(bdb.MetadataKey)

	metadata, err := unmarshal[dsm3.Metadata](metadataValue)
	if err != nil {
		return err
	}

	metadataBuf, err := marshal(metadata)
	if err != nil {
		return err
	}

	return common.SetKey(wtx, path, bdb.MetadataKey, metadataBuf)
}

// updateEncodingObjects, read values from read-only backup, write to new bucket.
func updateEncodingObjects() func(*zerolog.Logger, *bolt.DB, *bolt.DB) error {
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

				val, err := marshal(obj)
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

// updateEncodingRelations, read values from read-only backup, write to new bucket.
func updateEncodingRelations() func(*zerolog.Logger, *bolt.DB, *bolt.DB) error {
	return func(log *zerolog.Logger, roDB *bolt.DB, rwDB *bolt.DB) error {
		log.Info().Str("version", Version).Msg("updateRelations")

		if roDB == nil {
			log.Info().Bool("roDB", roDB == nil).Msg("updateRelations")
			return nil
		}

		if err := roDB.View(func(rtx *bolt.Tx) error {
			wtx, err := rwDB.Begin(true)
			if err != nil {
				return err
			}
			defer func() { _ = wtx.Rollback() }()

			b, err := common.SetBucket(rtx, bdb.RelationsObjPath)
			if err != nil {
				return err
			}

			c := b.Cursor()
			for key, value := c.First(); key != nil; key, value = c.Next() {
				rel, err := unmarshal[dsc3.Relation](value)
				if err != nil {
					return err
				}

				val, err := marshal(rel)
				if err != nil {
					return err
				}

				if err := common.SetKey(wtx, bdb.RelationsObjPath, ds.Relation(rel).ObjKey(), val); err != nil {
					return err
				}

				if err := common.SetKey(wtx, bdb.RelationsSubPath, ds.Relation(rel).SubKey(), val); err != nil {
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

var unmarshalOpts = protojson.UnmarshalOptions{
	DiscardUnknown: true,
}

func unmarshal[T any, M Message[T]](b []byte) (M, error) {
	var t T

	msg := M(&t)
	if err := unmarshalOpts.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return &t, nil
}

var marshalOpts = proto.MarshalOptions{
	AllowPartial:  false,
	Deterministic: false,
	UseCachedSize: false,
}

func marshal[T any, M Message[T]](t M) ([]byte, error) {
	return marshalOpts.Marshal(t)
}
