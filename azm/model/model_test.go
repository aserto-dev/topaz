package model_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	"github.com/aserto-dev/topaz/azm/model"
	v3 "github.com/aserto-dev/topaz/azm/v3"

	"github.com/hashicorp/go-multierror"
	"github.com/nsf/jsondiff"
	stretch "github.com/stretchr/testify/require"
)

var m1 = model.Model{
	Version: model.ModelVersion,
	Objects: map[model.ObjectName]*model.Object{
		model.ObjectName("user"): {},
		model.ObjectName("group"): {
			Relations: map[model.RelationName]*model.Relation{
				model.RelationName("member"): {
					Union: []*model.RelationRef{
						{Object: model.ObjectName("user")},
						{Object: model.ObjectName("group"), Relation: model.RelationName("member")},
					},
				},
			},
		},

		model.ObjectName("folder"): {
			Relations: map[model.RelationName]*model.Relation{
				model.RelationName("owner"): {
					Union: []*model.RelationRef{
						{Object: model.ObjectName("user")},
					},
				},
			},
			Permissions: map[model.RelationName]*model.Permission{
				model.RelationName("read"): {
					Union: []*model.PermissionTerm{{RelOrPerm: "owner"}},
				},
			},
		},
		model.ObjectName("document"): {
			Relations: map[model.RelationName]*model.Relation{
				model.RelationName("parent_folder"): {
					Union: []*model.RelationRef{{Object: model.ObjectName("folder")}},
				},
				model.RelationName("writer"): {
					Union: []*model.RelationRef{{Object: model.ObjectName("user")}},
				},
				model.RelationName("reader"): {
					Union: []*model.RelationRef{
						{Object: model.ObjectName("user")},
						{Object: model.ObjectName("user"), Relation: "*"},
					},
				},
			},
			Permissions: map[model.RelationName]*model.Permission{
				model.RelationName("edit"): {
					Union: []*model.PermissionTerm{{RelOrPerm: "writer"}},
				},
				model.RelationName("view"): {
					Union: []*model.PermissionTerm{
						{RelOrPerm: "reader"},
						{RelOrPerm: "writer"},
					},
				},
				model.RelationName("read_and_write"): {
					Intersection: []*model.PermissionTerm{
						{RelOrPerm: "reader"},
						{RelOrPerm: "writer"},
					},
				},
				model.RelationName("can_only_read"): {
					Exclusion: &model.ExclusionPermission{
						Include: &model.PermissionTerm{RelOrPerm: "reader"},
						Exclude: &model.PermissionTerm{RelOrPerm: "writer"},
					},
				},
				model.RelationName("read"): {
					Union: []*model.PermissionTerm{{Base: "parent_folder", RelOrPerm: "read"}},
				},
			},
		},
	},
}

func TestProgrammaticModel(t *testing.T) {
	stretch.NoError(t, m1.Validate())

	b1, err := json.Marshal(m1)
	stretch.NoError(t, err)

	w, err := os.Create("./model_test.json")
	stretch.NoError(t, err)
	t.Cleanup(func() { _ = w.Close() })

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(m1); err != nil {
		stretch.NoError(t, err)
	}

	b2, err := os.ReadFile("./testdata/model.json")
	stretch.NoError(t, err)

	m2 := model.Model{}
	if err := json.Unmarshal(b2, &m2); err != nil {
		stretch.NoError(t, err)
	}

	opts := jsondiff.DefaultJSONOptions()
	if diff, str := jsondiff.Compare(b1, b2, &opts); diff != jsondiff.FullMatch {
		stretch.Equal(t, jsondiff.FullMatch, diff, "diff: %s", str)
	}
}

