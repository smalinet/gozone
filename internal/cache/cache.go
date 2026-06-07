// Package cache provides a generic in-memory cache with per-item TTL expiration.
package cache

import (
	"sync"
	"time"
)

// entry holds a cached value and its expiration timestamp.
type entry[V any] struct {
	value  V
	expiry time.Time
}

// Cache is a generic in-memory cache with a fixed TTL for all entries.
// Zero value is not usable; create with New.
type Cache[V any] struct {
	mu    sync.RWMutex
	items map[string]*entry[V]
	ttl   time.Duration
}

// New creates a cache where every entry expires after ttl.
func New[V any](ttl time.Duration) *Cache[V] {
	return &Cache[V]{
		items: make(map[string]*entry[V]),
		ttl:   ttl,
	}
}

// Get returns the cached value and true if the key exists and has not expired.
func (c *Cache[V]) Get(key string) (V, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiry) {
		return *new(V), false
	}
	return e.value, true
}

// Set stores a value under key with the configured TTL.
func (c *Cache[V]) Set(key string, value V) {
	c.mu.Lock()
	c.items[key] = &entry[V]{value: value, expiry: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

// Delete removes a single key from the cache.
func (c *Cache[V]) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Clear removes all entries from the cache.
func (c *Cache[V]) Clear() {
	c.mu.Lock()
	c.items = make(map[string]*entry[V])
	c.mu.Unlock()
}

// Len returns the number of entries currently in the cache (including expired ones).
func (c *Cache[V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
