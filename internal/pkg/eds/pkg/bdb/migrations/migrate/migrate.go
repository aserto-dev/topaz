package migrate

import (
	"net/http"
	"os"

	cerr "github.com/aserto-dev/errors"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb/migrations/common"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb/migrations/mig004"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb/migrations/mig005"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb/migrations/mig006"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb/migrations/mig007"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb/migrations/mig008"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb/migrations/mig009"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/fs"

	"github.com/Masterminds/semver/v3"
	"github.com/rs/zerolog"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
)

type Migration func(*zerolog.Logger, *bolt.DB, *bolt.DB) error

// list of migration steps, keyed by version.
var migMap = map[string]Migration{
	mig004.Version: mig004.Migrate,
	mig005.Version: mig005.Migrate,
	mig006.Version: mig006.Migrate,
	mig007.Version: mig007.Migrate,
	mig008.Version: mig008.Migrate,
	mig009.Version: mig009.Migrate,
}

//nolint:lll // single line readability more important.
var (
	ErrDirectorySchemaVersionHigher  = cerr.NewAsertoError("E20054", codes.FailedPrecondition, http.StatusExpectationFailed, "directory schema version is higher than supported by engine")
	ErrDirectorySchemaUpdateRequired = cerr.NewAsertoError("E20055", codes.FailedPrecondition, http.StatusExpectationFailed, "directory schema update required")
	ErrUnknown                       = cerr.NewAsertoError("E99999", codes.Unknown, http.StatusInternalServerError, "unexpected error occurred")
)

// CheckSchemaVersion, validate schema version of the database file
// equal:  returns  true, nil
// lower:  returns false, nil
// higher  returns false, error
// errors: returns false, error.
func CheckSchemaVersion(config *bdb.Config, logger *zerolog.Logger, reqVersion *semver.Version) (bool, error) {
	if !fs.FileExists(config.DBPath) {
		if err := create(config, logger, reqVersion); err != nil {
			return false, err
		}

		return true, nil
	}

	boltdb, err := bdb.New(config, logger)
	if err != nil {
		return false, err
	}

	if err := boltdb.Open(); err != nil {
		return false, err
	}
	defer boltdb.Close()

	curVersion, err := common.GetVersion(boltdb.DB())
	if err != nil {
		return false, err
	}

	logger.Info().Str("current", curVersion.String()).Msg("schema_version")

	switch {
	case curVersion.Equal(reqVersion):
		return true, nil
	case curVersion.LessThan(reqVersion):
		return false, ErrDirectorySchemaUpdateRequired
	case curVersion.GreaterThan(reqVersion):
		return false, ErrDirectorySchemaVersionHigher.Msg(curVersion.String())
	default:
		return false, ErrUnknown
	}
}

func Migrate(config *bdb.Config, logger *zerolog.Logger, reqVersion *semver.Version) error {
	log := logger.With().Str("component", "migrate").Logger()

	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("recovered schema migration %s", r)
		}
	}()

	curVersion, err := getCurrent(config, &log)
	if err != nil {
		return err
	}

	// if current is equal to required, no further action required.
	if curVersion.Equal(reqVersion) {
		return nil
	}

	log.Info().Str("current", curVersion.String()).Str("required", reqVersion.String()).Msg("begin schema migration")

	for {
		nextVersion := curVersion.IncPatch()
		log.Info().Str("next", nextVersion.String()).Msg("starting")

		if err := migrate(config, &log, curVersion, &nextVersion); err != nil {
			log.Error().Err(err).Msg("migrate")
			return err
		}

		log.Info().Str("next", nextVersion.String()).Msg("finished")

		curVersion, err = getCurrent(config, logger)
		if err != nil {
			return err
		}

		log.Info().Str("current", curVersion.String()).Msg("updated current version")

		if curVersion.Equal(reqVersion) {
			break
		}
	}

	log.Info().Str("current", curVersion.String()).Str("required", reqVersion.String()).Msg("finished schema migration")

	return nil
}

func getCurrent(config *bdb.Config, logger *zerolog.Logger) (*semver.Version, error) {
	boltdb, err := bdb.New(config, logger)
	if err != nil {
		return nil, err
	}

	if err := boltdb.Open(); err != nil {
		return nil, err
	}

	defer boltdb.Close()

	return common.GetVersion(boltdb.DB())
}

func create(config *bdb.Config, log *zerolog.Logger, version *semver.Version) error {
	rwDB, err := common.OpenDB(config)
	if err != nil {
		return err
	}

	defer func() {
		log.Debug().Str("db_path", rwDB.Path()).Msg("close-rw")

		if err := rwDB.Close(); err != nil {
			log.Error().Err(err).Msg("close rwDB")
		}

		rwDB = nil
	}()

	// create flow is signaled by roDB == nil.
	if err := execute(log, nil, rwDB, version); err != nil {
		return err
	}

	if err := common.SetVersion(rwDB, version); err != nil {
		return err
	}

	if err := rwDB.Sync(); err != nil {
		return err
	}

	return nil
}

func migrate(config *bdb.Config, log *zerolog.Logger, curVersion, nextVersion *semver.Version) error {
	rwDB, err := common.OpenDB(config)
	if err != nil {
		return err
	}

	defer func() {
		log.Debug().Str("db_path", rwDB.Path()).Msg("close-rw")

		if err := rwDB.Close(); err != nil {
			log.Error().Err(err).Msg("close rwDB")
		}

		rwDB = nil
	}()

	if err := common.Backup(rwDB, curVersion); err != nil {
		return err
	}

	roDB, err := common.OpenReadOnlyDB(config, curVersion)
	if err != nil {
		return err
	}

	defer func() {
		log.Debug().Str("db_path", roDB.Path()).Msg("close-ro")

		if err := roDB.Close(); err != nil {
			log.Error().Err(err).Msg("close roDB")
		}

		roDB = nil
	}()

	if err := execute(log, roDB, rwDB, nextVersion); err != nil {
		return err
	}

	if err := common.SetVersion(rwDB, nextVersion); err != nil {
		return err
	}

	if err := rwDB.Sync(); err != nil {
		return err
	}

	return nil
}

func execute(logger *zerolog.Logger, roDB, rwDB *bolt.DB, newVersion *semver.Version) error {
	if fnMigrate, ok := migMap[newVersion.String()]; ok {
		return fnMigrate(logger, roDB, rwDB)
	}

	return os.ErrNotExist
}
