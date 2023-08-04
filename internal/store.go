package internal

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	MaxReadBuffSize  = 64
	MinWriteBuffSize = 4
	MaxWriteBuffSize = 1024
)

type RemoveReason uint8

const (
	REMOVED RemoveReason = iota
	EVICTED
	EXPIRED
)

type Shard[K comparable, V any] struct {
	dict       map[K]*Item[K, V]
	cap        int
	windowCap  int
	window     *Lru[K, V]
	doorkeeper *bloomFilter
	dkCounter  int
	mu         sync.RWMutex
}

func newShard[K comparable, V any](cap, windowCap int) *Shard[K, V] {
	return &Shard[K, V]{
		dict:       make(map[K]*Item[K, V], cap),
		doorkeeper: newBloomFilter(20*cap, 0.01),
		cap:        cap,
		windowCap:  windowCap,
		window:     NewLru[K, V](windowCap),
	}
}

func (s *Shard[K, V]) get(key K) (*Item[K, V], bool) {
	if item, ok := s.dict[key]; ok {
		return item, true
	}
	return nil, false
}

func (s *Shard[K, V]) set(i *Item[K, V]) {
	s.dict[i.key] = i
}

func (s *Shard[K, V]) delete(i *Item[K, V]) bool {
	var deleted bool
	exist, ok := s.dict[i.key]
	if ok && exist == i {
		delete(s.dict, i.key)
		deleted = true
	}
	return deleted
}

type Store[K comparable, V any] struct {
	cap             int
	shards          []*Shard[K, V]
	hash            *HashKey[K]
	shardNum        int
	policy          *TinyLFU[K, V]
	timerWheel      *TimerWheel[K, V]
	readBuf         *Queue[ReadBufItem[K, V]]
	readCounter     *atomic.Uint32
	writeBuf        chan WriteBufItem[K, V]
	itemPoll        sync.Pool
	mu              sync.Mutex
	closed          bool
	removalListener func(key K, value V, reason RemoveReason)
}

func NewStore[K comparable, V any](cap int) *Store[K, V] {
	hashKey := NewHash[K]()
	writeBufSize := cap / 100
	if writeBufSize < MinWriteBuffSize {
		writeBufSize = MinWriteBuffSize
	}
	if writeBufSize > MaxWriteBuffSize {
		writeBufSize = MaxWriteBuffSize
	}
	shardNum := 1
	for shardNum < runtime.NumCPU() {
		shardNum *= 2
	}

	shardSize := cap / shardNum
	windowSize := cap / 100 / shardNum
	if shardSize < 50 {
		shardSize = 100
	}
	mainCacheSize := cap - windowSize*shardNum

	s := &Store[K, V]{
		cap:        cap,
		shards:     make([]*Shard[K, V], 0, shardNum),
		shardNum:   shardNum,
		hash:       hashKey,
		policy:     NewTinyLFU[K, V](mainCacheSize, hashKey),
		readBuf:    NewQueue[ReadBufItem[K, V]](),
		writeBuf:   make(chan WriteBufItem[K, V], writeBufSize),
		itemPoll:   sync.Pool{New: func() interface{} { return &Item[K, V]{} }},
		timerWheel: NewTimerWheel[K, V](uint(cap)),
	}
	for i := 0; i < s.shardNum; i++ {
		s.shards = append(s.shards, newShard[K, V](shardSize, windowSize))
	}
	go s.maintenance()
	return s
}

// spread hash before get index
func (s *Store[K, V]) index(key K) (uint64, uint16) {
	base := s.hash.Hash(key)
	h := ((base >> 16) ^ base) * 0x45d9f3b
	h = ((h >> 16) ^ h) * 0x45d9f3b
	h = (h >> 16) ^ h
	return base, uint16(h & uint64(s.shardNum-1))
}

// drainRead drain read buffer, and access all items
func (s *Store[K, V]) drainRead() {
	s.mu.Lock()
	for {
		v, ok := s.readBuf.Pop()
		if !ok {
			break
		}
		s.policy.Access(v)
	}
	s.mu.Unlock()
	s.readCounter.Store(0)
}

func (s *Store[K, V]) Get(key K) (V, bool) {
	// tick，每次操作都会增加一个计数器
	s.policy.counter.Add(1)

	h, index := s.index(key)
	shard := s.shards[index]
	readCount := s.readCounter.Add(1)

	shard.mu.RLock()
	defer shard.mu.RUnlock()

	item, ok := shard.get(key)
	var res V
	if ok {
		expire := item.expire.Load()
		if expire != 0 && expire < s.timerWheel.clock.nowNano() {
			// 如果这是一个DDL的缓存项目，并且已经过期，那么Get失败
			ok = false
		} else {
			s.policy.hitCount.Add(1)
			res = item.val
		}
	}
	switch {
	case readCount < MaxReadBuffSize:
		var send ReadBufItem[K, V]
		send.hash = h
		if ok {
			send.item = item
		}
		s.readBuf.Push(send)
	case readCount == MaxReadBuffSize:
		s.drainRead()
	}
	return res, ok
}

