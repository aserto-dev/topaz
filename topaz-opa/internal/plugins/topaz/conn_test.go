package topaz_test

import (
	"context"
	"testing"
	"time"

	"github.com/aserto-dev/topaz/api/client"
	"github.com/aserto-dev/topaz/topaz-opa/internal/config"
	"github.com/aserto-dev/topaz/topaz-opa/internal/plugins/topaz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func enabledConfig() topaz.Config {
	return topaz.Config{
		Enabled:        true,
		Connection:     client.Config{Address: ":9292", Insecure: true},
		RequestTimeout: config.Duration(time.Second),
	}
}

func TestConnectionIsReused(t *testing.T) {
	topaz.SetConfig(enabledConfig())
	t.Cleanup(func() { topaz.CloseConn(); topaz.ResetConfig() })

	c1, err := topaz.GetConn()
	require.NoError(t, err)

	c2, err := topaz.GetConn()
	require.NoError(t, err)

	assert.Same(t, c1, c2)
}

func TestSetConfigAndCloseConnReleaseConnection(t *testing.T) {
	topaz.SetConfig(enabledConfig())
	t.Cleanup(func() { topaz.CloseConn(); topaz.ResetConfig() })

	c1, err := topaz.GetConn()
	require.NoError(t, err)

	topaz.SetConfig(enabledConfig())

	c2, err := topaz.GetConn()
	require.NoError(t, err)
	assert.NotSame(t, c1, c2, "config change must create a new connection")

	topaz.CloseConn()

	assert.Nil(t, topaz.CurrentConn())
}

func TestTimeoutInterceptor(t *testing.T) {
	invoker := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		dl, ok := ctx.Deadline()
		require.True(t, ok, "deadline must be set")
		assert.WithinDuration(t, time.Now().Add(time.Second), dl, 500*time.Millisecond)

		return nil
	}

	err := topaz.TimeoutInterceptor(config.Duration(time.Second))(context.Background(), "m", nil, nil, nil, invoker)
	require.NoError(t, err)
}

func TestTimeoutInterceptorDefault(t *testing.T) {
	invoker := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		dl, ok := ctx.Deadline()
		require.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(topaz.DefRequestTimeout), dl, time.Second)

		return nil
	}

	require.NoError(t, topaz.TimeoutInterceptor(0)(context.Background(), "m", nil, nil, nil, invoker))
}

func TestReconfigureUpdatesConfig(t *testing.T) {
	topaz.SetConfig(enabledConfig())
	t.Cleanup(func() { topaz.CloseConn(); topaz.ResetConfig() })

	p := &topaz.Plugin{}

	updated := enabledConfig()
	updated.Connection.Address = "other:9292"

	// the value type produced by topaz.PluginFactory.Validate
	p.Reconfigure(context.Background(), updated)

	assert.Equal(t, "other:9292", topaz.PluginConfig(p).Connection.Address)
	assert.Equal(t, "other:9292", topaz.GetConfig().Connection.Address)

	// unexpected types are ignored
	p.Reconfigure(context.Background(), &updated)
	p.Reconfigure(context.Background(), "bogus")
	assert.Equal(t, "other:9292", topaz.PluginConfig(p).Connection.Address)
}
