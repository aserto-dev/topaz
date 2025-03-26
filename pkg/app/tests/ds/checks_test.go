package ds_test

import (
	"context"
	"fmt"
	"testing"

	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/prop"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type checksTestCase struct {
	req  *dsr3.ChecksRequest
	resp *dsr3.ChecksResponse
	err  error
}

func testChecks(ctx context.Context, dsClient dsr3.ReaderClient) func(*testing.T) {
	return func(t *testing.T) {
		for i, tc := range checksTCs {
			t.Run(fmt.Sprintf("checks-%d", i), func(t *testing.T) {
				resp, err := dsClient.Checks(ctx, tc.req)
				if tc.err == nil {
					require.NoError(t, err)
				} else {
					require.ErrorIs(t, tc.err, err)
				}

				require.NotNil(t, resp, "response should not be nil")

				assert.Equal(t, len(tc.req.Checks), len(resp.Checks))

				for i := range tc.resp.Checks {
					require.Equal(t, tc.resp.Checks[i].GetCheck(), resp.Checks[i].GetCheck(), "i=%d", i)
					if tc.resp.Checks[i].Context == nil {
						continue
					}

					if tc.resp.Checks[i].Context.Fields == nil {
						return
					}

					if v, ok := tc.resp.Checks[i].Context.Fields["reason"]; ok {
						require.Equal(t, v.GetStringValue(), resp.Checks[i].Context.Fields["reason"].GetStringValue())
					}
				}
			})
		}
	}
}

