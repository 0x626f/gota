package cache

import (
	"sync"
	"time"
)

// Primary is a map-backed cache with optional per-item TTL.
type Primary[K comparable, D any] struct {
	mu   sync.Mutex
	data InMemory[K, *Pair[D, time.Duration]]
}

// NewPrimary creates a map-backed cache.
func NewPrimary[K comparable, D any]() *Primary[K, D] {
	return &Primary[K, D]{
		data: make(InMemory[K, *Pair[D, time.Duration]]),
	}
}

// Set stores a value. A positive TTL expires the value after that duration.
func (cache *Primary[K, D]) Set(key K, data D, ttl ...time.Duration) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	item := &Pair[D, time.Duration]{First: data}
	if len(ttl) > 0 && ttl[0] > 0 {
		item.Second = time.Duration(time.Now().Add(ttl[0]).UnixNano())
	}
	cache.data[key] = item
}

// Get retrieves a value unless it has expired.
func (cache *Primary[K, D]) Get(key K) (D, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	item, exists := cache.data[key]
	if !exists {
		var zero D
		return zero, false
	}
	if cache.expiredLocked(item) {
		cache.deleteLocked(key)
		var zero D
		return zero, false
	}

	return item.First, true
}

// Delete removes a value by key. It returns true even when the key is absent.
func (cache *Primary[K, D]) Delete(key K) bool {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.deleteLocked(key)
	return true
}

// Clear removes all items.
func (cache *Primary[K, D]) Clear() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	clear(cache.data)
}

// Flush removes expired items.
func (cache *Primary[K, D]) Flush() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for key, item := range cache.data {
		if cache.expiredLocked(item) {
			cache.deleteLocked(key)
		}
	}
}

func (cache *Primary[K, D]) deleteLocked(key K) {
	delete(cache.data, key)
}

func (cache *Primary[K, D]) expiredLocked(item *Pair[D, time.Duration]) bool {
	return item.Second > 0 && time.Duration(time.Now().UnixNano()) >= item.Second
}
