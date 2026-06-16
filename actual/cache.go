package actual

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data      any
	expiresAt time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
}

var (
	globalCache *Cache
	cacheMu     sync.Mutex
)

func InitCache(ttl time.Duration) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if globalCache == nil {
		globalCache = &Cache{
			entries: make(map[string]*cacheEntry),
			ttl:     ttl,
		}
	}
}

func GetCache() *Cache {
	return globalCache
}

func ClearCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if globalCache != nil {
		globalCache.mu.Lock()
		globalCache.entries = make(map[string]*cacheEntry)
		globalCache.mu.Unlock()
	}
}

func (c *Cache) get(key string) (any, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, false
	}
	return entry.data, true
}

func (c *Cache) set(key string, data any) {
	c.mu.Lock()
	c.entries[key] = &cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}
