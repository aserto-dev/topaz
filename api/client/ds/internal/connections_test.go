package internal_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/topaz/api/client"
	"github.com/aserto-dev/topaz/api/client/ds/internal"
)

func TestConnections(t *testing.T) {
	counter := &internal.ConnectCounter{}
	conns := internal.NewConnections()
	conns.Connect = counter.Connect

	cfg := &client.Config{Address: "localhost:8282"}

	t.Run("new connection", func(t *testing.T) {
		assert := require.New(t)

		conn, err := conns.Get(cfg)
		assert.NoError(err)
		assert.NotNil(conn)
		assert.Equal(1, counter.Count)
	})

	t.Run("cached connection", func(t *testing.T) {
		assert := require.New(t)
		conn, err := conns.Get(cfg)
		assert.NoError(err)
		assert.NotNil(conn)
		assert.Equal(1, counter.Count) // no new calls to `connect`
	})

	t.Run("second connection", func(t *testing.T) {
		assert := require.New(t)
		cfg := &client.Config{Address: "localhost:8282", APIKey: "foobar"}

		conn, err := conns.Get(cfg)
		assert.NoError(err)
		assert.NotNil(conn)
		assert.Equal(2, counter.Count) // new call to `connect`
	})
}
