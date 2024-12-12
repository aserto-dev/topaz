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
	assets_test "github.com/aserto-dev/topaz/pkg/app/tests/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestQuery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	t.Logf("\nTEST CONTAINER IMAGE: %q\n", tc.TestImage())

	req := testcontainers.ContainerRequest{
		Image:        tc.TestImage(),
		ExposedPorts: []string{"9292/tcp"},
		Env: map[string]string{
			"TOPAZ_CERTS_DIR":     "/certs",
			"TOPAZ_DB_DIR":        "/data",
			"TOPAZ_DECISIONS_DIR": "/decisions",
		},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            assets_test.ConfigReader(),
				ContainerFilePath: "/config/config.yaml",
				FileMode:          0x700,
			},
			{
				Reader:            assets_test.AcmecorpReader(),
				ContainerFilePath: "/data/test.db",
				FileMode:          0x700,
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
			client.WithInsecure(true),
		}

		azClient, err := azc.New(opts...)
		require.NoError(t, err)
		t.Cleanup(func() { _ = azClient.Close() })

		ctx, cancel := context.WithCancel(context.Background())
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
			Query: "x := opa.runtime()",
		},
		validate: func(t *testing.T, resp *authorizer.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *rt.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}

			require.NotNil(t, result)
			require.NotEmpty(t, result.Result)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, result.Result[0].Bindings["x"], "env")
		},
	},
	{
		name: "data",
		query: &authorizer.QueryRequest{
			Query: "x = data",
		},
		validate: func(t *testing.T, resp *authorizer.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *rt.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}
			require.NotNil(t, result)
			require.NotEmpty(t, result.Result)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, result.Result[0].Bindings["x"], "rebac")
		},
	},
	{
		name: "ds.user",
		query: &authorizer.QueryRequest{
			Query: `x = ds.user({"id": "euang@acmecorp.com"})`,
		},
		validate: func(t *testing.T, resp *authorizer.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *rt.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}

			require.NotNil(t, result)
			require.NotEmpty(t, result.Result)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, result.Result[0].Bindings["x"], "id")
			binding := result.Result[0].Bindings["x"].(map[string]interface{})
			require.Equal(t, "euang@acmecorp.com", binding["id"])
		},
	},
	{
		name: "ds.identity",
		query: &authorizer.QueryRequest{
			Query: `x = ds.identity({"id": "euang@acmecorp.com"})`,
		},
		validate: func(t *testing.T, resp *authorizer.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *rt.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}

			require.NotNil(t, result)
			require.NotEmpty(t, result.Result)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, "euang@acmecorp.com", result.Result[0].Bindings["x"])
		},
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
		validate: func(t *testing.T, resp *authorizer.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *rt.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}

			require.NotNil(t, result)
			require.NotEmpty(t, result.Result)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, result.Result[0].Bindings["x"], "identity")
			require.Contains(t, result.Result[0].Bindings["x"], "user")

			bindings := result.Result[0].Bindings["x"].(map[string]interface{})
			require.Contains(t, bindings["identity"], "type")
			require.Contains(t, bindings["user"], "id")
		},
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
		validate: func(t *testing.T, resp *authorizer.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *rt.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}

			require.NotNil(t, result)
			require.NotEmpty(t, result.Result)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, result.Result[0].Bindings["x"], "identity")
			require.Contains(t, result.Result[0].Bindings["x"], "user")

			bindings := result.Result[0].Bindings["x"].(map[string]interface{})
			require.Contains(t, bindings["identity"], "identity")
			require.Contains(t, bindings["identity"], "type")
			require.Equal(t, map[string]interface{}{}, bindings["user"])
		},
	},
}
