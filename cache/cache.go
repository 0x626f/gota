// Package cache provides in-memory cache implementations and helpers.
//
// Primary is a map-backed cache with optional per-item TTL. LRUCache evicts
// the least recently used item when capacity is exceeded. LFUCache evicts
// least frequently used items when Flush is called. Aggregator combines a cache
// with duplicate in-flight call suppression for string-keyed loads.
package cache

import "time"

// Cache defines the common operations for in-memory caches.
//
// Cache implementations are safe for concurrent use by multiple goroutines.
type Cache[D any, K comparable] interface {
	// Set stores a value in the cache with the specified key.
	// Existing keys are updated with the new value.
	// Implementations with a configured capacity may evict items when capacity is exceeded.
	// TTL is optional; implementations that do not support expiration ignore it.
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
	// Primary removes expired items, LRUCache removes least recently used items
	// until it fits capacity, and LFUCache removes least frequently used items
	// until it fits capacity.
	Flush()
}

// InMemory is a simple map-based cache used by cache implementations.
type InMemory[K comparable, D any] map[K]D