func TestModel(t *testing.T) {
	buf, err := os.ReadFile("./testdata/model.json")
	stretch.NoError(t, err)

	m := model.Model{}
	if err := json.Unmarshal(buf, &m); err != nil {
		stretch.NoError(t, err)
	}

	b1, err := json.Marshal(m)
	stretch.NoError(t, err)

	var m2 model.Model
	if err := json.Unmarshal(b1, &m2); err != nil {
		stretch.NoError(t, err)
	}

	b2, err := json.Marshal(m2)
	stretch.NoError(t, err)

	opts := jsondiff.DefaultConsoleOptions()
	if diff, str := jsondiff.Compare(buf, b1, &opts); diff != jsondiff.FullMatch {
		stretch.Equal(t, jsondiff.FullMatch, diff, "diff: %s", str)
	}

	if diff, str := jsondiff.Compare(buf, b2, &opts); diff != jsondiff.FullMatch {
		stretch.Equal(t, jsondiff.FullMatch, diff, "diff: %s", str)
	}

	if diff, str := jsondiff.Compare(b1, b2, &opts); diff != jsondiff.FullMatch {
		stretch.Equal(t, jsondiff.FullMatch, diff, "diff: %s", str)
	}
}

func TestValidation(t *testing.T) { //nolint:funlen
	tests := []struct {
		name           string
		manifest       string
		expectedErrors []string
	}{
		{
			"valid manifest",
			"./testdata/valid.yaml",
			[]string{},
		},
		{
			"invalid names",
			"./testdata/invalid_names.yaml",
			[]string{
				"invalid relation type: invalid name 'resource:_reader'",
				"invalid relation type: invalid name 'resource:reader.'",
				"invalid relation type: invalid name 'resource:1reader'",
				"invalid relation type: invalid name 'resource:r)e(d*e&r'",
				"invalid permission: invalid name 'resource:_can_read_'",
				"invalid permission: invalid name 'resource:Can+Reader'",
				"invalid permission: invalid name 'resource:@!#'",
				"invalid object type: invalid name '_user'",
				"invalid object type: invalid name '12user'",
				"invalid object type: invalid name 'u!s@e#r'",
			},
		},
		{
			"invalid definitions",
			"./testdata/invalid_terms.yaml",
			[]string{
				"unexpected '->' in 'user->user'",
				"unexpected '|' in '| user'",
				"relation 'resource:empty' has empty definition",
				"unexpected '$' in '$#@!'",
				"unexpected '!' in 'us!!er'",
				"unexpected '1' in '123user'",
				"unexpected '<EOF>' in 'user |'",
				"unexpected '|' in 'user | | user",
				"* unexpected '-' in '-user | user$ | _._': identifier expected\n" +
					"\t* unexpected '$' in '-user | user$ | _._': identifier expected\n" +
					"\t* unexpected '_' in '-user | user$ | _._'. expected ID: invalid expression",
				"unexpected '-' in '-user'",
				"unexpected '$' in 'user$'",
				"no viable alternative at input 'this*': parse error",
				"no viable alternative at input 'bad@': parse error",
				"no viable alternative at input 'base->bad$': parse error",
			},
		},
		{
			"relation/permission collision",
			"./testdata/rel_perm_collision.yaml",
			[]string{
				"permission name 'file:writer' conflicts with relation 'file:writer'",
				"relation 'file:bad' has empty definition",
			},
		},
		{
			"relations to undefined targets",
			"./testdata/undefined_rel_targets.yaml",
			[]string{
				"relation 'file:owner' references undefined object type 'person'",
				"relation 'file:reader' references undefined object type 'team'",
				"relation 'file:reader' references undefined object type 'project'",
				"relation 'file:writer' references undefined object type 'team'",
				"relation 'file:admin' references undefined relation type 'group#admin'",
			},
		},
		{
			"permissions to undefined targets",
			"./testdata/undefined_perm_targets.yaml",
			[]string{
				"permission 'folder:read' references undefined relation type 'folder:parent'",
				"permission 'folder:view' references undefined relation or permission 'folder:viewer'",
				"permission 'folder:view' references undefined relation or permission 'folder:guest'",
				"permission 'folder:write' references undefined relation or permission 'folder:editor'",
			},
		},
		{
			"cyclic relation definitions",
			"./testdata/invalid_cycles.yaml",
			[]string{
				"relation 'team:member' is circular and does not resolve to any object types",
				"relation 'team:owner' is circular and does not resolve to any object types",
				"relation 'project:owner' is circular and does not resolve to any object types",
			},
		},
		{
			"permissions with invalid targets",
			"./testdata/invalid_perms.yaml",
			[]string{
				"permission 'file:write' references 'owner->write', which can resolve to undefined relation or permission 'user:write'",
				"permission 'file:update' references 'parent->write', which can resolve to undefined relation or permission 'folder:write'",
			},
		},
		{
			"arrow permissions with wildcard at arrow base",
			"./testdata/wildcard_arrow.yaml",
			[]string{
				"wildcard relation 'resource:viewer' not allowed in the base of an arrow operator 'viewer->identifier' in permission 'resource:can_view'",
				"wildcard relation 'component:part' not allowed in the base of an arrow operator 'part->can_repair' in " +
					"permission 'component:can_repair'",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			assert := stretch.New(tt)
			m, err := v3.LoadFile(test.manifest)

			// Log the model for debugging purposes.
			var b bytes.Buffer

			enc := json.NewEncoder(&b)
			enc.SetIndent("", "  ")
			assert.NoError(enc.Encode(m))
			tt.Logf("model: %s", b.String())

			if len(test.expectedErrors) == 0 {
				assert.NoError(err)
				return
			}

			// verify that we got a load error.
			assert.Error(err)
			// verify that the error is of type ErrInvalidArgument
			aerr := derr.ErrInvalidArgument
			assert.ErrorAs(err, &aerr)
			assert.Equal("E20015", aerr.Code)

			var merr *multierror.Error

			assert.ErrorAs(aerr.Unwrap(), &merr)

			for _, expected := range test.expectedErrors {
				assert.ErrorContains(merr, expected)
			}

			// verify that we got the expected number of errors.
			assert.Len(merr.Errors, len(test.expectedErrors))
		})
	}
}

