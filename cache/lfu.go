package cache

import (
	"sync"
	"time"

	"github.com/0x626f/gota/collections/linkedlist"
)

// LFUCache evicts least frequently used items when capacity is reached.
type LFUCache[K comparable, D any] struct {
	mu sync.Mutex

	// capacity is the maximum number of items the cache can hold.
	// A capacity of 0 means unlimited.
	capacity int

	// size is the current number of cached items.
	size int

	// frequencies is a linked list of frequency buckets
	// Each bucket contains all items with the same access frequency
	frequencies *linkedlist.LinkedList[*Pair[uint, InMemory[K, D]]]

	// data maps frequency counts to their corresponding nodes in the frequencies list
	data InMemory[uint, *linkedlist.LinkedNode[*Pair[uint, InMemory[K, D]]]]

	// spot maps keys to their frequency bucket nodes for O(1) lookup
	spot InMemory[K, *linkedlist.LinkedNode[*Pair[uint, InMemory[K, D]]]]
}

// NewLFUCache creates an LFU cache. Capacity 0 or less means unlimited.
//
// Example:
//
//	cache := cache.NewLFUCache[string, int](100)
//	cache.Set("counter", 1)
//	cache.Get("counter") // Increases frequency
//	cache.Get("counter") // Increases frequency again
func NewLFUCache[K comparable, D any](capacity int) *LFUCache[K, D] {
	if capacity < 0 {
		capacity = 0
	}

	return &LFUCache[K, D]{
		capacity:    capacity,
		frequencies: linkedlist.NewLinkedList[*Pair[uint, InMemory[K, D]]](),
		data:        make(InMemory[uint, *linkedlist.LinkedNode[*Pair[uint, InMemory[K, D]]]]),
		spot:        make(InMemory[K, *linkedlist.LinkedNode[*Pair[uint, InMemory[K, D]]]]),
	}
}

// record returns the bucket for freq, creating it when needed.
func (cache *LFUCache[K, D]) record(freq uint) *linkedlist.LinkedNode[*Pair[uint, InMemory[K, D]]] {
	if cache.data[freq] == nil {
		entry := &Pair[uint, InMemory[K, D]]{First: freq, Second: make(InMemory[K, D])}
		node := cache.frequencies.Insert(entry)
		cache.data[freq] = node
	}
	return cache.data[freq]
}

func (cache *LFUCache[K, D]) removeBucket(node *linkedlist.LinkedNode[*Pair[uint, InMemory[K, D]]]) {
	if node == nil || len(node.Data.Second) != 0 {
		return
	}

	delete(cache.data, node.Data.First)
	cache.frequencies.Remove(node)
}

func (cache *LFUCache[K, D]) deleteFromBucket(key K, node *linkedlist.LinkedNode[*Pair[uint, InMemory[K, D]]]) {
	delete(node.Data.Second, key)
	delete(cache.spot, key)
	cache.size--
	cache.removeBucket(node)
}

func (cache *LFUCache[K, D]) evictLeastFrequent() {
	var victim *linkedlist.LinkedNode[*Pair[uint, InMemory[K, D]]]

	cache.frequencies.ForEach(func(_ int, nodeData *Pair[uint, InMemory[K, D]]) bool {
		node := cache.data[nodeData.First]
		if len(nodeData.Second) == 0 {
			cache.removeBucket(node)
			return true
		}
		if victim == nil || nodeData.First < victim.Data.First {
			victim = node
		}
		return true
	})

	if victim == nil {
		return
	}

	for key := range victim.Data.Second {
		cache.deleteFromBucket(key, victim)
		return
	}
}

// Set stores an item in the cache.
// New items start with a frequency of 1. Existing items keep their frequency.
func (cache *LFUCache[K, D]) Set(key K, item D, _ ...time.Duration) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if node, exists := cache.spot[key]; exists {
		node.Data.Second[key] = item
		return
	}

	node := cache.record(1)
	node.Data.Second[key] = item
	cache.spot[key] = node
	cache.size++
}

// Get retrieves an item and increments its access frequency.
func (cache *LFUCache[K, D]) Get(key K) (D, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	node, exists := cache.spot[key]

	if !exists {
		var zero D
		return zero, false
	}

	nextNode := cache.record(node.Data.First + 1)
	item := node.Data.Second[key]

	nextNode.Data.Second[key] = item
	delete(node.Data.Second, key)
	cache.spot[key] = nextNode
	cache.removeBucket(node)

	return item, true
}

// Delete removes an item by key. It returns true even when the key is absent.
func (cache *LFUCache[K, D]) Delete(key K) bool {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	node, exists := cache.spot[key]

	if !exists {
		return true
	}

	cache.deleteFromBucket(key, node)

	return true
}

// Flush removes least frequently used items until the cache fits its capacity.
// It is a no-op for unlimited caches.
func (cache *LFUCache[K, D]) Flush() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.capacity == 0 {
		return
	}

	for cache.size > cache.capacity {
		cache.evictLeastFrequent()
	}
}

// Clear removes all items while keeping the configured capacity.
func (cache *LFUCache[K, D]) Clear() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.frequencies.DeleteAll()
	clear(cache.data)
	clear(cache.spot)
	cache.size = 0
}
