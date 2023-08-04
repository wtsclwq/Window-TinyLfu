package internal

import (
	"sync/atomic"
)

type TinyLFU[K comparable, V any] struct {
	cap       int
	mainCache *SLru[K, V]
	sketch    *cmSketch

	hitCount atomic.Uint32

	counter   atomic.Uint32
	threshold uint32

	hashKey *HashKey[K]
}

func NewTinyLFU[K comparable, V any](cap int, hashKey *HashKey[K]) *TinyLFU[K, V] {
	return &TinyLFU[K, V]{
		cap:       cap,
		mainCache: newSLru[K, V](cap),
		sketch:    newCmSketch(int64(cap)),
		hashKey:   hashKey,
	}
}

func (t *TinyLFU[K, V]) Set(i *Item[K, V]) *Item[K, V] {
	// if it is a new item, add it to the main cache
	if i.isNew() {
		if victim := t.mainCache.maybeVictim(); victim != nil {
			freq := t.sketch.estimate(t.hashKey.Hash(i.key))
			victimFreq := t.sketch.estimate(t.hashKey.Hash(victim.key))
			if freq <= victimFreq {
				// 如果从Window淘汰的freq还不如mainCache淘汰的，直接返回
				return i
			}
		}
		evicted := t.mainCache.add(i)
		return evicted
	}
	return nil
}

// Access accesses an item in main cache
func (t *TinyLFU[K, V]) Access(ri ReadBufItem[K, V]) {
	if item := ri.item; item != nil {
		t.sketch.increment(ri.hash)
		if !item.isNew() {
			t.mainCache.access(item)
		}
	}
}

// Remove removes an item from main cache
func (t *TinyLFU[K, V]) Remove(i *Item[K, V]) {
	t.mainCache.remove(i)
}

func (t *TinyLFU[K, V]) EvictEntries() []*Item[K, V] {
	var removed []*Item[K, V]

	for t.mainCache.firstSegment.Len()+t.mainCache.secondSegment.Len() > t.mainCache.cap {
		entry := t.mainCache.firstSegment.PopBack()
		if entry == nil {
			break
		}
		removed = append(removed, entry)
	}
	for t.mainCache.firstSegment.Len()+t.mainCache.secondSegment.Len() > t.mainCache.cap {
		entry := t.mainCache.secondSegment.PopBack()
		if entry == nil {
			break
		}
		removed = append(removed, entry)
	}
	return removed
}
