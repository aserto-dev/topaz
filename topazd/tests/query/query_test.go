package query_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	client "github.com/aserto-dev/go-aserto"
	azc "github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	rt "github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/internal/pkg/fs"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/x"
	assets_test "github.com/aserto-dev/topaz/topazd/tests/assets"
	tc "github.com/aserto-dev/topaz/topazd/tests/common"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestQuery(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	t.Logf("\nTEST CONTAINER IMAGE: %q\n", tc.TestImage())

	req := testcontainers.ContainerRequest{
		Image:        tc.TestImage(),
		ExposedPorts: []string{"9292/tcp"},
		Env: map[string]string{
			x.EnvTopazCertsDir:  x.DefCertsDir,
			x.EnvTopazDBDir:     x.DefDBDir,
			x.EnvTopazDecisions: x.DefDecisionsDir,
		},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            assets_test.ConfigNoTLSReader(),
				ContainerFilePath: "/config/config.yaml",
				FileMode:          int64(fs.FileModeOwnerRWX),
			},
			{
				Reader:            assets_test.AcmecorpReader(),
				ContainerFilePath: "/data/test.db",
				FileMode:          int64(fs.FileModeOwnerRWX),
			},
		},
		WaitingFor: wait.ForAll(
			wait.ForExposedPort(),
			wait.ForLog("Starting 0.0.0.0:9292 gRPC server"),
		).WithStartupTimeoutDefault(300 * time.Second),
	}

	topaz, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          false,
	})
	require.NoError(t, err)

	if err := topaz.Start(ctx); err != nil {
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, topaz)
		cancel()
	})

	grpcAddr, err := tc.MappedAddr(ctx, topaz, "9292")
	require.NoError(t, err)

	t.Run("testQuery", testQuery(grpcAddr))
}

func testQuery(addr string) func(*testing.T) {
	return func(t *testing.T) {
		opts := []client.ConnectionOption{
			client.WithAddr(addr),
			client.WithNoTLS(true),
		}

		azClient, err := azc.New(opts...)
		require.NoError(t, err)
		t.Cleanup(func() { _ = azClient.Close() })

		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		for _, tc := range queryTests {
			f := func(t *testing.T) {
				resp, err := azClient.Query(ctx, tc.query)
				tc.validate(t, resp, err)
			}

			t.Run(tc.name, f)
		}
	}
}

var queryTests = []struct {
	name     string
	query    *authorizer.QueryRequest
	validate func(*testing.T, *authorizer.QueryResponse, error)
}{
	{
		name: "opa.runtime",
		query: &authorizer.QueryRequest{
			Query:           "x := opa.runtime()",
			IdentityContext: &api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_NONE},
		},
		validate: validateResult(
			contains("env"),
		),
	},
	{
		name: "data",
		query: &authorizer.QueryRequest{
			Query:           "x = data",
			IdentityContext: &api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_NONE},
		},
		validate: validateResult(
			contains("rebac"),
		),
	},
	{
		name: "ds.user",
		query: &authorizer.QueryRequest{
			Query:           `x = ds.user({"id": "euang@acmecorp.com"})`,
			IdentityContext: &api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_NONE},
		},
		validate: validateResult(
			contains("id"),
			func(t *testing.T, result *rt.Result) {
				binding, _ := result.Result[0].Bindings["x"].(map[string]any)
				require.Equal(t, "euang@acmecorp.com", binding["id"])
			},
		),
	},
	{
		name: "ds.identity",
		query: &authorizer.QueryRequest{
			Query:           `x = ds.identity({"id": "euang@acmecorp.com"})`,
			IdentityContext: &api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_NONE},
		},
		validate: validateResult(
			contains("euang@acmecorp.com"),
		),
	},
	{
		name: "identity_context_sub",
		query: &authorizer.QueryRequest{
			Query: "x = input",
			IdentityContext: &api.IdentityContext{
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
				Identity: "euang@acmecorp.com",
			},
		},
		validate: validateResult(
			contains("identity"),
			contains("user"),
			func(t *testing.T, result *rt.Result) {
				bindings, _ := result.Result[0].Bindings["x"].(map[string]any)
				require.Contains(t, bindings["identity"], "type")
				require.Contains(t, bindings["user"], "id")
			},
		),
	},
	{
		name: "identity_context_manual",
		query: &authorizer.QueryRequest{
			Query: "x = input",
			IdentityContext: &api.IdentityContext{
				Type:     api.IdentityType_IDENTITY_TYPE_MANUAL,
				Identity: "euang@acmecorp.com",
			},
		},
		validate: validateResult(
			contains("identity"),
			contains("user"),
			func(t *testing.T, result *rt.Result) {
				bindings, _ := result.Result[0].Bindings["x"].(map[string]any)
				require.Contains(t, bindings["identity"], "identity")
				require.Contains(t, bindings["identity"], "type")
				require.Equal(t, map[string]any{}, bindings["user"])
			},
		),
	},
}

func contains(val string) resultValidator {
	return func(t *testing.T, result *rt.Result) {
		require.Contains(t, result.Result[0].Bindings, "x")
		require.Contains(t, result.Result[0].Bindings["x"], val)
	}
}

type resultValidator func(t *testing.T, result *rt.Result)

func validateResult(check ...resultValidator) func(t *testing.T, resp *authorizer.QueryResponse, err error) {
	return func(t *testing.T, resp *authorizer.QueryResponse, err error) {
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.GetResponse())

		var result *rt.Result

		buf, err := resp.GetResponse().MarshalJSON()
		require.NoError(t, err)

		if err := json.Unmarshal(buf, &result); err != nil {
			require.NoError(t, err)
		}

		require.NotNil(t, result)
		require.NotEmpty(t, result.Result)

		for _, v := range check {
			v(t, result)
		}
	}
}
