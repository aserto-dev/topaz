package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/azm/model"
	v3 "github.com/aserto-dev/azm/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/fs"
	"github.com/rs/zerolog"

	"github.com/Masterminds/semver/v3"
	bolt "go.etcd.io/bbolt"
	berr "go.etcd.io/bbolt/errors"
)

const (
	versionKey  string = "version"
	baseVersion string = "0.0.4"
)

func SetBucket(tx *bolt.Tx, path bdb.Path) (*bolt.Bucket, error) {
	var b *bolt.Bucket

	for index, p := range path {
		if index == 0 {
			b = tx.Bucket([]byte(p))
		} else {
			b = b.Bucket([]byte(p))
		}

		if b == nil {
			return nil, os.ErrNotExist
		}
	}

	if b == nil {
		return nil, os.ErrNotExist
	}

	return b, nil
}

func SetKey(tx *bolt.Tx, path bdb.Path, key, value []byte) error {
	b, err := SetBucket(tx, path)
	if err != nil {
		return err
	}

	if b == nil {
		return os.ErrNotExist
	}

	return b.Put(key, value)
}

func CreateBucket(path bdb.Path) func(*zerolog.Logger, *bolt.DB, *bolt.DB) error {
	return func(log *zerolog.Logger, _ *bolt.DB, rwDB *bolt.DB) error {
		log.Info().Str("path", strings.Join(path, "/")).Msg("CreateBucket")

		if err := rwDB.Update(func(tx *bolt.Tx) error {
			var (
				b   *bolt.Bucket
				err error
			)

			for index, p := range path {
				if index == 0 {
					b, err = tx.CreateBucketIfNotExists([]byte(p))
				} else {
					b, err = b.CreateBucketIfNotExists([]byte(p))
				}
				if err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	}
}

func DeleteBucket(path bdb.Path) func(*zerolog.Logger, *bolt.DB, *bolt.DB) error {
	return func(log *zerolog.Logger, _ *bolt.DB, rwDB *bolt.DB) error {
		log.Info().Str("path", strings.Join(path, "/")).Msg("DeleteBucket")

		if err := rwDB.Update(func(tx *bolt.Tx) error {
			if len(path) == 1 {
				err := tx.DeleteBucket([]byte(path[0]))
				switch {
				case errors.Is(err, berr.ErrBucketNotFound):
					return nil
				case err != nil:
					return err
				default:
					return nil
				}
			}

			b, err := SetBucket(tx, path[:len(path)-1])
			if err != nil {
				return nil //nolint:nilerr // early return, when bucket does not exists.
			}

			err = b.DeleteBucket([]byte(path[len(path)-1]))
			switch {
			case errors.Is(err, berr.ErrBucketNotFound):
				return nil
			case err != nil:
				return err
			default:
				return nil
			}
		}); err != nil {
			return err
		}

		return nil
	}
}

func GetVersion(db *bolt.DB) (*semver.Version, error) {
	ver, _ := semver.NewVersion(baseVersion)

	err := db.View(func(tx *bolt.Tx) error {
		b, err := SetBucket(tx, bdb.SystemPath)
		if err != nil {
			return nil //nolint:nilerr // early return, when bucket does not exists.
		}

		v := b.Get([]byte(versionKey))

		// if key does not exist return base version.
		if v == nil {
			return nil
		}

		ver, _ = semver.NewVersion(string(v))

		return nil
	})

	return ver, err
}

func SetVersion(db *bolt.DB, version *semver.Version) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := SetBucket(tx, bdb.SystemPath)
		if err != nil {
			return nil //nolint:nilerr // early return, when bucket does not exists.
		}

		return b.Put([]byte(versionKey), []byte(version.String()))
	})
}

func EnsureBaseVersion(_ *zerolog.Logger, _, rwDB *bolt.DB) error {
	return rwDB.Update(func(tx *bolt.Tx) error {
		b, err := SetBucket(tx, bdb.SystemPath)
		if err != nil {
			return nil //nolint:nilerr // early return, when bucket does not exists.
		}

		return b.Put([]byte(versionKey), []byte(baseVersion))
	})
}

