package mempool

import (
	"sync"
)

const defaultSliceCapacity = 2048

type Pool[T any] struct {
	sync.Pool
}

func (p *Pool[T]) Get() T {
	t, _ := p.Pool.Get().(T)
	return t
}

func (p *Pool[T]) Put(x T) {
	p.Pool.Put(x)
}

func NewPool[T any](newF func() T) *Pool[T] {
	return &Pool[T]{
		Pool: sync.Pool{
			New: func() any {
				return newF()
			},
		},
	}
}

func NewSlicePool[T any](capacity int) *Pool[*[]T] {
	return NewPool(func() *[]T {
		s := make([]T, 0, capacity)
		return &s
	})
}

type Allocator[T any] interface {
	New() T
	Reset(m T)
}

type CollectionPool[T any] struct {
	slicePool *Pool[*[]T]
	msgPool   *Pool[T]
	alloc     Allocator[T]
}

func NewCollectionPool[T any](alloc Allocator[T]) *CollectionPool[T] {
	return &CollectionPool[T]{
		slicePool: NewSlicePool[T](defaultSliceCapacity),
		alloc:     alloc,
		msgPool: NewPool(func() T {
			return alloc.New()
		}),
	}
}

func (p CollectionPool[T]) GetSlice() *[]T {
	return p.slicePool.Get()
}

func (p *CollectionPool[T]) PutSlice(s *[]T) {
	for _, item := range *s {
		p.alloc.Reset(item)
		p.msgPool.Put(item)
	}

	*s = (*s)[:0]
	p.slicePool.Put(s)
}

func (p *CollectionPool[T]) Get() T {
	t, _ := p.msgPool.New().(T)
	return t
}

func (p *CollectionPool[T]) Put(t T) {
	p.msgPool.Put(t)
}
