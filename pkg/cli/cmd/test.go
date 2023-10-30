package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	az2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	dsr2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"

	"github.com/aserto-dev/go-directory-cli/client"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"

	"github.com/fatih/color"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	check           string = "check"
	checkRelation   string = "check_relation"
	checkPermission string = "check_permission"
	checkDecision   string = "check_decision"
	expected        string = "expected"
	passed          string = "PASS"
	failed          string = "FAIL"
	errored         string = "ERR "
)

type TestCmd struct {
	Exec     TestExecCmd     `cmd:"" help:"execute assertions"`
	Template TestTemplateCmd `cmd:"" help:"output assertions template"`
}

type TestExecCmd struct {
	File    string `arg:""  default:"assertions.json" help:"filepath to assertions file"`
	NoColor bool   `flag:"" default:"false" help:"disable colorized output"`
	Summary bool   `flag:"" default:"false" help:"display test summary"`
	results *testResults
	clients.Config
}

// nolint: funlen,gocyclo
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

	cmd.results = &testResults{
		total:   int32(len(assertions.Assertions)),
		passed:  0,
		failed:  0,
		errored: 0,
	}

	for i := 0; i < len(assertions.Assertions); i++ {
		var msg structpb.Struct
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(assertions.Assertions[i], &msg)
		if err != nil {
			return err
		}

		expected, ok := getBool(&msg, expected)
		if !ok {
			return fmt.Errorf("no expected outcome of assertion defined")
		}

		checkType := getCheckType(&msg)
		if checkType == CheckUnknown {
			return fmt.Errorf("unknown check type")
		}

		reqVersion := getReqVersion(msg.Fields[checkTypeMapStr[checkType]])
		if reqVersion == 0 {
			return fmt.Errorf("unknown request version")
		}

		var result *checkResult

		switch {
		case checkType == Check && reqVersion == 3:
			result = checkV3(c.Context, dsc, msg.Fields[checkTypeMapStr[checkType]])
		case checkType == CheckPermission && reqVersion == 3:
			result = checkPermissionV3(c.Context, dsc, msg.Fields[checkTypeMapStr[checkType]])
		case checkType == CheckRelation && reqVersion == 3:
			result = checkRelationV3(c.Context, dsc, msg.Fields[checkTypeMapStr[checkType]])
		case checkType == CheckPermission && reqVersion == 2:
			result = checkPermissionV2(c.Context, dsc, msg.Fields[checkTypeMapStr[checkType]])
		case checkType == CheckRelation && reqVersion == 2:
			result = checkRelationV2(c.Context, dsc, msg.Fields[checkTypeMapStr[checkType]])
		case checkType == CheckDecision:
			result = checkDecisionV2(c.Context, azc, msg.Fields[checkTypeMapStr[checkType]])
		}

		cmd.results.Passed(result.Outcome == expected)

		if result.Err != nil {
			cmd.results.IncrErrored()
		}

		fmt.Printf("%04d %-16s %v  %s [%s] (%s)\n",
			i+1,
			checkTypeMapStr[checkType],
			iff(expected == result.Outcome, color.GreenString(passed), color.RedString(failed)),
			result.Str,
			iff(result.Outcome, color.BlueString("%t", result.Outcome), color.YellowString("%t", result.Outcome)),
			result.Duration,
		)
	}

	if cmd.Summary {
		fmt.Print("\nTest Execution Summary:\n")
		fmt.Printf("%s\n", strings.Repeat("-", 23))
		fmt.Printf("total:   %d\n", cmd.results.total)
		fmt.Printf("passed:  %d\n", cmd.results.passed)
		fmt.Printf("failed:  %d\n", cmd.results.failed)
		fmt.Printf("errored: %d\n", cmd.results.errored)
		fmt.Println()
	}

	if cmd.results.errored > 0 || cmd.results.failed > 0 {
		return errors.New("one or more test errored or failed")
	}

	return nil
}

type checkResult struct {
	Outcome  bool
	Duration time.Duration
	Err      error
	Str      string
}

func getBool(msg *structpb.Struct, fieldName string) (bool, bool) {
	v, ok := msg.Fields[fieldName]
	return v.GetBoolValue(), ok
}

