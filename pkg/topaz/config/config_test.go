package config_test

import (
	"encoding/json"
	"os"
	"testing"

	topaz "github.com/aserto-dev/topaz/pkg/topaz/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:wsl
func TestLoadConfigV3(t *testing.T) {
	r, err := os.Open("../schema/config.yaml")
	require.NoError(t, err)

	cfg3, err := topaz.NewConfig(r)
	require.NoError(t, err)

	assert.Equal(t, 3, cfg3.Version)
	assert.NotEmpty(t, cfg3.Logging)

	// print interpreted json config.
	jEnc := json.NewEncoder(os.Stdout)
	jEnc.SetEscapeHTML(false)
	jEnc.SetIndent("", "  ")

	if err := jEnc.Encode(cfg3); err != nil {
		require.NoError(t, err)
	}

	// print interpreted yaml config.
	require.NoError(t,
		cfg3.Serialize(os.Stdout),
	)

	// opa, err := cfg3.Authorizer.OPA
	// require.NoError(t, err)

	// if err := yEnc.Encode(opa); err != nil {
	// 	require.NoError(t, err)
	// }

	// cfg3.Authorizer.OPA()

	// cfg3.Authorizer.OPA()
	// rCfg := &runtime.Config{}

	// b, err := json.Marshal(cfg3.Authorizer.OPA)
	// require.NoError(t, err)

	// if err := json.Unmarshal(b, rCfg); err != nil {
	// 	require.NoError(t, err)
	// }
}
