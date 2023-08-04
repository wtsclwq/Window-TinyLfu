package internal

import "sync/atomic"

type ListType uint8

const (
	ListProbation ListType = iota
	ListProtection
	ListWindow
	ListTimeWheel
	ListUnknown
)

const (
	NEW int8 = iota
	REMOVE
	UPDATE
)

type ReadBufItem[K comparable, V any] struct {
	item *Item[K, V]
	hash uint64
}

type WriteBufItem[K comparable, V any] struct {
	item       *Item[K, V]
	code       int8
	reSchedule bool
}

type Item[K comparable, V any] struct {
	// ListType is the type of list that the item belongs to
	belong ListType

	// shard num
	shardNum uint16

	// key and value
	key K
	val V

	// ddl expire time
	expire atomic.Int64

	// normal list meta data
	_list *List[K, V]
	next  *Item[K, V]
	pre   *Item[K, V]

	// timeWheel list meta data
	wheelPre  *Item[K, V]
	wheelNext *Item[K, V]
}

func NewItem[K comparable, V any](key K, val V, expire int64) *Item[K, V] {
	i := &Item[K, V]{
		key: key,
		val: val,
	}
	if expire > 0 {
		i.expire.Store(expire)
	}
	return i
}

func (i *Item[K, V]) isNew() bool {
	return i._list == nil && i.pre == nil && i.next == nil && i.belong == ListUnknown
}

func (i *Item[K, V]) isNewWheel() bool {
	return i._list == nil && i.wheelPre == nil && i.wheelNext == nil && i.belong == ListUnknown
}

func (i *Item[K, V]) Next(belong ListType) *Item[K, V] {
	switch belong {
	case ListProbation, ListProtection, ListWindow:
		n := i.next
		// because list is a ring list, the back item.next is list.root, but we want nil
		if i._list != nil && &i._list.root != n {
			return n
		}
	case ListTimeWheel:
		n := i.wheelNext
		// because list is a ring list, the back item.next is list.root, but we want nil
		if i._list != nil && &i._list.root != n {
			return n
		}
	}
	return nil
}

func (i *Item[K, V]) Pre(belong ListType) *Item[K, V] {

	switch belong {
	case ListProbation, ListProtection, ListWindow:
		p := i.pre
		// because list is a ring list, the front item.pre is list.root, but we want nil
		if i._list != nil && &i._list.root != p {
			return p
		}
	case ListTimeWheel:
		p := i.wheelPre
		// because list is a ring list, the front item.pre is list.root, but we want nil
		if i._list != nil && &i._list.root != p {
			return p
		}
	}
	return nil
}

func (i *Item[K, V]) setPre(pre *Item[K, V], belong ListType) {
	switch belong {
	case ListProbation, ListProtection, ListWindow:
		i.pre = pre
	case ListTimeWheel:
		i.wheelPre = pre
	}
}

func (i *Item[K, V]) setNext(next *Item[K, V], belong ListType) {
	switch belong {
	case ListProbation, ListProtection, ListWindow:
		i.next = next
	case ListTimeWheel:
		i.wheelNext = next
	}
}

func (i *Item[K, V]) getPrev(listType ListType) *Item[K, V] {
	switch listType {
	case ListProbation, ListProtection, ListWindow:
		return i.pre
	case ListTimeWheel:
		return i.wheelPre
	}
	return nil
}

func (i *Item[K, V]) getNext(listType ListType) *Item[K, V] {
	switch listType {
	case ListProbation, ListProtection, ListWindow:
		return i.next
	case ListTimeWheel:
		return i.wheelNext
	}
	return nil
}
