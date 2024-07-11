package authorizer

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
	"github.com/samber/lo"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"

	"github.com/fatih/color"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	checkDecision string = "check_decision"
	expected      string = "expected"
	passed        string = "PASS"
	failed        string = "FAIL"
	errored       string = "ERR "
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
	clients.AuthorizerConfig
}

// nolint: funlen,gocyclo
func (cmd *TestExecCmd) Run(c *cc.CommonCtx) error {
	r, err := os.Open(cmd.File)
	if err != nil {
		return err
	}
	defer r.Close()

	azc, err := clients.NewAuthorizerClient(c, &cmd.AuthorizerConfig)
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
		case checkType == CheckDecision:
			result = checkDecisionV2(c.Context, azc, msg.Fields[checkTypeMapStr[checkType]])
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
	CheckDecision
)

var checkTypeMap = map[string]checkType{
	checkDecision: CheckDecision,
}

var checkTypeMapStr = map[checkType]string{
	CheckDecision: checkDecision,
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

func checkDecisionV2(ctx context.Context, c az2.AuthorizerClient, msg *structpb.Value) *checkResult {
	var req az2.IsRequest
	if err := unmarshalReq(msg, &req); err != nil {
		return &checkResult{Err: err}
	}

	start := time.Now()

	resp, err := c.Is(ctx, &req)

	duration := time.Since(start)

	if err != nil {
		return &checkResult{
			Outcome:  false,
			Duration: duration,
			Err:      err,
			Str:      checkDecisionStringV2(&req),
		}
	}

	return &checkResult{
		Outcome:  lo.Ternary(err != nil, false, resp.Decisions[0].GetIs()),
		Duration: duration,
		Err:      err,
		Str:      checkDecisionStringV2(&req),
	}
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
	{"check_decision": {"identity_context": {"identity": "", "type": ""}, "resource_context": {}, "policy_context": {"path": "", "decisions": [""]}}, "expected":true},
  ]
}`

const assertionsTemplateV3 string = `{
  "assertions": [
	{"check_decision": {"identity_context": {"identity": "", "type": ""}, "resource_context": {}, "policy_context": {"path": "", "decisions": [""]}}, "expected":true},
  ]
}`

func (cmd *TestTemplateCmd) Run(c *cc.CommonCtx) error {
	if !cmd.Pretty {
		fmt.Fprintf(c.StdOut(), "%s\n", lo.Ternary(cmd.V2, assertionsTemplateV2, assertionsTemplateV3))
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

	enc := json.NewEncoder(c.StdOut())
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(template); err != nil {
		return err
	}

	return nil
}
