package internal

type List[K comparable, V any] struct {
	root Item[K, V] // sentinel node

	len int
}

func (l *List[K, V]) Init() *List[K, V] {
	l.root.next = &l.root
	l.root.pre = &l.root
	l.len = 0
	return l
}

func NewList[K comparable, V any]() *List[K, V] {
	return new(List[K, V]).Init()
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
	return l.root.next
}

func (l *List[K, V]) Back() *Item[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.root.pre
}

// insert inserts newItem after atItem
// a <-> b <-> c <-> d   newItem = x, atItem = c   ===>  a <-> b <-> c <->  [x]  <-> d
func (l *List[K, V]) insert(newItem, atItem *Item[K, V]) *Item[K, V] {
	newItem.pre = atItem
	newItem.next = atItem.next

	newItem.pre.next = newItem
	newItem.next.pre = newItem

	newItem._list = l
	l.len++

	return newItem
}

// remove removes i in list
func (l *List[K, V]) remove(i *Item[K, V]) {
	i.pre.next = i.next
	i.next.pre = i.pre

	i.next = nil
	i.pre = nil
	i._list = nil
	l.len--
}

// move moves i after at
// a <-> b <-> c <-> d   i = b, at = c   ===>  a  <-> c <-> a  <-> d
func (l *List[K, V]) move(i, at *Item[K, V]) {
	if i == at {
		return
	}

	i.pre.next = i.next
	i.next.pre = i.pre

	i.pre = at
	i.next = at.next

	i.pre.next = i
	i.next.pre = i
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
