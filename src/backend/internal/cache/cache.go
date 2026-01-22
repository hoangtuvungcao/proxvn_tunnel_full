package cache

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with expiry
type CacheItem struct {
	Value      interface{}
	Expiry     time.Time
	AccessTime time.Time
	Hits       int64
}

// Cache is an in-memory cache with TTL and LRU eviction
type Cache struct {
	items    map[string]*CacheItem
	mu       sync.RWMutex
	maxSize  int
	ttl      time.Duration
	cleanupC chan struct{}
}

// NewCache creates a new cache with size limit and TTL
func NewCache(maxSize int, ttl time.Duration) *Cache {
	c := &Cache{
		items:    make(map[string]*CacheItem),
		maxSize:  maxSize,
		ttl:      ttl,
		cleanupC: make(chan struct{}),
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Get retrieves a value from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check expiry
	if time.Now().After(item.Expiry) {
		return nil, false
	}

	// Update access stats
	item.AccessTime = time.Now()
	item.Hits++

	return item.Value, true
}

// Set stores a value in cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if at capacity
	if len(c.items) >= c.maxSize {
		c.evictLRU()
	}

	c.items[key] = &CacheItem{
		Value:      value,
		Expiry:     time.Now().Add(c.ttl),
		AccessTime: time.Now(),
		Hits:       0,
	}
}

// Delete removes a key from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheItem)
}

// Stats returns cache statistics
func (c *Cache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalHits int64
	for _, item := range c.items {
		totalHits += item.Hits
	}

	return map[string]interface{}{
		"size":       len(c.items),
		"max_size":   c.maxSize,
		"total_hits": totalHits,
		"ttl":        c.ttl.String(),
	}
}

// evictLRU removes least recently used item
func (c *Cache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, item := range c.items {
		if first || item.AccessTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.AccessTime
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// cleanup periodically removes expired items
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for key, item := range c.items {
				if now.After(item.Expiry) {
					delete(c.items, key)
				}
			}
			c.mu.Unlock()
		case <-c.cleanupC:
			return
		}
	}
}

// Close stops the cleanup goroutine
func (c *Cache) Close() {
	close(c.cleanupC)
}
