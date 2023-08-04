package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDeqeExpiure(t *testing.T) {
	store := NewStore[int, int](20000)

	expired := map[int]int{}
	store.removalListener = func(key, value int, reason RemoveReason) {
		if reason == EXPIRED {
			expired[key] = value
		}
	}
	_, index := store.index(123)
	expire := store.timerWheel.clock.expireNano(200 * time.Millisecond)
	for i := 0; i < 50; i++ {
		entry := &Item[int, int]{key: i}
		entry.expire.Store(expire)
		store.shards[index].window.Add(entry)
		store.shards[index].dict[i] = entry
	}
	require.True(t, len(expired) == 0)
	time.Sleep(1 * time.Second)
	store.Set(123, 123, 1*time.Second)
	require.True(t, len(expired) > 0)
}
