package servers_test

import (
	"slices"
	"testing"

	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name   string
	cfg    servers.Config
	verify func(*testing.T, error)
}

var testCases = []testCase{
	{
		"no port collisions",
		servers.Config{
			"directory":  withPorts("1234", "5678"),
			"authorizer": withPorts("4321", "8765"),
		},
		func(t *testing.T, err error) {
			require.NoError(t, err)
		},
	},
	{
		"port collision within a service",
		servers.Config{
			"directory": withPorts("1234", "1234"),
		},
		func(t *testing.T, err error) {
			require.Error(t, err)
			require.ErrorIs(t, err, servers.ErrPortCollision)
			assert.ErrorContains(t, err, "0.0.0.0:1234 [directory (grpc), directory (http)]")
		},
	},
	{
		"grpc collision",
		servers.Config{
			"directory":  withPorts("1234", "1"),
			"authorizer": withPorts("1234", "2"),
		},
		func(t *testing.T, err error) {
			require.Error(t, err)
			require.ErrorIs(t, err, servers.ErrPortCollision)
			assert.ErrorContains(t, err, "0.0.0.0:1234 [authorizer (grpc), directory (grpc)]")
		},
	},
	{
		"http collision",
		servers.Config{
			"directory":  withPorts("1", "1234"),
			"authorizer": withPorts("2", "1234"),
		},
		func(t *testing.T, err error) {
			require.Error(t, err)
			require.ErrorIs(t, err, servers.ErrPortCollision)
			assert.ErrorContains(t, err, "0.0.0.0:1234 [authorizer (http), directory (http)]")
		},
	},
	{
		"http/grpc collision",
		servers.Config{
			"directory":  withPorts("1", "1234"),
			"authorizer": withPorts("1234", "2"),
		},
		func(t *testing.T, err error) {
			require.Error(t, err)
			require.ErrorIs(t, err, servers.ErrPortCollision)
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

func TestEnabledServices(t *testing.T) {
	cfg := servers.Config{
		"directory":  &servers.Server{Services: []servers.ServiceName{"reader", "writer"}},
		"authorizer": &servers.Server{Services: []servers.ServiceName{"authorizer", "access"}},
	}

	assert.Subset(t, slices.Collect(cfg.EnabledServices()), []servers.ServiceName{"reader", "writer", "authorizer", "access"})
}

func TestListenAddresses(t *testing.T) {
	cfg := servers.Config{
		servers.ServerName("directory"): &servers.Server{
			GRPC: servers.GRPCServer{ListenAddress: "0.0.0.0:9292"},
			HTTP: servers.HTTPServer{ListenAddress: "0.0.0.0:9393"},
		},
		servers.ServerName("authorizer"): &servers.Server{
			GRPC: servers.GRPCServer{ListenAddress: "0.0.0.0:8282"},
			HTTP: servers.HTTPServer{ListenAddress: "0.0.0.0:8383"},
		},
	}

	expected := map[lo.Tuple2[servers.ServerName, servers.ListenAddress]]struct{}{
		lo.T2(servers.ServerName("directory"), servers.ListenAddress{"0.0.0.0:9292", servers.Kind.GRPC}):  {},
		lo.T2(servers.ServerName("directory"), servers.ListenAddress{"0.0.0.0:9393", servers.Kind.HTTP}):  {},
		lo.T2(servers.ServerName("authorizer"), servers.ListenAddress{"0.0.0.0:8282", servers.Kind.GRPC}): {},
		lo.T2(servers.ServerName("authorizer"), servers.ListenAddress{"0.0.0.0:8383", servers.Kind.HTTP}): {},
	}

	for name, address := range cfg.ListenAddresses() {
		key := lo.T2(name, address)
		assert.Contains(t, expected, key)
		delete(expected, key)
	}

	assert.Empty(t, expected)
}

func withPorts(grpc, http string) *servers.Server {
	return &servers.Server{
		GRPC: servers.GRPCServer{
			ListenAddress: listenAddr(grpc),
		},
		HTTP: servers.HTTPServer{
			ListenAddress: listenAddr(http),
		},
	}
}

func listenAddr(port string) string {
	return "0.0.0.0:" + port
}
