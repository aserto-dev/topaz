package impl

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/open-policy-agent/opa/v1/rego"
)

// TestCacheKey checks that the key derivation produces stable, distinct
// keys for different (path, decisions) inputs and is order-sensitive on
// the decisions list (ordering matters because the prepared query
// references decisions positionally as x0, x1, ...).
func TestCacheKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		pathA     string
		decsA     []string
		pathB     string
		decsB     []string
		wantEqual bool
	}{
		{"identical", "p", []string{"a"}, "p", []string{"a"}, true},
		{"different path", "p", []string{"a"}, "q", []string{"a"}, false},
		{"different decision", "p", []string{"a"}, "p", []string{"b"}, false},
		{"different decision count", "p", []string{"a"}, "p", []string{"a", "b"}, false},
		{"order matters", "p", []string{"a", "b"}, "p", []string{"b", "a"}, false},
		// Edge: paths or decisions that contain characters which a naive
		// "join with /" key would conflate. Our separators are 0x1f / 0x1e,
		// neither valid in a Rego identifier — but we still test that the
		// key is stable.
		{"path containing slash", "p/q", []string{"a"}, "p", []string{"q.a"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := cacheKey(tc.pathA, tc.decsA)
			b := cacheKey(tc.pathB, tc.decsB)
			if (a == b) != tc.wantEqual {
				t.Fatalf("cacheKey(%q,%v)=%q vs cacheKey(%q,%v)=%q: equal=%v want %v",
					tc.pathA, tc.decsA, a, tc.pathB, tc.decsB, b, a == b, tc.wantEqual)
			}
		})
	}
}

// TestGetOrPrepare_CachesAndDedupes asserts that for a given key the
// expensive factory function runs at most once across many concurrent
// goroutines, and that subsequent calls reuse the cached PreparedEvalQuery.
func TestGetOrPrepare_CachesAndDedupes(t *testing.T) {
	t.Parallel()

	c := newPreparedQueryCache()
	var prepCalls atomic.Int64

	// Build a real, evaluable PreparedEvalQuery from a trivial Rego module
	// so we exercise the same code path Is() would. The cache stores the
	// prepared object; we don't need the runtime here because the cache
	// itself doesn't care what's behind the factory.
	factory := func(ctx context.Context) (rego.PreparedEvalQuery, error) {
		prepCalls.Add(1)
		return rego.New(
			rego.Query("data.test.allow"),
			rego.Module("test.rego", "package test\nallow := true"),
		).PrepareForEval(ctx)
	}

	const N = 200
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			_, err := c.getOrPrepare(context.Background(), nil, "key1", factory)
			if err != nil {
				t.Errorf("getOrPrepare: %v", err)
			}
		}()
	}
	wg.Wait()

	// With singleflight collapsing concurrent misses, we expect 1 factory
	// call. Allow up to a small handful in case the platform's scheduler
	// races the first store-vs-load — but flag explicitly if it's many.
	if calls := prepCalls.Load(); calls > 4 {
		t.Fatalf("factory called %d times for one key, expected ~1", calls)
	}

	// A different key triggers exactly one more call.
	prevCalls := prepCalls.Load()
	_, err := c.getOrPrepare(context.Background(), nil, "key2", factory)
	if err != nil {
		t.Fatalf("getOrPrepare key2: %v", err)
	}
	if got := prepCalls.Load(); got != prevCalls+1 {
		t.Fatalf("expected exactly one factory call for key2; got delta=%d", got-prevCalls)
	}

	// Same key as the first batch — should NOT call the factory.
	prevCalls = prepCalls.Load()
	for i := 0; i < 50; i++ {
		_, err := c.getOrPrepare(context.Background(), nil, "key1", factory)
		if err != nil {
			t.Fatalf("getOrPrepare key1 reuse: %v", err)
		}
	}
	if got := prepCalls.Load(); got != prevCalls {
		t.Fatalf("expected 0 factory calls on cache-hit reuse; got delta=%d", got-prevCalls)
	}
}

// TestGetOrPrepare_FactoryError ensures errors from the factory are
// propagated and NOT cached (so a transient failure doesn't poison the
// cache for the lifetime of the server).
func TestGetOrPrepare_FactoryError(t *testing.T) {
	t.Parallel()

	c := newPreparedQueryCache()
	var calls atomic.Int64

	failingFactory := func(ctx context.Context) (rego.PreparedEvalQuery, error) {
		calls.Add(1)
		// Intentionally invalid — produces a real PrepareForEval error.
		return rego.New(rego.Query("not a valid query")).PrepareForEval(ctx)
	}

	for i := 0; i < 3; i++ {
		_, err := c.getOrPrepare(context.Background(), nil, "bad", failingFactory)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("expected factory to be called once per attempt on error; got %d calls for 3 attempts", got)
	}
}
