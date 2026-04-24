package graph_test

import (
	"strings"
	"testing"

	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
	azmgraph "github.com/aserto-dev/topaz/azm/graph"
	"github.com/aserto-dev/topaz/azm/mempool"
	"github.com/aserto-dev/topaz/azm/model"
	v3 "github.com/aserto-dev/topaz/azm/v3"
	"github.com/samber/lo"
	rq "github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestComputedSet(t *testing.T) {
	require := rq.New(t)

	m, err := v3.LoadFile("./computed_set.yaml")
	require.NoError(err)

	tests := []struct {
		check    string
		expected bool
	}{
		{"resource:album#can_view@user:frank", true},
		{"resource:album#can_view@user:seymour", false},
		{"resource:album#can_view@identity:zappa", true},
		{"resource:album#can_view@identity:duncan", false},
		{"resource:poster#can_view@user:frank", true},
		{"resource:poster#can_view@user:seymour", true},
		{"resource:poster#can_view@identity:zappa", false},
		{"resource:poster#can_view@identity:duncan", false},
		{"resource:t_shirt#can_view@user:frank", false},
		{"resource:t_shirt#can_view@user:seymour", false},
		{"resource:t_shirt#can_view@identity:zappa", true},
		{"resource:t_shirt#can_view@identity:duncan", true},
		{"resource:concert#can_view@user:frank", true},
		{"resource:concert#can_view@user:seymour", true},
		{"resource:concert#can_view@identity:zappa", true},
		{"resource:concert#can_view@identity:duncan", true},

		{"component:guitar#can_repair@identity:zappa", true},
		{"component:guitar#can_repair@user:frank", true},
		{"component:coil#can_repair@identity:zappa", false},
		{"component:coil#can_repair@user:frank", false},
		{"component:pickup#can_repair@identity:zappa", false},
		{"component:pickup#can_repair@user:frank", false},

		{"component:guitar#can_repair@identity:duncan", true},
		{"component:guitar#can_repair@user:seymour", true},
		{"component:pickup#can_repair@identity:duncan", true},
		{"component:pickup#can_repair@user:seymour", true},
		{"component:coil#can_repair@identity:duncan", true},
		{"component:coil#can_repair@user:seymour", true},
		{"component:magnet#can_repair@identity:duncan", false},
		{"component:magnet#can_repair@user:seymour", false},
	}

	pool := mempool.NewRelationsPool()

	for _, test := range tests {
		t.Run(test.check, func(tt *testing.T) {
			require := rq.New(tt)

			checker := azmgraph.NewCheck(m, checkReq(test.check, true), csRels.GetRelations, pool)

			res, err := checker.Check()
			require.NoError(err)
			tt.Log("trace:\n", strings.Join(checker.Trace(), "\n"))
			require.Equal(test.expected, res)
		})
	}
}

func TestComputedSetSearchSubjects(t *testing.T) {
	require := rq.New(t)
	m, err := v3.LoadFile("./computed_set.yaml")
	require.NoError(err)
	require.NotNil(m)

	tests := []searchTest{
		{"user:frank#identifier@identity:?", []object{{"identity", "zappa"}}},
		{"resource:album#can_view@user:?", []object{{"user", "frank"}, {"user", "unidentified"}}},
		{"resource:album#can_view@identity:?", []object{{"identity", "zappa"}}},
		{"resource:poster#can_view@user:?", []object{{"user", "*"}}},
		{"resource:poster#can_view@identity:?", []object{}},
		{"resource:t_shirt#can_view@user:?", []object{}},
		{"resource:t_shirt#can_view@identity:?", []object{{"identity", "*"}}},
		{"resource:concert#can_view@user:?", []object{{"user", "*"}}},
		{"resource:concert#can_view@identity:?", []object{{"identity", "*"}}},
		{"component:guitar#can_repair@user:?", []object{{"user", "seymour"}, {"user", "frank"}, {"user", "unidentified"}}},
		{"component:guitar#can_repair@identity:?", []object{{"identity", "duncan"}, {"identity", "zappa"}}},
		{"component:guitar#can_repair@group:?#member", []object{{"group", "guitarists"}}},
		{"component:pickup#can_repair@identity:?", []object{{"identity", "duncan"}}},
		{"component:pickup#can_repair@user:?", []object{{"user", "seymour"}}},
		{"component:coil#can_repair@user:?", []object{{"user", "seymour"}}},
		{"component:magnet#can_repair@user:?", []object{}},
		{"component:pickup#is_part_maintainer@user:?", []object{{"user", "seymour"}}},
		{"component:pickup#is_part_maintainer@identity:?", []object{{"identity", "duncan"}}},
		{"component:guitar#is_part_maintainer@user:?", []object{{"user", "seymour"}, {"user", "frank"}, {"user", "unidentified"}}},
		{"component:guitar#is_part_maintainer@identity:?", []object{{"identity", "duncan"}, {"identity", "zappa"}}},
	}

	pool := mempool.NewRelationsPool()
	for _, test := range tests {
		t.Run(test.search, testRunner(m, azmgraph.NewSubjectSearch, pool, &test))
	}
}