func BackupFilename(dbPath string, version *semver.Version) string {
	dir, file := filepath.Split(dbPath)
	ext := filepath.Ext(file)
	base := strings.TrimSuffix(file, ext)

	return filepath.Join(dir, fmt.Sprintf("%s-%s%s",
		base,
		version.String(),
		ext,
	))
}

func Backup(db *bolt.DB, version *semver.Version) error {
	dbPath := db.Path()

	return db.View(func(tx *bolt.Tx) error {
		w, err := os.Create(BackupFilename(dbPath, version))
		if err != nil {
			return err
		}

		defer func() { _ = w.Close() }()

		if _, err := tx.WriteTo(w); err != nil {
			return err
		}

		return nil
	})
}

func OpenDB(cfg *bdb.Config) (*bolt.DB, error) {
	dbDir := filepath.Dir(cfg.DBPath)
	if err := fs.EnsureDirPath(dbDir, fs.FileModeOwnerRWX); err != nil {
		return nil, err
	}

	db, err := bolt.Open(cfg.DBPath, fs.FileModeOwnerRW, &bolt.Options{
		Timeout: cfg.RequestTimeout,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func OpenReadOnlyDB(cfg *bdb.Config, version *semver.Version) (*bolt.DB, error) {
	db, err := bolt.Open(BackupFilename(cfg.DBPath, version), fs.FileModeOwnerRW, &bolt.Options{
		ReadOnly: true,
		Timeout:  cfg.RequestTimeout,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func MigrateModelV1(log *zerolog.Logger, roDB, rwDB *bolt.DB) error {
	// skip when roDB is nil.
	if roDB == nil {
		log.Debug().Msg("SKIP MigrateModel")
		return nil
	}

	ctx := context.Background()

	m, err := loadModelV1(ctx, roDB)
	if err != nil {
		return err
	}

	if err := rwDB.Update(func(tx *bolt.Tx) error {
		_, err := bdb.SetAny(ctx, tx, bdb.ManifestPathV1, bdb.ModelKey, m)
		return err
	}); err != nil {
		return err
	}

	return nil
}

func loadModelV1(ctx context.Context, roDB *bolt.DB) (*model.Model, error) {
	var m *model.Model
	if err := roDB.View(func(rtx *bolt.Tx) error {
		manifestBody, err := bdb.Get[dsm3.Body](ctx, rtx, bdb.ManifestPathV1, bdb.BodyKey)
		if err != nil {
			return err
		}

		m, err = v3.Load(bytes.NewReader(manifestBody.GetData()))
		return err
	}); err != nil {
		return m, err
	}

	return m, nil
}

func MigrateModelV2(log *zerolog.Logger, roDB, rwDB *bolt.DB) error {
	// skip when roDB is nil.
	if roDB == nil {
		log.Debug().Msg("SKIP MigrateModel")
		return nil
	}

	ctx := context.Background()

	m, err := loadModelV2(ctx, roDB)
	if err != nil {
		return err
	}

	if err := rwDB.Update(func(tx *bolt.Tx) error {
		_, err := bdb.SetAny(ctx, tx, bdb.ManifestPathV2, bdb.ModelKey, m)
		return err
	}); err != nil {
		return err
	}

	return nil
}

func loadModelV2(ctx context.Context, roDB *bolt.DB) (*model.Model, error) {
	var m *model.Model
	if err := roDB.View(func(rtx *bolt.Tx) error {
		manifestBody, err := bdb.Get[dsm3.Body](ctx, rtx, bdb.ManifestPathV2, bdb.BodyKey)
		if err != nil {
			return err
		}

		m, err = v3.Load(bytes.NewReader(manifestBody.GetData()))
		return err
	}); err != nil {
		return m, err
	}

	return m, nil
}
