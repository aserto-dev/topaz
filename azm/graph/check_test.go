package graph_test

import (
	"strings"
	"testing"

	azmgraph "github.com/aserto-dev/topaz/azm/graph"
	"github.com/aserto-dev/topaz/azm/mempool"
	v3 "github.com/aserto-dev/topaz/azm/v3"
	"github.com/rs/zerolog"
	assert "github.com/stretchr/testify/require"
)

type cycle bool

const (
	noCycle  cycle = false
	hasCycle cycle = true
)

var tests = []struct {
	check    string
	expected bool
	cycle    cycle
}{
	// Relations
	{"doc:doc1#owner@user:user1", false, noCycle},
	{"doc:doc1#viewer@user:user1", true, noCycle},
	{"doc:doc2#viewer@user:user1", true, noCycle},
	{"doc:doc2#viewer@user:userX", true, noCycle},
	{"doc:doc1#viewer@user:user2", true, noCycle},
	{"doc:doc1#viewer@user:user3", true, noCycle},
	{"doc:doc1#viewer@group:d1_viewers", false, noCycle},

	{"group:yin#member@user:yin_user", true, hasCycle},
	{"group:yin#member@user:yang_user", true, hasCycle},
	{"group:yang#member@user:yin_user", true, hasCycle},
	{"group:yang#member@user:yang_user", true, hasCycle},

	{"group:alpha#member@user:user1", false, hasCycle},

	// Permissions
	{"doc:doc1#can_change_owner@user:d1_owner", true, noCycle},
	{"doc:doc1#can_change_owner@user:user1", false, noCycle},
	{"doc:doc1#can_change_owner@user:userX", false, noCycle},

	{"doc:doc1#can_read@user:d1_owner", true, noCycle},
	{"doc:doc1#can_read@user:f1_owner", true, noCycle},
	{"doc:doc1#can_read@user:user1", true, noCycle},
	{"doc:doc1#can_read@user:f1_viewer", true, noCycle},
	{"doc:doc1#can_read@user:userX", false, noCycle},

	{"doc:doc1#can_write@user:d1_owner", true, noCycle},
	{"doc:doc1#can_write@user:f1_owner", true, noCycle},
	{"doc:doc1#can_write@user:user2", false, noCycle},

	{"folder:folder1#owner@user:f1_owner", true, noCycle},
	{"folder:folder1#can_create_file@user:f1_owner", true, noCycle},
	{"folder:folder1#can_share@user:f1_owner", true, noCycle},

	// intersection
	{"doc:doc1#can_share@user:d1_owner", false, noCycle},
	{"doc:doc1#can_share@user:f1_owner", true, noCycle},

	// negation
	{"folder:folder1#can_read@user:f1_owner", true, noCycle},
	{"doc:doc1#viewer@user:f1_owner", false, noCycle},
	{"doc:doc1#can_invite@user:f1_owner", true, noCycle},

	// cycles
	{"cycle:loop#can_delete@user:loop_owner", true, hasCycle},
	{"cycle:loop#can_delete@user:user1", false, noCycle},
}

func TestCheck(t *testing.T) {
	m, err := v3.LoadFile("./check_test.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, m)

	pool := mempool.NewRelationsPool()

	for _, test := range tests {
		t.Run(test.check, func(tt *testing.T) {
			assert := assert.New(tt)

			checker := azmgraph.NewCheck(m, checkReq(test.check, true), rels.GetRelations, pool)

			res, err := checker.Check()
			assert.NoError(err)
			tt.Log("trace:\n", strings.Join(checker.Trace(), "\n"))
			assert.Equal(test.expected, res)

			reason := checker.Reason()
			if test.cycle {
				assert.True(strings.HasPrefix(reason, "cycles detected:"), "expected 'reason' to begin with 'cycles detected:'")
				assert.Contains(reason, test.check)
			} else {
				assert.Empty(reason)
			}
		})
	}
}

func BenchmarkCheck(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	m, err := v3.LoadFile("./check_test.yaml")
	if err != nil {
		b.Fatalf("failed to load model: %s", err)
	}

	pool := mempool.NewRelationsPool()

	b.ResetTimer()

	for _, test := range tests {
		assert := assert.New(b)

		b.StopTimer()

		checker := azmgraph.NewCheck(m, checkReq(test.check, false), rels.GetRelations, pool)

		b.StartTimer()

		res, err := checker.Check()
		assert.NoError(err)
		assert.Equal(test.expected, res)
	}
}

var rels = NewRelationsReader(
	"folder:folder1#owner@user:f1_owner",
	"folder:folder1#viewer@group:f1_viewers#member",
	"group:f1_viewers#member@user:f1_viewer",
	"doc:doc1#parent@folder:folder1",
	"doc:doc1#owner@user:d1_owner",
	"doc:doc1#viewer@group:d1_viewers#member",
	"doc:doc1#viewer@user:user1",
	"group:d1_viewers#member@user:user2",
	"doc:doc2#viewer@user:*",

	"group:d1_viewers#member@group:d1_subviewers#member",
	"group:d1_subviewers#member@user:user3",

	// mutually recursive groups with users
	"group:yin#member@group:yang#member",
	"group:yang#member@group:yin#member",
	"group:yin#member@user:yin_user",
	"group:yang#member@user:yang_user",

	// mutually recursive groups with no users
	"group:alpha#member@group:omega#member",
	"group:omega#member@group:alpha#member",

	// cyclical permissions
	"cycle:loop#parent@cycle:loop",
	"cycle:loop#owner@user:loop_owner",
)