func TestComputedSetSearchObjects(t *testing.T) {
	require := rq.New(t)
	m, err := v3.LoadFile("./computed_set.yaml")
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

	tests := []searchTest{
		{"user:?#identifier@identity:zappa", []object{{"user", "frank"}}},
		{"resource:?#can_view@user:frank", []object{{"resource", "album"}, {"resource", "poster"}, {"resource", "concert"}}},
		{"resource:?#can_view@identity:zappa", []object{{"resource", "album"}, {"resource", "t_shirt"}, {"resource", "concert"}}},
		{"resource:?#can_view@user:*", []object{{"resource", "poster"}, {"resource", "concert"}}},
		{"resource:?#can_view@user:seymour", []object{{"resource", "poster"}, {"resource", "concert"}}},
		{"resource:?#can_view@identity:*", []object{{"resource", "t_shirt"}, {"resource", "concert"}}},
		{"resource:?#can_view@identity:duncan", []object{{"resource", "t_shirt"}, {"resource", "concert"}}},
		{"component:?#can_repair@user:seymour", []object{{"component", "guitar"}, {"component", "pickup"}, {"component", "coil"}}},
		{"component:?#can_repair@identity:duncan", []object{{"component", "guitar"}, {"component", "pickup"}, {"component", "coil"}}},
		{"component:?#can_repair@user:frank", []object{{"component", "guitar"}, {"component", "string"}}},
		{"component:?#can_repair@identity:zappa", []object{{"component", "guitar"}, {"component", "string"}}},
	}

	pool := mempool.NewRelationsPool()
	for _, test := range tests {
		t.Run(test.search, testRunner(m, azmgraph.NewObjectSearch, pool, &test))
	}
}

type searchable interface {
	Search() (*dsr.GraphResponse, error)
}

type searchFactory[T searchable] func(
	m *model.Model,
	req *dsr.GraphRequest,
	reader azmgraph.RelationReader,
	pool *mempool.RelationsPool,
) (T, error)

func testRunner[T searchable](m *model.Model, factory searchFactory[T], pool *mempool.RelationsPool, test *searchTest) func(*testing.T) {
	return func(tt *testing.T) {
		require := rq.New(tt)

		search, err := factory(m, graphReq(test.search), csRels.GetRelations, pool)
		require.NoError(err)

		res, err := search.Search()
		require.NoError(err)
		tt.Logf("explanation: +%v\n", res.GetExplanation().AsMap())
		tt.Logf("trace: +%v\n", res.GetTrace())

		subjects := lo.Map(res.GetResults(), func(s *dsc.ObjectIdentifier, _ int) object {
			return object{
				ObjectType: model.ObjectName(s.GetObjectType()),
				ObjectID:   model.ObjectID(s.GetObjectId()),
			}
		})

		for _, e := range test.expected {
			require.Contains(subjects, e)
		}

		require.Len(test.expected, len(subjects), subjects)
	}
}

var csRels = NewRelationsReader(
	"user:frank#identifier@identity:zappa",
	"group:guitarists#member@user:frank",
	"group:guitarists#member@user:unidentified",
	"group:musicians#member@group:guitarists#member",
	"resource:album#viewer@group:musicians#member",
	"resource:poster#public_viewer@user:*",
	"resource:t_shirt#public_viewer@identity:*",
	"resource:concert#public_viewer@user:*",
	"resource:concert#public_viewer@identity:*",

	"user:seymour#identifier@identity:duncan",
	"component:coil#maintainer@user:seymour",
	"component:string#maintainer@group:guitarists#member",
	"component:pickup#part@component:magnet",
	"component:pickup#part@component:coil",
	"component:guitar#part@component:pickup#part",
	"component:guitar#part@component:string",
)
