package bdb

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/aserto-dev/azm/cache"
	"github.com/aserto-dev/azm/model"
	cerr "github.com/aserto-dev/errors"
	"github.com/aserto-dev/topaz/internal/pkg/fs"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	berr "go.etcd.io/bbolt/errors"
	"google.golang.org/grpc/codes"
)

// Error codes returned by failures to parse an expression.
var (
	ErrPathNotFound    = cerr.NewAsertoError("E20050", codes.NotFound, http.StatusNotFound, "path not found")
	ErrKeyNotFound     = cerr.NewAsertoError("E20051", codes.NotFound, http.StatusNotFound, "key not found")
	ErrKeyExists       = cerr.NewAsertoError("E20052", codes.AlreadyExists, http.StatusConflict, "key already exists")
	ErrMultipleResults = cerr.NewAsertoError("E20053", codes.FailedPrecondition, http.StatusExpectationFailed, "multiple results for singleton request")
)

type Config struct {
	DBPath         string
	RequestTimeout time.Duration
	MaxBatchSize   int           `json:"-"` // obsolete bbolt configuration value.
	MaxBatchDelay  time.Duration `json:"-"` // obsolete bbolt configuration value.
}

// BoltDB based key-value store.
type BoltDB struct {
	logger *zerolog.Logger
	config *Config
	db     *bolt.DB
	mc     *cache.Cache
}

func New(config *Config, logger *zerolog.Logger) (*BoltDB, error) {
	newLogger := logger.With().Str("component", "kvs").Logger()
	db := BoltDB{
		config: config,
		logger: &newLogger,
		mc:     cache.New(&model.Model{}),
	}

	return &db, nil
}

// Open BoltDB key-value store instance.
func (s *BoltDB) Open() error {
	s.logger.Info().Str("db_path", s.config.DBPath).Msg("open")

	if s.config.DBPath == "" {
		return errors.New("store path not set")
	}

	dbDir := filepath.Dir(s.config.DBPath)
	if err := fs.EnsureDirPath(dbDir, fs.FileModeOwnerRWX); err != nil {
		return err
	}

	db, err := bolt.Open(s.config.DBPath, fs.FileModeOwnerRW, &bolt.Options{
		Timeout:      s.config.RequestTimeout,
		FreelistType: bolt.FreelistArrayType, // WARNING: using bolt.FreelistMapType resulted store corruptions in the migration path
	})
	if err != nil {
		return errors.Wrapf(err, "failed to open directory '%s'", s.config.DBPath)
	}

	s.db = db

	return nil
}

// Close closes BoltDB key-value store instance.
func (s *BoltDB) Close() {
	if s.db != nil {
		s.logger.Info().Str("db_path", s.config.DBPath).Msg("close")
		s.db.Close()
		s.db = nil
	}
}

func (s *BoltDB) DB() *bolt.DB {
	return s.db
}

func (s *BoltDB) Config() *Config {
	return s.config
}

// MC, model cache.
func (s *BoltDB) MC() *cache.Cache {
	return s.mc
}

// SetBucket, set bucket context to path.
func SetBucket(tx *bolt.Tx, path Path) (*bolt.Bucket, error) {
	var b *bolt.Bucket

	for index, p := range path {
		if index == 0 {
			b = tx.Bucket([]byte(p))
		} else {
			b = b.Bucket([]byte(p))
		}

		if b == nil {
			return nil, ErrPathNotFound
		}
	}

	if b == nil {
		return nil, ErrPathNotFound
	}

	return b, nil
}

// CreateBucket, create bucket path if not exists.
func CreateBucket(tx *bolt.Tx, path Path) (*bolt.Bucket, error) {
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
			return nil, err
		}
	}

	return b, nil
}

// DeleteBucket, delete tail bucket of path provided.
func DeleteBucket(tx *bolt.Tx, path Path) error {
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
		return nil //nolint:nilerr // early return when bucket does not exist, delete should not error.
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
}

// BucketExists, check if bucket path exists.
func BucketExists(tx *bolt.Tx, path Path) (bool, error) {
	_, err := SetBucket(tx, path)

	switch {
	case errors.Is(err, ErrPathNotFound):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

// ListBuckets, returns the bucket name underneath the path.
func ListBuckets(tx *bolt.Tx, path Path) ([]string, error) {
	results := []string{}

	b, err := SetBucket(tx, path)
	if err != nil {
		return results, err
	}

	iErr := b.ForEachBucket(func(k []byte) error {
		results = append(results, string(k))
		return nil
	})

	return results, iErr
}

// SetKey, set key and value in the path specified bucket.
func SetKey(tx *bolt.Tx, path Path, key, value []byte) error {
	b, err := SetBucket(tx, path)
	if err != nil {
		return err
	}

	if b == nil {
		return ErrPathNotFound
	}

	return b.Put(key, value)
}

// DeleteKey, delete key and value in path specified bucket, when it exists. None existing keys will not raise an error.
func DeleteKey(tx *bolt.Tx, path Path, key []byte) error {
	b, err := SetBucket(tx, path)
	if err != nil {
		return err
	}

	if b == nil {
		return ErrPathNotFound
	}

	return b.Delete(key)
}

// GetKey, get key and value from path specified bucket.
func GetKey(tx *bolt.Tx, path Path, key []byte) ([]byte, error) {
	b, err := SetBucket(tx, path)
	if err != nil {
		return []byte{}, err
	}

	if b == nil {
		return []byte{}, ErrPathNotFound
	}

	v := b.Get(key)
	if v == nil {
		return []byte{}, ErrKeyNotFound
	}

	return v, nil
}

// KeyExists, check if the key exists in the path specified bucket.
func KeyExists(tx *bolt.Tx, path Path, key []byte) (bool, error) {
	b, err := SetBucket(tx, path)
	if err != nil {
		return false, err
	}

	if b == nil {
		return false, ErrPathNotFound
	}

	v := b.Get(key)
	if v == nil {
		return false, nil
	}

	return true, nil
}
