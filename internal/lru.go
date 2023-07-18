package internal

type Lru[K comparable, V any] struct {
	list *List[K, V]
	cap  int
}

func NewLru[K comparable, V any](cap int) *Lru[K, V] {
	l := Lru[K, V]{
		list: NewList[K, V](),
		cap:  cap,
	}
	return &l
}

func (l *Lru[K, V]) access(i *Item[K, V]) {
	l.list.MoveToFront(i)
}

// Add try to add a new Item into lru list at front, and check if the lru list is full
// return true and evicted item if the lru list is full
func (l *Lru[K, V]) Add(i *Item[K, V]) (*Item[K, V], bool) {
	if l.list.Len() < l.cap {
		l.list.PushFront(i)
		return &Item[K, V]{}, false
	}
	evictItem := l.list.Back()
	l.list.PushFront(i)
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
	return l.cap
}
