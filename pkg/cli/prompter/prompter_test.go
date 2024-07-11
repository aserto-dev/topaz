package prompter_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/prompter"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPrompter(t *testing.T) {
	if enabled, err := strconv.ParseBool(os.Getenv("TEST_INTERACTIVE")); err != nil || !enabled {
		t.Skip("skip interactive tests")
	}

	reqs := []proto.Message{
		&reader.GetObjectRequest{
			Page: &common.PaginationRequest{
				Size:  100,
				Token: "",
			},
		},
		&reader.GetObjectsRequest{
			Page: &common.PaginationRequest{
				Size:  100,
				Token: "",
			},
		},
		&writer.SetObjectRequest{
			Object: &common.Object{
				Properties: &structpb.Struct{},
				CreatedAt:  &timestamppb.Timestamp{},
				UpdatedAt:  &timestamppb.Timestamp{},
			},
		},
		&writer.DeleteObjectRequest{},
		&reader.GetRelationRequest{},
		&reader.GetRelationsRequest{
			Page: &common.PaginationRequest{
				Size:  100,
				Token: "",
			},
		},
		&writer.SetRelationRequest{
			Relation: &common.Relation{},
		},
		&writer.DeleteRelationRequest{},

		&reader.CheckRequest{},
		&reader.GetGraphRequest{},
	}

	for _, req := range reqs {
		form := prompter.New(req)
		_ = form.Show()

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
