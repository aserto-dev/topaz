package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	az2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-directory-cli/client"
	dsr2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"

	"github.com/fatih/color"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	checkRelation   string = "check_relation"
	checkPermission string = "check_permission"
	checkDecision   string = "check_decision"
	expected        string = "expected"
	passed          string = "PASS"
	failed          string = "FAIL"
)

type TestCmd struct {
	Exec     TestExecCmd     `cmd:"" help:"execute assertions"`
	Template TestTemplateCmd `cmd:"" help:"output assertions template"`
}

type TestExecCmd struct {
	File    string `arg:""  default:"assertions.json" help:"filepath to assertions file"`
	NoColor bool   `flag:"" default:"false" help:"disable colorized output"`
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

	dsc, err := clients.NewDirectoryClient(c, &cmd.Config)
	if err != nil {
		return err
	}

	azc, err := clients.NewAuthorizerClient(c, &clients.AuthorizerConfig{
		Host:     iff(os.Getenv(clients.EnvTopazAuthorizerSvc) != "", os.Getenv(clients.EnvTopazAuthorizerSvc), ""),
		APIKey:   iff(os.Getenv(clients.EnvTopazAuthorizerKey) != "", os.Getenv(clients.EnvTopazAuthorizerKey), ""),
		Insecure: cmd.Config.Insecure,
		TenantID: cmd.Config.TenantID,
	})

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

	if cmd.NoColor {
		color.NoColor = true
	}

	for i := 0; i < len(assertions.Assertions); i++ {
		var msg structpb.Struct
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(assertions.Assertions[i], &msg)
		if err != nil {
			return err
		}

		expected := msg.Fields[expected].GetBoolValue()

		if field, ok := msg.Fields[checkRelation]; ok {
			if err := cmd.execCheckRelation(c, dsc, field, i, expected); err != nil {
				return err
			}
		}

		if field, ok := msg.Fields[checkPermission]; ok {
			if err := cmd.execCheckPermission(c, dsc, field, i, expected); err != nil {
				return err
			}
		}

		if field, ok := msg.Fields[checkDecision]; ok {
			if err := cmd.execCheckDecision(c, azc, field, i, expected); err != nil {
				return err
			}
		}
	}

	return nil
}

func (*TestExecCmd) execCheckRelation(c *cc.CommonCtx, dsc *client.Client, field *structpb.Value, i int, expected bool) error {
	var req dsr2.CheckRelationRequest
	if err := unmarshalReq(field, &req); err != nil {
		return err
	}

	if req.Relation.GetObjectType() == "" {
		req.Relation.ObjectType = req.Object.Type
	}

	start := time.Now()
	resp, err := dsc.Reader.CheckRelation(c.Context, &req)
	if err != nil {
		return err
	}
	duration := time.Since(start)
	outcome := resp.GetCheck()

	fmt.Printf("%04d %s %v  %s [%s] (%s)\n",
		i+1,
		"check-relation  ",
		iff(expected == outcome, color.GreenString(passed), color.RedString(failed)),
		checkRelationString(&req),
		iff(outcome, color.BlueString("%t", outcome), color.YellowString("%t", outcome)),
		duration,
	)

	return nil
}

func (*TestExecCmd) execCheckPermission(c *cc.CommonCtx, dsc *client.Client, field *structpb.Value, i int, expected bool) error {
	var req dsr2.CheckPermissionRequest
	if err := unmarshalReq(field, &req); err != nil {
		return err
	}

	start := time.Now()
	resp, err := dsc.Reader.CheckPermission(c.Context, &req)
	if err != nil {
		return err
	}
	duration := time.Since(start)
	outcome := resp.GetCheck()

	fmt.Printf("%04d %s %v  %s [%s] (%s)\n",
		i+1,
		"check-permission",
		iff(expected == resp.GetCheck(), color.GreenString(passed), color.RedString(failed)),
		checkPermissionString(&req),
		iff(outcome, color.BlueString("%t", outcome), color.YellowString("%t", outcome)),
		duration,
	)

	return nil
}

func (*TestExecCmd) execCheckDecision(c *cc.CommonCtx, azc az2.AuthorizerClient, field *structpb.Value, i int, expected bool) error {
	var req az2.IsRequest
	if err := unmarshalReq(field, &req); err != nil {
		return err
	}

	start := time.Now()
	resp, err := azc.Is(c.Context, &req)
	if err != nil {
		return err
	}
	duration := time.Since(start)
	decision := resp.Decisions[0]

	fmt.Printf("%04d %s %v  %s [%s] (%s)\n",
		i+1,
		"check-decision  ",
		iff(expected == decision.GetIs(), color.GreenString(passed), color.RedString(failed)),
		checkDecisionString(&req),
		iff(decision.GetIs(), color.BlueString("%t", decision.GetIs()), color.YellowString("%t", decision.GetIs())),
		duration,
	)

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

func checkDecisionString(req *az2.IsRequest) string {
	return fmt.Sprintf("%s/%s:%s",
		req.PolicyContext.GetPath(),
		req.PolicyContext.GetDecisions()[0],
		req.IdentityContext.Identity,
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
    {"check_permission":{"subject":{"type":"","key":""},"permission":{"name":""},"object":{"type":"","key":""}},"expected":false},

    {"check_decision":{"identity_context":{"identity":"","type":""},"resource_context":{},"policy_context":{"path":"","decisions":[""]}},"expected":true},
    {"check_decision":{"identity_context":{"identity":"","type":""},"resource_context":{},"policy_context":{"path":"","decisions":[""]}},"expected":false}
  ]
}`
