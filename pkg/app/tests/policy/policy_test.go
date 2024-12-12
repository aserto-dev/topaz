package policy_test

import (
	"context"
	"testing"
	"time"

	client "github.com/aserto-dev/go-aserto"
	azc "github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	assets_test "github.com/aserto-dev/topaz/pkg/app/tests/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestPolicy(t *testing.T) {
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

	t.Run("testPolicy", testPolicy(grpcAddr))
}

func testPolicy(addr string) func(*testing.T) {
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

		tests := []struct {
			name string
			test func(*testing.T)
		}{
			{"TestListPolicies", ListPolicies(ctx, azClient)},
			{"TestListPoliciesMasked", ListPoliciesMasked(ctx, azClient)},
			{"TestListPoliciesMaskedComposed", ListPoliciesMaskedComposed(ctx, azClient)},
			{"TestListPoliciesInvalidMask", ListPoliciesInvalidMask(ctx, azClient)},
			{"TestListPoliciesEmptyMask", ListPoliciesEmptyMask(ctx, azClient)},
			{"TestGetPolicies", GetPolicies(ctx, azClient)},
			{"TestGetPoliciesMasked", GetPoliciesMasked(ctx, azClient)},
			{"TestGetPoliciesMaskedComposed", GetPoliciesMaskedComposed(ctx, azClient)},
			{"TestGetPoliciesInvalidMask", GetPoliciesInvalidMask(ctx, azClient)},
			{"TestGetPoliciesEmptyMask", GetPoliciesEmptyMask(ctx, azClient)},
			{"TestGetPoliciesInvalidID", GetPoliciesInvalidID(ctx, azClient)},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, testCase.test)
		}
	}
}

func ListPolicies(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := azClient.ListPolicies(ctx, &authorizer.ListPoliciesRequest{})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.NotEmpty(listPoliciesResponse.Result)
	}
}

func ListPoliciesMasked(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := azClient.ListPolicies(ctx, &authorizer.ListPoliciesRequest{
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"raw",
				},
			},
		})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.NotEmpty(listPoliciesResponse.Result)
		assert.Nil(listPoliciesResponse.Result[0].Id)
		assert.NotNil(listPoliciesResponse.Result[0].Raw)
	}
}

func ListPoliciesMaskedComposed(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := azClient.ListPolicies(ctx, &authorizer.ListPoliciesRequest{
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"raw",
					"package_path",
				},
			},
		})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.NotEmpty(listPoliciesResponse.Result)
		assert.Nil(listPoliciesResponse.Result[0].Id)
		assert.NotNil(listPoliciesResponse.Result[0].Raw)
		assert.NotNil(listPoliciesResponse.Result[0].PackagePath)
	}
}

func ListPoliciesInvalidMask(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := azClient.ListPolicies(ctx, &authorizer.ListPoliciesRequest{
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"notexistantpath",
				},
			},
		})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.NotEmpty(listPoliciesResponse.Result)
		assert.NotNil(listPoliciesResponse.Result[0].Id)
		assert.NotNil(listPoliciesResponse.Result[0].Raw)
		assert.NotNil(listPoliciesResponse.Result[0].PackagePath)
		assert.NotNil(listPoliciesResponse.Result[0].Ast)
	}
}

func ListPoliciesEmptyMask(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := azClient.ListPolicies(ctx, &authorizer.ListPoliciesRequest{
			FieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{},
			},
		})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.NotEmpty(listPoliciesResponse.Result)
		assert.NotNil(listPoliciesResponse.Result[0].Id)
		assert.NotNil(listPoliciesResponse.Result[0].Raw)
		assert.NotNil(listPoliciesResponse.Result[0].PackagePath)
		assert.NotNil(listPoliciesResponse.Result[0].Ast)
	}
}

func GetPolicies(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		apiModule := getOneModule(ctx, azClient, t)
		getPoliciesResponse, err := azClient.GetPolicy(ctx, &authorizer.GetPolicyRequest{
			Id: *apiModule.Id,
		})

		assert.NoError(err)
		assert.NotNil(getPoliciesResponse.Result)
	}
}

func GetPoliciesMasked(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		apiModule := getOneModule(ctx, azClient, t)

		getPoliciesResponse, err := azClient.GetPolicy(ctx, &authorizer.GetPolicyRequest{
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

func GetPoliciesMaskedComposed(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		apiModule := getOneModule(ctx, azClient, t)

		getPoliciesResponse, err := azClient.GetPolicy(ctx, &authorizer.GetPolicyRequest{
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

func GetPoliciesInvalidMask(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)

		apiModule := getOneModule(ctx, azClient, t)

		getPoliciesResponse, err := azClient.GetPolicy(ctx, &authorizer.GetPolicyRequest{
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

func GetPoliciesEmptyMask(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)

		apiModule := getOneModule(ctx, azClient, t)

		getPoliciesResponse, err := azClient.GetPolicy(ctx, &authorizer.GetPolicyRequest{
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

func GetPoliciesInvalidID(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)

		_, err := azClient.GetPolicy(ctx, &authorizer.GetPolicyRequest{
			Id: "doesnotexist",
		})

		// TODO: replace this with aerr
		assert.Error(err)
	}
}

func getOneModule(ctx context.Context, azClient authorizer.AuthorizerClient, t *testing.T) *api.Module {
	assert := require.New(t)
	listPoliciesResponse, err := azClient.ListPolicies(ctx, &authorizer.ListPoliciesRequest{})
	if err != nil {
		assert.FailNow("failed to list policies", err.Error())
	}

	if len(listPoliciesResponse.Result) == 0 {
		assert.FailNow("no policy modules loaded")
	}
	return listPoliciesResponse.Result[0]
}
