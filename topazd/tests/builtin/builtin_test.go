package builtin_test

import (
	"context"
	"testing"

	"github.com/aserto-dev/topaz/api/authorizer/v2"
	"github.com/aserto-dev/topaz/api/authorizer/v2/api"
	client "github.com/aserto-dev/topaz/api/client"
	azc "github.com/aserto-dev/topaz/api/client/az"
	"github.com/aserto-dev/topaz/internal/fs"
	"github.com/aserto-dev/topaz/topaz/x"
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
		ExposedPorts: tc.TestExposedPorts,
		Env: map[string]string{
			x.EnvTopazCertsDir:     x.DefCertsDir,
			x.EnvTopazDBDir:        x.DefDBDir,
			x.EnvTopazDecisionsDir: x.DefDecisionsDir,
		},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            assets_test.PeoplefinderConfigReader(),
				ContainerFilePath: tc.TestConfigFilePath,
				FileMode:          int64(fs.FileModeOwnerRWX),
			},
			{
				Reader:            assets_test.AcmecorpReader(),
				ContainerFilePath: tc.TestDBFilePath,
				FileMode:          int64(fs.FileModeOwnerRWX),
			},
		},
		WaitingFor: wait.ForAll(
			wait.ForExposedPort(),
			wait.ForLog("Starting 0.0.0.0:9292 gRPC server"),
		).WithStartupTimeoutDefault(tc.TestStartupTimeout),
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

	t.Run("testBuiltins", testBuiltins(ctx, grpcAddr))
}

func testBuiltins(ctx context.Context, addr string) func(*testing.T) {
	return func(t *testing.T) {
		opts := []client.ConnectionOption{
			client.WithAddr(addr),
			client.WithInsecure(true),
		}

		azClient, err := azc.New(opts...)
		require.NoError(t, err)
		t.Cleanup(func() { _ = azClient.Close() })

		ctx, cancel := context.WithCancel(ctx)
		t.Cleanup(cancel)

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
		expected: map[string]any{"id": ""},
	},
	{
		name:     "ds.user",
		query:    "x = ds.user({})",
		expected: map[string]any{"id": ""},
	},
	{
		name:     "ds.check",
		query:    "x = ds.check({})",
		expected: map[string]any{"object_id": "", "object_type": "", "relation": "", "subject_id": "", "subject_type": "", "trace": false},
	},
	{
		name:     "ds.checks",
		query:    "x = ds.checks({})",
		expected: map[string]any{"checks": []any{map[string]any{"object_id": "", "object_type": "", "relation": "", "subject_id": "", "subject_type": "", "trace": false}}, "default": map[string]any{"object_id": "", "object_type": "", "relation": "", "subject_id": "", "subject_type": "", "trace": false}},
	},
	{
		name:     "ds.graph",
		query:    "x = ds.graph({})",
		expected: map[string]any{"explain": false, "object_id": "", "object_type": "", "relation": "", "subject_id": "", "subject_relation": "", "subject_type": "", "trace": false},
	},
	{
		name:     "ds.object",
		query:    "x = ds.object({})",
		expected: map[string]any{"object_id": "", "object_type": "", "with_relations": false},
	},
	{
		name:     "ds.relation",
		query:    "x = ds.relation({})",
		expected: map[string]any{"object_id": "", "object_type": "", "relation": "", "subject_id": "", "subject_relation": "", "subject_type": "", "with_objects": false},
	},
	{
		name:     "ds.relations",
		query:    "x = ds.relations({})",
		expected: map[string]any{"object_id": "", "object_type": "", "relation": "", "subject_id": "", "subject_relation": "", "subject_type": "", "with_empty_subject_relation": false, "with_objects": false},
	},
	{
		name:     "az.evaluation",
		query:    "x = az.evaluation({})",
		expected: map[string]any{"action": map[string]any{"name": "", "properties": map[string]any{}}, "context": map[string]any{}, "resource": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}, "subject": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}},
	},
	{
		name:     "az.evaluations",
		query:    "x = az.evaluations({})",
		expected: map[string]any{"action": map[string]any{"name": "", "properties": map[string]any{}}, "context": map[string]any{}, "evaluations": []any{map[string]any{"action": map[string]any{"name": "", "properties": map[string]any{}}, "context": map[string]any{}, "resource": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}, "subject": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}}}, "options": map[string]any{}, "resource": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}, "subject": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}},
	},
	{
		name:     "az.action_search",
		query:    "x = az.action_search({})",
		expected: map[string]any{"action": map[string]any{"name": "", "properties": map[string]any{}}, "context": map[string]any{}, "page": map[string]any{"limit": float64(100), "token": ""}, "resource": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}, "subject": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}},
	},
	{
		name:     "az.resource_search",
		query:    "x = az.resource_search({})",
		expected: map[string]any{"action": map[string]any{"name": "", "properties": map[string]any{}}, "context": map[string]any{}, "page": map[string]any{"limit": float64(100), "token": ""}, "resource": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}, "subject": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}},
	},
	{
		name:     "az.subject_search",
		query:    "x = az.subject_search({})",
		expected: map[string]any{"action": map[string]any{"name": "", "properties": map[string]any{}}, "context": map[string]any{}, "page": map[string]any{"limit": float64(100), "token": ""}, "resource": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}, "subject": map[string]any{"id": "", "properties": map[string]any{}, "type": ""}},
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
