package ds_test

import (
	"context"
	"fmt"
	"testing"

	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/prop"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"
	"github.com/stretchr/testify/require"
)

type checkTestCase struct {
	req  *dsr3.CheckRequest
	resp *dsr3.CheckResponse
	err  error
}

func testCheck(ctx context.Context, dsClient dsr3.ReaderClient) func(*testing.T) {
	return func(t *testing.T) {
		for i, tc := range checkTCs {
			t.Run(fmt.Sprintf("check-%d", i), func(t *testing.T) {
				resp, err := dsClient.Check(ctx, tc.req)
				if tc.err == nil {
					require.NoError(t, err)
				} else {
					require.ErrorIs(t, tc.err, err)
				}

				require.NotNil(t, resp, "response should not be nil")

				require.Equal(t, tc.resp.GetCheck(), resp.GetCheck())

				if tc.resp.Context == nil {
					return
				}

				if tc.resp.Context.Fields == nil {
					return
				}

				if v, ok := tc.resp.Context.Fields["reason"]; ok {
					require.Equal(t, v.GetStringValue(), resp.Context.Fields["reason"].GetStringValue())
				}
			})
		}
	}
}

var checkTCs []*checkTestCase = []*checkTestCase{
	{
		req: &dsr3.CheckRequest{
			ObjectType:  "folder",
			ObjectId:    "morty",
			Relation:    "owner",
			SubjectType: "user",
			SubjectId:   "morty@the-citadel.com",
			Trace:       false,
		},
		resp: &dsr3.CheckResponse{
			Check:   true,
			Context: tc.SetContext(prop.Reason, ""),
		},
		err: nil,
	},
	{
		req: &dsr3.CheckRequest{
			ObjectType:  "folder1",
			ObjectId:    "morty",
			Relation:    "owner",
			SubjectType: "user",
			SubjectId:   "morty@the-citadel.com",
			Trace:       false,
		},
		resp: &dsr3.CheckResponse{
			Check:   false,
			Context: tc.SetContext(prop.Reason, "E20026 object type not found: folder1"),
		},
		err: nil,
	},
	{
		req: &dsr3.CheckRequest{
			ObjectType:  "folder",
			ObjectId:    "morty1",
			Relation:    "owner",
			SubjectType: "user",
			SubjectId:   "morty@the-citadel.com",
			Trace:       false,
		},
		resp: &dsr3.CheckResponse{
			Check:   false,
			Context: tc.SetContext(prop.Reason, "E20025 object not found: object folder:morty1"),
		},
		err: nil,
	},
	{
		req: &dsr3.CheckRequest{
			ObjectType:  "folder",
			ObjectId:    "morty",
			Relation:    "owner1",
			SubjectType: "user",
			SubjectId:   "morty@the-citadel.com",
			Trace:       false,
		},
		resp: &dsr3.CheckResponse{
			Check:   false,
			Context: tc.SetContext(prop.Reason, "E20035 relation not found: relation: folder#owner1"),
		},
		err: nil,
	},
	{
		req: &dsr3.CheckRequest{
			ObjectType:  "folder",
			ObjectId:    "morty",
			Relation:    "owner",
			SubjectType: "user1",
			SubjectId:   "morty@the-citadel.com",
			Trace:       false,
		},
		resp: &dsr3.CheckResponse{
			Check:   false,
			Context: tc.SetContext(prop.Reason, "E20026 object type not found: user1"),
		},
		err: nil,
	},
	{
		req: &dsr3.CheckRequest{
			ObjectType:  "folder",
			ObjectId:    "morty",
			Relation:    "owner",
			SubjectType: "user",
			SubjectId:   "morty@the-citadel.com1",
			Trace:       false,
		},
		resp: &dsr3.CheckResponse{
			Check:   false,
			Context: tc.SetContext(prop.Reason, "E20025 object not found: subject user:morty@the-citadel.com1"),
		},
		err: nil,
	},
}
