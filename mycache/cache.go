package mycache

import (
	"mycache/eliminationstrategy"
	"sync"
)

type cache struct {
	mu         sync.Mutex
	es         int
	esCache    *eliminationstrategy.Cache
	cacheBytes int64
}

// Add() includes New()
func (c *cache) Add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.esCache == nil {
		c.esCache = eliminationstrategy.New(c.cacheBytes, c.es, nil)
	}
	c.esCache.Add(key, value)
}

func (c *cache) Get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.esCache == nil {
		return
	}

	if v, ok := c.esCache.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
