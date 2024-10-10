package common

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	cerr "github.com/aserto-dev/errors"
	az2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	ErrSkippedAuthorizerAssertion = cerr.NewAsertoError("T10001", codes.Internal, http.StatusInternalServerError, "no authorizer client")
	ErrSkippedDirectoryAssertion  = cerr.NewAsertoError("T10002", codes.Internal, http.StatusInternalServerError, "no directory client")
)

type TestExecCmd struct {
	Files   []string `arg:""  default:"assertions.json" help:"path to assertions file" sep:"none" optional:""`
	Stdin   bool     `flag:"" default:"false" help:"read assertions from --stdin"`
	Summary bool     `flag:"" default:"false" help:"display test summary"`
	Format  string   `flag:"" default:"table" help:"output format (table|csv)" enum:"table,csv"`
	Desc    string   `flag:"" default:"off" enum:"off,on,on-error" help:"output descriptions (off|on|on-error)"`
}

type TestRunner struct {
	cmd      *TestExecCmd
	azClient *azc.Client
	dsClient *dsc.Client
	results  *TestResults
}

func NewDirectoryTestRunner(c *cc.CommonCtx, cmd *TestExecCmd, dsConfig *dsc.Config) (*TestRunner, error) {
	dsClient, err := dsc.NewClient(c, dsConfig)
	if err != nil {
		return nil, err
	}

	return &TestRunner{
		cmd:      cmd,
		dsClient: dsClient,
	}, nil
}

func NewAuthorizerTestRunner(c *cc.CommonCtx, cmd *TestExecCmd, azConfig *azc.Config) (*TestRunner, error) {
	azClient, err := azc.NewClient(c, azConfig)
	if err != nil {
		return nil, err
	}

	return &TestRunner{
		cmd:      cmd,
		azClient: azClient,
	}, nil
}

func NewTestRunner(c *cc.CommonCtx, cmd *TestExecCmd, azConfig *azc.Config, dsConfig *dsc.Config) (*TestRunner, error) {
	dsClient, err := dsc.NewClient(c, dsConfig)
	if err != nil {
		return nil, err
	}

	azClient, err := azc.NewClient(c, azConfig)
	if err != nil {
		return nil, err
	}

	return &TestRunner{
		cmd:      cmd,
		azClient: azClient,
		dsClient: dsClient,
	}, nil
}

func (runner *TestRunner) Run(c *cc.CommonCtx) error {
	if runner.cmd.Stdin {
		return runner.exec(c, os.Stdin)
	}

	for _, file := range runner.cmd.Files {
		if err := runner.execFile(c, file); err != nil {
			return err
		}
	}

	return nil
}

func (runner *TestRunner) execFile(c *cc.CommonCtx, file string) error {
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()

	c.Con().Info().Msg(file)
	return runner.exec(c, r)
}

var pbUnmarshal = protojson.UnmarshalOptions{DiscardUnknown: true}

// nolint: gocyclo
func (runner *TestRunner) exec(c *cc.CommonCtx, r *os.File) error {
	csvWriter := csv.NewWriter(c.StdOut())

	var assertions struct {
		Assertions []json.RawMessage `json:"assertions"`
	}

	dec := json.NewDecoder(r)
	if err := dec.Decode(&assertions); err != nil {
		return err
	}

	runner.results = NewTestResults(assertions.Assertions)

	for i := 0; i < len(assertions.Assertions); i++ {
		var msg structpb.Struct
		if err := pbUnmarshal.Unmarshal(assertions.Assertions[i], &msg); err != nil {
			return err
		}

		expected, ok := GetBool(&msg, Expected)
		if !ok {
			return errors.Errorf("no expected outcome of assertion defined")
		}

		description, _ := GetString(&msg, Description)

		checkType := GetCheckType(&msg)
		if checkType == CheckUnknown {
			return errors.Errorf("unknown check type")
		}

		reqVersion := getReqVersion(msg.Fields[CheckTypeMapStr[checkType]])
		if reqVersion == 0 {
			return errors.Errorf("unknown request version")
		}

		var result *CheckResult

		switch {
		case checkType == Check && reqVersion == 3:
			result = checkV3(c.Context, runner.dsClient, msg.Fields[CheckTypeMapStr[checkType]])
		case checkType == CheckPermission && reqVersion == 3:
			result = checkPermissionV3(c.Context, runner.dsClient, msg.Fields[CheckTypeMapStr[checkType]])
		case checkType == CheckRelation && reqVersion == 3:
			result = checkRelationV3(c.Context, runner.dsClient, msg.Fields[CheckTypeMapStr[checkType]])
		case checkType == CheckDecision:
			result = checkDecisionV2(c.Context, runner.azClient, msg.Fields[CheckTypeMapStr[checkType]])
		default:
			continue
		}

		runner.results.Passed(result.Outcome == expected)

		if result.Err != nil {
			switch {
			case errors.Is(result.Err, ErrSkippedAuthorizerAssertion):
				result.Err = nil
				runner.results.IncrSkipped()
			case errors.Is(result.Err, ErrSkippedDirectoryAssertion):
				result.Err = nil
				runner.results.IncrSkipped()
			default:
				runner.results.IncrErrored()
			}
		}

		if runner.cmd.Format == TestOutputCSV {
			if i == 0 {
				result.PrintCSVHeader(csvWriter)
			}
			result.PrintCSV(csvWriter, i, expected, checkType, description)
		}

		if runner.cmd.Format == TestOutputTable {
			result.PrintTable(os.Stdout, i, expected, checkType, cc.NoColor())
			PrintDesc(runner.cmd.Desc, description, result, expected)
		}
	}

	if runner.cmd.Format == TestOutputCSV {
		csvWriter.Flush()
	}

	if runner.cmd.Summary {
		runner.results.PrintSummary(os.Stdout)
	}

	if runner.results.failed != 0 || runner.results.errored != 0 {
		return errors.Errorf("%d tests failed, %d tests errored", runner.results.failed, runner.results.errored)
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
		if _, ok := v.StructValue.Fields["object"]; ok {
			return 2
		}
		if _, ok := v.StructValue.Fields["identity_context"]; ok {
			return 2
		}
	}
	return 0
}

