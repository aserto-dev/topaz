package impl

import (
	"context"
	"strings"

	runtime "github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/internal/tsync"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage"
)

// preparedQueryCache memoizes rego.PreparedEvalQuery values keyed by the
// (policy path, decisions) tuple of an Is request.
//
// Why this exists: rego.New(...).PrepareForEval(ctx) parses the query and
// plans the topdown evaluation; for an Is() request the prepared query is a
// pure function of the policy path, decision names, the active OPA compiler
// (rebuilt on bundle reload), and the storage instance. Re-preparing per call
// burns CPU and serializes goroutines on the OPA compiler's internal
// structures — measurable under concurrent load.
//
// Invalidation: when the OPA plugin manager rebuilds the compiler (bundle
// reload / discovery update), we evict the entire cache. The watcher is
// registered lazily on first access so we don't need to plumb runtime
// lifecycle into the constructor.
//
// Concurrency: sync.Map for the read-mostly path; singleflight collapses
// concurrent misses for the same key into one PrepareForEval call so a
// thundering herd on first use doesn't multiply work.
type preparedQueryCache struct {
	entries     tsync.Map[string, *rego.PreparedEvalQuery] // key (string) -> *rego.PreparedEvalQuery
	prepGroup   tsync.Group[string, *rego.PreparedEvalQuery]
	watcherOnce tsync.Map[*plugins.Manager, struct{}] // key (*plugins.Manager) -> struct{} (one-time RegisterCompilerTrigger per runtime)
}

func newPreparedQueryCache() *preparedQueryCache {
	return &preparedQueryCache{}
}

// cacheKey returns a stable string for the (path, decisions) tuple.
// The decisions list ordering is preserved because the prepared query
// references them positionally as x0, x1, ... — reordering produces a
// semantically different query.
func cacheKey(path string, decisions []string) string {
	var b strings.Builder

	b.Grow(len(path) + 1 + len(decisions)*8)
	b.WriteString(path)
	b.WriteByte('\x1f') // unit separator: cannot appear in path or decision names

	for i, d := range decisions {
		if i > 0 {
			b.WriteByte('\x1e') // record separator
		}

		b.WriteString(d)
	}

	return b.String()
}

// getOrPrepare returns a PreparedEvalQuery for (path, decisions) against the
// given runtime. parsedQuery must be the already-validated AST body for the
// joined query string; the function does not re-parse.
//
// preparedQueryFactory builds the rego.New(...) options when called — only
// invoked on cache miss. This keeps the singleflight key path allocation-
// free for the hot read.
func (c *preparedQueryCache) getOrPrepare(
	ctx context.Context,
	rt *runtime.Runtime,
	key string,
	preparedQueryFactory func(ctx context.Context) (rego.PreparedEvalQuery, error),
) (rego.PreparedEvalQuery, error) {
	c.ensureCompilerWatcher(rt)

	if v, ok := c.entries.Load(key); ok {
		return *v, nil
	}

	// Collapse concurrent misses on the same key.
	v, err, _ := c.prepGroup.Do(key, func() (*rego.PreparedEvalQuery, error) {
		// Re-check under the singleflight: a concurrent call may have
		// just populated the entry while we were waiting to enter Do.
		if entry, ok := c.entries.Load(key); ok {
			return entry, nil
		}

		pq, err := preparedQueryFactory(context.WithoutCancel(ctx))
		if err != nil {
			return nil, err
		}

		stored := &pq
		c.entries.Store(key, stored)

		return stored, nil
	})
	if err != nil {
		return rego.PreparedEvalQuery{}, err
	}

	return *v, nil
}

// ensureCompilerWatcher registers (exactly once per runtime) a callback on
// the runtime's plugins manager that drops the entire cache whenever the
// OPA compiler is replaced. Compiler replacement happens on bundle activation,
// discovery updates, or any other policy mutation; the prepared queries
// reference compiler state that becomes invalid afterward.
func (c *preparedQueryCache) ensureCompilerWatcher(rt *runtime.Runtime) {
	if rt == nil {
		return
	}

	pm := rt.GetPluginsManager()
	if pm == nil {
		return
	}

	if _, alreadyRegistered := c.watcherOnce.LoadOrStore(pm, struct{}{}); alreadyRegistered {
		return
	}

	// Discard everything when the compiler rotates. We don't try to be
	// precise — bundle reloads are rare relative to Is() rate.
	pm.RegisterCompilerTrigger(func(_ storage.Transaction) {
		c.entries.Range(func(k string, _ *rego.PreparedEvalQuery) bool {
			c.entries.Clear()
			return true
		})
	})
}
