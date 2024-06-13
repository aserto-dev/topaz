package query_test

import (
	"context"
	"encoding/json"
	"testing"

	authz2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"
	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	harness := atesting.SetupOffline(t, func(cfg *config.Config) {
		cfg.Edge.DBPath = atesting.AssetAcmeDBFilePath()
		cfg.OPA.LocalBundles.Paths = []string{atesting.AssetLocalBundle()}
	})
	t.Cleanup(harness.Cleanup)

	client := harness.CreateGRPCClient()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	for _, tc := range queryTests {
		f := func(t *testing.T) {
			resp, err := client.Query(ctx, tc.query)
			tc.validate(t, resp, err)
		}

		t.Run(tc.name, f)
	}
}

var queryTests = []struct {
	name     string
	query    *authz2.QueryRequest
	validate func(*testing.T, *authz2.QueryResponse, error)
}{
	{
		name: "opa.runtime",
		query: &authz2.QueryRequest{
			Query: "x := opa.runtime()",
		},
		validate: func(t *testing.T, resp *authz2.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *runtime.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}

			require.NotNil(t, result)
			require.Greater(t, len(result.Result), 0)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, result.Result[0].Bindings["x"], "env")
		},
	},
	{
		name: "data",
		query: &authz2.QueryRequest{
			Query: "x = data",
		},
		validate: func(t *testing.T, resp *authz2.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *runtime.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}
			require.NotNil(t, result)
			require.Greater(t, len(result.Result), 0)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, result.Result[0].Bindings["x"], "mycars")
		},
	},
	{
		name: "ds.userV3",
		query: &authz2.QueryRequest{
			Query: `x = ds.user({"id": "CiQ3Y2VlOGU4NS1lM2NmLTRiNmYtODRlYy1mYWM4OTEwN2U5NTcSBWxvY2Fs"})`,
		},
		validate: func(t *testing.T, resp *authz2.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *runtime.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}

			require.NotNil(t, result)
			require.Greater(t, len(result.Result), 0)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, result.Result[0].Bindings["x"], "id")
			binding := result.Result[0].Bindings["x"].(map[string]interface{})
			require.Equal(t, binding["id"], "CiQ3Y2VlOGU4NS1lM2NmLTRiNmYtODRlYy1mYWM4OTEwN2U5NTcSBWxvY2Fs")
		},
	},
	{
		name: "ds.identityV3",
		query: &authz2.QueryRequest{
			Query: `x = ds.identity({"id": "euang@acmecorp.com"})`,
		},
		validate: func(t *testing.T, resp *authz2.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *runtime.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}

			require.NotNil(t, result)
			require.Greater(t, len(result.Result), 0)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, result.Result[0].Bindings["x"], "CiRkZmRhZGMzOS03MzM1LTQwNGQtYWY2Ni1jNzdjZjEzYTE1ZjgSBWxvY2Fs")
		},
	},
	{
		name: "identity_context_sub",
		query: &authz2.QueryRequest{
			Query: "x = input",
			IdentityContext: &api.IdentityContext{
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
				Identity: "euang@acmecorp.com",
			},
		},
		validate: func(t *testing.T, resp *authz2.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *runtime.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}

			require.NotNil(t, result)
			require.Greater(t, len(result.Result), 0)
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
		query: &authz2.QueryRequest{
			Query: "x = input",
			IdentityContext: &api.IdentityContext{
				Type:     api.IdentityType_IDENTITY_TYPE_MANUAL,
				Identity: "euang@acmecorp.com",
			},
		},
		validate: func(t *testing.T, resp *authz2.QueryResponse, err error) {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			var result *runtime.Result
			buf, err := resp.Response.MarshalJSON()
			require.NoError(t, err)

			if err := json.Unmarshal(buf, &result); err != nil {
				require.NoError(t, err)
			}

			require.NotNil(t, result)
			require.Greater(t, len(result.Result), 0)
			require.Contains(t, result.Result[0].Bindings, "x")
			require.Contains(t, result.Result[0].Bindings["x"], "identity")
			require.Contains(t, result.Result[0].Bindings["x"], "user")

			bindings := result.Result[0].Bindings["x"].(map[string]interface{})
			require.Contains(t, bindings["identity"], "identity")
			require.Contains(t, bindings["identity"], "type")
			require.Equal(t, bindings["user"], map[string]interface{}{})
		},
	},
}
