package directory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"

	"github.com/fatih/color"
	"github.com/samber/lo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	check           string = "check"
	checkRelation   string = "check_relation"
	checkPermission string = "check_permission"
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
	dsc.Config
}

// nolint: funlen,gocyclo
func (cmd *TestExecCmd) Run(c *cc.CommonCtx) error {
	r, err := os.Open(cmd.File)
	if err != nil {
		return err
	}
	defer r.Close()

	dsClient, err := dsc.NewClient(c, &cmd.Config)
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
			result = checkV3(c.Context, dsClient, msg.Fields[checkTypeMapStr[checkType]])
		case checkType == CheckPermission && reqVersion == 3:
			result = checkPermissionV3(c.Context, dsClient, msg.Fields[checkTypeMapStr[checkType]])
		case checkType == CheckRelation && reqVersion == 3:
			result = checkRelationV3(c.Context, dsClient, msg.Fields[checkTypeMapStr[checkType]])
		default:
			continue
		}

		cmd.results.Passed(result.Outcome == expected)

		if result.Err != nil {
			cmd.results.IncrErrored()
		}

		fmt.Printf("%04d %-16s %v  %s [%s] (%s)\n",
			i+1,
			checkTypeMapStr[checkType],
			lo.Ternary(expected == result.Outcome, color.GreenString(passed), color.RedString(failed)),
			result.Str,
			lo.Ternary(result.Outcome, color.BlueString("%t", result.Outcome), color.YellowString("%t", result.Outcome)),
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
}

var checkTypeMapStr = map[checkType]string{
	Check:           check,
	CheckRelation:   checkRelation,
	CheckPermission: checkPermission,
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

func checkV3(ctx context.Context, c *dsc.Client, msg *structpb.Value) *checkResult {
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

func checkPermissionV3(ctx context.Context, c *dsc.Client, msg *structpb.Value) *checkResult {
	var req dsr3.CheckPermissionRequest
	if err := unmarshalReq(msg, &req); err != nil {
		return &checkResult{Err: err}
	}

	start := time.Now()

	//nolint: staticcheck // SA1019: c.Reader.CheckPermission
	resp, err := c.Reader.CheckPermission(ctx, &req)

	duration := time.Since(start)

	return &checkResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkPermissionStringV3(&req),
	}
}

func checkRelationV3(ctx context.Context, c *dsc.Client, msg *structpb.Value) *checkResult {
	var req dsr3.CheckRelationRequest
	if err := unmarshalReq(msg, &req); err != nil {
		return &checkResult{Err: err}
	}

	start := time.Now()

	//nolint: staticcheck // SA1019: c.Reader.CheckRelation
	resp, err := c.Reader.CheckRelation(ctx, &req)

	duration := time.Since(start)

	return &checkResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkRelationStringV3(&req),
	}
}

func checkStringV3(req *dsr3.CheckRequest) string {
	return fmt.Sprintf("%s:%s#%s@%s:%s",
		req.GetObjectType(), req.GetObjectId(),
		req.GetRelation(),
		req.GetSubjectType(), req.GetSubjectId(),
	)
}

func checkRelationStringV3(req *dsr3.CheckRelationRequest) string {
	return fmt.Sprintf("%s:%s#%s@%s:%s",
		req.GetObjectType(), req.GetObjectId(),
		req.GetRelation(),
		req.GetSubjectType(), req.GetSubjectId(),
	)
}

func checkPermissionStringV3(req *dsr3.CheckPermissionRequest) string {
	return fmt.Sprintf("%s:%s#%s@%s:%s",
		req.GetObjectType(), req.GetObjectId(),
		req.GetPermission(),
		req.GetSubjectType(), req.GetSubjectId(),
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
	Pretty bool `flag:"" default:"false" help:"pretty print JSON"`
}

const assertionsTemplateV3 string = `{
  "assertions": [
  	{"check": {"object_type": "", "object_id": "", "relation": "", "subject_type": "", "subject_id": ""}, "expected": true}
  ]
}`

func (cmd *TestTemplateCmd) Run(c *cc.CommonCtx) error {
	if !cmd.Pretty {
		c.Out().Msg(assertionsTemplateV3)
		return nil
	}

	r := strings.NewReader(assertionsTemplateV3)
	dec := json.NewDecoder(r)

	var template interface{}
	if err := dec.Decode(&template); err != nil {
		return err
	}

	enc := json.NewEncoder(c.StdOut())
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(template); err != nil {
		return err
	}

	return nil
}
