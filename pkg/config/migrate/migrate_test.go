package migrate_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aserto-dev/topaz/pkg/config/migrate"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigV2(t *testing.T) {
	r, err := os.Open("/Users/gertd/.config/topaz/cfg/gdrive.yaml")
	require.NoError(t, err)

	cfg2, err := migrate.LoadConfigV2(r)
	require.NoError(t, err)
	require.NotNil(t, cfg2)
}

func TestMigrateConfig(t *testing.T) {
	r, err := os.Open("/Users/gertd/.config/topaz/cfg/gdrive.yaml")
	require.NoError(t, err)
	defer func() {
		if err := r.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}()

	cfg2, err := migrate.LoadConfigV2(r)
	require.NoError(t, err)

	cfg3, err := migrate.Migrate(cfg2)
	require.NoError(t, err)

	if err := cfg3.Generate(os.Stdout); err != nil {
		require.NoError(t, err)
	}
}
