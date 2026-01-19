package tests_test

import (
	"testing"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/go-directory/pkg/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestObjects(t *testing.T) {
	tcs := []*TestCase{}

	tcs = append(tcs, objectTestCasesWithID...)
	tcs = append(tcs, objectTestCasesWithoutID...)
	tcs = append(tcs, objectTestCasesStreamMode...)

	testRunner(t, tcs)
}

var objectTestCasesWithID = []*TestCase{
	{
		Name: "create test-obj-1",
		Req: &dsw.SetObjectRequest{
			Object: &dsc.Object{
				Type:        "user",
				Id:          "test-user@acmecorp.com",
				DisplayName: "test obj 1",
				Properties:  pb.NewStruct(),
				Etag:        "",
			},
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsw.SetObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
				t.Logf("resp etag:%s", resp.GetResult().GetEtag())

				assert.Equal(t, "user", resp.GetResult().GetType())
				assert.Equal(t, "test-user@acmecorp.com", resp.GetResult().GetId())
				assert.Equal(t, "test obj 1", resp.GetResult().GetDisplayName())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.Empty(t, resp.GetResult().GetProperties().GetFields())
				assert.NotEmpty(t, resp.GetResult().GetEtag())
				assert.Equal(t, "3016620182482667549", resp.GetResult().GetEtag())
			}
			return func(proto.Message) {}
		},
	},
	{
		Name: "get test-obj-1",
		Req: &dsr.GetObjectRequest{
			ObjectType: "user",
			ObjectId:   "test-user@acmecorp.com",
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsr.GetObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
				t.Logf("resp etag:%s", resp.GetResult().GetEtag())

				assert.Equal(t, "user", resp.GetResult().GetType())
				assert.Equal(t, "test-user@acmecorp.com", resp.GetResult().GetId())
				assert.Equal(t, "test obj 1", resp.GetResult().GetDisplayName())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.Empty(t, resp.GetResult().GetProperties().GetFields())
				assert.NotEmpty(t, resp.GetResult().GetEtag())
				assert.Equal(t, "3016620182482667549", resp.GetResult().GetEtag())
			}
			return func(req proto.Message) {}
		},
	},
	{
		Name: "update test-obj-1",
		Req: &dsw.SetObjectRequest{
			Object: &dsc.Object{
				Type:        "user",
				Id:          "test-user-11@acmecorp.com",
				DisplayName: "test obj 11",
				Etag:        "3016620182482667549",
			},
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsw.SetObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
				t.Logf("resp etag:%s", resp.GetResult().GetEtag())

				assert.Equal(t, "user", resp.GetResult().GetType())
				assert.Equal(t, "test-user-11@acmecorp.com", resp.GetResult().GetId())
				assert.Equal(t, "test obj 11", resp.GetResult().GetDisplayName())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.Empty(t, resp.GetResult().GetProperties().GetFields())
				assert.NotEmpty(t, resp.GetResult().GetEtag())
				assert.Equal(t, "2708540687187161441", resp.GetResult().GetEtag())
			}
			return func(req proto.Message) {}
		},
	},
	{
		Name: "get updated test-obj-11",
		Req: &dsr.GetObjectRequest{
			ObjectType: "user",
			ObjectId:   "test-user-11@acmecorp.com",
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsr.GetObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
				t.Logf("resp etag:%s", resp.GetResult().GetEtag())

				assert.Equal(t, "user", resp.GetResult().GetType())
				assert.Equal(t, "test-user-11@acmecorp.com", resp.GetResult().GetId())
				assert.Equal(t, "test obj 11", resp.GetResult().GetDisplayName())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.Empty(t, resp.GetResult().GetProperties().GetFields())
				assert.NotEmpty(t, resp.GetResult().GetEtag())
				assert.Equal(t, "2708540687187161441", resp.GetResult().GetEtag())
			}
			return func(req proto.Message) {}
		},
	},
	{
		Name: "delete test-obj-11",
		Req: &dsw.DeleteObjectRequest{
			ObjectType: "user",
			ObjectId:   "test-user-11@acmecorp.com",
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsw.DeleteObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
			}
			return func(req proto.Message) {}
		},
	},
	{
		Name: "get deleted test-obj-11",
		Req: &dsr.GetObjectRequest{
			ObjectType: "user",
			ObjectId:   "test-user-11@acmecorp.com",
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.Error(t, tErr)
			assert.Contains(t, tErr.Error(), "key not found")
			assert.Nil(t, msg)
			return func(req proto.Message) {}
		},
	},
	{
		Name: "delete deleted test-obj-11 by id",
		Req: &dsw.DeleteObjectRequest{
			ObjectType: "user",
			ObjectId:   "test-user-11@acmecorp.com",
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsw.DeleteObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
			}
			return func(req proto.Message) {}
		},
	},
}

