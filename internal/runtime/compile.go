package runtime

import (
	"context"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/metrics"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/server/types"
	"github.com/open-policy-agent/opa/v1/topdown"
	"github.com/pkg/errors"
)

// CompileResult contains the results of a Compile execution.
type CompileResult struct {
	Result      *any
	Metrics     map[string]any
	Explanation types.TraceV1
}

func (r *Runtime) Compile(
	ctx context.Context,
	qStr string,
	input map[string]any,
	unknowns []string,
	disableInlining []string,
	pretty, includeMetrics, includeInstrumentation bool,
	explain types.ExplainModeV1,
) (*CompileResult, error) {
	m := metrics.New()
	m.Timer(metrics.ServerHandler).Start()

	txn, err := r.storage.NewTransaction(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new OPA store transaction")
	}

	defer r.storage.Abort(ctx, txn)

	var buf *topdown.BufferTracer
	if explain != types.ExplainOffV1 {
		buf = topdown.NewBufferTracer()
	}

	m.Timer(metrics.RegoQueryParse).Start()

	parsedQuery, err := r.ValidateQuery(qStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate query")
	}

	m.Timer(metrics.RegoQueryParse).Stop()

	eval := rego.New(
		rego.Compiler(r.GetPluginsManager().GetCompiler()),
		rego.Store(r.storage),
		rego.Transaction(txn),
		rego.ParsedQuery(parsedQuery),
		rego.Input(input),
		rego.Unknowns(unknowns),
		rego.DisableInlining(disableInlining),
		rego.QueryTracer(buf),
		rego.Instrument(includeInstrumentation),
		rego.Metrics(m),
		rego.Runtime(r.pluginsManager.Info),
		rego.UnsafeBuiltins(unsafeBuiltinsMap),
		rego.InterQueryBuiltinCache(r.InterQueryCache),
	)

	pq, err := eval.Partial(ctx)
	if err != nil {
		var astErr ast.Errors
		if errors.As(err, &astErr) {
			return nil, errors.Wrap(astErr, "ast error")
		}

		return nil, err
	}

	m.Timer(metrics.ServerHandler).Stop()

	result := &CompileResult{}

	if includeMetrics || includeInstrumentation {
		result.Metrics = m.All()
	}

	if explain != types.ExplainOffV1 {
		result.Explanation = r.getExplainResponse(explain, *buf, pretty)
	}

	var i any = types.PartialEvaluationResultV1{
		Queries: pq.Queries,
		Support: pq.Support,
	}

	result.Result = &i

	return result, nil
}
