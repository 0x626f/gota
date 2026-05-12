package cache

import (
	"testing"
	"time"
)

func TestPrimary_SetAndGet(t *testing.T) {
	cache := NewPrimary[string, int]()

	cache.Set("key1", 100)

	val, exists := cache.Get("key1")
	if !exists {
		t.Fatal("key1 should exist")
	}
	if val != 100 {
		t.Errorf("expected 100, got %d", val)
	}
}

func TestPrimary_StoresValueAndExpirationInSingleMap(t *testing.T) {
	cache := NewPrimary[string, int]()

	cache.Set("key1", 100)
	cache.Set("key2", 200, time.Minute)

	item := cache.data["key1"]
	if item == nil {
		t.Fatal("key1 should be stored")
	}
	if item.First != 100 {
		t.Errorf("expected key1 value 100, got %d", item.First)
	}
	if item.Second != 0 {
		t.Errorf("expected key1 without TTL to have zero expiration, got %v", item.Second)
	}

	item = cache.data["key2"]
	if item == nil {
		t.Fatal("key2 should be stored")
	}
	if item.First != 200 {
		t.Errorf("expected key2 value 200, got %d", item.First)
	}
	if item.Second <= time.Duration(time.Now().UnixNano()) {
		t.Errorf("expected key2 expiration in the future, got %v", item.Second)
	}
}

func TestPrimary_SetUpdatesValue(t *testing.T) {
	cache := NewPrimary[string, int]()

	cache.Set("key1", 100)
	cache.Set("key1", 200)

	val, exists := cache.Get("key1")
	if !exists {
		t.Fatal("key1 should exist")
	}
	if val != 200 {
		t.Errorf("expected updated value 200, got %d", val)
	}
}

func TestPrimary_TTLExpiresValue(t *testing.T) {
	cache := NewPrimary[string, int]()

	cache.Set("key1", 100, 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	if _, exists := cache.Get("key1"); exists {
		t.Fatal("key1 should expire")
	}
}

func TestPrimary_NonPositiveTTLDoesNotExpire(t *testing.T) {
	cache := NewPrimary[string, int]()

	cache.Set("key1", 100, 0)
	cache.Set("key2", 200, -time.Second)
	time.Sleep(10 * time.Millisecond)

	if _, exists := cache.Get("key1"); !exists {
		t.Fatal("key1 should not expire")
	}
	if _, exists := cache.Get("key2"); !exists {
		t.Fatal("key2 should not expire")
	}
}

func TestPrimary_SetWithoutTTLClearsExistingTTL(t *testing.T) {
	cache := NewPrimary[string, int]()

	cache.Set("key1", 100, 10*time.Millisecond)
	cache.Set("key1", 200)
	time.Sleep(20 * time.Millisecond)

	val, exists := cache.Get("key1")
	if !exists {
		t.Fatal("key1 should not expire after reset without TTL")
	}
	if val != 200 {
		t.Errorf("expected updated value 200, got %d", val)
	}
}

func TestPrimary_FlushRemovesExpiredValues(t *testing.T) {
	cache := NewPrimary[string, int]()

	cache.Set("expired", 1, 10*time.Millisecond)
	cache.Set("live", 2)
	time.Sleep(20 * time.Millisecond)
	cache.Flush()

	if _, exists := cache.Get("expired"); exists {
		t.Fatal("expired key should be flushed")
	}
	if val, exists := cache.Get("live"); !exists || val != 2 {
		t.Fatal("live key should remain after flush")
	}
}

func TestPrimary_DeleteIsIdempotent(t *testing.T) {
	cache := NewPrimary[string, int]()

	cache.Set("key1", 100)
	if !cache.Delete("key1") {
		t.Fatal("delete should return true")
	}
	if !cache.Delete("key1") {
		t.Fatal("delete should return true when key is absent")
	}
}

func TestPrimary_Clear(t *testing.T) {
	cache := NewPrimary[string, int]()

	cache.Set("key1", 100)
	cache.Set("key2", 200, time.Minute)
	cache.Clear()

	if _, exists := cache.Get("key1"); exists {
		t.Fatal("key1 should be cleared")
	}
	if _, exists := cache.Get("key2"); exists {
		t.Fatal("key2 should be cleared")
	}
}

func TestPrimary_ImplementsCache(t *testing.T) {
	var _ Cache[int, string] = NewPrimary[string, int]()
}

func TestTTLIsIgnoredByLRUAndLFU(t *testing.T) {
	lru := NewLRUCache[string, int](10)
	lfu := NewLFUCache[string, int](10)

	lru.Set("key1", 100, 10*time.Millisecond)
	lfu.Set("key1", 100, 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	if val, exists := lru.Get("key1"); !exists || val != 100 {
		t.Fatal("LRU should ignore TTL")
	}
	if val, exists := lfu.Get("key1"); !exists || val != 100 {
		t.Fatal("LFU should ignore TTL")
	}
}
