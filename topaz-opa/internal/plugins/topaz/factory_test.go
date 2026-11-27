package topaz_test

import (
	"testing"

	"github.com/aserto-dev/topaz/topaz-opa/internal/config"
	"github.com/aserto-dev/topaz/topaz-opa/internal/plugins/topaz"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigBeforeSetConfig(t *testing.T) {
	topaz.ResetConfig()
	t.Cleanup(func() { topaz.ResetConfig() })

	assert.NotPanics(t, func() {
		cfg := topaz.GetConfig()
		assert.False(t, cfg.Enabled)
		assert.Equal(t, config.Duration(topaz.DefRequestTimeout), cfg.RequestTimeout)
	})
}

func TestGetDirectoryClientDisabledWithoutConfig(t *testing.T) {
	topaz.ResetConfig()
	t.Cleanup(func() { topaz.ResetConfig() })

	_, err := topaz.GetDirectoryClient()()
	assert.Error(t, err)
}
