package migrate_test

import (
	"os"
	"path"
	"testing"

	config2 "github.com/aserto-dev/topaz/pkg/cc/config"
	cnfg "github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/config/migrate"
	"github.com/aserto-dev/topaz/pkg/config/v3"

	req "github.com/stretchr/testify/require"
)

func TestMigrateConfig(t *testing.T) {
	require := req.New(t)

	cfg2 := loadV2(t)

	var cfg3 config.Config

	require.NoError(
		migrate.Migrate(cfg2, &cfg3),
	)

	tmp := path.Join(t.TempDir(), "config.yaml")
	serialize(t, &cfg3, tmp)
	printToLog(t, tmp)

	expected := loadV3(t, "config-v3.yaml")
	actual := loadV3(t, tmp)

	require.Equal(expected, actual)
}

func loadV2(t *testing.T) *config2.Config {
	t.Helper()

	r := open(t, "config-v2.yaml")

	require := req.New(t)

	cfg2, err := migrate.LoadConfigV2(r)
	require.NoError(err)
	require.NotNil(cfg2)

	return cfg2
}

func loadV3(t *testing.T, src string) *config.Config {
	t.Helper()

	r := open(t, src)

	require := req.New(t)

	cfg3, err := config.NewConfig(r)
	require.NoError(err)
	require.NotNil(cfg3)

	return cfg3
}

func open(t *testing.T, path string) *os.File {
	t.Helper()

	require := req.New(t)

	f, err := os.Open(path)
	require.NoError(err)

	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Logf("error closing file: %v", err)
		}
	})

	return f
}

func serialize(t *testing.T, cfg cnfg.Section, dest string) {
	t.Helper()

	require := req.New(t)

	f, err := os.Create(dest)
	require.NoError(err)

	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Logf("error closing file: %v", err)
		}
	})

	require.NoError(cfg.Serialize(f))
}

func printToLog(t *testing.T, path string) {
	t.Helper()

	b, err := os.ReadFile(path)
	req.NoError(t, err)

	t.Logf("Serialized config:\n%s", b)
}