type checkType int

const (
	CheckUnknown checkType = iota
	Check
	CheckRelation
	CheckPermission
	CheckDecision
)

var checkTypeMap = map[string]checkType{
	check:           Check,
	checkRelation:   CheckRelation,
	checkPermission: CheckPermission,
	checkDecision:   CheckDecision,
}

var checkTypeMapStr = map[checkType]string{
	Check:           check,
	CheckRelation:   checkRelation,
	CheckPermission: checkPermission,
	CheckDecision:   checkDecision,
}

func getCheckType(msg *structpb.Struct) checkType {
	for k, v := range checkTypeMap {
		if _, ok := msg.Fields[k]; ok {
			return v
		}
	}
	return CheckUnknown
}

func getReqVersion(val *structpb.Value) int {
	if val == nil {
		return 0
	}

	if v, ok := val.Kind.(*structpb.Value_StructValue); ok {
		if _, ok := v.StructValue.Fields["object_type"]; ok {
			return 3
		}
		if _, ok := v.StructValue.Fields["object"]; ok {
			return 2
		}
		if _, ok := v.StructValue.Fields["identity_context"]; ok {
			return 2
		}
	}
	return 0
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

func iff[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func checkV3(ctx context.Context, c *client.Client, msg *structpb.Value) *checkResult {
	var req dsr3.CheckRequest
	if err := unmarshalReq(msg, &req); err != nil {
		return &checkResult{Err: err}
	}

	start := time.Now()

	resp, err := c.Reader.Check(ctx, &req)

	duration := time.Since(start)

	return &checkResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkStringV3(&req),
	}
}

func checkPermissionV3(ctx context.Context, c *client.Client, msg *structpb.Value) *checkResult {
	var req dsr3.CheckPermissionRequest
	if err := unmarshalReq(msg, &req); err != nil {
		return &checkResult{Err: err}
	}

	start := time.Now()

	resp, err := c.Reader.CheckPermission(ctx, &req)

	duration := time.Since(start)

	return &checkResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkPermissionStringV3(&req),
	}
}

func checkRelationV3(ctx context.Context, c *client.Client, msg *structpb.Value) *checkResult {
	var req dsr3.CheckRelationRequest
	if err := unmarshalReq(msg, &req); err != nil {
		return &checkResult{Err: err}
	}

	start := time.Now()

	resp, err := c.Reader.CheckRelation(ctx, &req)

	duration := time.Since(start)

	return &checkResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkRelationStringV3(&req),
	}
}

func checkPermissionV2(ctx context.Context, c *client.Client, msg *structpb.Value) *checkResult {
	var req dsr2.CheckPermissionRequest
	if err := unmarshalReq(msg, &req); err != nil {
		return &checkResult{Err: err}
	}

	start := time.Now()

	resp, err := c.Reader2.CheckPermission(ctx, &req)

	duration := time.Since(start)

	return &checkResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkPermissionStringV2(&req),
	}
}

func checkRelationV2(ctx context.Context, c *client.Client, msg *structpb.Value) *checkResult {
	var req dsr2.CheckRelationRequest
	if err := unmarshalReq(msg, &req); err != nil {
		return &checkResult{Err: err}
	}

	start := time.Now()

	resp, err := c.Reader2.CheckRelation(ctx, &req)

	duration := time.Since(start)

	return &checkResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkRelationStringV2(&req),
	}
}

func checkDecisionV2(ctx context.Context, c az2.AuthorizerClient, msg *structpb.Value) *checkResult {
	var req az2.IsRequest
	if err := unmarshalReq(msg, &req); err != nil {
		return &checkResult{Err: err}
	}

	start := time.Now()

	resp, err := c.Is(ctx, &req)

	duration := time.Since(start)

	return &checkResult{
		Outcome:  resp.Decisions[0].GetIs(),
		Duration: duration,
		Err:      err,
		Str:      checkDecisionStringV2(&req),
	}
}

func checkStringV3(req *dsr3.CheckRequest) string {
	return fmt.Sprintf("%s:%s#%s@%s:%s",
		req.GetObjectType(), req.GetObjectId(),
		req.GetRelation(),
		req.GetSubjectType(), req.GetSubjectId(),
	)
}

