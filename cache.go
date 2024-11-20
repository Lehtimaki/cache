package cache

import (
	"context"
	"sync"
	"time"
)

type Bucket struct {
	ttl  time.Duration

	mu    sync.RWMutex
	store map[string]item
}

type item struct {
	value []byte
	ts    time.Time
	ttl   time.Duration
}

func New(ctx context.Context, ttl time.Duration) *Bucket {
	b := Bucket{
		ttl:   ttl,
		store: map[string]item{},
	}

	go b.gc(ctx, time.NewTicker(time.Second))

	return &b
}

func (b *Bucket) Put(k string, v []byte, opts ...Option) {
	c := item{
		value: v,
		ts:    time.Now(),
	}

	for _, opt := range opts {
		opt(&c)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.store[k] = c
}

func (b *Bucket) Get(k string) ([]byte, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if i, ok := b.store[k]; ok {
		return i.value, true
	}

	return []byte{}, false
}

func (b *Bucket) Del(k string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.store, k)
}

func (b *Bucket) gc(ctx context.Context, t *time.Ticker) {
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if keys, exp := b.collect(); exp {
				b.reclaim(keys)
			}
		case <-ctx.Done():
			break
		}
	}
}

func (b *Bucket) collect() ([]string, bool) {
	exp := []string{}

	b.mu.RLock()
	defer b.mu.RUnlock()

	now := time.Now()
	for k, v := range b.store {
		ttl := time.Duration(0)
		if v.ttl != 0 {
			ttl = v.ttl
		} else {
			ttl = b.ttl
		}

		if now.Sub(v.ts) > ttl {
			exp = append(exp, k)
		}
	}

	return exp, len(exp) > 0
}

func (b *Bucket) reclaim(keys []string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, k := range keys {
		delete(b.store, k)
	}
}
