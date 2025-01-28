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

				for i := 0; i < len(tc.resp.Checks); i++ {
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
					Context: tc.SetContext(prop.Reason, "E20025 object not found: E20026 object type not found: folder2: folder2: object_type"),
				},
			},
		},
		err: nil,
	},
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
					Context: tc.SetContext(prop.Reason, "E20035 relation not found: relation: folder#owner2"),
				},
			},
		},
		err: nil,
	},
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
					Context: tc.SetContext(prop.Reason, "E20025 object not found: E20026 object type not found: user2: user2: subject_type"),
				},
			},
		},
		err: nil,
	},
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
					Context: tc.SetContext(prop.Reason, "E20025 object not found: E20026 object type not found: folder2: folder2: object_type"),
				},
			},
		},
		err: nil,
	},
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
					Context: tc.SetContext(prop.Reason, "E20035 relation not found: relation: folder#owner2"),
				},
			},
		},
		err: nil,
	},
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
					Context: tc.SetContext(prop.Reason, "E20025 object not found: E20026 object type not found: user2: user2: subject_type"),
				},
			},
		},
		err: nil,
	},
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
}
