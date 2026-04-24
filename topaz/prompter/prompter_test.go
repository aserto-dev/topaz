package prompter_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	common "github.com/aserto-dev/topaz/api/directory/v4"
	"github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/api/directory/v4/writer"
	"github.com/aserto-dev/topaz/topaz/prompter"
	"github.com/aserto-dev/topaz/topaz/x"
	"github.com/authzen/access.go/api/access/v1"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPrompter(t *testing.T) {
	if enabled, err := strconv.ParseBool(os.Getenv("TEST_INTERACTIVE")); err != nil || !enabled {
		t.Skip("skip interactive tests")
	}

	reqs := directoryRequests()
	reqs = append(reqs, accessRequests()...)

	for _, req := range reqs {
		form := prompter.New(req)
		require.NotNil(t, form)

		if err := form.Show(); err != nil {
			require.ErrorIs(t, err, prompter.ErrCancelled)
		}

		fmt.Fprint(os.Stderr, protojson.MarshalOptions{
			Multiline:         true,
			Indent:            "  ",
			AllowPartial:      true,
			UseProtoNames:     true,
			UseEnumNumbers:    false,
			EmitUnpopulated:   true,
			EmitDefaultValues: true,
		}.Format(form.Req()))
	}
}

func directoryRequests() []proto.Message {
	reqs := []proto.Message{
		&reader.GetObjectRequest{},
		&reader.ListObjectsRequest{
			Page: &common.PaginationRequest{
				Size:  x.MaxPaginationSize,
				Token: "",
			},
		},
		&writer.SetObjectRequest{
			Object: &common.Object{
				Properties: &structpb.Struct{},
				UpdatedAt:  &timestamppb.Timestamp{},
			},
		},
		&writer.DeleteObjectRequest{},
		&reader.GetRelationRequest{},
		&reader.ListRelationsRequest{
			Page: &common.PaginationRequest{
				Size:  x.MaxPaginationSize,
				Token: "",
			},
		},
		&writer.SetRelationRequest{
			Relation: &common.Relation{},
		},
		&writer.DeleteRelationRequest{},
		&reader.CheckRequest{},
		&reader.ChecksRequest{
			Default: &reader.CheckRequest{},
			Checks:  []*reader.CheckRequest{},
		},
		&reader.GraphRequest{},
	}

	return reqs
}

//nolint:funlen
func accessRequests() []proto.Message {
	reqs := []proto.Message{
		&access.EvaluationRequest{
			Subject: &access.Subject{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Action: &access.Action{
				Name:       "",
				Properties: &structpb.Struct{},
			},
			Resource: &access.Resource{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Context: &structpb.Struct{},
		},
		&access.EvaluationsRequest{
			Subject: &access.Subject{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Action: &access.Action{
				Name:       "",
				Properties: &structpb.Struct{},
			},
			Resource: &access.Resource{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Context: &structpb.Struct{},
			Evaluations: []*access.EvaluationRequest{
				{
					Subject: &access.Subject{
						Type:       "",
						Id:         "",
						Properties: &structpb.Struct{},
					},
					Action: &access.Action{
						Name:       "",
						Properties: &structpb.Struct{},
					},
					Resource: &access.Resource{
						Type:       "",
						Id:         "",
						Properties: &structpb.Struct{},
					},
					Context: &structpb.Struct{},
				},
			},
			Options: &structpb.Struct{},
		},
		&access.ActionSearchRequest{
			Subject: &access.Subject{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Action: &access.Action{
				Name:       "",
				Properties: &structpb.Struct{},
			},
			Resource: &access.Resource{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Context: &structpb.Struct{},
			Page: &access.PaginationRequest{
				Token:      nil,
				Limit:      nil,
				Properties: nil,
			},
		},
		&access.ResourceSearchRequest{
			Subject: &access.Subject{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Action: &access.Action{
				Name:       "",
				Properties: &structpb.Struct{},
			},
			Resource: &access.Resource{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Context: &structpb.Struct{},
			Page: &access.PaginationRequest{
				Token:      nil,
				Limit:      nil,
				Properties: nil,
			},
		},
		&access.SubjectSearchRequest{
			Subject: &access.Subject{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Action: &access.Action{
				Name:       "",
				Properties: &structpb.Struct{},
			},
			Resource: &access.Resource{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Context: &structpb.Struct{},
			Page: &access.PaginationRequest{
				Token:      nil,
				Limit:      nil,
				Properties: nil,
			},
		},
	}

	return reqs
}