func TestResolution(t *testing.T) {
	assert := stretch.New(t)
	m, err := v3.LoadFile("./testdata/valid.yaml")
	assert.NoError(err)

	// Relations
	assert.Equal([]model.ObjectName{"user"}, m.Objects["team"].Relations["owner"].SubjectTypes)
	assert.Equal([]model.ObjectName{"team"}, m.Objects["group"].Relations["owner"].SubjectTypes)
	assert.Equal([]model.ObjectName{"group"}, m.Objects["group"].Relations["parent"].SubjectTypes)

	// - order-agnostic set comparison: a subset of equal length.
	assert.Len(m.Objects["team"].Relations["member"].SubjectTypes, 2)
	assert.Subset(m.Objects["team"].Relations["member"].SubjectTypes, []model.ObjectName{"user", "team"})

	assert.Len(m.Objects["group"].Relations["member"].SubjectTypes, 2)
	assert.Subset(m.Objects["group"].Relations["member"].SubjectTypes, []model.ObjectName{"user", "team"})

	assert.Len(m.Objects["group"].Relations["manager"].SubjectTypes, 2)
	assert.Subset(m.Objects["group"].Relations["manager"].SubjectTypes, []model.ObjectName{"user", "team"})

	// Permissions
	assert.Len(m.Objects["group"].Permissions["manage"].SubjectTypes, 2)
	assert.Subset(m.Objects["group"].Permissions["manage"].SubjectTypes, []model.ObjectName{"user", "team"})

	assert.Len(m.Objects["group"].Permissions["delete"].SubjectTypes, 2)
	assert.Subset(m.Objects["group"].Permissions["delete"].SubjectTypes, []model.ObjectName{"user", "team"})

	assert.Equal([]model.ObjectName{"team"}, m.Objects["group"].Permissions["purge"].SubjectTypes)
}
