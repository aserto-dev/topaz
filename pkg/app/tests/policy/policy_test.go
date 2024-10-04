package policy_test

import (
	"context"
	"os"
	"runtime"
	"testing"

	client "github.com/aserto-dev/go-aserto"
	azc "github.com/aserto-dev/go-aserto/az"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	assets_test "github.com/aserto-dev/topaz/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
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

		WaitingFor: wait.ForExposedPort(),
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

func TestPolicy(t *testing.T) {
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

func ListPolicies(ctx context.Context, azClient authorizer.AuthorizerClient) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		listPoliciesResponse, err := azClient.ListPolicies(ctx, &authorizer.ListPoliciesRequest{})

		assert.NoError(err)
		assert.NotNil(listPoliciesResponse.Result)
		assert.Greater(len(listPoliciesResponse.Result), 0)
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
		assert.Greater(len(listPoliciesResponse.Result), 0)
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
		assert.Greater(len(listPoliciesResponse.Result), 0)
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
		assert.Greater(len(listPoliciesResponse.Result), 0)
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
		assert.Greater(len(listPoliciesResponse.Result), 0)
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
