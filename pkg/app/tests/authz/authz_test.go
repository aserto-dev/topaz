package authz_test

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	client "github.com/aserto-dev/go-aserto"
	azc "github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	assets_test "github.com/aserto-dev/topaz/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

var addr string

func TestMain(m *testing.M) {
	rc := 0
	defer func() {
		os.Exit(rc)
	}()

	ctx := context.Background()

	h, err := tc.NewHarness(ctx, &testcontainers.ContainerRequest{
		Image:        "ghcr.io/aserto-dev/topaz:test-" + tc.CommitSHA() + "-" + runtime.GOARCH,
		ExposedPorts: []string{"9292/tcp", "9393/tcp"},
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
			wait.ForLog("Starting 0.0.0.0:9393 gateway server"),
		).WithStartupTimeoutDefault(120 * time.Second).WithDeadline(360 * time.Second),
	})
	if err != nil {
		rc = 99
		return
	}

	defer func() {
		if err := h.Close(ctx); err != nil {
			rc = 100
		}
	}()

	addr = h.AddrGRPC(ctx)

	rc = m.Run()
}

func TestWithMissingIdentity(t *testing.T) {
	opts := []client.ConnectionOption{
		client.WithAddr(addr),
		client.WithInsecure(true),
	}

	azClient, err := azc.New(opts...)
	require.NoError(t, err)
	t.Cleanup(func() { _ = azClient.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"TestDecisionTreeWithMissingIdentity", DecisionTreeWithMissingIdentity(ctx, azClient)},
		{"TestDecisionTreeWithUserID", DecisionTreeWithUserID(ctx, azClient)},
		{"TestIsWithMissingIdentity", IsWithMissingIdentity(ctx, azClient)},
		{"TestQueryWithMissingIdentity", QueryWithMissingIdentity(ctx, azClient)},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, testCase.test)
	}
}

func DecisionTreeWithMissingIdentity(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := azClient.DecisionTree(ctx, &authorizer.DecisionTreeRequest{
			PolicyContext: &api.PolicyContext{
				Path:      "",
				Decisions: []string{},
			},
			IdentityContext: &api.IdentityContext{
				Identity: "noexisting-user@acmecorp.com",
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
			},
			Options:         &authorizer.DecisionTreeOptions{},
			ResourceContext: &structpb.Struct{},
		})

		if errX != nil {
			t.Logf("ERR >>> %s\n", errX)
		}

		if assert.Error(t, errX) {
			s, ok := status.FromError(errX)
			assert.Equal(t, ok, true)
			assert.Equal(t, s.Code(), codes.NotFound)
		}
		assert.Nil(t, respX, "response object should be nil")
	}
}

func DecisionTreeWithUserID(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := azClient.DecisionTree(ctx, &authorizer.DecisionTreeRequest{
			PolicyContext: &api.PolicyContext{
				Path:      "peoplefinder.GET",
				Decisions: []string{"allowed"},
			},
			IdentityContext: &api.IdentityContext{
				Identity: "CiQyYmZhYTU1Mi1kOWE1LTQxZTktYTZjMy01YmU2MmI0NDMzYzgSBWxvY2Fs", // April Stewart
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
			},
			Options:         &authorizer.DecisionTreeOptions{},
			ResourceContext: &structpb.Struct{},
		})

		if errX != nil {
			t.Logf("ERR >>> %s\n", errX)
		}

		assert.NoError(t, errX)
		assert.NotNil(t, respX, "response object should not be nil")
		assert.Equal(t, "peoplefinder.GET", respX.PathRoot)

		path := respX.Path.AsMap()
		assert.Len(t, path, 2)
	}
}

func IsWithMissingIdentity(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := azClient.Is(ctx, &authorizer.IsRequest{
			PolicyContext: &api.PolicyContext{
				Path:      "peoplefinder.POST.api.users.__id",
				Decisions: []string{"allowed"},
			},
			IdentityContext: &api.IdentityContext{
				Identity: "noexisting-user@acmecorp.com",
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
			},
			ResourceContext: &structpb.Struct{},
		})

		if errX != nil {
			t.Logf("ERR >>> %s\n", errX)
		}

		if assert.Error(t, errX) {
			s, ok := status.FromError(errX)
			assert.Equal(t, ok, true)
			assert.Equal(t, s.Code(), codes.NotFound)
		}
		assert.Nil(t, respX, "response object should be nil")
	}
}

func QueryWithMissingIdentity(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := azClient.Query(ctx, &authorizer.QueryRequest{
			IdentityContext: &api.IdentityContext{
				Identity: "noexisting-user@acmecorp.com",
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
			},
			Query: "x = true",
			Input: "",
			Options: &authorizer.QueryOptions{
				Metrics:      false,
				Instrument:   false,
				Trace:        authorizer.TraceLevel_TRACE_LEVEL_OFF,
				TraceSummary: false,
			},
			PolicyContext: &api.PolicyContext{
				Path:      "",
				Decisions: []string{},
			},
			ResourceContext: &structpb.Struct{},
		})

		if errX != nil {
			t.Logf("ERR >>> %s\n", errX)
		}

		if assert.Error(t, errX) {
			s, ok := status.FromError(errX)
			assert.Equal(t, ok, true)
			assert.Equal(t, s.Code(), codes.NotFound)
		}
		assert.Nil(t, respX, "response object should be nil")
	}
}
