//nolint:forcetypeassert
package tsync

import (
	"golang.org/x/sync/singleflight"
)

// Result wraps the singleflight return values with strict types.
type Result[V any] struct {
	Val    V
	Err    error
	Shared bool
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type Group[K comparable, V any] struct {
	internal singleflight.Group
}

// Do executes and suppresses duplicate calls for the given key.
func (g *Group[K, V]) Do(key K, fn func() (V, error)) (V, error, bool) {
	// Convert key to string representation for internal singleflight
	// You can use a dedicated stringifier function or fmt.Sprint if K is complex
	stringKey := any(key).(string)

	val, err, shared := g.internal.Do(stringKey, func() (any, error) {
		return fn()
	})

	if val == nil {
		var zero V
		return zero, err, shared
	}

	return val.(V), err, shared
}

// DoChan is like Do but returns a channel that will receive the
// results when they become ready.
func (g *Group[K, V]) DoChan(key K, fn func() (V, error)) <-chan Result[V] {
	stringKey := any(key).(string)

	internalCh := g.internal.DoChan(stringKey, func() (any, error) {
		return fn()
	})

	typedCh := make(chan Result[V], 1)

	// Stream internal untyped results into our typed channel asynchronously
	go func() {
		defer close(typedCh)

		res := <-internalCh

		var typedVal V
		if res.Val != nil {
			typedVal = res.Val.(V)
		}

		typedCh <- Result[V]{
			Val:    typedVal,
			Err:    res.Err,
			Shared: res.Shared,
		}
	}()

	return typedCh
}

// Forget tells the singleflight to forget about a key.
func (g *Group[K, V]) Forget(key K) {
	stringKey := any(key).(string)
	g.internal.Forget(stringKey)
}