func checkV3(ctx context.Context, c *dsc.Client, msg *structpb.Value) *CheckResult {
	if c == nil {
		return &CheckResult{
			Outcome:  false,
			Duration: 0,
			Err:      ErrSkippedDirectoryAssertion,
			Str:      "SKIPPED",
		}
	}

	var req dsr3.CheckRequest
	if err := UnmarshalReq(msg, &req); err != nil {
		return &CheckResult{Err: err}
	}

	start := time.Now()

	resp, err := c.Reader.Check(ctx, &req)

	duration := time.Since(start)

	return &CheckResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkStringV3(&req),
	}
}

func checkPermissionV3(ctx context.Context, c *dsc.Client, msg *structpb.Value) *CheckResult {
	if c == nil {
		return &CheckResult{
			Outcome:  false,
			Duration: 0,
			Err:      ErrSkippedDirectoryAssertion,
			Str:      "SKIPPED",
		}
	}

	var req dsr3.CheckPermissionRequest
	if err := UnmarshalReq(msg, &req); err != nil {
		return &CheckResult{Err: err}
	}

	start := time.Now()

	//nolint: staticcheck // SA1019: c.Reader.CheckPermission
	resp, err := c.Reader.CheckPermission(ctx, &req)

	duration := time.Since(start)

	return &CheckResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkPermissionStringV3(&req),
	}
}

func checkRelationV3(ctx context.Context, c *dsc.Client, msg *structpb.Value) *CheckResult {
	if c == nil {
		return &CheckResult{
			Outcome:  false,
			Duration: 0,
			Err:      ErrSkippedDirectoryAssertion,
			Str:      "SKIPPED",
		}
	}

	var req dsr3.CheckRelationRequest
	if err := UnmarshalReq(msg, &req); err != nil {
		return &CheckResult{Err: err}
	}

	start := time.Now()

	//nolint: staticcheck // SA1019: c.Reader.CheckRelation
	resp, err := c.Reader.CheckRelation(ctx, &req)

	duration := time.Since(start)

	return &CheckResult{
		Outcome:  resp.GetCheck(),
		Duration: duration,
		Err:      err,
		Str:      checkRelationStringV3(&req),
	}
}

func checkDecisionV2(ctx context.Context, c *azc.Client, msg *structpb.Value) *CheckResult {
	if c == nil {
		return &CheckResult{
			Outcome:  false,
			Duration: 0,
			Err:      ErrSkippedAuthorizerAssertion,
			Str:      "SKIPPED",
		}
	}

	var req az2.IsRequest
	if err := UnmarshalReq(msg, &req); err != nil {
		return &CheckResult{Err: err}
	}

	start := time.Now()

	resp, err := c.Authorizer.Is(ctx, &req)

	duration := time.Since(start)

	if err != nil {
		return &CheckResult{
			Outcome:  false,
			Duration: duration,
			Err:      err,
			Str:      checkDecisionStringV2(&req),
		}
	}

	return &CheckResult{
		Outcome:  lo.Ternary(err != nil, false, resp.Decisions[0].GetIs()),
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

func checkDecisionStringV2(req *az2.IsRequest) string {
	return fmt.Sprintf("%s/%s:%s",
		req.PolicyContext.GetPath(),
		req.PolicyContext.GetDecisions()[0],
		req.IdentityContext.Identity,
	)
}
