package graph_test

import (
	"strings"
	"testing"

	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	"github.com/aserto-dev/topaz/azm/graph"
	"github.com/aserto-dev/topaz/azm/mempool"
	"github.com/aserto-dev/topaz/azm/model"
	v3 "github.com/aserto-dev/topaz/azm/v3"
	"github.com/samber/lo"
	req "github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSearchObjects(t *testing.T) {
	require := req.New(t)

	rels := relations()

	m, err := v3.LoadFile("./check_test.yaml")
	require.NoError(err)
	require.NotNil(m)

	im := m.Invert()
	mnfst := manifest(im)

	b, err := yaml.Marshal(mnfst)
	require.NoError(err)

	t.Logf("inverted model:\n%s\n", b)

	require.NoError(
		im.Validate(model.AllowPermissionInArrowBase),
	)

	pool := mempool.NewRelationsPool()

	for _, test := range searchObjectsTests {
		t.Run(test.search, func(tt *testing.T) {
			assert := req.New(tt)

			objSearch, err := graph.NewObjectSearch(m, graphReq(test.search), rels.GetRelations, pool)
			assert.NoError(err)

			res, err := objSearch.Search()
			assert.NoError(err)
			tt.Logf("explanation: +%v\n", res.GetExplanation().AsMap())
			tt.Logf("trace: +%v\n", res.GetTrace())

			subjects := lo.Map(res.GetResults(), func(s *dsc.ObjectIdentifier, _ int) object {
				return object{
					ObjectType: model.ObjectName(s.GetObjectType()),
					ObjectID:   model.ObjectID(s.GetObjectId()),
				}
			})

			for _, e := range test.expected {
				assert.Contains(subjects, e)
			}

			assert.Len(test.expected, len(subjects), subjects)
		})
	}
}

type object struct {
	ObjectType model.ObjectName
	ObjectID   model.ObjectID
}

type searchTest struct {
	search   string
	expected []object
}

