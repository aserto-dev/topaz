package runtime_test

import (
	"testing"

	runtime "github.com/aserto-dev/topaz/internal/runtime"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/require"
)

func TestDeepCopy(t *testing.T) {
	assert := require.New(t)
	runtimeConfig := &runtime.Config{}

	_, err := copystructure.Copy(runtimeConfig)

	assert.NoError(err)
}
