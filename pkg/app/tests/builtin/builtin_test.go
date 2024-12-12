package builtin_test

import (
	"context"
	"testing"
	"time"

	client "github.com/aserto-dev/go-aserto"
	azc "github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	assets_test "github.com/aserto-dev/topaz/pkg/app/tests/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestBuiltins(t *testing.T) {
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
				Reader:            assets_test.PeoplefinderConfigReader(),
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

	t.Run("testBuiltins", testBuiltins(grpcAddr))
}

func testBuiltins(addr string) func(*testing.T) {
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

		// BuiltinHelptests(ctx, client)
		for _, tc := range BuiltinHelpTests {
			f := func(t *testing.T) {
				resp, err := azClient.Query(ctx, &authorizer.QueryRequest{
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

		// BuiltinNotFoundErrTests
		for _, tc := range BuiltinNotFoundErrTests {
			f := func(t *testing.T) {
				resp, err := azClient.Query(ctx, &authorizer.QueryRequest{
					Query: tc.query,
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Response)

				r := resp.Response.AsMap()
				require.NotNil(t, r)
			}

			t.Run(tc.name, f)
		}
	}
}

var BuiltinHelpTests = []struct {
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
			},
		},
	},
	{
		name:  "ds.user",
		query: "x = ds.user({})",
		expected: map[string]interface{}{
			"ds.user": map[string]interface{}{
				"id": "",
			},
		},
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
		name:  "ds.checks",
		query: "x = ds.checks({})",
		expected: map[string]interface{}{
			"ds.checks": map[string]interface{}{
				"default": map[string]interface{}{
					"object_id":    "",
					"object_type":  "",
					"relation":     "",
					"subject_id":   "",
					"subject_type": "",
					"trace":        false,
				},
				"checks": []interface{}{
					(map[string]interface{}{
						"object_id":    "",
						"object_type":  "",
						"relation":     "",
						"subject_id":   "",
						"subject_type": "",
						"trace":        false,
					}),
				},
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
				"object_type":      "",
				"object_id":        "",
				"relation":         "",
				"subject_type":     "",
				"subject_id":       "",
				"subject_relation": "",
				"explain":          false,
				"trace":            false,
			},
		},
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
			},
		},
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
			},
		},
	},
	{
		name:  "ds.relations",
		query: "x = ds.relations({})",
		expected: map[string]interface{}{
			"ds.relations": map[string]interface{}{
				"object_id":                   "",
				"object_type":                 "",
				"page":                        nil,
				"relation":                    "",
				"subject_id":                  "",
				"subject_relation":            "",
				"subject_type":                "",
				"with_objects":                false,
				"with_empty_subject_relation": false,
			},
		},
	},
}

var BuiltinNotFoundErrTests = []struct {
	name     string
	query    string
	expected map[string]interface{}
}{
	{
		name:  "ds.identity",
		query: `x = ds.identity({"id": "no_existing_identifier"})`,
	},
	{
		name:  "ds.user",
		query: `x = ds.user({"id": "none_existing_user_object_id"})`,
	},
	{
		name:  "ds.object",
		query: `x = ds.object({"object_type": "none_existing_type", "object_id": "none_existing_id"})`,
	},
	{
		name: "ds.relation",
		query: `x = ds.relation({
			"object_type": "none_existing_object_type",
			"object_id": "none_existing_object_id",
			"relation": "none_existing_relation",
			"subject_type": "none_existing_subject_type",
			"subject_id": "none_existing_subject_id",
			})`,
	},
	{
		name: "ds.relation.with.subject_relation",
		query: `x = ds.relation({
			"object_type": "none_existing_object_type",
			"object_id": "none_existing_object_id",
			"relation": "none_existing_relation",
			"subject_type": "none_existing_subject_type",
			"subject_id": "none_existing_subject_id",
			"subject_relation": "none_existing_subject_relation",
			})`,
	},
}
