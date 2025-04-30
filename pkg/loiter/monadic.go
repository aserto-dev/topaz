// Package loiter does for iterators what github.com/samber/lo does for slices and maps.
package loiter

import (
	"iter"
	"slices"
)

// Append takes a sequence and returns a new sequence with additional values at the end.
func Append[T any](s iter.Seq[T], vals ...T) iter.Seq[T] {
	return Chain(s, slices.Values(vals))
}

// Chain concatenates multiple sequences into one.
func Chain[T any](s ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, it := range s {
			for v := range it {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func ContainsAny[T comparable](s iter.Seq[T], vals ...T) bool {
	lookup := make(map[T]struct{}, len(vals))
	for _, v := range vals {
		lookup[v] = struct{}{}
	}

	for t := range s {
		if _, ok := lookup[t]; ok {
			return true
		}
	}

	return false
}

// Filter returns a sequence that only yields the items that satisfy the predicate.
func Filter[T any](s iter.Seq[T], predicate func(item T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for t := range s {
			if predicate(t) && !yield(t) {
				return
			}
		}
	}
}

// Map transforms a sequence of one type to another.
func Map[T any, R any](s iter.Seq[T], transform func(T) R) iter.Seq[R] {
	return func(yield func(R) bool) {
		for t := range s {
			if !yield(transform(t)) {
				return
			}
		}
	}
}

// FilterMap returns a sequence obtained after both filtering and mapping using the given transform function.
func FilterMap[T any, R any](s iter.Seq[T], transform func(t T) (R, bool)) iter.Seq[R] {
	return func(yield func(R) bool) {
		for t := range s {
			if r, ok := transform(t); ok && !yield(r) {
				return
			}
		}
	}
}

// FlatMap explodes a sequence by transforming each value into a sequence of another type.
func FlatMap[T any, R any](src iter.Seq[T], transform func(T) iter.Seq[R]) iter.Seq[R] {
	return func(yield func(R) bool) {
		for t := range src {
			for r := range transform(t) {
				if !yield(r) {
					return
				}
			}
		}
	}
}

func Pairs1[T any, R any](s iter.Seq[T], transform func(T) R) iter.Seq2[T, R] {
	return func(yield func(T, R) bool) {
		for t := range s {
			if !yield(t, transform(t)) {
				return
			}
		}
	}
}

func Pairs2[T any, R any](s iter.Seq[T], transform func(T) R) iter.Seq2[R, T] {
	return func(yield func(R, T) bool) {
		for t := range s {
			if !yield(transform(t), t) {
				return
			}
		}
	}
}

// Reduce reduces a sequence to a value using an accumulator function.
func Reduce[T any, R any](s iter.Seq[T], accumulator func(R, T) R, initial R) R {
	for t := range s {
		initial = accumulator(initial, t)
	}

	return initial
}

// Times returns a sequence of count elements produced by repeated calls to generator.
func Times[T any](count int, generator func(int) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := range count {
			if !yield(generator(i)) {
				return
			}
		}
	}
}

// Uniq returns a duplicate-free version of a sequence.
func Uniq[T comparable](s iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[T]struct{})

		for t := range s {
			if _, ok := seen[t]; ok {
				continue
			}

			seen[t] = struct{}{}

			if !yield(t) {
				return
			}
		}
	}
}

// UniqBy returns a duplicate-free version of a sequence, in which the uniqueness of elements is determined by a key
// function.
func UniqBy[T any, U comparable](s iter.Seq[T], key func(T) U) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[U]struct{})

		for t := range s {
			k := key(t)

			if _, ok := seen[k]; ok {
				continue
			}

			seen[k] = struct{}{}

			if !yield(t) {
				return
			}
		}
	}
}