func checkRelationStringV2(req *dsr2.CheckRelationRequest) string {
	return fmt.Sprintf("%s:%s#%s@%s:%s",
		req.Object.GetType(), req.Object.GetKey(),
		req.Relation.GetName(),
		req.Subject.GetType(), req.Subject.GetKey(),
	)
}

func checkRelationStringV3(req *dsr3.CheckRelationRequest) string {
	return fmt.Sprintf("%s:%s#%s@%s:%s",
		req.GetObjectType(), req.GetObjectId(),
		req.GetRelation(),
		req.GetSubjectType(), req.GetSubjectId(),
	)
}

func checkPermissionStringV2(req *dsr2.CheckPermissionRequest) string {
	return fmt.Sprintf("%s:%s#%s@%s:%s",
		req.Object.GetType(), req.Object.GetKey(),
		req.Permission.GetName(),
		req.Subject.GetType(), req.Subject.GetKey(),
	)
}

func checkPermissionStringV3(req *dsr3.CheckPermissionRequest) string {
	return fmt.Sprintf("%s:%s#%s@%s:%s",
		req.GetObjectType(), req.GetObjectId(),
		req.GetPermission(),
		req.GetSubjectType(), req.GetObjectId(),
	)
}

func checkDecisionStringV2(req *az2.IsRequest) string {
	return fmt.Sprintf("%s/%s:%s",
		req.PolicyContext.GetPath(),
		req.PolicyContext.GetDecisions()[0],
		req.IdentityContext.Identity,
	)
}

type testResults struct {
	total   int32
	passed  int32
	failed  int32
	errored int32
}

func (t *testResults) IncrTotal() {
	atomic.AddInt32(&t.total, 1)
}

func (t *testResults) IncrPassed() {
	atomic.AddInt32(&t.passed, 1)
}

func (t *testResults) IncrFailed() {
	atomic.AddInt32(&t.failed, 1)
}

func (t *testResults) IncrErrored() {
	atomic.AddInt32(&t.errored, 1)
}

func (t *testResults) Passed(passed bool) {
	if passed {
		t.IncrPassed()
		return
	}
	t.IncrFailed()
}

type TestTemplateCmd struct {
	V2     bool `flag:"" default:"false" help:"use v2 template"`
	Pretty bool `flag:"" default:"false" help:"pretty print JSON"`
}

const assertionsTemplateV2 string = `{
  "assertions": [
    {"check_relation": {"object": {"type": "", "key": ""}, "relation": {"name": ""}, "subject": {"type": "", "key": ""}}, "expected": true},
    {"check_permission": {"object": {"type": "", "key": ""}, "permission": {"name": ""}, "subject": {"type": "", "key": ""}}, "expected": true},
	{"check_decision": {"identity_context": {"identity": "", "type": ""}, "resource_context": {}, "policy_context": {"path": "", "decisions": [""]}}, "expected":true},
  ]
}`

const assertionsTemplateV3 string = `{
  "assertions": [
	{"check": {"object_type": "", "object_id": "", "relation": "", "subject_type": "", "subject_id": ""}, "expected": true},
	{"check_relation": {"object_type": "", "object_id": "", "relation": "", "subject_type": "", "subject_id": ""}, "expected": true},
	{"check_permission": {"object_type": "", "object_id": "", "permission": "", "subject_type": "", "subject_id": ""}, "expected": true},
	{"check_decision": {"identity_context": {"identity": "", "type": ""}, "resource_context": {}, "policy_context": {"path": "", "decisions": [""]}}, "expected":true},
  ]
}`

func (cmd *TestTemplateCmd) Run(c *cc.CommonCtx) error {
	if !cmd.Pretty {
		fmt.Fprintf(c.UI.Output(), "%s\n", iff(cmd.V2, assertionsTemplateV2, assertionsTemplateV3))
		return nil
	}

	r := strings.NewReader(assertionsTemplateV3)
	if cmd.V2 {
		r = strings.NewReader(assertionsTemplateV2)
	}

	dec := json.NewDecoder(r)

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
