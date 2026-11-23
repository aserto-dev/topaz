package ds //nolint:testpackage

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/topaz/api/client"
	"github.com/aserto-dev/topaz/api/client/ds/internal"
)

const (
	base      = `{"address": "localhost:8282"}`
	noBase    = `{"reader": {"address": "localhost:9292"}, "writer": {"address": "localhost:9393"}}`
	overrides = `{"address": "localhost:8282", "reader": {"address": "localhost:9292"}}`
)

func TestUnmarshalConfig(t *testing.T) {
	t.Run("base only", func(t *testing.T) {
		assert := require.New(t)

		cfg := Config{}
		if err := json.Unmarshal([]byte(base), &cfg); err != nil {
			assert.FailNow("failed to unmarshal config", err)
		}

		assert.NotNil(cfg.Config)
		assert.Equal("localhost:8282", cfg.Address)
		assert.Nil(cfg.Reader)

		assert.NoError(cfg.Validate())
	})

	t.Run("no base", func(t *testing.T) {
		assert := require.New(t)

		var cfg Config
		if err := json.Unmarshal([]byte(noBase), &cfg); err != nil {
			assert.FailNow("failed to unmarshal config", err)
		}

		assert.Nil(cfg.Config)
		assert.Nil(cfg.Importer)
		assert.Nil(cfg.Exporter)
		assert.NotNil(cfg.Reader)
		assert.NotNil(cfg.Writer)
		assert.Equal("localhost:9292", cfg.Reader.Address)
		assert.Equal("localhost:9393", cfg.Writer.Address)
	})

	t.Run("overrides", func(t *testing.T) {
		assert := require.New(t)

		var cfg Config
		if err := json.Unmarshal([]byte(overrides), &cfg); err != nil {
			assert.FailNow("failed to unmarshal config", err)
		}

		assert.NotNil(cfg.Config)
		assert.NotNil(cfg.Reader)
		assert.Nil(cfg.Writer)
		assert.Nil(cfg.Importer)
		assert.Nil(cfg.Exporter)
		assert.Equal("localhost:8282", cfg.Address)
		assert.Equal("localhost:9292", cfg.Reader.Address)
	})
}

func TestConnect(t *testing.T) {
	t.Run("base only", func(t *testing.T) {
		assert := require.New(t)

		conns, counter := mockConns()

		cfg := &Config{
			Config: &client.Config{Address: "localhost:8282"},
		}

		dir, err := connect(conns, cfg)
		assert.NoError(err)
		assert.NotNil(dir)
		assert.NotNil(dir.Reader)
		assert.NotNil(dir.Writer)
		assert.NotNil(dir.Importer)
		assert.NotNil(dir.Exporter)
		assert.Equal(1, counter.Count)
	})

	t.Run("base with overrides", func(t *testing.T) {
		assert := require.New(t)

		conns, counter := mockConns()

		cfg := &Config{
			Config: &client.Config{Address: "localhost:1234"},
			Reader: &client.Config{Address: "localhost:4321"},
		}

		dir, err := connect(conns, cfg)
		assert.NoError(err)
		assert.NotNil(dir)
		assert.NotNil(dir.Reader)
		assert.NotNil(dir.Writer)
		assert.NotNil(dir.Importer)
		assert.NotNil(dir.Exporter)
		assert.Equal(2, counter.Count)
	})

	t.Run("no base", func(t *testing.T) {
		assert := require.New(t)

		conns, counter := mockConns()

		cfg := &Config{
			Reader: &client.Config{Address: "localhost:9292"},
		}

		dir, err := connect(conns, cfg)
		assert.NoError(err)
		assert.NotNil(dir)
		assert.NotNil(dir.Reader)
		assert.Nil(dir.Writer)
		assert.Nil(dir.Importer)
		assert.Nil(dir.Exporter)
		assert.Equal(1, counter.Count)
	})
}

func mockConns() (*internal.Connections, *internal.ConnectCounter) {
	counter := &internal.ConnectCounter{}
	conns := internal.NewConnections()
	conns.Connect = counter.Connect

	return conns, counter
}
