package wtlfu

import (
	"container/list"
	"wtlfu/basic"
	"wtlfu/bf"
	"wtlfu/cm_sketch"
	"wtlfu/lru"
	"wtlfu/slru"
)

const perItemCountNum = 4
const countRowNum = 4

type WTinyLFU struct {
	cmSketch   *cm_sketch.CmSketch
	doorKeeper *bf.BloomFilter
	tick       int
	threshold  int
	window     *lru.Lru
	mainCatch  *slru.SLru
	data       map[string]*list.Element
}

func NewWTinyLFU(totalSize, threshold int) *WTinyLFU {
	windowSize := totalSize / 100
	if windowSize < 1 {
		windowSize = 1
	}

	mainSize := totalSize - windowSize
	if mainSize < 5 {
		mainSize = 5
	}

	probationSize := mainSize / 5
	protectedSize := mainSize - probationSize

	data := make(map[string]*list.Element, totalSize)

	return &WTinyLFU{
		cmSketch:   cm_sketch.NewCmSketch(uint64(totalSize*4), countRowNum),
		doorKeeper: bf.NewBloomFilter(countRowNum, uint32(totalSize)),
		tick:       0,
		threshold:  threshold,
		window:     lru.NewLru(data, windowSize),
		mainCatch:  slru.NewSLru(data, probationSize, protectedSize),
		data:       data,
	}
}

func (w *WTinyLFU) Get(key string) (string, bool) {
	w.tick++
	if w.tick == w.threshold {
		w.cmSketch.Reset()
		w.doorKeeper.Reset()
		w.tick = 0
	}

	w.cmSketch.Increase(key)

	v, ok := w.data[key]
	if !ok {
		return "", false
	}

	item := v.Value.(*basic.Item)
	value := item.Val

	if item.Belong == basic.Window {
		w.window.Access(v)
	} else {
		w.mainCatch.Access(v)
	}
	return value, true
}

func (w *WTinyLFU) Set(key, value string) {
	if e, ok := w.data[key]; ok {
		item := e.Value.(*basic.Item)
		item.Val = value
		w.cmSketch.Increase(key)

		if item.Belong == basic.Window {
			w.window.Access(e)
		} else {
			w.mainCatch.Access(e)
		}
		return
	}

	newItem := basic.Item{
		Belong: basic.Window,
		Key:    key,
		Val:    value,
	}

	evictItem, evicted := w.window.Add(newItem)
	if !evicted {
		return
	}

	mainVictim := w.mainCatch.Victim()
	if mainVictim == nil {
		w.mainCatch.Add(evictItem)
	}

	if !w.doorKeeper.Check(key) {
		w.doorKeeper.Set(key)
		return
	}

	newNodeCount := w.cmSketch.Estimate(key)
	victimCount := w.cmSketch.Estimate(mainVictim.Key)

	if newNodeCount < victimCount {
		return
	}
	w.mainCatch.Add(newItem)
}
