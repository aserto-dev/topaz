package config_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/aserto-dev/topaz/pkg/config"
	topaz "github.com/aserto-dev/topaz/pkg/topaz/config"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateV2toV3(t *testing.T) {
	// c2 := cfg2.Config{}
	// c3 := cfg3.Config{}
}

// func loadConfigV2(r io.Reader) (*cfg2.Config, error) {
// 	return nil, nil
// }

//nolint:wsl
func TestLoadConfigV3(t *testing.T) {
	r, err := os.Open("./schema/config.yaml")
	require.NoError(t, err)

	cfg3, err := loadConfigV3(r)
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

func loadConfigV3(r io.Reader) (*topaz.Config, error) {
	init := &topaz.ConfigV3{}

	v := config.NewViper()
	v.SetEnvPrefix("TOPAZ")
	v.AutomaticEnv()

	v.ReadConfig(r)

	if err := v.Unmarshal(init); err != nil {
		return nil, err
	}

	// config version check.
	if init.Version != topaz.Version {
		return nil, errors.Errorf("config version mismatch (got %d, expected %d)", init.Version, topaz.Version)
	}

	// logging settings validation.
	if err := init.Logging.ParseLogLevel(zerolog.Disabled); err != nil {
		return nil, errors.Wrap(err, "config log level")
	}

	cfg := &topaz.Config{}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
