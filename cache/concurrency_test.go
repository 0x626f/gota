package cache

import (
	"sync"
	"testing"
	"time"
)

func TestPrimary_ConcurrentOperations(t *testing.T) {
	cache := NewPrimary[int, int]()
	runConcurrentCacheOperations(t, cache, true)
}

func TestLRUCache_ConcurrentOperations(t *testing.T) {
	cache := NewLRUCache[int, int](25)
	runConcurrentCacheOperations(t, cache, false)
}

func TestLFUCache_ConcurrentOperations(t *testing.T) {
	cache := NewLFUCache[int, int](25)
	runConcurrentCacheOperations(t, cache, false)
}

func runConcurrentCacheOperations(t *testing.T, cache Cache[int, int], useTTL bool) {
	t.Helper()

	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		worker := worker
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < 200; i++ {
				key := (worker * 200) + i
				if useTTL && i%3 == 0 {
					cache.Set(key, i, time.Millisecond)
				} else {
					cache.Set(key, i)
				}
				cache.Get(key)
				if i%5 == 0 {
					cache.Delete(key)
				}
				if i%11 == 0 {
					cache.Flush()
				}
			}
		}()
	}
	wg.Wait()
}
