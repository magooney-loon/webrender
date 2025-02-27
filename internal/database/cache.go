package database

import (
	"container/list"
	"sync"
	"time"
)

// CacheKey represents a unique key for a cache entry.
type CacheKey struct {
	Query string
	Args  string // JSON serialized arguments
}

// CacheEntry represents a cached query result.
type CacheEntry struct {
	Key       CacheKey
	Value     interface{}
	CreatedAt time.Time
	Size      int64
}

// Cache is an LRU cache for database query results.
type Cache struct {
	// Configuration
	maxItems     int
	ttl          time.Duration
	maxSizeBytes int64
	enabled      bool

	// Cache storage
	items     map[CacheKey]*list.Element
	evictList *list.List
	size      int64

	// Metrics for monitoring
	hits      uint64
	misses    uint64
	evictions uint64

	// Synchronization
	mu sync.RWMutex
}

// NewCache creates a new LRU cache with the specified configuration.
func NewCache(maxItems int, ttl time.Duration, enabled bool) *Cache {
	return &Cache{
		maxItems:     maxItems,
		ttl:          ttl,
		maxSizeBytes: 64 * 1024 * 1024, // 64MB default max size
		enabled:      enabled,
		items:        make(map[CacheKey]*list.Element),
		evictList:    list.New(),
	}
}

// Get retrieves a value from the cache.
func (c *Cache) Get(key CacheKey) (interface{}, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	element, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	// Get the entry and check if it's expired
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := element.Value.(*CacheEntry)
	if time.Since(entry.CreatedAt) > c.ttl {
		// Entry has expired, remove it
		c.evictList.Remove(element)
		delete(c.items, key)
		c.size -= entry.Size
		c.evictions++
		return nil, false
	}

	// Move to front (recently used)
	c.evictList.MoveToFront(element)
	c.hits++
	return entry.Value, true
}

// Set adds a value to the cache.
func (c *Cache) Set(key CacheKey, value interface{}, size int64) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the key exists
	if element, exists := c.items[key]; exists {
		c.evictList.MoveToFront(element)
		entry := element.Value.(*CacheEntry)
		c.size = c.size - entry.Size + size
		entry.Value = value
		entry.CreatedAt = time.Now()
		entry.Size = size
		return
	}

	// Enforce size limit by removing items
	for c.maxItems > 0 && c.evictList.Len() >= c.maxItems {
		c.removeOldest()
	}

	// Enforce byte size limit
	for c.maxSizeBytes > 0 && c.size+size > c.maxSizeBytes && c.evictList.Len() > 0 {
		c.removeOldest()
	}

	// Add new entry
	entry := &CacheEntry{
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
		Size:      size,
	}
	element := c.evictList.PushFront(entry)
	c.items[key] = element
	c.size += size
}

// Remove removes a key from the cache.
func (c *Cache) Remove(key CacheKey) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, exists := c.items[key]; exists {
		entry := element.Value.(*CacheEntry)
		c.size -= entry.Size
		c.evictList.Remove(element)
		delete(c.items, key)
	}
}

// removeOldest removes the oldest item from the cache.
func (c *Cache) removeOldest() {
	element := c.evictList.Back()
	if element == nil {
		return
	}

	entry := element.Value.(*CacheEntry)
	c.size -= entry.Size
	c.evictList.Remove(element)
	delete(c.items, entry.Key)
	c.evictions++
}

// Clear empties the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[CacheKey]*list.Element)
	c.evictList.Init()
	c.size = 0
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.evictList.Len()
}

// Size returns the cache size in bytes.
func (c *Cache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

// Stats returns cache statistics.
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Hits:          c.hits,
		Misses:        c.misses,
		Evictions:     c.evictions,
		Items:         c.evictList.Len(),
		SizeBytes:     c.size,
		HitRatio:      calculateHitRatio(c.hits, c.misses),
		MemoryUsagePC: float64(c.size) / float64(c.maxSizeBytes) * 100,
	}
}

// CacheStats holds cache performance statistics.
type CacheStats struct {
	Hits          uint64
	Misses        uint64
	Evictions     uint64
	Items         int
	SizeBytes     int64
	HitRatio      float64
	MemoryUsagePC float64
}

// calculateHitRatio calculates the cache hit ratio.
func calculateHitRatio(hits, misses uint64) float64 {
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}
