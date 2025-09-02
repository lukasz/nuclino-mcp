package cache

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value      interface{}
	ExpiresAt  time.Time
	AccessedAt time.Time
	HitCount   int64
}

// IsExpired checks if the cache item has expired
func (c *CacheItem) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// Touch updates the access time and increments hit count
func (c *CacheItem) Touch() {
	c.AccessedAt = time.Now()
	c.HitCount++
}

// Cache represents an in-memory cache with TTL and LRU eviction
type Cache struct {
	mu         sync.RWMutex
	items      map[string]*CacheItem
	maxSize    int
	defaultTTL time.Duration
	stats      CacheStats
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Expirations int64
}

// HitRate returns the cache hit rate as a percentage
func (s *CacheStats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total) * 100
}

// NewCache creates a new cache instance
func NewCache(maxSize int, defaultTTL time.Duration) *Cache {
	cache := &Cache{
		items:      make(map[string]*CacheItem),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	if item.IsExpired() {
		c.stats.Misses++
		c.stats.Expirations++
		// Don't delete here to avoid write lock, let cleanup handle it
		return nil, false
	}

	item.Touch()
	c.stats.Hits++
	return item.Value, true
}

// Set stores a value in the cache with default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	item := &CacheItem{
		Value:      value,
		ExpiresAt:  now.Add(ttl),
		AccessedAt: now,
		HitCount:   0,
	}

	// If cache is at capacity, evict LRU item
	if len(c.items) >= c.maxSize {
		c.evictLRU()
	}

	c.items[key] = item
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
}

// Size returns the current number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Stats returns cache performance statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.stats
}

// evictLRU removes the least recently used item (assumes lock is held)
func (c *Cache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.AccessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.AccessedAt
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
		c.stats.Evictions++
	}
}

// cleanup removes expired items periodically
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		for key, item := range c.items {
			if item.IsExpired() {
				delete(c.items, key)
				c.stats.Expirations++
			}
		}

		c.mu.Unlock()
	}
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	MaxSize       int
	DefaultTTL    time.Duration
	ItemTTL       time.Duration
	WorkspaceTTL  time.Duration
	CollectionTTL time.Duration
	SearchTTL     time.Duration
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxSize:       1000,
		DefaultTTL:    5 * time.Minute,
		ItemTTL:       10 * time.Minute,
		WorkspaceTTL:  30 * time.Minute,
		CollectionTTL: 15 * time.Minute,
		SearchTTL:     2 * time.Minute,
	}
}
