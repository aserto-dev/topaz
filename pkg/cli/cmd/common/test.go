package common

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"github.com/samber/lo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	TestOutputTable string = "table"
	TestOutputCSV   string = "csv"
	Expected        string = "expected"
	Description     string = "description"
	DescOff         string = "off"
	DescOn          string = "on"
	DescOnError     string = "on-error"
	Passed          string = "PASS"
	Failed          string = "FAIL"
	Errored         string = "ERR "
)

type CheckType int

const (
	CheckUnknown CheckType = iota
	Check
	CheckRelation
	CheckPermission
	CheckDecision
)

const (
	CheckStr           string = "check"
	CheckRelationStr   string = "check_relation"
	CheckPermissionStr string = "check_permission"
	CheckDecisionStr   string = "check_decision"
)

type CheckResult struct {
	Outcome  bool
	Duration time.Duration
	Err      error
	Str      string
}

var CheckTypeMap = map[string]CheckType{
	CheckStr:           Check,
	CheckRelationStr:   CheckRelation,
	CheckPermissionStr: CheckPermission,
	CheckDecisionStr:   CheckDecision,
}

var CheckTypeMapStr = map[CheckType]string{
	Check:           CheckStr,
	CheckRelation:   CheckRelationStr,
	CheckPermission: CheckPermissionStr,
	CheckDecision:   CheckDecisionStr,
}

func GetCheckType(msg *structpb.Struct) CheckType {
	for k, v := range CheckTypeMap {
		if _, ok := msg.Fields[k]; ok {
			return v
		}
	}
	return CheckUnknown
}

func (cr *CheckResult) PrintTable(w io.Writer, index int, expected bool, checkType CheckType, noColor bool) {
	if noColor {
		fmt.Fprintf(w,
			"%04d %-16s %v  %s [%s] (%s)\n",
			index+1,
			CheckTypeMapStr[checkType],
			lo.Ternary(expected == cr.Outcome, Passed, Failed),
			cr.Str,
			strconv.FormatBool(cr.Outcome),
			cr.Duration,
		)
	} else {
		fmt.Fprintf(w,
			"%04d %-16s %v  %s [%s] (%s)\n",
			index+1,
			CheckTypeMapStr[checkType],
			lo.Ternary(expected == cr.Outcome, color.GreenString(Passed), color.RedString(Failed)),
			cr.Str,
			lo.Ternary(cr.Outcome, color.BlueString("%t", cr.Outcome), color.YellowString("%t", cr.Outcome)),
			cr.Duration,
		)
	}
}

func (cr *CheckResult) PrintCSV(w *csv.Writer, index int, expected bool, checkType CheckType, description string) {
	_ = w.Write([]string{
		strconv.Itoa(index + 1),
		CheckTypeMapStr[checkType],
		lo.Ternary(expected == cr.Outcome, Passed, Failed),
		cr.Str,
		lo.Ternary(cr.Outcome, strconv.FormatBool(cr.Outcome), strconv.FormatBool(cr.Outcome)),
		strconv.FormatInt(int64(cr.Duration/time.Microsecond), 10),
		description,
	})
}

func (cr *CheckResult) PrintCSVHeader(w *csv.Writer) {
	_ = w.Write([]string{
		"id",
		"type",
		"expected",
		"result",
		"outcome",
		"duration",
		"description",
	})
}

func PrintDesc(descFlag, description string, result *CheckResult, expected bool) {
	if descFlag != DescOff && description != "" {
		if descFlag == DescOn {
			fmt.Printf(">>>> %s\n", description)
		}
		if descFlag == DescOnError && result.Outcome != expected {
			fmt.Printf("!!!! %s\n", description)
		}
	}
}

func NewTestResults(assertions []json.RawMessage) *TestResults {
	return &TestResults{
		total:   int32(len(assertions)), //nolint: gosec // G115: integer overflow conversion int -> int32.
		passed:  0,
		failed:  0,
		errored: 0,
	}
}

type TestResults struct {
	total   int32
	passed  int32
	failed  int32
	errored int32
	skipped int32
}

func (t *TestResults) IncrTotal() {
	atomic.AddInt32(&t.total, 1)
}

func (t *TestResults) IncrPassed() {
	atomic.AddInt32(&t.passed, 1)
}

func (t *TestResults) IncrFailed() {
	atomic.AddInt32(&t.failed, 1)
}

func (t *TestResults) IncrErrored() {
	atomic.AddInt32(&t.errored, 1)
}

func (t *TestResults) IncrSkipped() {
	atomic.AddInt32(&t.skipped, 1)
}

func (t *TestResults) Passed(passed bool) {
	if passed {
		t.IncrPassed()
		return
	}
	t.IncrFailed()
}

func (t *TestResults) PrintSummary(w io.Writer) {
	fmt.Fprintf(w, "\nTest Execution Summary:\n")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 23))
	fmt.Fprintf(w, "total:   %d\n", t.total)
	fmt.Fprintf(w, "passed:  %d\n", t.passed)
	fmt.Fprintf(w, "failed:  %d\n", t.failed)
	fmt.Fprintf(w, "skipped: %d\n", t.skipped)
	fmt.Fprintf(w, "errored: %d\n", t.errored)
	fmt.Fprintln(w)
}

func (t *TestResults) Errored() int32 {
	return t.errored
}

func (t *TestResults) Failed() int32 {
	return t.failed
}

func GetBool(msg *structpb.Struct, fieldName string) (bool, bool) {
	v, ok := msg.Fields[fieldName]
	return v.GetBoolValue(), ok
}

func GetString(msg *structpb.Struct, fieldName string) (string, bool) {
	v, ok := msg.Fields[fieldName]
	return v.GetStringValue(), ok
}

func UnmarshalReq(value *structpb.Value, msg proto.Message) error {
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
