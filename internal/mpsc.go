package internal

import (
	"sync"
	"sync/atomic"
)

type node[V any] struct {
	next *node[V]
	val  V
}

type Queue[V any] struct {
	head, tail atomic.Value // 使用 atomic.Value 替代 atomic.Pointer
	nodePool   sync.Pool
}

func NewQueue[V any]() *Queue[V] {
	q := &Queue[V]{nodePool: sync.Pool{New: func() any {
		return new(node[V])
	}}}
	stub := &node[V]{}
	q.head.Store(stub) // 使用 Store 方法替代直接赋值
	q.tail.Store(stub) // 使用 Store 方法替代直接赋值
	return q
}

func (q *Queue[V]) Push(x V) {
	n := q.nodePool.Get().(*node[V])
	n.val = x
	// current producer acquires head node
	prev := q.head.Swap(n).(*node[V]) // 使用 Swap 方法替代直接赋值

	// release node to consumer
	prev.next = n // 不再需要使用 atomic.StorePointer
}

func (q *Queue[V]) Pop() (V, bool) {
	tail := q.tail.Load().(*node[V]) // 使用 Load 方法替代直接赋值
	next := tail.next
	if next != nil {
		var null V
		v := next.val
		next.val = null
		q.tail.Store(next) // 使用 Store 方法替代直接赋值
		tail.next = nil
		q.nodePool.Put(tail)
		return v, true
	}
	var null V
	return null, false
}

func (q *Queue[V]) Empty() bool {
	tail := q.tail.Load().(*node[V]) // 使用 Load 方法替代直接赋值
	return tail.next == nil
}
