package internal

type Lru[K comparable, V any] struct {
	list *List[K, V]
}

func NewLru[K comparable, V any](cap int) *Lru[K, V] {
	l := Lru[K, V]{
		list: NewList[K, V](cap, ListWindow),
	}
	return &l
}

func (l *Lru[K, V]) access(i *Item[K, V]) {
	l.list.MoveToFront(i)
}

// Add try to add a new Item into lru list at front, and check if the lru list is full
// return true and evicted item if the lru list is full
func (l *Lru[K, V]) Add(i *Item[K, V]) (*Item[K, V], bool) {
	i.belong = l.list.listType
	evictItem := l.list.PushFront(i)
	if evictItem == nil {
		return nil, false
	}
	return evictItem, true
}

// Remove removes an Item from lru list
func (l *Lru[K, V]) Remove(i *Item[K, V]) {
	l.list.Remove(i)
}

func (l *Lru[K, V]) Len() int {
	return l.list.Len()
}

func (l *Lru[K, V]) Cap() int {
	return l.list.Cap()
}