// create object without id.
var objectTestCasesWithoutID = []*TestCase{
	{
		Name: "create test-obj-2 with no-id",
		Req: &dsw.SetObjectRequest{
			Object: &dsc.Object{
				Type:        "user",
				Id:          "test-user-2@acmecorp.com",
				DisplayName: "test obj 2",
				Properties:  pb.NewStruct(),
				Etag:        "",
			},
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsw.SetObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
				t.Logf("resp etag:%s", resp.GetResult().GetEtag())

				assert.Equal(t, "user", resp.GetResult().GetType())
				assert.Equal(t, "test-user-2@acmecorp.com", resp.GetResult().GetId())
				assert.Equal(t, "test obj 2", resp.GetResult().GetDisplayName())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.Empty(t, resp.GetResult().GetProperties().GetFields())
				assert.NotEmpty(t, resp.GetResult().GetEtag())
				assert.Greater(t, len(resp.GetResult().GetEtag()), 4)

				return func(req proto.Message) {
					lastHash := resp.GetResult().GetEtag()

					switch r := req.(type) {
					case *dsw.SetObjectRequest:
						r.Object.Etag = lastHash
					}
					t.Logf("propagated hash:%s", lastHash)
				}
			}
			return func(req proto.Message) {}
		},
	},
	{
		Name: "get test-obj-2",
		Req: &dsr.GetObjectRequest{
			ObjectType: "user",
			ObjectId:   "test-user-2@acmecorp.com",
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsr.GetObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
				t.Logf("resp etag:%s", resp.GetResult().GetEtag())

				assert.Equal(t, "user", resp.GetResult().GetType())
				assert.Equal(t, "test-user-2@acmecorp.com", resp.GetResult().GetId())
				assert.Equal(t, "test obj 2", resp.GetResult().GetDisplayName())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.Empty(t, resp.GetResult().GetProperties().GetFields())
				assert.NotEmpty(t, resp.GetResult().GetEtag())
				assert.Greater(t, len(resp.GetResult().GetEtag()), 4)

				return func(req proto.Message) {
					lastHash := resp.GetResult().GetEtag()

					switch r := req.(type) {
					case *dsw.SetObjectRequest:
						r.Object.Etag = lastHash
					}
					t.Logf("propagated hash:%s", lastHash)
				}
			}
			return func(req proto.Message) {}
		},
	},
	{
		Name: "update test-obj-2",
		Req: &dsw.SetObjectRequest{
			Object: &dsc.Object{
				Type:        "user",
				Id:          "test-user-2@acmecorp.com",
				DisplayName: "test obj 22",
			},
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsw.SetObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
				t.Logf("resp etag:%s", resp.GetResult().GetEtag())

				assert.Equal(t, "user", resp.GetResult().GetType())
				assert.Equal(t, "test-user-2@acmecorp.com", resp.GetResult().GetId())
				assert.Equal(t, "test obj 22", resp.GetResult().GetDisplayName())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.Empty(t, resp.GetResult().GetProperties().GetFields())
				assert.NotEmpty(t, resp.GetResult().GetEtag())
				assert.Greater(t, len(resp.GetResult().GetEtag()), 4)
			}
			return func(req proto.Message) {}
		},
	},
	{
		Name: "get updated test-obj-2",
		Req: &dsr.GetObjectRequest{
			ObjectType: "user",
			ObjectId:   "test-user-2@acmecorp.com",
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsr.GetObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
				t.Logf("resp etag:%s", resp.GetResult().GetEtag())

				assert.Equal(t, "user", resp.GetResult().GetType())
				assert.Equal(t, "test-user-2@acmecorp.com", resp.GetResult().GetId())
				assert.Equal(t, "test obj 22", resp.GetResult().GetDisplayName())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.Empty(t, resp.GetResult().GetProperties().GetFields())
				assert.NotEmpty(t, resp.GetResult().GetEtag())
				assert.Greater(t, len(resp.GetResult().GetEtag()), 4)
			}
			return func(req proto.Message) {}
		},
	},
	{
		Name: "delete test-obj-2",
		Req: &dsw.DeleteObjectRequest{
			ObjectType: "user",
			ObjectId:   "test-user-2@acmecorp.com",
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.NotNil(t, msg)
			switch resp := msg.(type) {
			case *dsw.DeleteObjectResponse:
				require.NoError(t, tErr)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetResult())
			}
			return func(req proto.Message) {}
		},
	},
	{
		Name: "get deleted test-obj-2",
		Req: &dsr.GetObjectRequest{
			ObjectType: "user",
			ObjectId:   "test-user-2@acmecorp.com",
		},
		Checks: func(t *testing.T, msg proto.Message, tErr error) func(proto.Message) {
			require.Error(t, tErr)
			assert.Contains(t, tErr.Error(), "key not found")
			assert.Nil(t, msg)
			return func(req proto.Message) {}
		},
	},
}

var objectTestCasesStreamMode = []*TestCase{}
