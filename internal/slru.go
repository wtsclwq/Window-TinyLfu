package internal

type SLru[K comparable, V any] struct {
	firstSegment  *List[K, V]
	secondSegment *List[K, V]
	firstCap      int
	secondCap     int
}

func newSLru[K comparable, V any](cap int) *SLru[K, V] {
	fc := cap / 5
	sc := cap - fc
	slru := SLru[K, V]{
		firstSegment:  NewList[K, V](),
		secondSegment: NewList[K, V](),
		firstCap:      fc,
		secondCap:     sc,
	}
	return &slru
}

// add try to add a new Item into probation at front , and check if the probation list is full
func (s *SLru[K, V]) add(i *Item[K, V]) *Item[K, V] {
	i.belong = Probation
	var evicted *Item[K, V]
	if s.firstSegment.Len() >= s.firstCap {
		evicted = s.firstSegment.Back()
	}
	s.firstSegment.PushFront(i)
	return evicted
}

// access accesses an item and update the order
func (s *SLru[K, V]) access(i *Item[K, V]) {
	// 如果访问的是cache队列的，就单纯调整一下位置
	switch i.belong {
	case Probation:
		// If access an item in probation segment, just move it to the protection segment
		s.firstSegment.remove(i)
		// If probation segment is full, evict one into probation segment
		if s.secondSegment.Len() >= s.secondCap {
			evictedItem := s.secondSegment.Back()
			evictedItem.belong = Probation
			s.firstSegment.PushFront(evictedItem)
		}
		i.belong = Protection
		s.secondSegment.PushFront(i)
	case Protection:
		// If access an item in protection segment, adjust the order
		s.secondSegment.PushFront(i)
	case Window:
		panic("error")
	}
}

// maybeVictim returns the victim item if the slru is full
func (s *SLru[K, V]) maybeVictim() *Item[K, V] {
	if s.len() < s.firstCap+s.secondCap {
		return nil
	}
	return s.firstSegment.Back()
}

// len returns the number of items both in probation and protection
func (s *SLru[K, V]) len() int {
	return s.firstSegment.Len() + s.secondSegment.Len()
}