var searchObjectsTests = []searchTest{
	// Relations
	{"folder:?#owner@user:f1_owner", []object{{"folder", "folder1"}}},
	{"folder:?#viewer@group:f1_viewers#member", []object{{"folder", "folder1"}}},
	{"group:?#member@user:user2", []object{{"group", "d1_viewers"}}},
	{"group:?#member@group:d1_subviewers#member", []object{{"group", "d1_viewers"}}},
	{"group:?#member@user:f1_viewer", []object{{"group", "f1_viewers"}}},
	{"folder:?#parent@folder:folder1", []object{{"folder", "folder2"}}},
	{"doc:?#owner@user:d1_owner", []object{{"doc", "doc1"}}},
	{"doc:?#viewer@user:f1_viewer", []object{{"doc", "doc2"}}},
	{"doc:?#viewer@group:d1_viewers#member", []object{{"doc", "doc1"}}},
	{"doc:?#parent@folder:folder1", []object{{"doc", "doc1"}, {"doc", "doc2"}}},
	{"doc:?#in_parent_chain@folder:folder1", []object{{"doc", "doc1"}, {"doc", "doc2"}, {"doc", "doc3"}}},
	{"group:?#member@group:leaf#member", []object{{"group", "branch"}, {"group", "trunk"}, {"group", "root"}}},
	{"doc:?#viewer@group:leaf#member", []object{{"doc", "doc_tree"}}},
	{"group:?#member@group:yang#member", []object{{"group", "yin"}, {"group", "yang"}}},
	{"group:?#member@user:user3", []object{{"group", "d1_subviewers"}, {"group", "d1_viewers"}}},
	{"group:?#member@user:yin_user", []object{{"group", "yin"}, {"group", "yang"}}},
	{"doc:?#viewer@group:d1_subviewers#member", []object{{"doc", "doc1"}}},
	{"doc:?#auditor@user:boss", []object{{"doc", "doc1"}}},
	{"doc:?#auditor@user:employee#manager", []object{{"doc", "doc1"}}},

	// wildcard
	{"doc:?#viewer@user:user1", []object{{"doc", "doc1"}, {"doc", "doc2"}}},
	{"doc:?#viewer@user:f1_owner", []object{{"doc", "doc1"}, {"doc", "doc2"}}},
	{"doc:?#viewer@user:user2", []object{{"doc", "doc1"}, {"doc", "doc2"}}},
	{"doc:?#viewer@user:*", []object{{"doc", "doc2"}}},
	{"doc:?#viewer@user:some_user", []object{{"doc", "doc2"}}},

	// Permissions
	{"folder:?#is_owner@user:f1_owner", []object{{"folder", "folder1"}, {"folder", "folder2"}}},
	{"folder:?#can_create_file@user:f1_owner", []object{{"folder", "folder1"}, {"folder", "folder2"}}},
	{"folder:?#can_read@user:f1_owner", []object{{"folder", "folder1"}, {"folder", "folder2"}}},
	{"folder:?#can_share@user:f1_owner", []object{{"folder", "folder1"}, {"folder", "folder2"}}},
	{"doc:?#can_change_owner@user:f1_owner", []object{{"doc", "doc1"}, {"doc", "doc2"}, {"doc", "doc3"}}},
	{"doc:?#can_write@user:f1_owner", []object{{"doc", "doc1"}, {"doc", "doc2"}, {"doc", "doc3"}}},
	{"doc:?#can_read@user:f1_owner", []object{{"doc", "doc1"}, {"doc", "doc2"}, {"doc", "doc3"}}},
	{"doc:?#can_share@user:f1_owner", []object{{"doc", "doc1"}, {"doc", "doc2"}, {"doc", "doc3"}}},
	{"doc:?#can_invite@user:f1_owner", []object{{"doc", "doc2"}, {"doc", "doc3"}}},
	{"folder:?#is_owner@group:f1_viewers", []object{}},
	{"folder:?#can_create_file@group:f1_viewers", []object{}},
	{"folder:?#can_read@group:f1_viewers#member", []object{{"folder", "folder1"}, {"folder", "folder2"}}},
	{"folder:?#can_read@group:f1_viewers", []object{}},
	{"folder:?#can_share@group:f1_viewers", []object{}},
	{"doc:?#can_change_owner@group:f1_viewers", []object{}},
	{"doc:?#can_write@group:f1_viewers", []object{}},
	{"doc:?#can_read@group:f1_viewers", []object{}},
	{"doc:?#can_read@group:f1_viewers#member", []object{{"doc", "doc1"}, {"doc", "doc2"}, {"doc", "doc3"}}},
	{"doc:?#can_share@group:f1_viewers#member", []object{}},
	{"doc:?#can_invite@group:f1_viewers#member", []object{{"doc", "doc1"}, {"doc", "doc2"}, {"doc", "doc3"}}},
	{"folder:?#is_owner@user:f1_viewer", []object{}},
	{"folder:?#can_create_file@user:f1_viewer", []object{}},
	{"folder:?#can_read@user:f1_viewer", []object{{"folder", "folder1"}, {"folder", "folder2"}}},
	{"folder:?#can_share@user:f1_viewer", []object{}},
	{"doc:?#can_change_owner@user:f1_viewer", []object{}},
	{"doc:?#can_write@user:f1_viewer", []object{}},
	{"doc:?#can_read@user:f1_viewer", []object{{"doc", "doc1"}, {"doc", "doc2"}, {"doc", "doc3"}}},
	{"doc:?#can_share@user:f1_viewer", []object{}},
	{"doc:?#can_invite@user:f1_viewer", []object{{"doc", "doc1"}, {"doc", "doc2"}, {"doc", "doc3"}}},
	{"folder:?#is_owner@user:d1_owner", []object{}},
	{"folder:?#can_create_file@user:d1_owner", []object{}},
	{"folder:?#can_read@user:d1_owner", []object{}},
	{"folder:?#can_share@user:d1_owner", []object{}},
	{"doc:?#can_change_owner@user:d1_owner", []object{{"doc", "doc1"}}},
	{"doc:?#can_write@user:d1_owner", []object{{"doc", "doc1"}}},
	{"doc:?#can_read@user:d1_owner", []object{{"doc", "doc1"}, {"doc", "doc2"}}},
	{"doc:?#can_share@user:d1_owner", []object{}},
	{"doc:?#can_invite@user:d1_owner", []object{}},
	{"folder:?#is_owner@group:d1_viewers", []object{}},
	{"folder:?#can_create_file@group:d1_viewers", []object{}},
	{"folder:?#can_read@group:d1_viewers#member", []object{}},
	{"folder:?#can_read@group:d1_viewers#member", []object{}},
	{"folder:?#can_share@group:d1_viewers#member", []object{}},
	{"doc:?#can_change_owner@group:d1_viewers#member", []object{}},
	{"doc:?#can_write@group:d1_viewers#member", []object{}},
	{"doc:?#can_read@group:d1_viewers#member", []object{{"doc", "doc1"}}},
	{"doc:?#can_share@group:d1_viewers#member", []object{}},
	{"doc:?#can_invite@group:d1_viewers#member", []object{}},
	{"folder:?#is_owner@user:user1", []object{}},
	{"folder:?#can_create_file@user:user1", []object{}},
	{"folder:?#can_read@user:user1", []object{}},
	{"folder:?#can_share@user:user1", []object{}},
	{"doc:?#can_change_owner@user:user1", []object{}},
	{"doc:?#can_write@user:user1", []object{}},
	{"doc:?#can_read@user:user1", []object{{"doc", "doc1"}, {"doc", "doc2"}}},
	{"doc:?#can_share@user:user1", []object{}},
	{"doc:?#can_invite@user:user1", []object{}},
	{"folder:?#is_owner@user:user2", []object{}},
	{"folder:?#can_create_file@user:user2", []object{}},
	{"folder:?#can_read@user:user2", []object{}},
	{"folder:?#can_share@user:user2", []object{}},
	{"doc:?#can_change_owner@user:user2", []object{}},
	{"doc:?#can_write@user:user2", []object{}},
	{"doc:?#can_read@user:user2", []object{{"doc", "doc1"}, {"doc", "doc2"}}},
	{"doc:?#can_share@user:user2", []object{}},
	{"doc:?#can_invite@user:user2", []object{}},
	{"folder:?#is_owner@user:user3", []object{}},
	{"folder:?#can_create_file@user:user3", []object{}},
	{"folder:?#can_read@user:user3", []object{}},
	{"folder:?#can_share@user:user3", []object{}},
	{"doc:?#can_change_owner@user:user3", []object{}},
	{"doc:?#can_write@user:user3", []object{}},
	{"doc:?#can_read@user:user3", []object{{"doc", "doc1"}, {"doc", "doc2"}}},
	{"doc:?#can_share@user:user3", []object{}},
	{"doc:?#can_invite@user:user3", []object{}},
	{"doc:?#can_view@user:some_user", []object{{"doc", "doc2"}}},
}

