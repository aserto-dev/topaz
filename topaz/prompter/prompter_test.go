package prompter_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/topaz/prompter"
	"github.com/aserto-dev/topaz/topaz/x"
	dsa "github.com/authzen/access.go/api/access/v1"

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
		&dsr.GetObjectRequest{},
		&dsr.GetObjectsRequest{
			Page: &dsc.PaginationRequest{
				Size:  x.MaxPaginationSize,
				Token: "",
			},
		},
		&dsw.SetObjectRequest{
			Object: &dsc.Object{
				Properties: &structpb.Struct{},
				CreatedAt:  &timestamppb.Timestamp{},
				UpdatedAt:  &timestamppb.Timestamp{},
			},
		},
		&dsw.DeleteObjectRequest{},
		&dsr.GetRelationRequest{},
		&dsr.GetRelationsRequest{
			Page: &dsc.PaginationRequest{
				Size:  x.MaxPaginationSize,
				Token: "",
			},
		},
		&dsw.SetRelationRequest{
			Relation: &dsc.Relation{},
		},
		&dsw.DeleteRelationRequest{},
		&dsr.CheckRequest{},
		&dsr.ChecksRequest{
			Default: &dsr.CheckRequest{},
			Checks:  []*dsr.CheckRequest{},
		},
		&dsr.GetGraphRequest{},
	}

	return reqs
}

//nolint:funlen
func accessRequests() []proto.Message {
	reqs := []proto.Message{
		&dsa.EvaluationRequest{
			Subject: &dsa.Subject{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Action: &dsa.Action{
				Name:       "",
				Properties: &structpb.Struct{},
			},
			Resource: &dsa.Resource{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Context: &structpb.Struct{},
		},
		&dsa.EvaluationsRequest{
			Subject: &dsa.Subject{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Action: &dsa.Action{
				Name:       "",
				Properties: &structpb.Struct{},
			},
			Resource: &dsa.Resource{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Context: &structpb.Struct{},
			Evaluations: []*dsa.EvaluationRequest{
				{
					Subject: &dsa.Subject{
						Type:       "",
						Id:         "",
						Properties: &structpb.Struct{},
					},
					Action: &dsa.Action{
						Name:       "",
						Properties: &structpb.Struct{},
					},
					Resource: &dsa.Resource{
						Type:       "",
						Id:         "",
						Properties: &structpb.Struct{},
					},
					Context: &structpb.Struct{},
				},
			},
			Options: &structpb.Struct{},
		},
		&dsa.ActionSearchRequest{
			Subject: &dsa.Subject{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Action: &dsa.Action{
				Name:       "",
				Properties: &structpb.Struct{},
			},
			Resource: &dsa.Resource{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Context: &structpb.Struct{},
			Page: &dsa.PaginationRequest{
				Token:      nil,
				Limit:      nil,
				Properties: nil,
			},
		},
		&dsa.ResourceSearchRequest{
			Subject: &dsa.Subject{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Action: &dsa.Action{
				Name:       "",
				Properties: &structpb.Struct{},
			},
			Resource: &dsa.Resource{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Context: &structpb.Struct{},
			Page: &dsa.PaginationRequest{
				Token:      nil,
				Limit:      nil,
				Properties: nil,
			},
		},
		&dsa.SubjectSearchRequest{
			Subject: &dsa.Subject{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Action: &dsa.Action{
				Name:       "",
				Properties: &structpb.Struct{},
			},
			Resource: &dsa.Resource{
				Type:       "",
				Id:         "",
				Properties: &structpb.Struct{},
			},
			Context: &structpb.Struct{},
			Page: &dsa.PaginationRequest{
				Token:      nil,
				Limit:      nil,
				Properties: nil,
			},
		},
	}

	return reqs
}
