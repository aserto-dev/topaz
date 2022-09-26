package engine_test

import (
	"context"
	"testing"

	api_v2 "github.com/aserto-dev/go-authorizer/aserto/api/v2"
	authz2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-eds/pkg/pb"
	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	policy "github.com/aserto-dev/go-grpc/aserto/authorizer/policy/v1"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWithMissingIdentity(t *testing.T) {
	harness := atesting.SetupOnline(t, func(cfg *config.Config) {
		cfg.Directory.Path = atesting.AssetAcmeEBBFilePath()
	})
	defer harness.Cleanup()

	policyID, err := getPolicyID(harness, "peoplefinder")
	require.NoError(t, err, "getPolicyID")

	client := harness.CreateGRPCClientV2()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"TestDecisionTreeWithMissingIdentity", DecisionTreeWithMissingIdentity(ctx, client, policyID)},
		{"TestIsWithMissingIdentity", IsWithMissingIdentity(ctx, client, policyID)},
		{"TestQueryWithMissingIdentity", QueryWithMissingIdentity(ctx, client, policyID)},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, testCase.test)
	}
}

func DecisionTreeWithMissingIdentity(ctx context.Context, client authz2.AuthorizerClient, policyID string) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := client.DecisionTree(ctx, &authz2.DecisionTreeRequest{
			PolicyContext: &api_v2.PolicyContext{
				Path:      "",
				Decisions: []string{},
			},
			IdentityContext: &api.IdentityContext{
				Identity: "noexisting-user@acmecorp.com",
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
			},
			Options:         &authz2.DecisionTreeOptions{},
			ResourceContext: pb.NewStruct(),
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

func IsWithMissingIdentity(ctx context.Context, client authz2.AuthorizerClient, policyID string) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := client.Is(ctx, &authz2.IsRequest{
			PolicyContext: &api_v2.PolicyContext{
				Path:      "peoplefinder.POST.api.users.__id",
				Decisions: []string{"allowed"},
			},
			IdentityContext: &api.IdentityContext{
				Identity: "noexisting-user@acmecorp.com",
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
			},
			ResourceContext: pb.NewStruct(),
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

func QueryWithMissingIdentity(ctx context.Context, client authz2.AuthorizerClient, policyID string) func(*testing.T) {
	return func(t *testing.T) {
		respX, errX := client.Query(ctx, &authz2.QueryRequest{
			IdentityContext: &api.IdentityContext{
				Identity: "noexisting-user@acmecorp.com",
				Type:     api.IdentityType_IDENTITY_TYPE_SUB,
			},
			Query: "x = true",
			Input: "",
			Options: &authz2.QueryOptions{
				Metrics:      false,
				Instrument:   false,
				Trace:        authz2.TraceLevel_TRACE_LEVEL_OFF,
				TraceSummary: false,
			},
			PolicyContext: &api_v2.PolicyContext{
				Path:      "",
				Decisions: []string{},
			},
			ResourceContext: pb.NewStruct(),
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

func getPolicyID(harness *atesting.EngineHarness, name string) (string, error) {
	client := harness.CreateGRPCClient().Policy

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := client.ListPolicies(ctx, &policy.ListPoliciesRequest{})
	if err != nil {
		return "", errors.Wrap(err, "ListPolicies")
	}

	for _, v := range resp.Results {
		if v.Name == name {
			return v.Id, nil
		}
	}

	return "", errors.Errorf("policy name not found [%s]", name)
}
