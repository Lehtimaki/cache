package cache

import (
	"context"
	"sync"
	"time"
)

type Cache interface {
	Put(string, []byte)
	Get(string) ([]byte, bool)
	Del(string)
}

type cache struct {
	ttl  time.Duration

	mu    sync.RWMutex
	store map[string]cell
}

type cell struct {
	value []byte
	ts    time.Time
}

func New(ctx context.Context, ttl time.Duration) Cache {
	c := cache{
		ttl:   ttl,
		store: map[string]cell{},
	}

	if c.ttl > 0 {
		go c.gc(ctx, time.NewTicker(time.Second))
	}

	return &c
}

func (c *cache) Put(k string, v []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[k] = cell{
		value: v,
		ts:    time.Now(),
	}
}

func (c *cache) Get(k string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, ok := c.store[k]; ok {
		return v.value, true
	}

	return []byte{}, false
}

func (c *cache) Del(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, k)
}

func (c *cache) gc(ctx context.Context, t *time.Ticker) {
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if keys, exp := c.collect(); exp {
				c.reclaim(keys)
			}
		case <-ctx.Done():
			break
		}
	}
}

func (c *cache) collect() ([]string, bool) {
	exp := []string{}

	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	for k, v := range c.store {
		if now.Sub(v.ts) > c.ttl {
			exp = append(exp, k)
		}
	}

	return exp, len(exp) > 0
}

func (c *cache) reclaim(keys []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, k := range keys {
		delete(c.store, k)
	}
}
