package internal

import "sync/atomic"

type ListType uint8

const (
	PROBATION ListType = iota
	PROTECTION
	WINDOW
	TIMEWHELL
	UNKNOWN
)

const (
	NEW int8 = iota
	REMOVE
	UPDATE
)

type ReadBufItem[K comparable, V any] struct {
	entry *Item[K, V]
	hash  uint64
}

type WriteBufItem[K comparable, V any] struct {
	entry     *Item[K, V]
	code      int8
	rechedule bool
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

	// timewhell list meta data
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
	return i._list == nil && i.pre == nil && i.next == nil && i.belong == UNKNOWN
}

func (i *Item[K, V]) Next(belong ListType) *Item[K, V] {
	switch belong {
	case PROBATION, PROTECTION, WINDOW:
		n := i.next
		// because list is a ring list, the back item.next is list.root, but we want nil
		if i._list != nil && &i._list.root != n {
			return n
		}
	case TIMEWHELL:
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
	case PROBATION, PROTECTION, WINDOW:
		p := i.pre
		// because list is a ring list, the front item.pre is list.root, but we want nil
		if i._list != nil && &i._list.root != p {
			return p
		}
	case TIMEWHELL:
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
	case PROBATION, PROTECTION, WINDOW:
		i.pre = pre
	case TIMEWHELL:
		i.wheelPre = pre
	}
}

func (i *Item[K, V]) setNext(next *Item[K, V], belong ListType) {
	switch belong {
	case PROBATION, PROTECTION, WINDOW:
		i.next = next
	case TIMEWHELL:
		i.wheelNext = next
	}
}
