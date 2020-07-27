package geecache

import (
	"sync"

	"geecache/lru"
)

type cache struct {
	mu sync.Mutex
	lru *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, bv ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, bv)
}

func (c *cache) get(key string) (bv ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		bv, _ = v.(ByteView)
		return
	}
	return
}