package directory

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type TestCmd struct {
	Exec     TestExecCmd     `cmd:"" help:"execute assertions"`
	Template TestTemplateCmd `cmd:"" help:"output assertions template"`
}

type TestExecCmd struct {
	File    string `arg:""  default:"assertions.json" help:"filepath to assertions file"`
	Summary bool   `flag:"" default:"false" help:"display test summary"`
	Format  string `flag:"" default:"table" help:"output format (table|csv)" enum:"table,csv"`
	Desc    string `flag:"" default:"off" enum:"off,on,on-error" help:"output descriptions (off|on|on-error)"`

	results *common.TestResults
	dsc.Config
}

// nolint: gocyclo
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

	csvWriter := csv.NewWriter(c.StdOut())

	var assertions struct {
		Assertions []json.RawMessage `json:"assertions"`
	}

	dec := json.NewDecoder(r)
	if err := dec.Decode(&assertions); err != nil {
		return err
	}

	cmd.results = common.NewTestResults(assertions.Assertions)

	for i := 0; i < len(assertions.Assertions); i++ {
		var msg structpb.Struct
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(assertions.Assertions[i], &msg)
		if err != nil {
			return err
		}

		expected, ok := common.GetBool(&msg, common.Expected)
		if !ok {
			return fmt.Errorf("no expected outcome of assertion defined")
		}

		description, _ := common.GetString(&msg, common.Description)

		checkType := common.GetCheckType(&msg)
		if checkType == common.CheckUnknown {
			return fmt.Errorf("unknown check type")
		}

		reqVersion := getReqVersion(msg.Fields[common.CheckTypeMapStr[checkType]])
		if reqVersion == 0 {
			return fmt.Errorf("unknown request version")
		}

		var result *common.CheckResult

		switch {
		case checkType == common.Check && reqVersion == 3:
			result = checkV3(c.Context, dsClient, msg.Fields[common.CheckTypeMapStr[checkType]])
		case checkType == common.CheckPermission && reqVersion == 3:
			result = checkPermissionV3(c.Context, dsClient, msg.Fields[common.CheckTypeMapStr[checkType]])
		case checkType == common.CheckRelation && reqVersion == 3:
			result = checkRelationV3(c.Context, dsClient, msg.Fields[common.CheckTypeMapStr[checkType]])
		default:
			continue
		}

		cmd.results.Passed(result.Outcome == expected)

		if result.Err != nil {
			cmd.results.IncrErrored()
		}

		if cmd.Format == common.TestOutputCSV {
			if i == 0 {
				result.PrintCSVHeader(csvWriter)
			}
			result.PrintCSV(csvWriter, i, expected, checkType, description)
		}

		if cmd.Format == common.TestOutputTable {
			result.PrintTable(os.Stdout, i, expected, checkType, cc.NoColor())
			common.PrintDesc(cmd.Desc, description, result, expected)
		}
	}

	if cmd.Format == common.TestOutputCSV {
		csvWriter.Flush()
	}

	if cmd.Summary {
		cmd.results.PrintSummary(os.Stdout)
	}

	if cmd.results.Errored() > 0 || cmd.results.Failed() > 0 {
		return errors.New("one or more test errored or failed")
	}

	return nil
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

func checkV3(ctx context.Context, c *dsc.Client, msg *structpb.Value) *common.CheckResult {
	var req dsr3.CheckRequest
	if err := common.UnmarshalReq(msg, &req); err != nil {
		return &common.CheckResult{Err: err}
	}

	start := time.Now()

	resp, err := c.Reader.Check(ctx, &req)

	duration := time.Since(start)

	return &common.CheckResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkStringV3(&req),
	}
}

func checkPermissionV3(ctx context.Context, c *dsc.Client, msg *structpb.Value) *common.CheckResult {
	var req dsr3.CheckPermissionRequest
	if err := common.UnmarshalReq(msg, &req); err != nil {
		return &common.CheckResult{Err: err}
	}

	start := time.Now()

	//nolint: staticcheck // SA1019: c.Reader.CheckPermission
	resp, err := c.Reader.CheckPermission(ctx, &req)

	duration := time.Since(start)

	return &common.CheckResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkPermissionStringV3(&req),
	}
}

func checkRelationV3(ctx context.Context, c *dsc.Client, msg *structpb.Value) *common.CheckResult {
	var req dsr3.CheckRelationRequest
	if err := common.UnmarshalReq(msg, &req); err != nil {
		return &common.CheckResult{Err: err}
	}

	start := time.Now()

	//nolint: staticcheck // SA1019: c.Reader.CheckRelation
	resp, err := c.Reader.CheckRelation(ctx, &req)

	duration := time.Since(start)

	return &common.CheckResult{
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

type TestTemplateCmd struct {
	Pretty bool `flag:"" default:"false" help:"pretty print JSON"`
}

const assertionsTemplateV3 string = `{
  "assertions": [
  	{"check": {"object_type": "", "object_id": "", "relation": "", "subject_type": "", "subject_id": ""}, "expected": true, "description": ""}
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
