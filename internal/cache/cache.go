package cache

import (
	"context"
	"hash/maphash"
	"sync"
	"time"
)

type item struct {
	val       string
	expiresAt time.Time
}

type shard struct {
	mu   sync.RWMutex
	data map[string]item
}

const shardCount = 32 // Количество шардов

type Cache struct {
	seed   maphash.Seed
	shards [shardCount]*shard
}

func (c *Cache) cleanByTTL(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, s := range c.shards {
				s.removeExpired()
			}
		}
	}
}

// Каждый шард блокируется отдельно и ненадолго — остальные шарды в это время доступны.
func (s *shard) removeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for k, v := range s.data {
		if now.After(v.expiresAt) {
			delete(s.data, k)
		}
	}
}

func (c *Cache) Set(key string, val string, TTL time.Duration) {
	s := c.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = item{val, time.Now().Add(TTL)}
}

func (c *Cache) Get(key string) (string, bool) {
	s := c.getShard(key)
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok || time.Now().After(v.expiresAt) {
		return "", false
	}
	return v.val, true
}

func (c *Cache) getShard(key string) *shard {
	h := maphash.String(c.seed, key)
	return c.shards[h%shardCount]
}

func New(ctx context.Context, interval time.Duration) *Cache {
	c := &Cache{seed: maphash.MakeSeed()}

	for i := range c.shards {
		c.shards[i] = &shard{data: make(map[string]item)}
	}

	go c.cleanByTTL(ctx, interval)

	return c
}
