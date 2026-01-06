package builtin_test

import (
	"context"
	"testing"
	"time"

	client "github.com/aserto-dev/go-aserto"
	azc "github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/x"
	assets_test "github.com/aserto-dev/topaz/topazd/tests/assets"
	tc "github.com/aserto-dev/topaz/topazd/tests/common"

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
				FileMode:          0o700,
			},
			{
				Reader:            assets_test.AcmecorpReader(),
				ContainerFilePath: "/data/test.db",
				FileMode:          0o700,
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
			client.WithNoTLS(true),
			client.WithInsecure(false),
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

var BuiltinHelpTests = []struct {
	name     string
	query    string
	expected any
}{
	{
		name:  "ds.identity",
		query: "x = ds.identity({})",
		expected: `ds.identity({
	"id": ""
})`,
	},
	{
		name:  "ds.user",
		query: "x = ds.user({})",
		expected: `ds.user({
	"id": ""
})`,
	},
	{
		name:  "ds.check",
		query: "x = ds.check({})",
		expected: `ds.check({
	"object_type": "",
	"object_id": "",
	"relation": "",
	"subject_type": "",
	"subject_id": "",
	"trace": false
})`,
	},
	{
		name:  "ds.checks",
		query: "x = ds.checks({})",
		expected: `ds.checks({
	"default": {
		"object_id": "",
		"object_type": "",
		"relation": "",
		"subject_id": "",
		"subject_type": "",
		"trace": false
	},
	"checks": [
		{
			"object_id": "",
			"object_type": "",
			"relation": "",
			"subject_id": "",
			"subject_type": "",
			"trace": false
		}
	]
})`,
	},
	{
		name:  "ds.check_relation",
		query: "x = ds.check_relation({})",
		expected: `ds.check_relation({
	"object_id": "",
	"object_type": "",
	"relation": "",
	"subject_id": "",
	"subject_type": "",
	"trace": false
})`,
	},
	{
		name:  "ds.check_permission",
		query: "x = ds.check_permission({})",
		expected: `ds.check_permission({
	"object_id": "",
	"object_type": "",
	"permission": "",
	"subject_id": "",
	"subject_type": "",
	"trace": false
})`,
	},
	{
		name:  "ds.graph",
		query: "x = ds.graph({})",
		expected: `ds.graph({
	"object_type": "",
	"object_id": "",
	"relation": "",
	"subject_type": "",
	"subject_id": "",
	"subject_relation": "",
	"explain": false,
	"trace": false
})`,
	},
	{
		name:  "ds.object",
		query: "x = ds.object({})",
		expected: `ds.object({
	"object_type": "",
	"object_id": "",
	"with_relation": false
})`,
	},
	{
		name:  "ds.relation",
		query: "x = ds.relation({})",
		expected: `ds.relation({
	"object_id": "",
	"object_type": "",
	"relation": "",
	"subject_id": "",
	"subject_relation": "",
	"subject_type": "",
	"with_objects": false
	})`,
	},
	{
		name:  "ds.relations",
		query: "x = ds.relations({})",
		expected: `ds.relations({
	object_type: "",
	object_id: "",
	relation: "",
	subject_type: "",
	subject_id: "",
	subject_relation: "",
	with_objects: false,
	with_empty_subject_relation: false
})`,
	},
	{
		name:  "az.evaluation",
		query: "x = az.evaluation({})",
		expected: `az.evaluation({
  "subject": {"type": "", "id": "", "properties": {}}, 
  "action": {"name": "", "properties": {}}, 
  "resource": {"type": "", "id": "", "properties": {}}, 
  "context": {}
})`,
	},
	{
		name:  "az.evaluations",
		query: "x = az.evaluations({})",
		expected: `az.evaluations({
	"subject": {"type": "", "id": "", "properties": {}},
	"action": {"name": "", "properties": {}},
	"resource": {"type": "", "id": "", "properties": {}},
  "context": {},
	"options": {},
	"evaluations": [
		{
			"subject": {"type": "", "id": "", "properties": {}},
			"action": {"name": "", "properties": {}},
			"resource": {"type": "", "id": "", "properties": {}},
			"context": {}
		},
		{
			"subject": {"type": "", "id": "", "properties": {}},
			"action": {"name": "", "properties": {}},
			"resource": {"type": "", "id": "", "properties": {}},
			"context": {},
		}
	]
})`,
	},
	{
		name:  "az.action_search",
		query: "x = az.action_search({})",
		expected: `az.action_search({
	"subject": {"type": "", "id": "", "properties": {}},
	"action": {"name": "", "properties": {}},
	"resource": {"type": "", "id": "", "properties": {}},
	"context": {},
	"page": {"next_token": ""}
})`,
	},
	{
		name:  "az.resource_search",
		query: "x = az.resource_search({})",
		expected: `az.resource_search({
	"subject": {"type": "", "id": "", "properties": {}},
	"action": {"name": "", "properties": {}},
	"resource": {"type": "", "id": "", "properties": {}},
	"context": {},
	"page": {"next_token": ""}
})`,
	},
	{
		name:  "az.subject_search",
		query: "x = az.subject_search({})",
		expected: `az.subject_search({
	"subject": {"type": "", "properties": {}},
	"action": {"name": "", "properties": {}},
	"resource": {"type": "", "id": "", "properties": {}},
	"context": {},
	"page": {"next_token": ""}
})`,
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