func (s *Store[K, V]) Set(key K, val V, ttl time.Duration) bool {
	s.policy.counter.Add(1)

	h, index := s.index(key)
	shard := s.shards[index]

	var expire int64
	if ttl > 0 {
		// 计算到期时间
		expire = s.timerWheel.clock.expireNano(ttl)
	}

	shard.mu.Lock()
	defer shard.mu.Unlock()

	item, ok := shard.get(key)
	if ok {
		// 如果存在，那么更新
		var reScheduler bool
		item.val = val
		if expire != 0 {
			// 原子操作，更新过期时间
			oldExpire := item.expire.Swap(expire)
			// 如果过期时间不一样，那么需要重新调度
			if oldExpire != expire {
				reScheduler = true
			}
		}
		if reScheduler {
			s.writeBuf <- WriteBufItem[K, V]{
				item:       item,
				code:       UPDATE,
				reSchedule: true,
			}
		}
		return true
	}
	// 如果不存在，需要先加入window
	// 非更新的set操作，需要判断是否触发保鲜机制
	if shard.dkCounter >= shard.cap {
		// 触发保险机制
		shard.doorkeeper.reset()
		shard.dkCounter = 0
	}

	// 如果这是第一次插入，直接忽略，无法通过doorkeeper
	hit := shard.doorkeeper.insert(h)
	if !hit {
		shard.dkCounter++
		return false
	}

	// 如果通过了doorkeeper，那么就可以插入了
	item = s.itemPoll.Get().(*Item[K, V])
	item.key = key
	item.val = val
	item.expire.Store(expire)
	item.shardNum = index
	shard.set(item)

	if evicted, isEvicted := shard.window.Add(item); isEvicted {
		// 如果window满了，那么需要将evicted的item从shard中删除并且尝试假如到policy中
		expire := evicted.expire.Load()
		if expire > 0 && expire < s.timerWheel.clock.nowNano() {
			// 如果被window剔除的已经过期，那么直接删除，放回到poll
			deleted := shard.delete(evicted)
			if deleted {
				s.postDelete(item)
				if s.removalListener != nil {
					s.removalListener(evicted.key, evicted.val, EXPIRED)
				}
			}
		} else {
			// 如果没有过期，那么需要尝试加入到policy中
			s.writeBuf <- WriteBufItem[K, V]{
				item: evicted,
				code: NEW,
			}
		}
	}
	return true
}

func (s *Store[K, V]) Delete(key K) {
	_, index := s.index(key)
	shard := s.shards[index]

	shard.mu.Lock()
	defer shard.mu.Unlock()

	item, ok := shard.get(key)
	if ok {
		shard.delete(item)
		s.writeBuf <- WriteBufItem[K, V]{
			item: item,
			code: REMOVE,
		}
	}
}

func (s *Store[K, V]) postDelete(item *Item[K, V]) {
	var zero V
	item.val = zero
	s.itemPoll.Put(item)
}

// remove item from cache/policy/timeWheel and add back to pool
func (s *Store[K, V]) removeItem(item *Item[K, V], reason RemoveReason) {
	if !item.isNew() {
		s.policy.Remove(item)
	}
	if !item.isNewWheel() {
		s.timerWheel.deSchedule(item)
	}

	var k K
	var v V
	switch reason {
	case EVICTED, EXPIRED:
		shard := s.shards[item.shardNum]
		shard.mu.Lock()
		deleted := shard.delete(item)
		shard.mu.Unlock()
		if deleted {
			k, v = item.key, item.val
			if s.removalListener != nil {
				s.removalListener(k, v, reason)
			}
			s.postDelete(item)
		}
	// already removed from shard map
	case REMOVED:
		shard := s.shards[item.shardNum]
		shard.mu.RLock()
		k, v = item.key, item.val
		shard.mu.RUnlock()
		if s.removalListener != nil {
			s.removalListener(k, v, reason)
		}
	}
}

func (s *Store[K, V]) maintenance() {
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			s.mu.Lock()
			if s.closed {
				s.mu.Unlock()
				return
			}
			s.timerWheel.advance(0, s.removeItem)
			s.mu.Unlock()
		}
	}()

	for writeItem := range s.writeBuf {
		s.mu.Lock()
		item := writeItem.item
		if item == nil {
			s.mu.Unlock()
			continue
		}

		// lock free because store API never read/modify item metadata
		switch writeItem.code {
		case NEW:
			if item.expire.Load() != 0 {
				s.timerWheel.schedule(item)
			}
			evicted := s.policy.Set(item)
			if evicted != nil {
				s.removeItem(evicted, EVICTED)
			}
			removed := s.policy.EvictEntries()
			for _, e := range removed {
				s.removeItem(e, EVICTED)
			}
		case REMOVE:
			s.removeItem(item, REMOVED)
		case UPDATE:
			if writeItem.reSchedule {
				s.timerWheel.schedule(item)
			}
		}
		writeItem.item = nil
		s.mu.Unlock()
	}
}

func (s *Store[K, V]) Close() {
	for _, shard := range s.shards {
		shard.mu.RLock()
		shard.dict = nil
		shard.mu.RUnlock()
	}
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	close(s.writeBuf)
}
