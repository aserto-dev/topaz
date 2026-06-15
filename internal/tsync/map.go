//nolint:forcetypeassert
package tsync

import "sync"

type Map[K comparable, V any] struct {
	internal sync.Map
}

func (m *Map[K, V]) Load(key K) (V, bool) {
	val, ok := m.internal.Load(key)
	if !ok {
		var zero V
		return zero, false
	}

	return val.(V), true
}

func (m *Map[K, V]) Store(key K, value V) {
	m.internal.Store(key, value)
}

func (m *Map[K, V]) Clear() {
	m.internal.Clear()
}

func (m *Map[K, V]) Delete(key K) {
	m.internal.Delete(key)
}

func (m *Map[K, V]) LoadAndDelete(key K) (V, bool) {
	val, loaded := m.internal.LoadAndDelete(key)
	if !loaded {
		var zero V
		return zero, false
	}

	return val.(V), true
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (V, bool) {
	actualVal, loaded := m.internal.LoadOrStore(key, value)

	return actualVal.(V), loaded
}

func (m *Map[K, V]) Swap(key K, value V) (V, bool) {
	prevVal, loaded := m.internal.Swap(key, value)
	if !loaded {
		var zero V
		return zero, false
	}

	return prevVal.(V), true
}

//nolint:predeclared
func (m *Map[K, V]) CompareAndSwap(key K, old, new V) bool {
	return m.internal.CompareAndSwap(key, old, new)
}

func (m *Map[K, V]) CompareAndDelete(key K, old V) bool {
	return m.internal.CompareAndDelete(key, old)
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.internal.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}
