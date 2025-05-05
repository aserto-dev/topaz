package topaz_test

import (
	"context"
	"testing"

	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/pkg/topaz"
	"github.com/stretchr/testify/require"
)

func TestTopazRun(t *testing.T) {
	ctx := context.Background()

	t.Setenv(x.EnvTopazDBDir, t.TempDir())

	topazApp, err := topaz.NewTopaz(ctx, "./schema/config.yaml")
	require.NoError(t, err)

	_, err = topazApp.Start(ctx)
	require.NoError(t, err)
}
