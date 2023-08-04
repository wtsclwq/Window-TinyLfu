package internal

type List[K comparable, V any] struct {
	listType ListType

	root Item[K, V] // sentinel node

	len int
	cap int
}

func (l *List[K, V]) Init() *List[K, V] {
	l.root = Item[K, V]{} // sentinel node
	l.root.setPre(&l.root, l.listType)
	l.root.setNext(&l.root, l.listType)
	l.root._list = l
	l.len = 0
	return l
}

func NewList[K comparable, V any](cap int, listType ListType) *List[K, V] {
	l := &List[K, V]{
		listType: listType,
		cap:      cap,
	}
	return l.Init()
}

func (l *List[K, V]) layInit() {
	if l.root.next == nil {
		l.Init()
	}
}

func (l *List[K, V]) Len() int {
	return l.len
}

func (l *List[K, V]) Front() *Item[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.root.Next(l.listType)
}

func (l *List[K, V]) Back() *Item[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.root.Pre(l.listType)
}

// insert inserts newItem after atItem
// a <-> b <-> c <-> d   newItem = x, atItem = c   ===>  a <-> b <-> c <->  [x]  <-> d
func (l *List[K, V]) insert(newItem, atItem *Item[K, V]) *Item[K, V] {
	var evicted *Item[K, V]
	if l.len >= l.cap {
		evicted = l.PopBack()
	}

	newItem.belong = l.listType

	newItem.setPre(atItem, l.listType)
	newItem.setNext(atItem.getNext(l.listType), l.listType)

	newItem.getPrev(l.listType).setNext(newItem, l.listType)
	newItem.getNext(l.listType).setPre(newItem, l.listType)

	l.len++

	return evicted
}

// remove removes i in list
func (l *List[K, V]) remove(i *Item[K, V]) {
	i.getPrev(l.listType).setNext(i.getNext(l.listType), l.listType)
	i.getNext(l.listType).setPre(i.getPrev(l.listType), l.listType)

	i.setPre(nil, l.listType)
	i.setNext(nil, l.listType)

	i.belong = ListUnknown

	l.len--
}

// move moves i after at
// a <-> b <-> c <-> d   i = b, at = c   ===>  a  <-> c <-> a  <-> d
func (l *List[K, V]) move(i, at *Item[K, V]) {
	if i == at {
		return
	}

	i.getPrev(l.listType).setNext(i.getNext(l.listType), l.listType)
	i.getNext(l.listType).setPre(i.getPrev(l.listType), l.listType)

	i.setPre(at, l.listType)
	i.setNext(at.getNext(l.listType), l.listType)

	i.getPrev(l.listType).setNext(i, l.listType)
	i.getNext(l.listType).setPre(i, l.listType)
}

// Remove removes i from l if the i is in the list l
func (l *List[K, V]) Remove(i *Item[K, V]) *Item[K, V] {
	if i._list == l {
		l.remove(i)
	}
	return i
}

// PushFront insert a new item i at the front of the list l and return i
func (l *List[K, V]) PushFront(i *Item[K, V]) *Item[K, V] {
	l.layInit()
	return l.insert(i, &l.root)
}

// PushBack insert a new item i at the back of the list l and return i
func (l *List[K, V]) PushBack(i *Item[K, V]) *Item[K, V] {
	l.layInit()
	return l.insert(i, l.root.pre)
}

// MoveToFront moves i to front of list
func (l *List[K, V]) MoveToFront(i *Item[K, V]) {
	if i._list != l || l.root.next == i {
		return
	}
	l.move(i, &l.root)
}

// MoveToBack moves i to back of list
func (l *List[K, V]) MoveToBack(i *Item[K, V]) {
	if i._list != l || l.root.pre == i {
		return
	}
	l.move(i, l.root.pre)
}

func (l *List[K, V]) PopBack() *Item[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.Remove(l.Back())
}

func (l *List[K, V]) PopFront() *Item[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.Remove(l.Front())
}

func (l *List[K, V]) Cap() int {
	return l.cap
}
