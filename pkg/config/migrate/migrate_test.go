package migrate_test

import (
	"os"
	"testing"

	"github.com/aserto-dev/topaz/pkg/config/migrate"
	"github.com/aserto-dev/topaz/pkg/config/v3"

	req "github.com/stretchr/testify/require"
)

func TestLoadConfigV2(t *testing.T) {
	require := req.New(t)

	r, err := os.Open("config-v2.yaml")
	require.NoError(err)

	cfg2, err := migrate.LoadConfigV2(r)
	require.NoError(err)
	require.NotNil(cfg2)
}

func TestMigrateConfig(t *testing.T) {
	require := req.New(t)

	r, err := os.Open("config-v2.yaml")
	require.NoError(err)

	t.Cleanup(func() {
		if err := r.Close(); err != nil {
			t.Logf("error closing file: %v", err)
		}
	})

	cfg, err := config.NewConfig(r)
	require.NoError(err)

	require.NoError(
		cfg.Serialize(os.Stdout),
	)
}
