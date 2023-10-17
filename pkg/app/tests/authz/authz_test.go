package authz_test

import (
	"context"
	"testing"

	authz2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	authz_api_v2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestWithMissingIdentity(t *testing.T) {
	harness := atesting.SetupOnline(t, func(cfg *config.Config) {
		cfg.Edge.DBPath = atesting.AssetAcmeDBFilePath()
	})
	defer harness.Cleanup()

	client := harness.CreateGRPCClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"TestDecisionTreeWithMissingIdentity", DecisionTreeWithMissingIdentity(ctx, client)},
		{"TestDecisionTreeWithUserID", DecisionTreeWithUserID(ctx, client)},
		{"TestIsWithMissingIdentity", IsWithMissingIdentity(ctx, client)},
		{"TestQueryWithMissingIdentity", QueryWithMissingIdentity(ctx, client)},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, testCase.test)
	}
}

func DecisionTreeWithMissingIdentity(ctx context.Context, client authz2.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := client.DecisionTree(ctx, &authz2.DecisionTreeRequest{
			PolicyContext: &authz_api_v2.PolicyContext{
				Path:      "",
				Decisions: []string{},
			},
			IdentityContext: &authz_api_v2.IdentityContext{
				Identity: "noexisting-user@acmecorp.com",
				Type:     authz_api_v2.IdentityType_IDENTITY_TYPE_SUB,
			},
			Options:         &authz2.DecisionTreeOptions{},
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

func DecisionTreeWithUserID(ctx context.Context, client authz2.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := client.DecisionTree(ctx, &authz2.DecisionTreeRequest{
			PolicyContext: &authz_api_v2.PolicyContext{
				Path:      "peoplefinder.GET",
				Decisions: []string{"allowed"},
			},
			IdentityContext: &authz_api_v2.IdentityContext{
				Identity: "CiQyYmZhYTU1Mi1kOWE1LTQxZTktYTZjMy01YmU2MmI0NDMzYzgSBWxvY2Fs", // April Stewart
				Type:     authz_api_v2.IdentityType_IDENTITY_TYPE_SUB,
			},
			Options:         &authz2.DecisionTreeOptions{},
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

func IsWithMissingIdentity(ctx context.Context, client authz2.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := client.Is(ctx, &authz2.IsRequest{
			PolicyContext: &authz_api_v2.PolicyContext{
				Path:      "peoplefinder.POST.api.users.__id",
				Decisions: []string{"allowed"},
			},
			IdentityContext: &authz_api_v2.IdentityContext{
				Identity: "noexisting-user@acmecorp.com",
				Type:     authz_api_v2.IdentityType_IDENTITY_TYPE_SUB,
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

func QueryWithMissingIdentity(ctx context.Context, client authz2.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := client.Query(ctx, &authz2.QueryRequest{
			IdentityContext: &authz_api_v2.IdentityContext{
				Identity: "noexisting-user@acmecorp.com",
				Type:     authz_api_v2.IdentityType_IDENTITY_TYPE_SUB,
			},
			Query: "x = true",
			Input: "",
			Options: &authz2.QueryOptions{
				Metrics:      false,
				Instrument:   false,
				Trace:        authz2.TraceLevel_TRACE_LEVEL_OFF,
				TraceSummary: false,
			},
			PolicyContext: &authz_api_v2.PolicyContext{
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
