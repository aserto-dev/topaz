package authorizer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
)

type TestCmd struct {
	Exec     TestExecCmd     `cmd:"" help:"execute assertions"`
	Template TestTemplateCmd `cmd:"" help:"output assertions template"`
}

type TestExecCmd struct {
	common.TestExecCmd
	azc.Config
}

// nolint: gocyclo
func (cmd *TestExecCmd) Run(c *cc.CommonCtx) error {
	runner, err := common.NewAuthorizerTestRunner(
		c,
		&common.TestExecCmd{
			File:    cmd.File,
			Summary: cmd.Summary,
			Format:  cmd.Format,
			Desc:    cmd.Desc,
		},
		&cmd.Config,
	)
	if err != nil {
		return err
	}

	return runner.Run(c)
}

// 	r, err := os.Open(cmd.File)
// 	if err != nil {
// 		return err
// 	}
// 	defer r.Close()

// 	azClient, err := azc.NewClient(c, &cmd.Config)
// 	if err != nil {
// 		return err
// 	}

// 	var assertions struct {
// 		Assertions []json.RawMessage `json:"assertions"`
// 	}

// 	dec := json.NewDecoder(r)
// 	if err := dec.Decode(&assertions); err != nil {
// 		return err
// 	}

// 	cmd.results = common.NewTestResults(assertions.Assertions)

// 	csvWriter := csv.NewWriter(c.StdOut())

// 	for i := 0; i < len(assertions.Assertions); i++ {
// 		var msg structpb.Struct
// 		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(assertions.Assertions[i], &msg)
// 		if err != nil {
// 			return err
// 		}

// 		expected, ok := common.GetBool(&msg, common.Expected)
// 		if !ok {
// 			return fmt.Errorf("no expected outcome of assertion defined")
// 		}

// 		description, _ := common.GetString(&msg, common.Description)

// 		checkType := common.GetCheckType(&msg)
// 		if checkType == common.CheckUnknown {
// 			return fmt.Errorf("unknown check type")
// 		}

// 		reqVersion := getReqVersion(msg.Fields[common.CheckTypeMapStr[checkType]])
// 		if reqVersion == 0 {
// 			return fmt.Errorf("unknown request version")
// 		}

// 		var result *common.CheckResult

// 		switch {
// 		case checkType == common.CheckDecision:
// 			result = checkDecisionV2(c.Context, azClient.Authorizer, msg.Fields[common.CheckTypeMapStr[checkType]])
// 		default:
// 			continue
// 		}

// 		cmd.results.Passed(result.Outcome == expected)

// 		if result.Err != nil {
// 			cmd.results.IncrErrored()
// 		}

// 		if cmd.Format == common.TestOutputCSV {
// 			if i == 0 {
// 				result.PrintCSVHeader(csvWriter)
// 			}

// 			result.PrintCSV(csvWriter, i, expected, checkType, description)
// 		}

// 		if cmd.Format == common.TestOutputTable {
// 			result.PrintTable(os.Stdout, i, expected, checkType, cc.NoColor())
// 			common.PrintDesc(cmd.Desc, description, result, expected)
// 		}
// 	}

// 	if cmd.Format == common.TestOutputCSV {
// 		csvWriter.Flush()
// 	}

// 	if cmd.Summary {
// 		cmd.results.PrintSummary(os.Stdout)
// 	}

// 	if cmd.results.Errored() > 0 || cmd.results.Failed() > 0 {
// 		return errors.New("one or more test errored or failed")
// 	}

// 	return nil
// }

// func getReqVersion(val *structpb.Value) int {
// 	if val == nil {
// 		return 0
// 	}

// 	if v, ok := val.Kind.(*structpb.Value_StructValue); ok {
// 		if _, ok := v.StructValue.Fields["object_type"]; ok {
// 			return 3
// 		}
// 		if _, ok := v.StructValue.Fields["object"]; ok {
// 			return 2
// 		}
// 		if _, ok := v.StructValue.Fields["identity_context"]; ok {
// 			return 2
// 		}
// 	}
// 	return 0
// }

// func checkDecisionV2(ctx context.Context, c az2.AuthorizerClient, msg *structpb.Value) *common.CheckResult {
// 	var req az2.IsRequest
// 	if err := common.UnmarshalReq(msg, &req); err != nil {
// 		return &common.CheckResult{Err: err}
// 	}

// 	start := time.Now()

// 	resp, err := c.Is(ctx, &req)

// 	duration := time.Since(start)

// 	if err != nil {
// 		return &common.CheckResult{
// 			Outcome:  false,
// 			Duration: duration,
// 			Err:      err,
// 			Str:      checkDecisionStringV2(&req),
// 		}
// 	}

// 	return &common.CheckResult{
// 		Outcome:  lo.Ternary(err != nil, false, resp.Decisions[0].GetIs()),
// 		Duration: duration,
// 		Err:      err,
// 		Str:      checkDecisionStringV2(&req),
// 	}
// }

// func checkDecisionStringV2(req *az2.IsRequest) string {
// 	return fmt.Sprintf("%s/%s:%s",
// 		req.PolicyContext.GetPath(),
// 		req.PolicyContext.GetDecisions()[0],
// 		req.IdentityContext.Identity,
// 	)
// }

type TestTemplateCmd struct {
	Pretty bool `flag:"" default:"false" help:"pretty print JSON"`
}

const assertionsTemplate string = `{
  "assertions": [
	{"check_decision": {"identity_context": {"identity": "", "type": ""}, "resource_context": {}, "policy_context": {"path": "", "decisions": [""]}}, "expected":true, "description": ""},
  ]
}`

func (cmd *TestTemplateCmd) Run(c *cc.CommonCtx) error {
	if !cmd.Pretty {
		fmt.Fprintln(c.StdOut(), assertionsTemplate)
		return nil
	}

	r := strings.NewReader(assertionsTemplate)

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
