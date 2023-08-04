package internal

type SLru[K comparable, V any] struct {
	firstSegment  *List[K, V]
	secondSegment *List[K, V]
	cap           int
}

func newSLru[K comparable, V any](cap int) *SLru[K, V] {
	fc := cap / 5
	sc := cap - fc
	slru := SLru[K, V]{
		firstSegment:  NewList[K, V](fc, ListProbation),
		secondSegment: NewList[K, V](sc, ListProtection),
		cap:           cap,
	}
	return &slru
}

// add try to add a new Item into probation at front , and check if the probation list is full
func (s *SLru[K, V]) add(i *Item[K, V]) *Item[K, V] {
	var evicted *Item[K, V]
	if s.firstSegment.Len()+s.secondSegment.Len() >= s.cap {
		evicted = s.firstSegment.PopBack()
	}
	i.belong = ListProbation
	evicted = s.firstSegment.PushFront(i)
	return evicted
}

// access accesses an item and update the order
func (s *SLru[K, V]) access(i *Item[K, V]) {
	// 如果访问的是cache队列的，就单纯调整一下位置
	switch i.belong {
	case ListProbation:
		// If access an item in probation segment, just move it to the protection segment
		s.firstSegment.remove(i)
		// If probation segment is full, evict one into probation segment
		i.belong = ListProtection
		evicted := s.secondSegment.PushFront(i)
		if evicted != nil {
			evicted.belong = ListProbation
			s.secondSegment.PushFront(evicted)
		}
	case ListProtection:
		// If access an item in protection segment, adjust the order
		s.secondSegment.MoveToFront(i)
	default:
		panic("error")
	}
}

// maybeVictim returns the victim item if the slru is full
func (s *SLru[K, V]) maybeVictim() *Item[K, V] {
	if s.firstSegment.Len()+s.secondSegment.Len() < s.cap {
		return nil
	}
	return s.firstSegment.Back()
}

// remove removes an item from slru
func (s *SLru[K, V]) remove(i *Item[K, V]) {
	switch i.belong {
	case ListProbation:
		s.firstSegment.remove(i)
	case ListProtection:
		s.secondSegment.remove(i)
	}
}

// len returns the number of items both in probation and protection
func (s *SLru[K, V]) len() int {
	return s.firstSegment.Len() + s.secondSegment.Len()
}
