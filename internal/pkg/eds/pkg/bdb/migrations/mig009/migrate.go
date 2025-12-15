package mig009

import (
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb/migrations/common"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bolt "go.etcd.io/bbolt"
)

/*
mig009

update:

remove hardcoded "0.0.1" ManifestVersion from ManifestPath

* ManifestPath      Path = []string{"_manifest", ManifestName, ManifestVersion}
* ManifestPathV2    Path = []string{"_manifest", ManifestName}

migrate to use ManifestPathV2
* ManifestPath => ManifestPathV2
* path _manifest/default/0.0.1/body => _manifest/default/body
* path _manifest/default/0.0.1/model => _manifest/default/model
* path _manifest/default/0.0.1/metadata => _manifest/default/metadata
*/

const (
	Version string = "0.0.9"
)

var fnMap = []func(*zerolog.Logger, *bolt.DB, *bolt.DB) error{
	common.CreateBucket(bdb.SystemPath),

	common.DeleteBucket(bdb.ManifestPathV1),
	common.CreateBucket(bdb.ManifestPathV2),
	updateManifest(bdb.ManifestPathV1, bdb.ManifestPathV2),

	common.CreateBucket(bdb.ObjectsPath),
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

//nolint:gocognit
func updateManifest(oldPath, newPath bdb.Path) func(*zerolog.Logger, *bolt.DB, *bolt.DB) error {
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

			rb, err := common.SetBucket(rtx, oldPath)
			if err != nil {
				return err
			}

			wb, err := common.SetBucket(wtx, newPath)
			if err != nil {
				return err
			}

			bodyValue := rb.Get(bdb.BodyKey)
			if bodyValue == nil {
				return status.Errorf(codes.NotFound, "%s not found", string(bdb.BodyKey))
			}

			if err := wb.Put(bdb.BodyKey, bodyValue); err != nil {
				return err
			}

			modelValue := rb.Get(bdb.ModelKey)
			if modelValue == nil {
				return status.Errorf(codes.NotFound, "%s not found", string(bdb.ModelKey))
			}

			if err := wb.Put(bdb.ModelKey, modelValue); err != nil {
				return err
			}

			metadataValue := rb.Get(bdb.MetadataKey)
			if metadataValue == nil {
				return status.Errorf(codes.NotFound, "%s not found", string(bdb.MetadataKey))
			}

			if err := wb.Put(bdb.MetadataKey, metadataValue); err != nil {
				return err
			}

			return wtx.Commit()
		}); err != nil {
			return err
		}

		return nil
	}
}
