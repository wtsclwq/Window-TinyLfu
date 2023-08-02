package internal

import (
	"hash"

	"github.com/spaolacci/murmur3"
)

const hashNum = 4
const seed uint32 = 64

type WTinyLFU[K comparable, V any] struct {
	window    *Lru[K, V]
	mainCache *SLru[K, V]

	cmSketch   *CmSketch
	doorKeeper *BloomFilter

	tick      int
	threshold int

	hashKey *HashKey[K]
}

func NewWTinyLFU[K comparable, V any](totalSize, threshold int) *WTinyLFU[K, V] {
	windowSize := totalSize / 100
	if windowSize < 1 {
		windowSize = 1
	}

	mainSize := totalSize - windowSize
	if mainSize < 5 {
		mainSize = 5
	}

	var hashHelpers = make([]hash.Hash32, 0, hashNum)
	for i := 0; i < hashNum; i++ {
		hashHelpers = append(hashHelpers, murmur3.New32WithSeed(seed+uint32(i)))
	}

	return &WTinyLFU[K, V]{
		window:    NewLru[K, V](windowSize),
		mainCache: newSLru[K, V](mainSize),

		cmSketch:   NewCmSketch(totalSize*2, hashHelpers),
		doorKeeper: NewBloomFilter(hashNum, uint32(totalSize)),

		tick:      0,
		threshold: threshold,

		hashKey: NewHash[K](),
	}
}

// Access will try to access an item and update the order
func (w *WTinyLFU[K, V]) Access(i *Item[K, V]) {
	w.tick++
	if w.tick == w.threshold {
		w.cmSketch.Reset()
		w.doorKeeper.Reset()
		w.tick = 0
	}
	if i.isNew() {
		return
	}
	w.cmSketch.Increase(i.hashKey)
	if i.belong == Window {
		w.window.access(i)
	} else {
		w.mainCache.access(i)
	}
}

// Set will try to add a new item into W-TinyLFU
func (w *WTinyLFU[K, V]) Set(i *Item[K, V]) {
	w.tick++
	if w.tick > w.threshold {
		w.cmSketch.Reset()
		w.doorKeeper.Reset()
		w.tick = 0
	}

	if !i.isNew() {
		return
	}

	i.hashKey = w.hashKey.Hash(i.key)

	evictItem, hasEvicted := w.window.Add(i)
	if !hasEvicted {
		return
	}

	// If main cache is not full, i.e. main cache will not return a victim
	mainVictim := w.mainCache.maybeVictim()
	if mainVictim == nil {
		w.mainCache.add(evictItem)
	}

	// If main cache is full, i.e. main will return a victim.
	// We need to compare the evicted item which from window with the mainVictim

	// If doorKeeper can't find the record of the evicted item, i.e. the evicted item have not been access
	if !w.doorKeeper.Check(evictItem.hashKey) {
		w.doorKeeper.Set(i.hashKey)
		return
	}

	// Compare frequency
	newNodeCount := w.cmSketch.Estimate(evictItem.hashKey)
	victimCount := w.cmSketch.Estimate(mainVictim.hashKey)
	if newNodeCount >= victimCount {
		w.mainCache.add(evictItem)
	}
}
