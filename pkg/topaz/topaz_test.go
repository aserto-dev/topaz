package topaz_test

import (
	"context"
	"testing"

	"github.com/aserto-dev/topaz/pkg/topaz"
	"github.com/stretchr/testify/require"
)

func TestTopazRun(t *testing.T) {
	ctx := context.Background()

	topazApp, err := topaz.NewTopaz("./schema/config.yaml")
	require.NoError(t, err)

	require.NoError(t,
		topazApp.Run(ctx),
	)
}
