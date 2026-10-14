package mig010

import (
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	m33 "github.com/aserto-dev/go-directory/aserto/directory/migrate/v33"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/eds/pkg/bdb/migrations/common"

	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
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
	common.CreateBucket(bdb.ManifestPathV3),
	updateManifestKeys(),

	common.CreateBucket(bdb.ModelPathV3),
	updateModelKeys(),

	common.CreateBucket(bdb.ObjectsPath),
	updateObjects(),

	common.CreateBucket(bdb.RelationsObjPath),
	common.CreateBucket(bdb.RelationsSubPath),
}

var _ = &m33.Object33{}

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

func updateManifestKeys() func(*zerolog.Logger, *bolt.DB, *bolt.DB) error {
	return func(log *zerolog.Logger, roDB *bolt.DB, rwDB *bolt.DB) error {
		log.Info().Str("version", Version).Msg("updateManifestKeys")

		wtx, err := rwDB.Begin(true)
		if err != nil {
			return err
		}

		defer func() { _ = wtx.Rollback() }()

		if err := common.SetKey(wtx, bdb.ManifestPathV3, bdb.ManifestKeyV3, []byte{}); err != nil {
			return err
		}

		return wtx.Commit()
	}
}

func updateModelKeys() func(*zerolog.Logger, *bolt.DB, *bolt.DB) error {
	return func(log *zerolog.Logger, roDB *bolt.DB, rwDB *bolt.DB) error {
		log.Info().Str("version", Version).Msg("updateModelKeys")

		wtx, err := rwDB.Begin(true)
		if err != nil {
			return err
		}

		defer func() { _ = wtx.Rollback() }()

		if err := common.SetKey(wtx, bdb.ModelPathV3, bdb.ModelKeyV3, []byte{}); err != nil {
			return err
		}

		if err := common.SetKey(wtx, bdb.ModelPathV3, bdb.InvertedModelKeyV3, []byte{}); err != nil {
			return err
		}

		return wtx.Commit()
	}
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
				obj, err := unmarshal[m33.Object33](value)
				if err != nil {
					return err
				}

				val, err := marshal(patchObjectWrite(obj))
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
		Resolver:       nil,
		RecursionLimit: 0,
		NoLazyDecoding: false,
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

const (
	propDisplayName string = "display_name"
)

// PatchObjectRead: transfers Objects.Properties["display_name"] to Object.DisplayName (BACKWARD COMPATIBILITY).
// func patchObjectRead(i *m33.Object33) *dsc.Object {
// fields := i.GetProperties().GetFields()

// if displayName, ok := fields[propDisplayName]; ok {
// 	i.DisplayName = displayName.GetStringValue() //nolint:staticcheck // Marked as deprecated
// }

// return i
// 	return &dsc.Object{
// 		ObjectType: i.GetType(),
// 		ObjectId:   i.GetId(),
// 		Properties: i.GetProperties(),
// 		UpdatedAt:  i.GetUpdatedAt(),
// 		Etag:       i.GetEtag(),
// 	}
// }

// PatchObjectWrite: transfers Object.DisplayName to Objects.Properties["display_name"] (FORWARD COMPATIBILITY).
func patchObjectWrite(i *m33.Object33) *dsc.Object {
	if displayName := i.GetDisplayName(); len(displayName) > 0 {
		if i.GetProperties().Fields == nil {
			i.Properties.Fields = map[string]*structpb.Value{}
		}

		i.Properties.Fields[propDisplayName] = structpb.NewStringValue(displayName)
	}

	// return i
	return &dsc.Object{
		ObjectType: i.GetType(),
		ObjectId:   i.GetId(),
		Properties: i.GetProperties(),
		UpdatedAt:  i.GetUpdatedAt(),
		Etag:       i.GetEtag(),
	}
}
