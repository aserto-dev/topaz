package builtin_test

import (
	"context"
	"testing"
	"time"

	client "github.com/aserto-dev/go-aserto"
	azc "github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	assets_test "github.com/aserto-dev/topaz/pkg/app/tests/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"
	"github.com/aserto-dev/topaz/pkg/cli/x"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestBuiltins(t *testing.T) {
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

		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		// BuiltinHelptests(ctx, client)
		for _, tc := range BuiltinHelpTests {
			f := func(t *testing.T) {
				resp, err := azClient.Query(ctx, &authorizer.QueryRequest{
					Query:           tc.query,
					IdentityContext: &api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_NONE},
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.GetResponse())

				r := resp.GetResponse().AsMap()

				v1, ok := r["result"].([]any)
				require.True(t, ok)
				v2, ok := v1[0].(map[string]any)
				require.True(t, ok)
				v3, ok := v2["bindings"].(map[string]any)
				require.True(t, ok)

				v := v3["x"]

				assert.Equal(t, tc.expected, v)
			}

			t.Run(tc.name, f)
		}

		// BuiltinNotFoundErrTests
		for _, tc := range BuiltinNotFoundErrTests {
			f := func(t *testing.T) {
				resp, err := azClient.Query(ctx, &authorizer.QueryRequest{
					Query:           tc.query,
					IdentityContext: &api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_NONE},
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.GetResponse())

				r := resp.GetResponse().AsMap()
				require.NotNil(t, r)
			}

			t.Run(tc.name, f)
		}
	}
}

//nolint:lll
var BuiltinHelpTests = []struct {
	name     string
	query    string
	expected any
}{
	{
		name:     "ds.identity",
		query:    "x = ds.identity({})",
		expected: "ds.identity({\n\t\"id\": \"\"\n})",
	},
	{
		name:     "ds.user",
		query:    "x = ds.user({})",
		expected: "ds.user({\n\t\"id\": \"\"\n})",
	},
	{
		name:     "ds.check",
		query:    "x = ds.check({})",
		expected: "ds.check({\n\t\"object_type\": \"\",\n\t\"object_id\": \"\",\n\t\"relation\": \"\",\n\t\"subject_type\": \"\"\n\t\"subject_id\": \"\",\n\t\"trace\": false\n})",
	},
	{
		name:     "ds.checks",
		query:    "x = ds.checks({})",
		expected: "ds.checks({\n\t\"default\": {\n\t\t\"object_id\": \"\",\n\t\t\"object_type\": \"\",\n\t\t\"relation\": \"\",\n\t\t\"subject_id\": \"\",\n\t\t\"subject_type\": \"\",\n\t\t\"trace\": false\n\t},\n\t\"checks\": [\n\t\t{\n\t\t\t\"object_id\": \"\",\n\t\t\t\"object_type\": \"\",\n\t\t\t\"relation\": \"\",\n\t\t\t\"subject_id\": \"\",\n\t\t\t\"subject_type\": \"\",\n\t\t\t\"trace\": false\n\t\t}\n\t]\n})",
	},
	{
		name:     "ds.check_relation",
		query:    "x = ds.check_relation({})",
		expected: "ds.check_relation({\n\t\"object_id\": \"\",\n\t\"object_type\": \"\",\n\t\"relation\": \"\",\n\t\"subject_id\": \"\",\n\t\"subject_type\": \"\",\n\t\"trace\": false\n})",
	},
	{
		name:     "ds.check_permission",
		query:    "x = ds.check_permission({})",
		expected: "ds.check_permission({\n\t\"object_id\": \"\",\n\t\"object_type\": \"\",\n\t\"permission\": \"\",\n\t\"subject_id\": \"\",\n\t\"subject_type\": \"\",\n\t\"trace\": false\n})",
	},
	{
		name:     "ds.graph",
		query:    "x = ds.graph({})",
		expected: "ds.graph({\n\t\"object_type\": \"\",\n\t\"object_id\": \"\",\n\t\"relation\": \"\",\n\t\"subject_type\": \"\",\n\t\"subject_id\": \"\",\n\t\"subject_relation\": \"\",\n\t\"explain\": false,\n\t\"trace\": false\n})",
	},
	{
		name:     "ds.object",
		query:    "x = ds.object({})",
		expected: "ds.object({\n\t\"object_type\": \"\",\n\t\"object_id\": \"\",\n\t\"with_relation\": false\n})",
	},
	{
		name:     "ds.relation",
		query:    "x = ds.relation({})",
		expected: "ds.relation({\n\t\"object_id\": \"\",\n\t\"object_type\": \"\",\n\t\"relation\": \"\",\n\t\"subject_id\": \"\",\n\t\"subject_relation\": \"\",\n\t\"subject_type\": \"\",\n\t\"with_objects\": false\n\t})",
	},
	{
		name:     "ds.relations",
		query:    "x = ds.relations({})",
		expected: "ds.relations({\n\tobject_type: \"\",\n\tobject_id: \"\",\n\trelation: \"\",\n\tsubject_type: \"\",\n\tsubject_id: \"\",\n\tsubject_relation: \"\",\n\twith_objects: false,\n\twith_empty_subject_relation: false\n})",
	},
	{
		name:     "az.evaluation",
		query:    "x = az.evaluation({})",
		expected: "az.evaluation({\n  \"subject\": {\"type\": \"\", \"id\": \"\", \"properties\": {}}, \n  \"action\": {\"name\": \"\", \"properties\": {}}, \n  \"resource\": {\"type\": \"\", \"id\": \"\", \"properties\": {}}, \n  \"context\": {}\n})",
	},
	{
		name:     "az.evaluations",
		query:    "x = az.evaluations({})",
		expected: "az.evaluations({\n\t\"subject\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n\t\"action\": {\"name\": \"\", \"properties\": {}},\n\t\"resource\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n  \"context\": {},\n\t\"options\": {},\n\t\"evaluations\": [\n\t\t{\n\t\t\t\"subject\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n\t\t\t\"action\": {\"name\": \"\", \"properties\": {}},\n\t\t\t\"resource\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n\t\t\t\"context\": {}\n\t\t},\n\t\t{\n\t\t\t\"subject\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n\t\t\t\"action\": {\"name\": \"\", \"properties\": {}},\n\t\t\t\"resource\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n\t\t\t\"context\": {},\n\t\t}\n\t]\n})",
	},
	{
		name:     "az.action_search",
		query:    "x = az.action_search({})",
		expected: "az.action_search({\n\t\"subject\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n\t\"action\": {\"name\": \"\", \"properties\": {}},\n\t\"resource\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n\t\"context\": {},\n\t\"page\": {\"next_token\": \"\"}\n})",
	},
	{
		name:     "az.resource_search",
		query:    "x = az.resource_search({})",
		expected: "az.resource_search({\n\t\"subject\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n\t\"action\": {\"name\": \"\", \"properties\": {}},\n\t\"resource\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n\t\"context\": {},\n\t\"page\": {\"next_token\": \"\"}\n})",
	},
	{
		name:     "az.subject_search",
		query:    "x = az.subject_search({})",
		expected: "az.subject_search({\n\t\"subject\": {\"type\": \"\", \"properties\": {}},\n\t\"action\": {\"name\": \"\", \"properties\": {}},\n\t\"resource\": {\"type\": \"\", \"id\": \"\", \"properties\": {}},\n\t\"context\": {},\n\t\"page\": {\"next_token\": \"\"}\n})\n",
	},
}

var BuiltinNotFoundErrTests = []struct {
	name     string
	query    string
	expected map[string]any
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
