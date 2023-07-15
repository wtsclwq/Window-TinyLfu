package lru

import (
	"container/list"
	"wtlfu/basic"
)

type Lru struct {
	dataDict map[string]*list.Element
	cap      int
	list     *list.List
}

func NewLru(dataDict map[string]*list.Element, cap int) *Lru {
	l := Lru{
		dataDict: dataDict,
		cap:      cap,
		list:     list.New(),
	}
	return &l
}

func (l *Lru) Access(elem *list.Element) {
	l.list.MoveToFront(elem)
}

func (l *Lru) Add(newItem basic.Item) (basic.Item, bool) {
	if l.list.Len() < l.cap {
		l.dataDict[newItem.Key] = l.list.PushFront(&newItem)
		return basic.Item{}, false
	}

	backElem := l.list.Back()
	backItem := backElem.Value.(*basic.Item)

	delete(l.dataDict, backItem.Key)
	evictItem := *backItem
	// 复用backItem在链表中的node，避免Push
	*backItem = newItem

	l.dataDict[backItem.Key] = backElem
	l.list.MoveToFront(backElem)

	return evictItem, true
}

func (l *Lru) Len() int {
	return len(l.dataDict)
}

func (l *Lru) Remove(key string) (string, bool) {
	v, ok := l.dataDict[key]
	if !ok {
		return "", false
	}
	item := v.Value.(*basic.Item)
	l.list.Remove(v)
	delete(l.dataDict, key)
	return item.Val, true
}
