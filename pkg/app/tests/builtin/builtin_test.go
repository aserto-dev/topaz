package builtin_test

import (
	"context"
	"testing"

	authz2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var builtinHelptests = []struct {
	name     string
	query    string
	expected map[string]interface{}
}{
	{
		name:  "ds.identity",
		query: "x = ds.identity({})",
		expected: map[string]interface{}{
			"ds.identity": map[string]interface{}{
				"id": "",
			}},
	},
	{
		name:  "ds.user",
		query: "x = ds.user({})",
		expected: map[string]interface{}{
			"ds.user": map[string]interface{}{
				"id": "",
			}},
	},
	{
		name:  "ds.check",
		query: "x = ds.check({})",
		expected: map[string]interface{}{
			"ds.check": map[string]interface{}{
				"object_type":  "",
				"object_id":    "",
				"relation":     "",
				"subject_type": "",
				"subject_id":   "",
				"trace":        false,
			},
		},
	},
	{
		name:  "ds.check_relation",
		query: "x = ds.check_relation({})",
		expected: map[string]interface{}{
			"ds.check_relation": map[string]interface{}{
				"object_type":  "",
				"object_id":    "",
				"relation":     "",
				"subject_type": "",
				"subject_id":   "",
				"trace":        false,
			},
		},
	},
	{
		name:  "ds.check_permission",
		query: "x = ds.check_permission({})",
		expected: map[string]interface{}{
			"ds.check_permission": map[string]interface{}{
				"object_type":  "",
				"object_id":    "",
				"permission":   "",
				"subject_type": "",
				"subject_id":   "",
				"trace":        false,
			},
		},
	},
	{
		name:  "ds.graph",
		query: "x = ds.graph({})",
		expected: map[string]interface{}{
			"ds.graph": map[string]interface{}{
				"anchor_type":      "",
				"anchor_id":        "",
				"object_type":      "",
				"object_id":        "",
				"relation":         "",
				"subject_type":     "",
				"subject_id":       "",
				"subject_relation": "",
			}},
	},
	{
		name:  "ds.object",
		query: "x = ds.object({})",
		expected: map[string]interface{}{
			"ds.object": map[string]interface{}{
				"object_type":    "",
				"object_id":      "",
				"page":           nil,
				"with_relations": false,
			}},
	},
	{
		name:  "ds.relation",
		query: "x = ds.relation({})",
		expected: map[string]interface{}{
			"ds.relation": map[string]interface{}{
				"object_id":        "",
				"object_type":      "",
				"relation":         "",
				"subject_id":       "",
				"subject_relation": "",
				"subject_type":     "",
				"with_objects":     false,
			}},
	},
	{
		name:  "ds.relations",
		query: "x = ds.relations({})",
		expected: map[string]interface{}{
			"ds.relations": map[string]interface{}{
				"object_id":        "",
				"object_type":      "",
				"page":             nil,
				"relation":         "",
				"subject_id":       "",
				"subject_relation": "",
				"subject_type":     "",
				"with_objects":     false,
			}},
	},
}

func TestBuiltinsHelp(t *testing.T) {
	harness := atesting.SetupOnline(t, func(cfg *config.Config) {
		cfg.Edge.DBPath = atesting.AssetAcmeDBFilePath()
	})
	defer harness.Cleanup()

	client := harness.CreateGRPCClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, tc := range builtinHelptests {
		f := func(t *testing.T) {
			resp, err := client.Query(ctx, &authz2.QueryRequest{
				Query: tc.query,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Response)

			r := resp.Response.AsMap()

			v1 := r["result"].([]interface{})
			v2 := v1[0].(map[string]interface{})
			v3 := v2["bindings"].(map[string]interface{})
			v := v3["x"]

			assert.Equal(t, v, tc.expected)
		}

		t.Run(tc.name, f)
	}
}
