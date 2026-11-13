package cache_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"cachettl/internal/cache"
)

type entry struct {
	key string
	val string
	ttl time.Duration
}

type lookup struct {
	val string
	ok  bool
}

func numberedKeys(n int) []string {
	keys := make([]string, n)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}
	return keys
}

func TestCache(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		before   []entry
		pause    time.Duration
		after    []entry
		wait     time.Duration
		wantGet  map[string]lookup
	}{
		{
			name:     "ключ не записан - промах",
			interval: time.Hour,
			wantGet:  map[string]lookup{"a": {"", false}},
		},
		{
			name:     "до истечения TTL - значение отдаётся",
			interval: time.Hour,
			before:   []entry{{"a", "v1", 2 * time.Second}},
			wait:     time.Second,
			wantGet:  map[string]lookup{"a": {"v1", true}},
		},
		{
			name:     "ровно в момент истечения - значение ещё отдаётся",
			interval: time.Hour,
			before:   []entry{{"a", "v1", 2 * time.Second}},
			wait:     2 * time.Second,
			wantGet:  map[string]lookup{"a": {"v1", true}},
		},
		{
			name:     "истёк TTL, очистки ещё не было - промах",
			interval: time.Hour,
			before:   []entry{{"a", "v1", 2 * time.Second}},
			wait:     3 * time.Second,
			wantGet:  map[string]lookup{"a": {"", false}},
		},
		{
			name:     "перезапись ключа - новое значение и TTL от момента перезаписи",
			interval: time.Hour,
			before:   []entry{{"a", "v1", 2 * time.Second}},
			pause:    time.Second,
			after:    []entry{{"a", "v2", 2 * time.Second}},
			wait:     1500 * time.Millisecond,
			wantGet:  map[string]lookup{"a": {"v2", true}},
		},
		{
			name:     "после очистки истёкшие записи - промах, живые отдаются",
			interval: time.Second,
			before: []entry{
				{"a", "v1", 1500 * time.Millisecond},
				{"b", "v2", 5 * time.Second},
			},
			wait: 2 * time.Second,
			wantGet: map[string]lookup{
				"a": {"", false},
				"b": {"v2", true},
			},
		},
		{
			name:     "запись истекает ровно в момент тика - значение ещё отдаётся",
			interval: time.Second,
			before:   []entry{{"a", "v1", time.Second}},
			wait:     time.Second,
			wantGet:  map[string]lookup{"a": {"v1", true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				c := cache.New(t.Context(), tt.interval)

				for _, e := range tt.before {
					c.Set(e.key, e.val, e.ttl)
				}
				time.Sleep(tt.pause)
				for _, e := range tt.after {
					c.Set(e.key, e.val, e.ttl)
				}
				time.Sleep(tt.wait)
				synctest.Wait()

				got := map[string]lookup{}
				for key := range tt.wantGet {
					val, ok := c.Get(key)
					got[key] = lookup{val, ok}
				}
				require.Equal(t, tt.wantGet, got)
			})
		})
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := cache.New(t.Context(), time.Millisecond)

		var wg sync.WaitGroup
		for i := range 8 {
			wg.Go(func() {
				key := strconv.Itoa(i)
				for range 100 {
					c.Set(key, "tmp", time.Millisecond)
					c.Get(key)
					time.Sleep(time.Millisecond)
				}
				c.Set(key, "final-"+key, time.Hour)
			})
		}
		wg.Wait()
		time.Sleep(time.Second)
		synctest.Wait()

		want := map[string]lookup{}
		got := map[string]lookup{}
		for i := range 8 {
			key := strconv.Itoa(i)
			want[key] = lookup{"final-" + key, true}
			val, ok := c.Get(key)
			got[key] = lookup{val, ok}
		}
		require.Equal(t, want, got)
	})
}

const keysCount = 1 << 16

func newFilledCache(b *testing.B, keys []string, interval time.Duration) *cache.Cache {
	ctx, cancel := context.WithCancel(context.Background())
	b.Cleanup(cancel)

	c := cache.New(ctx, interval)
	for _, k := range keys {
		c.Set(k, "val", time.Hour)
	}
	return c
}

func BenchmarkCache_Get(b *testing.B) {
	keys := numberedKeys(keysCount)
	c := newFilledCache(b, keys, 5*time.Millisecond)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Get(keys[i&(keysCount-1)])
			i++
		}
	})
}

func BenchmarkCache_Set(b *testing.B) {
	keys := numberedKeys(keysCount)
	c := newFilledCache(b, keys, 5*time.Millisecond)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(keys[i&(keysCount-1)], "val", time.Hour)
			i++
		}
	})
}

func BenchmarkCache_ReadHeavy(b *testing.B) {
	keys := numberedKeys(keysCount)
	c := newFilledCache(b, keys, 5*time.Millisecond)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			k := keys[i&(keysCount-1)]
			c.Set(k, "val", time.Hour)
			for range 9 {
				c.Get(k)
			}
			i++
		}
	})
}

func BenchmarkCache_GetDuringCleanup(b *testing.B) {
	keys := numberedKeys(keysCount)
	c := newFilledCache(b, keys, 50*time.Microsecond)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Get(keys[i&(keysCount-1)])
			i++
		}
	})
}
