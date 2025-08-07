package loiter

import (
	"iter"

	"github.com/samber/lo"
)

// Chain2 concatenates multiple sequences into one.
func Chain2[K any, V any](s ...iter.Seq2[K, V]) iter.Seq2[K, V] {
	return func(yield Yields2[K, V]) {
		for _, it := range s {
			for k, v := range it {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// Collect is simlar to maps.Collect, but takes a replacement function that decides which value to keep
// when duplicate keys are encountered.
// The function takes the key, the previously seen value, and the current value. If it return false, the previous vcalue
// is kept, otherwise the current value is used.
func Collect[K comparable, V any](s iter.Seq2[K, V], replace func(key K, prev, current V) bool) map[K]V {
	m := make(map[K]V)
	for k, v := range s {
		if prev, ok := m[k]; !ok || replace(k, prev, v) {
			m[k] = v
		}
	}

	return m
}

func ExplodeValues[K any, V any, R any](s iter.Seq2[K, V], transform func(K, V) iter.Seq[R]) iter.Seq2[K, R] {
	return func(yield Yields2[K, R]) {
		for k, v := range s {
			for r := range transform(k, v) {
				if !yield(k, r) {
					return
				}
			}
		}
	}
}

// Flatten transforms a dyadic sequence (Seq2[K, V]) to a monadic one (Seq[R]).
func Flatten[K any, V any, R any](s iter.Seq2[K, V], transform func(K, V) R) iter.Seq[R] {
	return func(yield Yields[R]) {
		for k, v := range s {
			if !yield(transform(k, v)) {
				return
			}
		}
	}
}

// MapKeys transforms the first-position elements (keys) in a dyadic sequence.
func MapKeys[K any, V any, R any](s iter.Seq2[K, V], transform func(K, V) R) iter.Seq2[R, V] {
	return func(yield Yields2[R, V]) {
		for k, v := range s {
			if !yield(transform(k, v), v) {
				return
			}
		}
	}
}

// MapValues transforms the second-position elements (values) of a dyadic sequence.
func MapValues[K any, V any, R any](s iter.Seq2[K, V], transform func(K, V) R) iter.Seq2[K, R] {
	return func(yield Yields2[K, R]) {
		for k, v := range s {
			if !yield(k, transform(k, v)) {
				return
			}
		}
	}
}

// MapEntries transform a dyadic sequence to a sequence of another type.
func MapEntries[K1 any, V1 any, K2 any, V2 any](s iter.Seq2[K1, V1], transform func(K1, V1) (K2, V2)) iter.Seq2[K2, V2] {
	return func(yield Yields2[K2, V2]) {
		for k, v := range s {
			if !yield(transform(k, v)) {
				return
			}
		}
	}
}

func RejectKey[K comparable, V any](s iter.Seq2[K, V], reject K) iter.Seq2[K, V] {
	return func(yield Yields2[K, V]) {
		for k, v := range s {
			if k != reject && !yield(k, v) {
				return
			}
		}
	}
}

func RejectValue[K any, V comparable](s iter.Seq2[K, V], reject V) iter.Seq2[K, V] {
	return func(yield Yields2[K, V]) {
		for k, v := range s {
			if v != reject && !yield(k, v) {
				return
			}
		}
	}
}

func Seq2[K any, V any](s ...lo.Tuple2[K, V]) iter.Seq2[K, V] {
	return func(yield Yields2[K, V]) {
		for _, t := range s {
			if !yield(t.A, t.B) {
				return
			}
		}
	}
}

// Transpose swaps the keys and values of a dyadic sequence.
func Transpose[K any, V any](s iter.Seq2[K, V]) iter.Seq2[V, K] {
	return func(yield Yields2[V, K]) {
		for k, v := range s {
			if !yield(v, k) {
				return
			}
		}
	}
}

func WithKey[K any, V any](s iter.Seq[V], key K) iter.Seq2[K, V] {
	return func(yield Yields2[K, V]) {
		for v := range s {
			if !yield(key, v) {
				return
			}
		}
	}
}

func WithValue[K any, V any](s iter.Seq[K], value V) iter.Seq2[K, V] {
	return func(yield Yields2[K, V]) {
		for k := range s {
			if !yield(k, value) {
				return
			}
		}
	}
}
