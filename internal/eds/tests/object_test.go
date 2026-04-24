package tests_test

import (
	"testing"

	"github.com/aserto-dev/topaz/api/directory/pkg/pb"
	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
	dsw "github.com/aserto-dev/topaz/api/directory/v4/writer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestObjects(t *testing.T) {
	tcs := make([]*TestCase, 0, len(objectTestCasesWithID)+len(objectTestCasesWithoutID)+len(objectTestCasesStreamMode))

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
				ObjectType: "user",
				ObjectId:   "test-user@acmecorp.com",
				Properties: setProperty(nil, propDisplayName, "test obj 1"),
				Etag:       "",
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

				assert.Equal(t, "user", resp.GetResult().GetObjectType())
				assert.Equal(t, "test-user@acmecorp.com", resp.GetResult().GetObjectId())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.NotEmpty(t, resp.GetResult().GetProperties().GetFields())
				assert.Equal(t, "test obj 1", getProperty(resp.GetResult().GetProperties(), propDisplayName)) //nolint:staticcheck
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

				assert.Equal(t, "user", resp.GetResult().GetObjectType())
				assert.Equal(t, "test-user@acmecorp.com", resp.GetResult().GetObjectId())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.NotEmpty(t, resp.GetResult().GetProperties().GetFields())
				assert.Equal(t, "test obj 1", getProperty(resp.GetResult().GetProperties(), propDisplayName)) //nolint:staticcheck
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
				ObjectType: "user",
				ObjectId:   "test-user-11@acmecorp.com",
				Properties: setProperty(nil, propDisplayName, "test obj 11"),
				Etag:       "3016620182482667549",
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

				assert.Equal(t, "user", resp.GetResult().GetObjectType())
				assert.Equal(t, "test-user-11@acmecorp.com", resp.GetResult().GetObjectId())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.NotEmpty(t, resp.GetResult().GetProperties().GetFields())
				assert.Equal(t, "test obj 11", getProperty(resp.GetResult().GetProperties(), propDisplayName)) //nolint:staticcheck
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

				assert.Equal(t, "user", resp.GetResult().GetObjectType())
				assert.Equal(t, "test-user-11@acmecorp.com", resp.GetResult().GetObjectId())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.NotEmpty(t, resp.GetResult().GetProperties().GetFields())
				assert.Equal(t, "test obj 11", getProperty(resp.GetResult().GetProperties(), propDisplayName)) //nolint:staticcheck
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
				ObjectType: "user",
				ObjectId:   "test-user-2@acmecorp.com",
				Properties: setProperty(pb.NewStruct(), propDisplayName, "test obj 2"),
				Etag:       "",
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

				assert.Equal(t, "user", resp.GetResult().GetObjectType())
				assert.Equal(t, "test-user-2@acmecorp.com", resp.GetResult().GetObjectId())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.NotEmpty(t, resp.GetResult().GetProperties().GetFields())
				assert.Equal(t, "test obj 2", getProperty(resp.GetResult().GetProperties(), propDisplayName))

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

				assert.Equal(t, "user", resp.GetResult().GetObjectType())
				assert.Equal(t, "test-user-2@acmecorp.com", resp.GetResult().GetObjectId())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.NotEmpty(t, resp.GetResult().GetProperties().GetFields())
				assert.Equal(t, "test obj 2", getProperty(resp.GetResult().GetProperties(), propDisplayName))
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
				ObjectType: "user",
				ObjectId:   "test-user-2@acmecorp.com",
				Properties: setProperty(nil, propDisplayName, "test obj 22"),
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

				assert.Equal(t, "user", resp.GetResult().GetObjectType())
				assert.Equal(t, "test-user-2@acmecorp.com", resp.GetResult().GetObjectId())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.NotEmpty(t, resp.GetResult().GetProperties().GetFields())
				assert.Equal(t, "test obj 22", getProperty(resp.GetResult().GetProperties(), propDisplayName)) //nolint:staticcheck

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

				assert.Equal(t, "user", resp.GetResult().GetObjectType())
				assert.Equal(t, "test-user-2@acmecorp.com", resp.GetResult().GetObjectId())
				assert.NotNil(t, resp.GetResult().GetProperties())
				assert.NotEmpty(t, resp.GetResult().GetProperties().GetFields())
				assert.Equal(t, "test obj 22", getProperty(resp.GetResult().GetProperties(), propDisplayName)) //nolint:staticcheck

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

const propDisplayName string = "display_name"

func setProperty(s *structpb.Struct, k, v string) *structpb.Struct {
	if s == nil {
		s = pb.NewStruct()
	}

	s.Fields[k] = structpb.NewStringValue(v)

	return s
}

func getProperty(s *structpb.Struct, k string) string {
	if s == nil {
		return ""
	}

	return s.Fields[k].GetStringValue()
}
