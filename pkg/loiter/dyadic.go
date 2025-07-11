package loiter

import "iter"

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
