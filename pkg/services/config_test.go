package services_test

import (
	"testing"

	"github.com/aserto-dev/topaz/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name   string
	cfg    services.Config
	verify func(*testing.T, error)
}

var testCases = []testCase{
	{
		"no port collisions",
		services.Config{
			"directory":  withPorts("1234", "5678"),
			"authorizer": withPorts("4321", "8765"),
		},
		func(t *testing.T, err error) {
			require.NoError(t, err)
		},
	},
	{
		"port collision within a service",
		services.Config{
			"directory": withPorts("1234", "1234"),
		},
		func(t *testing.T, err error) {
			require.Error(t, err)
			require.ErrorIs(t, err, services.ErrPortCollision)
			assert.ErrorContains(t, err, "0.0.0.0:1234 [directory (grpc), directory (http)]")
		},
	},
	{
		"grpc collision",
		services.Config{
			"directory":  withPorts("1234", "1"),
			"authorizer": withPorts("1234", "2"),
		},
		func(t *testing.T, err error) {
			require.Error(t, err)
			require.ErrorIs(t, err, services.ErrPortCollision)
			assert.ErrorContains(t, err, "0.0.0.0:1234 [authorizer (grpc), directory (grpc)]")
		},
	},
	{
		"http collision",
		services.Config{
			"directory":  withPorts("1", "1234"),
			"authorizer": withPorts("2", "1234"),
		},
		func(t *testing.T, err error) {
			require.Error(t, err)
			require.ErrorIs(t, err, services.ErrPortCollision)
			assert.ErrorContains(t, err, "0.0.0.0:1234 [authorizer (http), directory (http)]")
		},
	},
	{
		"http/grpc collision",
		services.Config{
			"directory":  withPorts("1", "1234"),
			"authorizer": withPorts("1234", "2"),
		},
		func(t *testing.T, err error) {
			require.Error(t, err)
			require.ErrorIs(t, err, services.ErrPortCollision)
			assert.ErrorContains(t, err, "0.0.0.0:1234 [authorizer (grpc), directory (http)]")
		},
	},
}

func TestValidate(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.verify(t, tc.cfg.Validate())
		})
	}
}

func withPorts(grpc, http string) *services.Service {
	return &services.Service{
		GRPC: services.GRPCService{
			ListenAddress: listenAddr(grpc),
		},
		Gateway: services.GatewayService{
			ListenAddress: listenAddr(http),
		},
	}
}

func listenAddr(port string) string {
	return "0.0.0.0:" + port
}
