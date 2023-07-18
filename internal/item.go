package internal

type ListType uint8

const (
	Probation ListType = iota
	Protection
	Window
	UnKnown
)

type Item[K comparable, V any] struct {
	belong  ListType
	key     K
	val     V
	hashKey string

	_list *List[K, V]
	next  *Item[K, V]
	pre   *Item[K, V]
}

func NewItem[K comparable, V any](key K, val V) *Item[K, V] {
	return &Item[K, V]{key: key, val: val, belong: UnKnown}
}

func (i *Item[K, V]) isNew() bool {
	return i._list == nil && i.pre == nil && i.next == nil && i.belong == UnKnown
}

func (i *Item[K, V]) Next() *Item[K, V] {
	n := i.next

	// because list is a ring list, the back item.next is list.root, but we want nil
	if i._list != nil && &i._list.root != n {
		return n
	}

	return nil
}

func (i *Item[K, V]) Pre() *Item[K, V] {
	p := i.pre

	// because list is a ring list, the front item.pre is list.root, but we want nil
	if i._list != nil && &i._list.root != p {
		return p
	}

	return nil
}
