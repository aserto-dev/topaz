// nolint
package config_test

import (
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	cfg3 "github.com/aserto-dev/topaz/pkg/config"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMigrateV2toV3(t *testing.T) {
	// c2 := cfg2.Config{}
	// c3 := cfg3.Config{}
}

// func loadConfigV2(r io.Reader) (*cfg2.Config, error) {
// 	return nil, nil
// }

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
	yEnc := yaml.NewEncoder(os.Stdout)
	yEnc.SetIndent(2)
	if err := yEnc.Encode(cfg3); err != nil {
		require.NoError(t, err)
	}

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

func loadConfigV3(r io.Reader) (*cfg3.Config, error) {
	init := &cfg3.ConfigV3{}

	v := viper.NewWithOptions(viper.EnvKeyReplacer(newReplacer()))
	v.SetConfigType("yaml")
	v.SetEnvPrefix("TOPAZ")
	v.AutomaticEnv()

	v.ReadConfig(r)

	if err := v.Unmarshal(init, func(dc *mapstructure.DecoderConfig) { dc.TagName = "json" }); err != nil {
		return nil, err
	}

	// config version check.
	if init.Version != cfg3.Version {
		return nil, errors.Errorf("config version mismatch (got %d, expected %d)", init.Version, cfg3.Version)
	}

	// logging settings validation.
	if err := init.Logging.ParseLogLevel(zerolog.Disabled); err != nil {
		return nil, errors.Wrap(err, "config log level")
	}

	cfg := &cfg3.Config{}

	if err := v.Unmarshal(cfg, func(dc *mapstructure.DecoderConfig) { dc.TagName = "json" }); err != nil {
		return nil, err
	}

	return cfg, nil
}

type replacer struct {
	r *strings.Replacer
}

func newReplacer() *replacer {
	return &replacer{r: strings.NewReplacer(".", "_")}
}

func (r replacer) Replace(s string) string {
	if s == "TOPAZ_VERSION" {
		// Prevent the `version` field from be overridden by env vars.
		return ""
	}

	return r.r.Replace(s)
}

var envRegex = regexp.MustCompile(`(?U:\${.*})`)

// subEnvVars will look for any environment variables in the passed in string
// with the syntax of ${VAR_NAME} and replace that string with ENV[VAR_NAME].
func subEnvVars(s string) string {
	updatedConfig := envRegex.ReplaceAllStringFunc(strings.ReplaceAll(s, `"`, `'`), func(s string) string {
		// Trim off the '${' and '}'
		if len(s) <= 3 {
			// This should never happen..
			return ""
		}
		varName := s[2 : len(s)-1]

		// Lookup the variable in the environment. We play by
		// bash rules.. if its undefined we'll treat it as an
		// empty string instead of raising an error.
		return os.Getenv(varName)
	})

	return updatedConfig
}
