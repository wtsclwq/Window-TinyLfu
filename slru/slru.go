package slru

import (
	"container/list"
	"wtlfu/basic"
	"wtlfu/util"
)

type SLru struct {
	dataDict       map[string]*list.Element
	probationList  *list.List
	protectionList *list.List
	probationCap   int
	protectionCap  int
}

func NewSLru(dataDict map[string]*list.Element, freeListCap int, cacheListCap int) *SLru {
	slru := SLru{
		dataDict:       dataDict,
		probationList:  list.New(),
		protectionList: list.New(),
		probationCap:   freeListCap,
		protectionCap:  cacheListCap,
	}
	return &slru
}

func (s *SLru) Access(elem *list.Element) {
	item := elem.Value.(*basic.Item)
	// 如果访问的是cache队列的，就单纯调整一下位置
	if item.Belong == basic.Probation {
		s.protectionList.MoveToFront(elem)
		return
	}

	util.Assert(item.Belong == basic.Probation)
	pbElem := elem
	pbItem := item

	// 如果访问的是probation队列，需要将其升级到protection
	if s.protectionList.Len() < s.protectionCap {
		s.probationList.Remove(pbElem)
		pbItem.Belong = basic.Protection
		s.dataDict[pbItem.Key] = s.protectionList.PushFront(pbElem)
		return
	}

	// 当然如果protection满了，就需要从protection下放一个，当前item升级上去
	ptBackElem := s.protectionList.Back()
	ptBackItem := ptBackElem.Value.(*basic.Item)

	*ptBackItem, *pbItem = *pbItem, *ptBackItem
	// 注意下面的ptBackItem和pbItem已经交换
	ptBackItem.Belong = basic.Protection
	pbItem.Belong = basic.Probation

	s.dataDict[item.Key] = elem
	s.dataDict[ptBackItem.Key] = ptBackElem

	s.probationList.MoveToFront(elem)
	// 下放的ptBachItem在pb中仍然是队头（大厂毕业去小厂还能当leader？hah）
	s.protectionList.MoveToFront(ptBackElem)
}

func (s *SLru) Add(newItem basic.Item) {
	newItem.Belong = 1
	if s.probationList.Len() < s.probationCap {
		s.dataDict[newItem.Key] = s.probationList.PushFront(&newItem)
	}

	backElem := s.probationList.Back()
	backItem := backElem.Value.(*basic.Item)
	delete(s.dataDict, backItem.Key)

	*backItem = newItem
	s.dataDict[backItem.Key] = backElem
	s.probationList.MoveToFront(backElem)
}

func (s *SLru) Len() int {
	return s.probationList.Len() + s.protectionList.Len()
}

func (s *SLru) Remove(key string) (string, bool) {
	v, ok := s.dataDict[key]
	if !ok {
		return "", false
	}

	item := v.Value.(*basic.Item)

	if item.Belong == basic.Protection {
		s.protectionList.Remove(v)
	} else {
		s.probationList.Remove(v)
	}
	delete(s.dataDict, key)
	return item.Val, true
}

func (s *SLru) Victim() *basic.Item {
	if s.Len() < s.probationCap+s.protectionCap {
		return nil
	}
	v := s.probationList.Back()
	return v.Value.(*basic.Item)
}
