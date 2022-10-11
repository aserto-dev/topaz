package engine_test

import (
	"context"
	"testing"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestPolicy(t *testing.T) {
	harness := atesting.SetupOnline(t, func(cfg *config.Config) {
		cfg.Directory.EdgeConfig.DBPath = atesting.AssetAcmeEBBFilePath()
		cfg.Directory.Remote.Addr = "localhost:12346"
	})
	defer harness.Cleanup()

	client := harness.CreateGRPCClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"TestListPolicies", ListPolicies(ctx, client)},
		{"TestListPoliciesMasked", ListPoliciesMasked(ctx, client)},
		{"TestListPoliciesMaskedComposed", ListPoliciesMaskedComposed(ctx, client)},
		{"TestListPoliciesInvalidMask", ListPoliciesInvalidMask(ctx, client)},
		{"TestListPoliciesEmptyMask", ListPoliciesEmptyMask(ctx, client)},
		{"TestGetPolicies", GetPolicies(ctx, client)},
		{"TestGetPoliciesMasked", GetPoliciesMasked(ctx, client)},
		{"TestGetPoliciesMaskedComposed", GetPoliciesMaskedComposed(ctx, client)},
		{"TestGetPoliciesInvalidMask", GetPoliciesInvalidMask(ctx, client)},
		{"TestGetPoliciesEmptyMask", GetPoliciesEmptyMask(ctx, client)},
		{"TestGetPoliciesInvalidID", GetPoliciesInvalidID(ctx, client)},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, testCase.test)
	}
}

func ListPolicies(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := client.ListPolicies(ctx, &authorizer.ListPoliciesRequest{})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.Greater(len(listPoliciesResponse.Result), 0)
	}
}

func ListPoliciesMasked(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := client.ListPolicies(ctx, &authorizer.ListPoliciesRequest{
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"raw",
				},
			},
		})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.Greater(len(listPoliciesResponse.Result), 0)
		assert.Nil(listPoliciesResponse.Result[0].Id)
		assert.NotNil(listPoliciesResponse.Result[0].Raw)
	}
}

func ListPoliciesMaskedComposed(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := client.ListPolicies(ctx, &authorizer.ListPoliciesRequest{
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"raw",
					"package_path",
				},
			},
		})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.Greater(len(listPoliciesResponse.Result), 0)
		assert.Nil(listPoliciesResponse.Result[0].Id)
		assert.NotNil(listPoliciesResponse.Result[0].Raw)
		assert.NotNil(listPoliciesResponse.Result[0].PackagePath)
	}
}

func ListPoliciesInvalidMask(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := client.ListPolicies(ctx, &authorizer.ListPoliciesRequest{
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"notexistantpath",
				},
			},
		})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.Greater(len(listPoliciesResponse.Result), 0)
		assert.NotNil(listPoliciesResponse.Result[0].Id)
		assert.NotNil(listPoliciesResponse.Result[0].Raw)
		assert.NotNil(listPoliciesResponse.Result[0].PackagePath)
		assert.NotNil(listPoliciesResponse.Result[0].Ast)
	}
}

func ListPoliciesEmptyMask(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := client.ListPolicies(ctx, &authorizer.ListPoliciesRequest{
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{},
			},
		})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.Greater(len(listPoliciesResponse.Result), 0)
		assert.NotNil(listPoliciesResponse.Result[0].Id)
		assert.NotNil(listPoliciesResponse.Result[0].Raw)
		assert.NotNil(listPoliciesResponse.Result[0].PackagePath)
		assert.NotNil(listPoliciesResponse.Result[0].Ast)
	}
}

func GetPolicies(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		apiModule := getOneModule(ctx, client, t)
		getPoliciesResponse, err := client.GetPolicy(ctx, &authorizer.GetPolicyRequest{
			Id: *apiModule.Id,
		})

		assert.NoError(err)
		assert.NotNil(getPoliciesResponse.Result)
	}
}

func GetPoliciesMasked(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		apiModule := getOneModule(ctx, client, t)

		getPoliciesResponse, err := client.GetPolicy(ctx, &authorizer.GetPolicyRequest{
			Id: *apiModule.Id,
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"raw",
				},
			},
		})

		assert.NoError(err)
		assert.NotNil(getPoliciesResponse.Result)
		assert.Nil(getPoliciesResponse.Result.Id)
		assert.NotNil(getPoliciesResponse.Result.Raw)
	}
}

func GetPoliciesMaskedComposed(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		apiModule := getOneModule(ctx, client, t)

		getPoliciesResponse, err := client.GetPolicy(ctx, &authorizer.GetPolicyRequest{
			Id: *apiModule.Id,
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"raw",
					"package_path",
				},
			},
		})

		assert.NoError(err)
		assert.NotNil(getPoliciesResponse.Result)
		assert.Nil(getPoliciesResponse.Result.Id)
		assert.NotNil(getPoliciesResponse.Result.Raw)
		assert.NotNil(getPoliciesResponse.Result.PackagePath)
	}
}

func GetPoliciesInvalidMask(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)

		apiModule := getOneModule(ctx, client, t)

		getPoliciesResponse, err := client.GetPolicy(ctx, &authorizer.GetPolicyRequest{
			Id: *apiModule.Id,
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"notexistantpath",
				},
			},
		})

		assert.NoError(err)
		assert.NotNil(getPoliciesResponse.Result)
		assert.NotNil(getPoliciesResponse.Result.Id)
		assert.NotNil(getPoliciesResponse.Result.Raw)
		assert.NotNil(getPoliciesResponse.Result.PackagePath)
		assert.NotNil(getPoliciesResponse.Result.Ast)
	}
}

func GetPoliciesEmptyMask(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)

		apiModule := getOneModule(ctx, client, t)

		getPoliciesResponse, err := client.GetPolicy(ctx, &authorizer.GetPolicyRequest{
			Id: *apiModule.Id,
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{},
			},
		})

		assert.NoError(err)
		assert.NotNil(getPoliciesResponse.Result)
		assert.NotNil(getPoliciesResponse.Result.Id)
		assert.NotNil(getPoliciesResponse.Result.Raw)
		assert.NotNil(getPoliciesResponse.Result.PackagePath)
		assert.NotNil(getPoliciesResponse.Result.Ast)
	}
}

func GetPoliciesInvalidID(ctx context.Context, client authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)

		_, err := client.GetPolicy(ctx, &authorizer.GetPolicyRequest{
			Id: "doesnotexist",
		})

		// TODO: replace this with aerr
		assert.Error(err)
	}
}

func getOneModule(ctx context.Context, client authorizer.AuthorizerClient, t *testing.T) *api.Module {
	assert := require.New(t)
	listPoliciesResponse, err := client.ListPolicies(ctx, &authorizer.ListPoliciesRequest{})
	if err != nil {
		assert.FailNow("failed to list policies", err.Error())
	}

	if len(listPoliciesResponse.Result) == 0 {
		assert.FailNow("no policy modules loaded")
	}
	return listPoliciesResponse.Result[0]
}
