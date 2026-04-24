package tests_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	dsc3 "github.com/aserto-dev/topaz/api/directory/v4"
	dsr3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/api/directory/v4/writer"
	dsw3 "github.com/aserto-dev/topaz/api/directory/v4/writer"
	"github.com/aserto-dev/topaz/internal/eds/pkg/server"
	"github.com/aserto-dev/topaz/internal/fs"

	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/stretchr/testify/require"
)

// const blockSize = 1024 // test with 1KiB block size to exercise chunking.

// func TestManifestV2(t *testing.T) {
// 	client, closer := testInit()
// 	t.Cleanup(closer)

// 	manifest, err := os.ReadFile("./manifest_v2_test.yaml")
// 	require.NoError(t, err)

// 	t.Run("set-manifest", testSetManifest(client, manifest))
// 	t.Run("get-manifest", testGetManifest(client, "./manifest_v2_test.yaml"))
// 	t.Run("delete-manifest", testDeleteManifest(client))
// }

func TestManifestV3(t *testing.T) {
	client, closer := testInit()
	t.Cleanup(closer)

	manifest, err := os.ReadFile("./manifest_v3_test.yaml")
	require.NoError(t, err)

	t.Run("set-manifest", testSetManifest(client, manifest))
	t.Run("get-manifest", testGetManifest(client, "./manifest_v3_test.yaml"))
	t.Run("get-model", testGetModel(client))
	t.Run("delete-manifest", testDeleteManifest(client))
}

func TestManifestDiff(t *testing.T) {
	client, closer := testInit()
	t.Cleanup(closer)

	m1, err := os.ReadFile("./manifest_v3_test.yaml")
	require.NoError(t, err)

	require.NoError(t, setManifest(client, m1))
	require.NoError(t, loadData(client, "./diff_test.json"))

	tests := []struct {
		name     string
		manifest string
		check    func(*require.Assertions, error)
	}{
		{
			"delete object in use", removeObjectInUse, func(assert *require.Assertions, err error) {
				assert.Error(err)
				assert.ErrorContains(err, "object type in use: user")
			},
		},
		{
			"delete relation in use", removeRelationInUse, func(assert *require.Assertions, err error) {
				assert.Error(err)
				assert.ErrorContains(err, "relation type in use: user#manager")
			},
		},
		{
			"delete direct assignment in use", removeDirectAssignemntInUse, func(assert *require.Assertions, err error) {
				assert.Error(err)
				assert.ErrorContains(err, "relation type in use: user#manager@user")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			assert := require.New(tt)
			err := setManifest(client, []byte(test.manifest))
			test.check(assert, err)
		})
	}
}

func testSetManifest(client *server.TestEdgeClient, manifest []byte) func(*testing.T) {
	return func(t *testing.T) {
		require.NoError(t, setManifest(client, manifest))
	}
}

func setManifest(client *server.TestEdgeClient, manifest []byte) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := client.V3.Writer.SetManifest(ctx, &writer.SetManifestRequest{
		Manifest: &dsc3.Manifest{
			Id:      "",
			Content: manifest,
		},
	})

	_ = resp

	return err
}

func getManifest(client *server.TestEdgeClient) ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := client.V3.Reader.GetManifest(ctx, &dsr3.GetManifestRequest{
		Id: "",
	})

	return resp.Manifest.Content, err
}

func testGetManifest(client *server.TestEdgeClient, manifest string) func(*testing.T) {
	return func(t *testing.T) {
		data, err := getManifest(client)
		require.NoError(t, err)

		tempManifest := path.Join(os.TempDir(), "manifest.yaml")
		if err := os.WriteFile(tempManifest, data, fs.FileModeOwnerRW); err != nil {
			require.NoError(t, err)
		}

		input1, err := ytbx.LoadFile(manifest)
		require.NoError(t, err)

		input2, err := ytbx.LoadFile(tempManifest)
		require.NoError(t, err)

		// compare
		opts := []dyff.CompareOption{dyff.IgnoreOrderChanges(true)}
		report, err := dyff.CompareInputFiles(input1, input2, opts...)
		require.NoError(t, err)

		for _, diff := range report.Diffs {
			t.Log(diff.Path.ToDotStyle())
		}
	}
}

func testGetModel(client *server.TestEdgeClient) func(*testing.T) {
	return func(t *testing.T) {
		ctx := t.Context()

		resp, err := client.V3.Reader.GetModel(ctx, &dsr3.GetModelRequest{})
		if err != nil {
			require.NoError(t, err)
		}

		fmt.Println(resp.GetModel().GetModel())
	}
}

func testDeleteManifest(client *server.TestEdgeClient) func(*testing.T) {
	return func(t *testing.T) {
		require.NoError(t, deleteManifest(client))
	}
}

func deleteManifest(client *server.TestEdgeClient) error {
	_, err := client.V3.Writer.DeleteManifest(
		context.Background(),
		&dsw3.DeleteManifestRequest{},
	)

	return err
}

type testData struct {
	Objects   []*dsc3.Object   `json:"objects"`
	Relations []*dsc3.Relation `json:"relations"`
}

func loadData(client *server.TestEdgeClient, dataFile string) error {
	bin, err := os.ReadFile(dataFile)
	if err != nil {
		return err
	}

	var td testData
	if err := json.Unmarshal(bin, &td); err != nil {
		return err
	}

	ctx := context.Background()

	for _, obj := range td.Objects {
		if _, err := client.V3.Writer.SetObject(ctx, &dsw3.SetObjectRequest{Object: obj}); err != nil {
			return err
		}
	}

	for _, rel := range td.Relations {
		if _, err := client.V3.Writer.SetRelation(ctx, &dsw3.SetRelationRequest{Relation: rel}); err != nil {
			return err
		}
	}

	return nil
}

const (
	removeObjectInUse = `
model:
  version: 3

types: {}
`

	removeRelationInUse = `
model:
  version: 3

types:
  user: {}
`

	removeDirectAssignemntInUse = `
model:
  version: 3

types:
  user:
    relations:
      manager: group

  group:
    relations:
      member: user
`
)