var checksTCs []*checksTestCase = []*checksTestCase{
	// id = 0
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{
				ObjectType:  "folder",
				ObjectId:    "morty",
				Relation:    "owner",
				SubjectType: "user",
				SubjectId:   "morty@the-citadel.com",
				Trace:       false,
			},
			Checks: []*dsr3.CheckRequest{
				{},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   true,
					Context: tc.SetContext(prop.Reason, ""),
				},
			},
		},
		err: nil,
	},
	// id = 1
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{
				ObjectType:  "folder2",
				ObjectId:    "morty",
				Relation:    "owner",
				SubjectType: "user",
				SubjectId:   "morty@the-citadel.com",
				Trace:       false,
			},
			Checks: []*dsr3.CheckRequest{
				{},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20026 object type not found: folder2: object_type"),
				},
			},
		},
		err: nil,
	},
	// id = 2
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{
				ObjectType:  "folder",
				ObjectId:    "morty2",
				Relation:    "owner",
				SubjectType: "user",
				SubjectId:   "morty@the-citadel.com",
				Trace:       false,
			},
			Checks: []*dsr3.CheckRequest{
				{},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20025 object not found: object folder:morty2"),
				},
			},
		},
		err: nil,
	},
	// id = 3
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{
				ObjectType:  "folder",
				ObjectId:    "morty",
				Relation:    "owner2",
				SubjectType: "user",
				SubjectId:   "morty@the-citadel.com",
				Trace:       false,
			},
			Checks: []*dsr3.CheckRequest{
				{},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20036 relation type not found: relation: folder#owner2"),
				},
			},
		},
		err: nil,
	},
	// id = 4
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{
				ObjectType:  "folder",
				ObjectId:    "morty",
				Relation:    "owner",
				SubjectType: "user2",
				SubjectId:   "morty@the-citadel.com",
				Trace:       false,
			},
			Checks: []*dsr3.CheckRequest{
				{},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20026 object type not found: user2: subject_type"),
				},
			},
		},
		err: nil,
	},
	// id = 5
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{
				ObjectType:  "folder",
				ObjectId:    "morty",
				Relation:    "owner",
				SubjectType: "user",
				SubjectId:   "morty@the-citadel.com2",
				Trace:       false,
			},
			Checks: []*dsr3.CheckRequest{
				{},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20025 object not found: subject user:morty@the-citadel.com2"),
				},
			},
		},
		err: nil,
	},
	// id = 6
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{},
			Checks: []*dsr3.CheckRequest{
				{
					ObjectType:  "folder",
					ObjectId:    "morty",
					Relation:    "owner",
					SubjectType: "user",
					SubjectId:   "morty@the-citadel.com",
					Trace:       false,
				},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   true,
					Context: tc.SetContext(prop.Reason, ""),
				},
			},
		},
		err: nil,
	},
	// id = 7
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{},
			Checks: []*dsr3.CheckRequest{
				{
					ObjectType:  "folder2",
					ObjectId:    "morty",
					Relation:    "owner",
					SubjectType: "user",
					SubjectId:   "morty@the-citadel.com",
					Trace:       false,
				},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20026 object type not found: folder2: object_type"),
				},
			},
		},
		err: nil,
	},
	// id = 8
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{},
			Checks: []*dsr3.CheckRequest{
				{
					ObjectType:  "folder",
					ObjectId:    "morty2",
					Relation:    "owner",
					SubjectType: "user",
					SubjectId:   "morty@the-citadel.com",
					Trace:       false,
				},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20025 object not found: object folder:morty2"),
				},
			},
		},
		err: nil,
	},
	// id = 9
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{},
			Checks: []*dsr3.CheckRequest{
				{
					ObjectType:  "folder",
					ObjectId:    "morty",
					Relation:    "owner2",
					SubjectType: "user",
					SubjectId:   "morty@the-citadel.com",
					Trace:       false,
				},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20036 relation type not found: relation: folder#owner2"),
				},
			},
		},
		err: nil,
	},
	// id = 10
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{},
			Checks: []*dsr3.CheckRequest{
				{
					ObjectType:  "folder",
					ObjectId:    "morty",
					Relation:    "owner",
					SubjectType: "user2",
					SubjectId:   "morty@the-citadel.com",
					Trace:       false,
				},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20026 object type not found: user2: subject_type"),
				},
			},
		},
		err: nil,
	},
	// id = 11
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{},
			Checks: []*dsr3.CheckRequest{
				{
					ObjectType:  "folder",
					ObjectId:    "morty",
					Relation:    "owner",
					SubjectType: "user",
					SubjectId:   "morty@the-citadel.com2",
					Trace:       false,
				},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20025 object not found: subject user:morty@the-citadel.com2"),
				},
			},
		},
		err: nil,
	},
	// id = 12
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{
				ObjectType:  "folder",
				ObjectId:    "morty",
				Relation:    "",
				SubjectType: "user",
				SubjectId:   "morty@the-citadel.com",
				Trace:       false,
			},
			Checks: []*dsr3.CheckRequest{
				{
					Relation: "owner",
				},
				{
					Relation: "can_read",
				},
				{
					Relation: "can_write",
				},
				{
					Relation: "can_share",
				},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   true,
					Context: tc.SetContext(prop.Reason, ""),
				},
				{
					Check:   true,
					Context: tc.SetContext(prop.Reason, ""),
				},
				{
					Check:   true,
					Context: tc.SetContext(prop.Reason, ""),
				},
				{
					Check:   true,
					Context: tc.SetContext(prop.Reason, ""),
				},
			},
		},
		err: nil,
	},
	// id = 13 - default checks request, no fields set.
	{
		req:  &dsr3.ChecksRequest{},
		resp: &dsr3.ChecksResponse{},
		err:  nil,
	},
	// id = 14 - default checks request, with empty "default" field.
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{},
		},
		resp: &dsr3.ChecksResponse{},
		err:  nil,
	},
	// id 15 - default checks request, with empty "checks" field.
	{
		req: &dsr3.ChecksRequest{
			Checks: []*dsr3.CheckRequest{},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{},
		},
		err: nil,
	},
	// id = 16 - default checks request, with empty "checks" field.
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{},
			Checks:  []*dsr3.CheckRequest{{}},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check:   false,
					Context: tc.SetContext(prop.Reason, "E20046 invalid argument object identifier: type: object_type"),
				},
			},
		},
		err: nil,
	},
	// id = 17 - default sub-request is nil
	{
		req: &dsr3.ChecksRequest{
			Checks: []*dsr3.CheckRequest{
				{
					ObjectType:  "folder",
					ObjectId:    "morty",
					Relation:    "owner",
					SubjectType: "user",
					SubjectId:   "morty@the-citadel.com",
					Trace:       false,
				},
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{
				{
					Check: true,
				},
			},
		},
		err: nil,
	},
	// id = 18 - checks array in nil
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{
				ObjectType:  "folder",
				ObjectId:    "morty",
				Relation:    "owner",
				SubjectType: "user",
				SubjectId:   "morty@the-citadel.com",
				Trace:       false,
			},
		},
		resp: &dsr3.ChecksResponse{
			Checks: []*dsr3.CheckResponse{},
		},
		err: nil,
	},
	// id = 19 - checks array is empty
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{
				ObjectType:  "folder",
				ObjectId:    "morty",
				Relation:    "owner",
				SubjectType: "user",
				SubjectId:   "morty@the-citadel.com",
				Trace:       false,
			},
			Checks: []*dsr3.CheckRequest{},
		},
		resp: &dsr3.ChecksResponse{},
		err:  nil,
	},
	// id = 20 - default + checks == nil
	{
		req: &dsr3.ChecksRequest{
			Default: nil,
			Checks:  nil,
		},
		resp: &dsr3.ChecksResponse{},
		err:  nil,
	},
	// id = 20 - default = empty, checks = nil
	{
		req: &dsr3.ChecksRequest{
			Default: &dsr3.CheckRequest{},
			Checks:  nil,
		},
		resp: &dsr3.ChecksResponse{},
		err:  nil,
	},
	// id = 20 - default = nil;, checks = empty
	{
		req: &dsr3.ChecksRequest{
			Default: nil,
			Checks:  []*dsr3.CheckRequest{},
		},
		resp: &dsr3.ChecksResponse{},
		err:  nil,
	},
}
