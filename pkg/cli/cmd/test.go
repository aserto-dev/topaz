package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	dsr2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	checkRelation   string = "check_relation"
	checkPermission string = "check_permission"
	expected        string = "expected"
	passed          string = "PASSED"
	failed          string = "FAILED"
)

type TestCmd struct {
	Exec     TestExecCmd     `cmd:"" help:"execute assertions"`
	Template TestTemplateCmd `cmd:"" help:"output assertions template"`
}

type TestExecCmd struct {
	File string `arg:""  default:"assertions.json" help:"filepath to assertions file"`
	clients.Config
}

type TestTemplateCmd struct {
	Pretty bool `arg:"" default:"false" help:"pretty print JSON"`
}

func (cmd *TestExecCmd) Run(c *cc.CommonCtx) error {
	r, err := os.Open(cmd.File)
	if err != nil {
		return err
	}
	defer r.Close()

	dsClient, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	var assertions struct {
		Assertions []json.RawMessage `json:"assertions"`
	}

	dec := json.NewDecoder(r)
	if err := dec.Decode(&assertions); err != nil {
		return err
	}

	for i := 0; i < len(assertions.Assertions); i++ {
		var msg structpb.Struct
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(assertions.Assertions[i], &msg)
		if err != nil {
			return err
		}

		expected := msg.Fields[expected].GetBoolValue()

		if field, ok := msg.Fields[checkRelation]; ok {
			var req dsr2.CheckRelationRequest
			if err := unmarshalReq(field, &req); err != nil {
				return err
			}

			start := time.Now()
			resp, err := dsClient.Reader.CheckRelation(c.Context, &req)
			if err != nil {
				return err
			}
			duration := time.Since(start)

			fmt.Printf("%04d %s %s  %s (%s)\n", i, "check-relation  ", iff(expected == resp.GetCheck(), passed, failed), checkRelationString(&req), duration)
		}

		if field, ok := msg.Fields[checkPermission]; ok {
			var req dsr2.CheckPermissionRequest
			if err := unmarshalReq(field, &req); err != nil {
				return err
			}

			start := time.Now()
			resp, err := dsClient.Reader.CheckPermission(c.Context, &req)
			if err != nil {
				return err
			}
			duration := time.Since(start)

			fmt.Printf("%04d %s %s  %s (%s)\n", i, "check-permission", iff(expected == resp.GetCheck(), passed, failed), checkPermissionString(&req), duration)
		}
	}

	return nil
}

func unmarshalReq(value *structpb.Value, msg proto.Message) error {
	b, err := value.MarshalJSON()
	if err != nil {
		return err
	}

	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(b, msg)
	if err != nil {
		return err
	}

	return nil
}

func ProtoToStr(msg proto.Message) string {
	return protojson.MarshalOptions{
		Multiline:       false,
		Indent:          "  ",
		AllowPartial:    false,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
	}.Format(msg)
}

func iff[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func checkRelationString(req *dsr2.CheckRelationRequest) string {
	return fmt.Sprintf("%s:%s#%s@%s:%s",
		req.Object.GetType(), req.Object.GetKey(),
		req.Relation.GetName(),
		req.Subject.GetType(), req.Subject.GetKey(),
	)
}

func checkPermissionString(req *dsr2.CheckPermissionRequest) string {
	return fmt.Sprintf("%s:%s#%s@%s:%s",
		req.Object.GetType(), req.Object.GetKey(),
		req.Permission.GetName(),
		req.Subject.GetType(), req.Subject.GetKey(),
	)
}

func (cmd *TestTemplateCmd) Run(c *cc.CommonCtx) error {
	if !cmd.Pretty {
		fmt.Fprintf(c.UI.Output(), "%s\n", assertionsTemplate)
		return nil
	}

	dec := json.NewDecoder(strings.NewReader(assertionsTemplate))

	var template interface{}
	if err := dec.Decode(&template); err != nil {
		return err
	}

	enc := json.NewEncoder(c.UI.Output())
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(template); err != nil {
		return err
	}

	return nil
}

const assertionsTemplate string = `{
  "assertions": [
    {"check_relation":{"subject":{"type":"","key":""},"relation":{"name":""},"object":{"type":"","key":""}},"expected":true},
    {"check_relation":{"subject":{"type":"","key":""},"relation":{"name":""},"object":{"type":"","key":""}},"expected":false},

    {"check_permission":{"subject":{"type":"","key":""},"permission":{"name":""},"object":{"type":"","key":""}},"expected":true},
    {"check_permission":{"subject":{"type":"","key":""},"permission":{"name":""},"object":{"type":"","key":""}},"expected":false}
  ]
}`
