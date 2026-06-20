package runtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/metrics"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/server/types"
	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/open-policy-agent/opa/v1/topdown"
	"github.com/open-policy-agent/opa/v1/topdown/lineage"
	"github.com/pkg/errors"
)

// map of unsafe builtins.
var unsafeBuiltinsMap = map[string]struct{}{ast.HTTPSend.Name: {}}

// Result contains the results of a Query execution.
type Result struct {
	Result      rego.ResultSet `json:"result"`
	Metrics     map[string]any `json:"metrics"`
	Explanation types.TraceV1  `json:"explanation"`
	DecisionID  string         `json:"decision_id"`
}

// Query executes a REGO query against the Aserto OPA Runtime
// explain can be "notes", "full" or "off".
func (r *Runtime) Query(
	ctx context.Context,
	qStr string,
	input map[string]any,
	pretty, includeMetrics, includeInstrumentation bool,
	explain types.ExplainModeV1,
) (*Result, error) {
	m := metrics.New()

	decisionID := uuid.New().String()

	parsedQuery, err := r.ValidateQuery(qStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate query")
	}

	txn, err := r.storage.NewTransaction(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new OPA store transaction")
	}

	defer r.storage.Abort(ctx, txn)

	results, err := r.execQuery(ctx, txn, decisionID, parsedQuery, input, m, explain, includeMetrics, includeInstrumentation, pretty)
	if err != nil {
		return nil, errors.Wrapf(err, "query execution failed, decision-id: [%s], query: [%s]", decisionID, qStr)
	}

	return results, nil
}

func (r *Runtime) ValidateQuery(query string) (ast.Body, error) {
	var body ast.Body

	body, err := ast.ParseBody(query)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (r *Runtime) ValidateRule(rule string) (bool, error) {
	ref, err := ast.ParseRef(rule)
	if err != nil {
		return false, fmt.Errorf("invalid ref: %w", err)
	}

	compiler := r.pluginsManager.GetCompiler()
	rules := compiler.GetRules(ref)

	return len(rules) > 0, nil
}

func (r *Runtime) execQuery(
	ctx context.Context,
	txn storage.Transaction,
	decisionID string,
	parsedQuery ast.Body,
	input map[string]any,
	m metrics.Metrics,
	explainMode types.ExplainModeV1,
	includeMetrics, includeInstrumentation, pretty bool,
) (*Result, error) {
	var buf *topdown.BufferTracer
	if explainMode != types.ExplainOffV1 {
		buf = topdown.NewBufferTracer()
	}

	opts := r.builtins

	compiler := r.pluginsManager.GetCompiler()

	opts = append(opts,
		rego.Store(r.storage),
		rego.Transaction(txn),
		rego.Compiler(compiler),
		rego.ParsedQuery(parsedQuery),
		rego.Metrics(m),
		rego.Instrument(includeInstrumentation),
		rego.QueryTracer(buf),
		rego.Trace(true),
		rego.Runtime(r.pluginsManager.Info),
		rego.UnsafeBuiltins(unsafeBuiltinsMap),
		rego.InterQueryBuiltinCache(r.InterQueryCache),
		rego.Input(input),
		rego.Imports(r.imports),
	)

	for _, r := range r.pluginsManager.GetWasmResolvers() {
		for _, entrypoint := range r.Entrypoints() {
			opts = append(opts, rego.Resolver(entrypoint, r))
		}
	}

	regoQuery := rego.New(opts...)

	output, err := regoQuery.Eval(ctx)
	if err != nil {
		r.Logger.Warn().
			Err(err).Str("decisionID", decisionID).
			Str("query", parsedQuery.String()).
			Interface("input", input).
			Msg("error evaluating query")

		return nil, errors.Wrap(err, "failed to evaluate rego query")
	}

	results := &Result{
		Result:     output,
		DecisionID: decisionID,
	}

	if includeMetrics || includeInstrumentation {
		results.Metrics = m.All()
	}

	if explainMode != types.ExplainOffV1 {
		results.Explanation = r.getExplainResponse(explainMode, *buf, pretty)
	}

	r.Logger.Debug().
		Err(err).Str("decisionID", decisionID).
		Str("query", parsedQuery.String()).
		Interface("input", input).
		Msg("query evaluated")

	return results, err
}

func (r *Runtime) getExplainResponse(explainMode types.ExplainModeV1, trace []*topdown.Event, pretty bool) types.TraceV1 {
	switch explainMode {
	case types.ExplainNotesV1:
		if explanation, err := types.NewTraceV1(lineage.Notes(trace), pretty); err == nil {
			return explanation
		}

		return nil

	case types.ExplainFailsV1:
		if explanation, err := types.NewTraceV1(lineage.Fails(trace), pretty); err == nil {
			return explanation
		}

		return nil

	case types.ExplainDebugV1:
		if explanation, err := types.NewTraceV1(lineage.Debug(trace), pretty); err == nil {
			return explanation
		}

		return nil

	case types.ExplainFullV1:
		if explanation, err := types.NewTraceV1(trace, pretty); err == nil {
			return explanation
		}

		return nil

	case types.ExplainOffV1:
		return nil
	}

	return nil
}
