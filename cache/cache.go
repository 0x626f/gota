// Package cache provides implementations of various caching strategies
// including LRU (Least Recently Used) and LFU (Least Frequently Used) caches.
//
// Caching helps improve performance by storing frequently accessed data
// in memory for quick retrieval, while automatically evicting less important
// data when capacity limits are reached.
package cache

import "time"

// Cache defines the common operations for in-memory caches.
//
// Cache implementations are safe for concurrent use by multiple goroutines.
type Cache[D any, K comparable] interface {
	// Set stores a value in the cache with the specified key.
	// Existing-key behavior depends on the implementation.
	// Implementations may evict an existing item when capacity is reached.
	// TTL is optional and only used by Primary.
	Set(key K, data D, ttl ...time.Duration)

	// Get retrieves a value from the cache by its key.
	// It returns the cached value and true, or a zero value and false.
	Get(key K) (D, bool)

	// Delete removes a value from the cache by its key.
	// It returns true even when the key is already absent.
	Delete(key K) bool

	// Clear removes all items while keeping the configured capacity.
	Clear()

	// Flush removes extra items according to the cache eviction policy.
	Flush()
}

// InMemory is a simple map-based cache used by cache implementations.
type InMemory[K comparable, D any] map[K]D
