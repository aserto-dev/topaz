package migrate_test

import (
	"os"
	"testing"

	config2 "github.com/aserto-dev/topaz/pkg/cc/config"
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

	f, err := os.Create("config-v3.yaml")
	require.NoError(err)

	require.NoError(
		cfg3.Serialize(f),
	)
}

func loadV2(t *testing.T) *config2.Config {
	t.Helper()

	require := req.New(t)

	r, err := os.Open("config-v2.yaml")
	require.NoError(err)

	defer func() {
		if err := r.Close(); err != nil {
			t.Logf("error closing file: %v", err)
		}
	}()

	cfg2, err := migrate.LoadConfigV2(r)
	require.NoError(err)
	require.NotNil(cfg2)

	return cfg2
}
