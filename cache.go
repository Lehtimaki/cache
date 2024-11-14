package cache

import (
	"sync"
)

type Cache interface {
	Put(string, []byte)
	Get(string) ([]byte, bool)
	Del(string)
}

type cache struct {
	mu    sync.RWMutex
	store map[string][]byte
}

func New() Cache {
	return &cache{
		store: map[string][]byte{},
	}
}

func (c *cache) Put(k string, v []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[k] = v
}

func (c *cache) Get(k string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, ok := c.store[k]; ok {
		return v, true
	}

	return []byte{}, false
}

func (c *cache) Del(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, k)
}
