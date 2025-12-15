package directory

import (
	"context"
	"errors"
	"sync"
	"time"

	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	dsa1 "github.com/authzen/access.go/api/access/v1"

	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb/migrations/migrate"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/datasync"
	v3 "github.com/aserto-dev/topaz/internal/pkg/eds/pkg/directory/v3"

	"github.com/Masterminds/semver/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// required minimum schema version, when the current version is lower,
// migration will be invoked to update to the minimum schema version required.
const (
	schemaVersion   string = "0.0.9"
	manifestVersion int    = 2
	manifestName    string = "edge"
)

type Config struct {
	DBPath         string        `json:"db_path"`
	RequestTimeout time.Duration `json:"request_timeout"`
	Seed           bool          `json:"seed_metadata"`
	EnableV2       bool          `json:"enable_v2"`
}

type Directory struct {
	config    *Config
	logger    *zerolog.Logger
	store     *bdb.BoltDB
	exporter3 dse3.ExporterServer
	importer3 dsi3.ImporterServer
	model3    dsm3.ModelServer
	reader3   dsr3.ReaderServer
	writer3   dsw3.WriterServer
	access1   dsa1.AccessServer
}

var (
	directory *Directory
	once      sync.Once
)

func Get() (*Directory, error) {
	if directory != nil {
		return directory, nil
	}

	return nil, status.Error(codes.Internal, "directory not initialized")
}

func New(ctx context.Context, config *Config, logger *zerolog.Logger) (*Directory, error) {
	var err error

	once.Do(func() {
		directory, err = newDirectory(ctx, config, logger)
	})

	return directory, err
}

func newDirectory(_ context.Context, config *Config, logger *zerolog.Logger) (*Directory, error) {
	newLogger := logger.With().Str("component", "directory").Logger()

	cfg := bdb.Config{
		DBPath:         config.DBPath,
		RequestTimeout: config.RequestTimeout,
	}

	if ok, err := migrate.CheckSchemaVersion(&cfg, logger, semver.MustParse(schemaVersion)); !ok {
		switch {
		case errors.Is(err, migrate.ErrDirectorySchemaUpdateRequired):
			if err := migrate.Migrate(&cfg, logger, semver.MustParse(schemaVersion)); err != nil {
				return nil, err
			}
		case errors.Is(err, migrate.ErrDirectorySchemaVersionHigher):
			return nil, err
		default:
			return nil, err
		}

		if ok, err := migrate.CheckSchemaVersion(&cfg, logger, semver.MustParse(schemaVersion)); !ok {
			return nil, err
		}
	}

	store, err := bdb.New(&bdb.Config{
		DBPath:         config.DBPath,
		RequestTimeout: config.RequestTimeout,
	},
		&newLogger,
	)
	if err != nil {
		return nil, err
	}

	if err := store.Open(); err != nil {
		return nil, err
	}

	reader3 := v3.NewReader(logger, store)
	writer3 := v3.NewWriter(logger, store)
	exporter3 := v3.NewExporter(logger, store)
	importer3 := v3.NewImporter(logger, store)

	access1 := v3.NewAccess(logger, reader3)

	dir := &Directory{
		config:    config,
		logger:    &newLogger,
		store:     store,
		model3:    v3.NewModel(logger, store),
		reader3:   reader3,
		writer3:   writer3,
		exporter3: exporter3,
		importer3: importer3,
		access1:   access1,
	}

	if err := store.LoadModel(); err != nil {
		return nil, err
	}

	return dir, nil
}

func (s *Directory) Close() {
	if s.store != nil {
		s.store.Close()
		s.store = nil
	}
}

func (s *Directory) Exporter3() dse3.ExporterServer {
	return s.exporter3
}

func (s *Directory) Importer3() dsi3.ImporterServer {
	return s.importer3
}

func (s *Directory) Model3() dsm3.ModelServer {
	return s.model3
}

func (s *Directory) Reader3() dsr3.ReaderServer {
	return s.reader3
}

func (s *Directory) Writer3() dsw3.WriterServer {
	return s.writer3
}

func (s *Directory) Access1() dsa1.AccessServer {
	return s.access1
}

func (s *Directory) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Directory) Config() Config {
	return *s.config
}

func (s *Directory) DataSyncClient() datasync.SyncClient {
	return datasync.New(s.logger, s.store)
}