func relations() RelationsReader {
	return NewRelationsReader(
		"folder:folder1#owner@user:f1_owner",
		"folder:folder2#parent@folder:folder1",
		"folder:folder1#viewer@group:f1_viewers#member",
		"group:f1_viewers#member@user:f1_viewer",
		"doc:doc1#parent@folder:folder1",
		"doc:doc1#owner@user:d1_owner",
		"doc:doc1#viewer@group:d1_viewers#member",
		"doc:doc1#viewer@user:user1",
		"doc:doc1#viewer@user:f1_owner",
		"group:d1_viewers#member@user:user2",
		"doc:doc2#parent@folder:folder1",
		"doc:doc2#viewer@user:*",
		"doc:doc2#viewer@user:user2",
		"doc:doc3#parent@folder:folder2",
		"user:employee#manager@user:boss",
		"doc:doc1#auditor@user:employee#manager",

		"group:d1_viewers#member@group:d1_subviewers#member",
		"group:d1_subviewers#member@user:user3",
		"group:f1_viewers#member@group:f1_subviewers#member",
		"group:d1_subviewers#member@user:user4",

		// nested groups
		"group:leaf#member@user:leaf_user",
		"group:branch#member@group:leaf#member",
		"group:trunk#member@group:branch#member",
		"group:root#member@group:trunk#member",
		"doc:doc_tree#viewer@group:root#member",

		// mutually recursive groups with users
		"group:yin#member@group:yang#member",
		"group:yang#member@group:yin#member",
		"group:yin#member@user:yin_user",
		"group:yang#member@user:yang_user",

		// mutually recursive groups with no users
		"group:alpha#member@group:omega#member",
		"group:omega#member@group:alpha#member",
	)
}

func manifest(m *model.Model) *v3.Manifest {
	mnfst := v3.Manifest{
		ModelInfo: &v3.ModelInfo{Version: v3.SchemaVersion(v3.SupportedSchemaVersion)},
		ObjectTypes: lo.MapEntries(m.Objects, func(on model.ObjectName, o *model.Object) (v3.ObjectTypeName, *v3.ObjectType) {
			return v3.ObjectTypeName(on), &v3.ObjectType{
				Relations: lo.MapEntries(o.Relations, func(rn model.RelationName, r *model.Relation) (v3.RelationName, string) {
					return v3.RelationName(rn), strings.Join(
						lo.Map(r.Union, func(rr *model.RelationRef, _ int) string {
							return rr.String()
						}),
						" | ",
					)
				}),
				Permissions: lo.MapEntries(o.Permissions, func(pn model.RelationName, p *model.Permission) (v3.PermissionName, string) {
					name := v3.PermissionName(pn)

					var (
						terms    []*model.PermissionTerm
						operator string
					)

					switch {
					case p.IsUnion():
						terms = p.Union
						operator = " | "
					case p.IsIntersection():
						terms = p.Intersection
						operator = " & "
					case p.IsExclusion():
						terms = []*model.PermissionTerm{p.Exclusion.Include, p.Exclusion.Exclude}
						operator = " - "
					}

					return name, strings.Join(lo.Map(terms, func(pt *model.PermissionTerm, _ int) string {
						return pt.String()
					}), operator)
				}),
			}
		}),
	}

	return &mnfst
}
