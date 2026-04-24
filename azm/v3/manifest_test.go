package v3_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/aserto-dev/topaz/azm/model"
	v3 "github.com/aserto-dev/topaz/azm/v3"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestManifestUnmarshal(t *testing.T) {
	buf, err := os.ReadFile("./manifest.yaml")
	require.NoError(t, err)

	manifest := v3.Manifest{}
	if err := yaml.Unmarshal(buf, &manifest); err != nil {
		require.NoError(t, err)
	}

	enc := yaml.NewEncoder(os.Stderr)
	enc.SetIndent(2)

	if err := enc.Encode(&manifest); err != nil {
		require.NoError(t, err)
	}
}

func TestLoadModel(t *testing.T) {
	m1, err := v3.LoadFile("./manifest.yaml")
	require.NoError(t, err)
	require.NotNil(t, m1)

	b1, err := json.Marshal(m1)
	require.NoError(t, err)
	t.Logf("model:\n%s\n", b1)

	b2, err := os.ReadFile("../model/testdata/model.json")
	require.NoError(t, err)

	m2 := model.Model{}
	if err := json.Unmarshal(b2, &m2); err != nil {
		require.NoError(t, err)
	}

	opts := jsondiff.DefaultJSONOptions()
	if diff, str := jsondiff.Compare(b2, b1, &opts); diff != jsondiff.FullMatch {
		require.Equal(t, jsondiff.FullMatch, diff, "diff: %s", str)
	}
}

func TestLoadEmptyManifest(t *testing.T) {
	r := bytes.NewReader([]byte{})

	m1, err := v3.Load(r)
	require.NoError(t, err)
	require.NotNil(t, m1)

	b1, err := json.Marshal(m1)
	require.NoError(t, err)
	require.NotNil(t, b1)
}

func TestLoadManifestWithNullObject(t *testing.T) {
	m1, err := v3.LoadFile("./manifest_empty_user.yaml")
	require.NoError(t, err)
	require.NotNil(t, m1)

	require.Len(t, m1.Objects, 2)
	require.Empty(t, m1.Objects["user"].Permissions)
	require.Empty(t, m1.Objects["user"].Relations)
}
