package internal

import (
	"math/bits"
	"time"
)

type Clock struct {
	start time.Time
}

func (c *Clock) nowNano() int64 {
	return time.Since(c.start).Nanoseconds()
}

func (c *Clock) expireNano(ttl time.Duration) int64 {
	return c.nowNano() + ttl.Nanoseconds()
}

type TimerWheel[K comparable, V any] struct {
	buckets []uint
	spans   []int64
	shift   []uint
	wheel   [][]*List[K, V]
	clock   *Clock
	nanos   int64
}

func NewTimerWheel[K comparable, V any](size uint) *TimerWheel[K, V] {
	clock := &Clock{start: time.Now()}
	buckets := []uint{64, 64, 32, 4, 1}
	spans := []int64{
		next2Power(((1 * time.Second).Nanoseconds())),
		next2Power(((1 * time.Minute).Nanoseconds())),
		next2Power(((1 * time.Hour).Nanoseconds())),
		next2Power(((24 * time.Hour).Nanoseconds())),
		next2Power(((24 * time.Hour).Nanoseconds())) * 4,
		next2Power(((24 * time.Hour).Nanoseconds())) * 4,
	}

	shift := []uint{
		uint(bits.TrailingZeros64(uint64(spans[0]))),
		uint(bits.TrailingZeros64(uint64(spans[1]))),
		uint(bits.TrailingZeros64(uint64(spans[2]))),
		uint(bits.TrailingZeros64(uint64(spans[3]))),
		uint(bits.TrailingZeros64(uint64(spans[4]))),
	}

	wheel := [][]*List[K, V]{}
	for i := 0; i < 5; i++ {
		tmp := []*List[K, V]{}
		for j := 0; j < int(buckets[i]); j++ {
			tmp = append(tmp, NewList[K, V](0, TIMEWHELL))
		}
		wheel = append(wheel, tmp)
	}

	return &TimerWheel[K, V]{
		buckets: buckets,
		spans:   spans,
		shift:   shift,
		wheel:   wheel,
		nanos:   clock.nowNano(),
		clock:   clock,
	}

}

func (tw *TimerWheel[K, V]) findIndex(expire int64) (int, int) {
	duration := expire - tw.nanos
	for i := 0; i < 5; i++ {
		if duration < int64(tw.spans[i+1]) {
			ticks := expire >> int(tw.shift[i])
			slot := int(ticks) & (int(tw.buckets[i]) - 1)
			return i, slot
		}
	}
	return 4, 0
}

func (tw *TimerWheel[K, V]) deschedule(item *Item[K, V]) {
	item.Pre(TIMEWHELL).setNext(item.Next(TIMEWHELL), TIMEWHELL)
	item.Next(TIMEWHELL).setPre(item.Pre(TIMEWHELL), TIMEWHELL)
	item.setNext(nil, TIMEWHELL)
	item.setPre(nil, TIMEWHELL)
}

func (tw *TimerWheel[K, V]) schedule(item *Item[K, V]) {
	if item.wheelPre != nil {
		tw.deschedule(item)
	}
	x, y := tw.findIndex(item.expire.Load())
	tw.wheel[x][y].PushFront(item)
}

func (tw *TimerWheel[K, V]) advance(now int64, remove func(item *Item[K, V], reason RemoveReason)) {
	if now == 0 {
		now = tw.clock.nowNano()
	}
	previous := tw.nanos
	tw.nanos = now

	for i := 0; i < 5; i++ {
		prevTicks := previous >> int64(tw.shift[i])
		currentTicks := tw.nanos >> int64(tw.shift[i])
		if currentTicks <= prevTicks {
			break
		}
		tw.expire(i, prevTicks, currentTicks-prevTicks, remove)
	}
}

func (tw *TimerWheel[K, V]) expire(index int, prevTicks int64, delta int64, remove func(item *Item[K, V], reason RemoveReason)) {
	mask := tw.buckets[index] - 1
	steps := tw.buckets[index]
	if delta < int64(steps) {
		steps = uint(delta)
	}
	start := prevTicks & int64(mask)
	end := start + int64(steps)
	for i := start; i < end; i++ {
		list := tw.wheel[index][i&int64(mask)]
		item := list.Front()
		for item != nil {
			next := item.Next(TIMEWHELL)
			if item.expire.Load() <= tw.nanos {
				tw.deschedule(item)
				remove(item, EXPIRED)
			} else {
				tw.schedule(item)
			}
			item = next

		}
		list.Init()
	}
}
