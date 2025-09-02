package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(10, 1*time.Minute)

	// Test setting and getting a value
	cache.Set("key1", "value1")

	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Test getting non-existent key
	_, found = cache.Get("nonexistent")
	assert.False(t, found)
}

func TestCache_TTL(t *testing.T) {
	cache := NewCache(10, 100*time.Millisecond)

	cache.Set("key1", "value1")

	// Should be available immediately
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = cache.Get("key1")
	assert.False(t, found)
}

func TestCache_CustomTTL(t *testing.T) {
	cache := NewCache(10, 1*time.Minute)

	// Set with custom short TTL
	cache.SetWithTTL("key1", "value1", 100*time.Millisecond)

	// Should be available immediately
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = cache.Get("key1")
	assert.False(t, found)
}

func TestCache_LRUEviction(t *testing.T) {
	cache := NewCache(2, 1*time.Minute) // Small cache size

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Both should be present
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	assert.True(t, found1)
	assert.True(t, found2)

	// Access key1 to make it more recently used
	cache.Get("key1")

	// Add third item, should evict key2 (least recently used)
	cache.Set("key3", "value3")

	// key2 should be evicted
	_, found2 = cache.Get("key2")
	assert.False(t, found2)

	// key1 and key3 should still be present
	_, found1 = cache.Get("key1")
	_, found3 := cache.Get("key3")
	assert.True(t, found1)
	assert.True(t, found3)
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(10, 1*time.Minute)

	cache.Set("key1", "value1")

	// Should exist
	_, found := cache.Get("key1")
	assert.True(t, found)

	// Delete it
	cache.Delete("key1")

	// Should not exist
	_, found = cache.Get("key1")
	assert.False(t, found)
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(10, 1*time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Both should exist
	assert.Equal(t, 2, cache.Size())

	// Clear cache
	cache.Clear()

	// Should be empty
	assert.Equal(t, 0, cache.Size())
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	assert.False(t, found1)
	assert.False(t, found2)
}

func TestCache_Stats(t *testing.T) {
	cache := NewCache(10, 1*time.Minute)

	// Initial stats
	stats := cache.Stats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)

	// Set a value
	cache.Set("key1", "value1")

	// Hit
	cache.Get("key1")
	stats = cache.Stats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)

	// Miss
	cache.Get("nonexistent")
	stats = cache.Stats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)

	// Hit rate should be 50%
	assert.Equal(t, 50.0, stats.HitRate())
}

func TestCache_HitCount(t *testing.T) {
	cache := NewCache(10, 1*time.Minute)

	cache.Set("key1", "value1")

	// Access multiple times
	cache.Get("key1")
	cache.Get("key1")
	cache.Get("key1")

	// Check internal hit count (accessing internal state for testing)
	cache.mu.RLock()
	item := cache.items["key1"]
	cache.mu.RUnlock()

	assert.Equal(t, int64(3), item.HitCount)
}

func TestCache_Concurrent(t *testing.T) {
	cache := NewCache(100, 1*time.Minute)

	// Test concurrent access
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 50; i++ {
			cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 50; i++ {
			cache.Get(fmt.Sprintf("key%d", i))
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Should not panic and should have some items
	assert.Greater(t, cache.Size(), 0)
}

func TestCacheItem_IsExpired(t *testing.T) {
	item := &CacheItem{
		Value:     "test",
		ExpiresAt: time.Now().Add(-1 * time.Second), // Already expired
	}

	assert.True(t, item.IsExpired())

	item.ExpiresAt = time.Now().Add(1 * time.Second) // Not expired
	assert.False(t, item.IsExpired())
}

func TestCacheItem_Touch(t *testing.T) {
	item := &CacheItem{
		Value:      "test",
		ExpiresAt:  time.Now().Add(1 * time.Minute),
		AccessedAt: time.Now().Add(-1 * time.Minute),
		HitCount:   0,
	}

	oldAccessTime := item.AccessedAt
	oldHitCount := item.HitCount

	item.Touch()

	assert.True(t, item.AccessedAt.After(oldAccessTime))
	assert.Equal(t, oldHitCount+1, item.HitCount)
}

func TestDefaultCacheConfig(t *testing.T) {
	config := DefaultCacheConfig()

	assert.Equal(t, 1000, config.MaxSize)
	assert.Equal(t, 5*time.Minute, config.DefaultTTL)
	assert.Equal(t, 10*time.Minute, config.ItemTTL)
	assert.Equal(t, 30*time.Minute, config.WorkspaceTTL)
	assert.Equal(t, 15*time.Minute, config.CollectionTTL)
	assert.Equal(t, 2*time.Minute, config.SearchTTL)
}
