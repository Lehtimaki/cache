package cache

import (
	"context"
	"sync"
	"time"
)

type Bucket struct {
	ttl  time.Duration

	mu    sync.RWMutex
	store map[string]Item
}

type Item struct {
	value []byte
	ts    time.Time
	ttl   time.Duration
}

// create and return a new cache bucket.
func New(ctx context.Context, ttl time.Duration) *Bucket {
	b := Bucket{
		ttl:   ttl,
		store: map[string]Item{},
	}

	go b.gc(ctx, time.NewTicker(time.Second))

	return &b
}

// Puts a new key/value into the bucket. Overwrites an existing item
// if the key already exists.
func (b *Bucket) Put(key string, value []byte, options ...Option) {
	item := Item{
		value: value,
		ts:    time.Now(),
	}

	for _, option := range options {
		option(&item)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.store[key] = item
}

// Gets a key from the bucket and returns the items contents.
// Returns a secondary bool value to let the caller know if the
// key was found or not.
func (b *Bucket) Get(key string) ([]byte, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if item, ok := b.store[key]; ok {
		return item.value, true
	}

	return []byte{}, false
}

// Deletes the given key from the bucket.
func (b *Bucket) Del(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.store, key)
}

// The garbage collector mostly handles the ticker. It will call
// `collect` which only uses a read-lock on every tick,
// and `reclaim` which uses a write-lock when necessary.
func (b *Bucket) gc(ctx context.Context, t *time.Ticker) {
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if keys, hasExpiredKeys := b.collect(); hasExpiredKeys {
				b.reclaim(keys)
			}
		case <-ctx.Done():
			break
		}
	}
}

// Collects a list of expired keys from the bucket.
func (b *Bucket) collect() ([]string, bool) {
	expiredKeys := []string{}

	b.mu.RLock()
	defer b.mu.RUnlock()

	now := time.Now()
	for key, item := range b.store {
		ttl := b.ttl
		if item.ttl != 0 {
			ttl = item.ttl
		}

		if ttl > 0 && now.Sub(item.ts) > ttl {
			expiredKeys = append(expiredKeys, key)
		}
	}

	return expiredKeys, len(expiredKeys) > 0
}

// Deletes keys in bulk.
func (b *Bucket) reclaim(keys []string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, key := range keys {
		delete(b.store, key)
	}
}
