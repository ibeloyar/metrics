package pool

import (
	"container/list"
	"sync"
)

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	mu    sync.Mutex
	items *list.List
}

func New[T Resetter]() *Pool[T] {
	return &Pool[T]{
		mu:    sync.Mutex{},
		items: list.New(),
	}
}

func (p *Pool[T]) Put(obj T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	obj.Reset()

	p.items.PushBack(obj)
}

func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if first := p.items.Front(); first != nil {
		p.items.Remove(first)

		return first.Value.(T)
	}

	return *new(T)
}

func (p *Pool[T]) Len() int {
	return p.items.Len()
}
