package authorizer_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/config"
)

func TestMarshaling(t *testing.T) {
	t.Skip("too sensitive to whitespace")

	for _, tc := range []struct {
		name   string
		cfg    string
		verify func(*testing.T, *authorizer.Config)
	}{
		{"opa", opaConfig, func(t *testing.T, c *authorizer.Config) {
			assert.Equal(t, "instance", c.OPA.InstanceID)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := config.NewViper()
			v.ReadConfig(
				strings.NewReader(tc.cfg),
			)

			var c authorizer.Config
			err := v.Unmarshal(&c)
			require.NoError(t, err)

			tc.verify(t, &c)

			var out bytes.Buffer

			require.NoError(t,
				c.Serialize(&out),
			)

			assert.Equal(t, config.TrimN(preamble)+config.Indent(tc.cfg, 2), out.String())
		})
	}
}

const (
	preamble = `
# authorizer configuration.
authorizer:
`

	opaConfig = `
# Open Policy Agent configuration.
opa:
  instance_id: 'instance'
  graceful_shutdown_period_seconds: 1
  max_plugin_wait_time_seconds: 10

# decision logger configuration.
decision_logger:
  enabled: false

# control plane configuration
controller:
  enabled: false

# jwt validation configuration
jwt:
  acceptable_time_skew: 2s
`
)
