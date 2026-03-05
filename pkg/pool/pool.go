package pool

import (
	"sync"
)

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	items sync.Pool
}

func New[T Resetter]() *Pool[T] {
	return &Pool[T]{
		items: sync.Pool{
			New: func() any {
				return new(T)
			},
		},
	}
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()

	p.items.Put(&obj)
}

func (p *Pool[T]) Get() T {
	obj := p.items.Get().(*T)

	return *obj
}
