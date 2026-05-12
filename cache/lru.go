package cache

import (
	"sync"
	"time"

	"github.com/0x626f/gota/collections/linkedlist"
)

// LRUCache evicts the least recently used item when capacity is reached.
type LRUCache[K comparable, D any] struct {
	mu sync.Mutex

	// capacity is the maximum number of items the cache can hold
	// A capacity of 0 means unlimited
	capacity int

	// recent is a linked list maintaining items in access order
	// Most recently accessed items are at the front
	recent *linkedlist.LinkedList[*Pair[K, D]]

	// data maps keys to their corresponding nodes in the linked list
	// for O(1) lookup and access
	data InMemory[K, *linkedlist.LinkedNode[*Pair[K, D]]]
}

// NewLRUCache creates an LRU cache. Capacity 0 or less means unlimited.
//
// Example:
//
//	cache := cache.NewLRUCache[string, int](100)
//	cache.Set("user:123", 42)
//	value, found := cache.Get("user:123")
func NewLRUCache[K comparable, D any](capacity int) *LRUCache[K, D] {
	if capacity < 0 {
		capacity = 0
	}

	return &LRUCache[K, D]{
		capacity: capacity,
		recent:   linkedlist.NewLinkedList[*Pair[K, D]](),
		data:     make(map[K]*linkedlist.LinkedNode[*Pair[K, D]]),
	}
}

// Set adds an item to the cache.
// If the key already exists, this method does nothing (existing value is preserved).
// If the cache is at capacity, the least recently used item is evicted to make room.
func (cache *LRUCache[K, D]) Set(key K, item D, _ ...time.Duration) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if _, exists := cache.data[key]; exists {
		return
	}

	node := cache.recent.InsertFront(&Pair[K, D]{First: key, Second: item})
	cache.data[key] = node

	if cache.capacity != 0 && cache.recent.Size() > cache.capacity {
		retired := cache.recent.PopRight()
		delete(cache.data, retired.First)
	}
}

// Get retrieves an item and marks it as most recently used.
func (cache *LRUCache[K, D]) Get(key K) (D, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if node, exists := cache.data[key]; exists {
		cache.recent.MoveToFront(node)
		return node.Data.Second, true
	}

	var zero D
	return zero, false
}

// Delete removes an item by key. It returns true even when the key is absent.
func (cache *LRUCache[K, D]) Delete(key K) bool {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if node, exists := cache.data[key]; exists {
		cache.recent.Remove(node)
		delete(cache.data, key)
	}
	return true
}

// Flush removes least recently used items until the cache fits its capacity.
// It is a no-op for unlimited caches.
func (cache *LRUCache[K, D]) Flush() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.capacity == 0 {
		return
	}

	if cache.recent.Size() > cache.capacity {
		cache.recent.ForEach(func(index int, data *Pair[K, D]) bool {
			if (index + 1) > cache.capacity {
				delete(cache.data, data.First)
			}
			return true
		})
		cache.recent.Shrink(cache.capacity)
	}
}

// Clear removes all items while keeping the configured capacity.
func (cache *LRUCache[K, D]) Clear() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.recent.DeleteAll()
	clear(cache.data)
}
