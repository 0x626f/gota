package cache

import (
	"testing"
)

// ============================================================================
// COMPREHENSIVE TEST SUITE FOR LRU CACHE
// ============================================================================

// ----------------------------------------------------------------------------
// Edge Cases: Basic Operations
// ----------------------------------------------------------------------------

func TestLRUCache_NewCache(t *testing.T) {
	cache := NewLRUCache[string, int](5)
	if cache == nil {
		t.Fatal("NewLRUCache returned nil")
	}
	if cache.capacity != 5 {
		t.Errorf("Expected capacity 5, got %d", cache.capacity)
	}
}

func TestLRUCache_NewCache_ZeroCapacity(t *testing.T) {
	cache := NewLRUCache[string, int](0)
	if cache == nil {
		t.Fatal("NewLRUCache returned nil")
	}

	// Should allow unlimited items with zero capacity
	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	val, exists := cache.Get("key1")
	if !exists || val != 1 {
		t.Error("Should be able to store items with zero capacity")
	}
}

func TestLRUCache_NewCache_NegativeCapacity(t *testing.T) {
	cache := NewLRUCache[string, int](-1)
	if cache.capacity != 0 {
		t.Errorf("Expected negative capacity to become 0, got %d", cache.capacity)
	}

	cache.Set("key1", 1)
	cache.Flush()

	if val, exists := cache.Get("key1"); !exists || val != 1 {
		t.Error("negative capacity should behave as unlimited")
	}
}

func TestLRUCache_SetAndGet_SingleItem(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("key1", 100)

	val, exists := cache.Get("key1")
	if !exists {
		t.Error("Expected key1 to exist")
	}
	if val != 100 {
		t.Errorf("Expected value 100, got %d", val)
	}
}

func TestLRUCache_Get_NonExistentKey(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	val, exists := cache.Get("nonexistent")
	if exists {
		t.Error("Expected key to not exist")
	}
	if val != 0 {
		t.Errorf("Expected zero value, got %d", val)
	}
}

func TestLRUCache_Set_UpdatesExistingKey(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("key1", 100)
	cache.Set("key1", 200)

	val, exists := cache.Get("key1")
	if !exists {
		t.Error("Expected key1 to exist")
	}
	if val != 200 {
		t.Errorf("Expected updated value 200, got %d", val)
	}
}

func TestLRUCache_SetExistingKeyMarksAsRecent(t *testing.T) {
	cache := NewLRUCache[string, int](2)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key1", 10)
	cache.Set("key3", 3)

	if val, exists := cache.Get("key1"); !exists || val != 10 {
		t.Fatalf("expected key1 to remain with updated value 10, got %d, exists %t", val, exists)
	}
	if _, exists := cache.Get("key2"); exists {
		t.Fatal("expected key2 to be evicted as least recently used")
	}
	if val, exists := cache.Get("key3"); !exists || val != 3 {
		t.Fatalf("expected key3 to exist with value 3, got %d, exists %t", val, exists)
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Capacity and Eviction
// ----------------------------------------------------------------------------

func TestLRUCache_EvictLeastRecentlyUsed(t *testing.T) {
	cache := NewLRUCache[string, int](3)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	// All should exist initially
	if _, exists := cache.Get("key1"); !exists {
		t.Error("key1 should exist")
	}

	// Add fourth item, should evict key2 (least recently used after key1 was accessed)
	cache.Set("key4", 4)

	// key1 should still exist (was refreshed by Get)
	if _, exists := cache.Get("key1"); !exists {
		t.Error("key1 should still exist")
	}

	// key2 should be evicted (least recently used)
	if _, exists := cache.Get("key2"); exists {
		t.Error("key2 should have been evicted")
	}

	// key3 and key4 should still exist
	if _, exists := cache.Get("key3"); !exists {
		t.Error("key3 should exist")
	}
	if _, exists := cache.Get("key4"); !exists {
		t.Error("key4 should exist")
	}
}

func TestLRUCache_GetRefreshesRecency(t *testing.T) {
	cache := NewLRUCache[string, int](3)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	// Access key1 to refresh it
	cache.Get("key1")

	// Add key4, should evict key2 (now least recently used)
	cache.Set("key4", 4)

	// key1 should still exist (was refreshed)
	if _, exists := cache.Get("key1"); !exists {
		t.Error("key1 should exist (was refreshed)")
	}

	// key2 should be evicted
	if _, exists := cache.Get("key2"); exists {
		t.Error("key2 should have been evicted")
	}
}

func TestLRUCache_FillToCapacity(t *testing.T) {
	cache := NewLRUCache[int, string](5)

	for i := 1; i <= 5; i++ {
		cache.Set(i, "value")
	}

	// All 5 should exist
	for i := 1; i <= 5; i++ {
		if _, exists := cache.Get(i); !exists {
			t.Errorf("Key %d should exist", i)
		}
	}

	// Add 6th item
	cache.Set(6, "value6")

	// First item should be evicted
	if _, exists := cache.Get(1); exists {
		t.Error("Key 1 should have been evicted")
	}

	// 6th should exist
	if _, exists := cache.Get(6); !exists {
		t.Error("Key 6 should exist")
	}
}

func TestLRUCache_CapacityOne(t *testing.T) {
	cache := NewLRUCache[string, int](1)

	cache.Set("key1", 1)
	val, exists := cache.Get("key1")
	if !exists || val != 1 {
		t.Error("key1 should exist")
	}

	cache.Set("key2", 2)

	// key1 should be evicted
	if _, exists := cache.Get("key1"); exists {
		t.Error("key1 should have been evicted")
	}

	// key2 should exist
	val2, exists := cache.Get("key2")
	if !exists || val2 != 2 {
		t.Error("key2 should exist")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Delete Operations
// ----------------------------------------------------------------------------

func TestLRUCache_Delete_ExistingKey(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("key1", 100)

	deleted := cache.Delete("key1")
	if !deleted {
		t.Error("Delete should return true for existing key")
	}

	_, exists := cache.Get("key1")
	if exists {
		t.Error("key1 should not exist after deletion")
	}
}

func TestLRUCache_Delete_NonExistentKey(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	deleted := cache.Delete("nonexistent")
	if !deleted {
		t.Error("Delete should return true for non-existent key")
	}
}

func TestLRUCache_Delete_AllItems(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	cache.Delete("key1")
	cache.Delete("key2")
	cache.Delete("key3")

	if _, exists := cache.Get("key1"); exists {
		t.Error("key1 should not exist")
	}
	if _, exists := cache.Get("key2"); exists {
		t.Error("key2 should not exist")
	}
	if _, exists := cache.Get("key3"); exists {
		t.Error("key3 should not exist")
	}
}

func TestLRUCache_Delete_ThenReAdd(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("key1", 100)
	cache.Delete("key1")
	cache.Set("key1", 200)

	val, exists := cache.Get("key1")
	if !exists {
		t.Error("key1 should exist after re-adding")
	}
	if val != 200 {
		t.Errorf("Expected value 200, got %d", val)
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Complex Access Patterns
// ----------------------------------------------------------------------------

func TestLRUCache_AccessPattern_FIFO(t *testing.T) {
	cache := NewLRUCache[int, string](3)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	// Access in FIFO order (no refresh)
	cache.Set(4, "four")

	// 1 should be evicted
	if _, exists := cache.Get(1); exists {
		t.Error("1 should have been evicted")
	}
}

func TestLRUCache_AccessPattern_MostRecentlyUsed(t *testing.T) {
	cache := NewLRUCache[int, string](3)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	// Keep accessing 1
	cache.Get(1)
	cache.Set(4, "four")
	cache.Get(1)
	cache.Set(5, "five")
	cache.Get(1)
	cache.Set(6, "six")

	// 1 should still exist (always refreshed)
	if _, exists := cache.Get(1); !exists {
		t.Error("1 should still exist")
	}

	// 2, 3, 4 should be evicted
	if _, exists := cache.Get(2); exists {
		t.Error("2 should have been evicted")
	}
}

func TestLRUCache_AccessPattern_RoundRobin(t *testing.T) {
	cache := NewLRUCache[int, string](3)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	// Access in round-robin
	cache.Get(1)
	cache.Get(2)
	cache.Get(3)
	cache.Get(1)
	cache.Get(2)

	// Add new item, 3 should be evicted (least recently accessed)
	cache.Set(4, "four")

	if _, exists := cache.Get(3); exists {
		t.Error("3 should have been evicted")
	}

	if _, exists := cache.Get(1); !exists {
		t.Error("1 should exist")
	}
	if _, exists := cache.Get(2); !exists {
		t.Error("2 should exist")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Flush Operations
// ----------------------------------------------------------------------------

func TestLRUCache_Refresh_NoEffect(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	cache.Flush()

	// All should still exist
	if _, exists := cache.Get("key1"); !exists {
		t.Error("key1 should exist")
	}
	if _, exists := cache.Get("key2"); !exists {
		t.Error("key2 should exist")
	}
	if _, exists := cache.Get("key3"); !exists {
		t.Error("key3 should exist")
	}
}

func TestLRUCache_Refresh_OverCapacity(t *testing.T) {
	cache := NewLRUCache[int, string](3)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")
	cache.Set(4, "four")
	cache.Set(5, "five")

	// Now cache has 5 items but capacity is 3
	cache.Flush()

	// After refresh, should have only 3 items (most recent)
	// The Flush method calls Shrink which keeps the first N elements
	// Since items 1 and 2 were evicted during Set operations, we have 3, 4, 5
	// Shrink to 3 should keep first 3 elements in the list

	// Verify that refresh completed without error and cache is functional
	cache.Set(6, "six")
	if val, exists := cache.Get(6); !exists || val != "six" {
		t.Error("Cache should be functional after refresh")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Different Data Types
// ----------------------------------------------------------------------------

func TestLRUCache_IntKey_StringValue(t *testing.T) {
	cache := NewLRUCache[int, string](5)

	cache.Set(1, "one")
	cache.Set(2, "two")

	val, exists := cache.Get(1)
	if !exists || val != "one" {
		t.Error("Failed with int key, string value")
	}
}

func TestLRUCache_StringKey_StructValue(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	cache := NewLRUCache[string, User](5)

	cache.Set("user1", User{"Alice", 30})
	cache.Set("user2", User{"Bob", 25})

	val, exists := cache.Get("user1")
	if !exists {
		t.Error("user1 should exist")
	}
	if val.Name != "Alice" || val.Age != 30 {
		t.Error("User data incorrect")
	}
}

func TestLRUCache_PointerValues(t *testing.T) {
	type Data struct {
		Value int
	}

	cache := NewLRUCache[string, *Data](5)

	data1 := &Data{100}
	cache.Set("key1", data1)

	val, exists := cache.Get("key1")
	if !exists {
		t.Error("key1 should exist")
	}
	if val.Value != 100 {
		t.Error("Value incorrect")
	}

	// Modify through pointer
	data1.Value = 200
	val2, _ := cache.Get("key1")
	if val2.Value != 200 {
		t.Error("Pointer modification should be reflected")
	}
}

func TestLRUCache_NilValues(t *testing.T) {
	cache := NewLRUCache[string, *int](5)

	cache.Set("key1", nil)

	val, exists := cache.Get("key1")
	if !exists {
		t.Error("key1 should exist")
	}
	if val != nil {
		t.Error("Value should be nil")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Stress Tests
// ----------------------------------------------------------------------------

func TestLRUCache_LargeCapacity(t *testing.T) {
	cache := NewLRUCache[int, int](1000)

	for i := 0; i < 1000; i++ {
		cache.Set(i, i*2)
	}

	// All should exist
	for i := 0; i < 1000; i++ {
		val, exists := cache.Get(i)
		if !exists {
			t.Errorf("Key %d should exist", i)
		}
		if val != i*2 {
			t.Errorf("Expected %d, got %d", i*2, val)
		}
	}

	// Add one more, first should be evicted
	cache.Set(1000, 2000)
	if _, exists := cache.Get(0); exists {
		t.Error("Key 0 should have been evicted")
	}
}

func TestLRUCache_ManyOperations(t *testing.T) {
	cache := NewLRUCache[int, int](10)

	// Perform many mixed operations
	for i := 0; i < 100; i++ {
		cache.Set(i, i)
		if i%3 == 0 {
			cache.Get(i / 2)
		}
		if i%5 == 0 {
			cache.Delete(i / 3)
		}
	}

	// Cache should still be functional
	cache.Set(200, 200)
	val, exists := cache.Get(200)
	if !exists || val != 200 {
		t.Error("Cache should still work after many operations")
	}
}

func TestLRUCache_AlternatingSetGet(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	for i := 0; i < 20; i++ {
		key := "key" + string(rune('A'+i%5))
		cache.Set(key, i)
		cache.Get(key)
	}

	// Should have 5 most recent keys
	if _, exists := cache.Get("keyE"); !exists {
		t.Error("keyE should exist")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Sequential Eviction
// ----------------------------------------------------------------------------

func TestLRUCache_SequentialEviction(t *testing.T) {
	cache := NewLRUCache[int, int](3)

	// Add 1, 2, 3
	cache.Set(1, 1)
	cache.Set(2, 2)
	cache.Set(3, 3)

	// Add 4, evicts 1
	cache.Set(4, 4)
	if _, exists := cache.Get(1); exists {
		t.Error("1 should be evicted")
	}

	// Add 5, evicts 2
	cache.Set(5, 5)
	if _, exists := cache.Get(2); exists {
		t.Error("2 should be evicted")
	}

	// Add 6, evicts 3
	cache.Set(6, 6)
	if _, exists := cache.Get(3); exists {
		t.Error("3 should be evicted")
	}

	// 4, 5, 6 should exist
	if _, exists := cache.Get(4); !exists {
		t.Error("4 should exist")
	}
	if _, exists := cache.Get(5); !exists {
		t.Error("5 should exist")
	}
	if _, exists := cache.Get(6); !exists {
		t.Error("6 should exist")
	}
}

func TestLRUCache_MultipleGetsSameKey(t *testing.T) {
	cache := NewLRUCache[string, int](3)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	// Get key1 multiple times
	for i := 0; i < 10; i++ {
		cache.Get("key1")
	}

	// Add new keys
	cache.Set("key4", 4)
	cache.Set("key5", 5)

	// key1 should still exist (was frequently accessed)
	if _, exists := cache.Get("key1"); !exists {
		t.Error("key1 should still exist")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Empty Cache Operations
// ----------------------------------------------------------------------------

func TestLRUCache_EmptyCache_Get(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	_, exists := cache.Get("anything")
	if exists {
		t.Error("Get on empty cache should return false")
	}
}

func TestLRUCache_EmptyCache_Delete(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	deleted := cache.Delete("anything")
	if !deleted {
		t.Error("Delete on empty cache should return true")
	}
}

func TestLRUCache_EmptyCache_Refresh(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	// Should not panic
	cache.Flush()
}

// ----------------------------------------------------------------------------
// Edge Cases: Zero and Negative Values
// ----------------------------------------------------------------------------

func TestLRUCache_ZeroValue(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("zero", 0)

	val, exists := cache.Get("zero")
	if !exists {
		t.Error("zero should exist")
	}
	if val != 0 {
		t.Errorf("Expected 0, got %d", val)
	}
}

func TestLRUCache_NegativeValue(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("negative", -100)

	val, exists := cache.Get("negative")
	if !exists {
		t.Error("negative should exist")
	}
	if val != -100 {
		t.Errorf("Expected -100, got %d", val)
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: String Keys with Special Characters
// ----------------------------------------------------------------------------

func TestLRUCache_EmptyStringKey(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("", 100)

	val, exists := cache.Get("")
	if !exists {
		t.Error("Empty string key should exist")
	}
	if val != 100 {
		t.Errorf("Expected 100, got %d", val)
	}
}

func TestLRUCache_SpecialCharacterKeys(t *testing.T) {
	cache := NewLRUCache[string, string](10)

	specialKeys := []string{
		"key with spaces",
		"key\twith\ttabs",
		"key\nwith\nnewlines",
		"key!@#$%^&*()",
		"unicode-键值",
	}

	for _, key := range specialKeys {
		cache.Set(key, "value-"+key)
	}

	for _, key := range specialKeys {
		val, exists := cache.Get(key)
		if !exists {
			t.Errorf("Key '%s' should exist", key)
		}
		expected := "value-" + key
		if val != expected {
			t.Errorf("For key '%s': expected '%s', got '%s'", key, expected, val)
		}
	}
}

func TestLRUCache_Clear_EmptyCache(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	// Clear on empty cache should not panic
	cache.Clear()

	// Verify cache is empty
	if _, exists := cache.Get("anything"); exists {
		t.Error("Cache should be empty after Clear")
	}
}

func TestLRUCache_Clear_WithItems(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	// Clear the cache
	cache.Clear()

	// All items should be removed
	if _, exists := cache.Get("key1"); exists {
		t.Error("key1 should not exist after Clear")
	}
	if _, exists := cache.Get("key2"); exists {
		t.Error("key2 should not exist after Clear")
	}
	if _, exists := cache.Get("key3"); exists {
		t.Error("key3 should not exist after Clear")
	}
}

func TestLRUCache_Clear_AfterEviction(t *testing.T) {
	cache := NewLRUCache[int, string](3)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")
	cache.Set(4, "four") // Evicts 1

	// Clear the cache
	cache.Clear()

	// All remaining items should be removed
	for i := 1; i <= 4; i++ {
		if _, exists := cache.Get(i); exists {
			t.Errorf("Key %d should not exist after Clear", i)
		}
	}
}

func TestLRUCache_Clear_ThenReuse(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Clear()

	// Cache should be functional after Clear
	cache.Set("key3", 3)
	cache.Set("key4", 4)

	val, exists := cache.Get("key3")
	if !exists || val != 3 {
		t.Error("Cache should be functional after Clear")
	}
	val, exists = cache.Get("key4")
	if !exists || val != 4 {
		t.Error("Cache should be functional after Clear")
	}

	// Old items should not exist
	if _, exists := cache.Get("key1"); exists {
		t.Error("key1 should not exist")
	}
	if _, exists := cache.Get("key2"); exists {
		t.Error("key2 should not exist")
	}
}

func TestLRUCache_Clear_MultipleTimes(t *testing.T) {
	cache := NewLRUCache[int, int](5)

	cache.Set(1, 100)
	cache.Clear()
	cache.Clear()
	cache.Clear()

	// Cache should still be functional
	cache.Set(2, 200)
	val, exists := cache.Get(2)
	if !exists || val != 200 {
		t.Error("Cache should be functional after multiple Clears")
	}
}

func TestLRUCache_Clear_WithLargeDataset(t *testing.T) {
	cache := NewLRUCache[int, int](100)

	// Add many items
	for i := 0; i < 100; i++ {
		cache.Set(i, i*2)
	}

	// Clear all
	cache.Clear()

	// Verify all removed
	for i := 0; i < 100; i++ {
		if _, exists := cache.Get(i); exists {
			t.Errorf("Key %d should not exist after Clear", i)
		}
	}

	// Cache should still be functional
	cache.Set(999, 1998)
	val, exists := cache.Get(999)
	if !exists || val != 1998 {
		t.Error("Cache should be functional after Clear")
	}
}

func TestLRUCache_Clear_ZeroCapacity(t *testing.T) {
	cache := NewLRUCache[string, int](0)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	// Clear the cache
	cache.Clear()

	// All items should be removed
	if _, exists := cache.Get("key1"); exists {
		t.Error("key1 should not exist after Clear")
	}
	if _, exists := cache.Get("key2"); exists {
		t.Error("key2 should not exist after Clear")
	}
	if _, exists := cache.Get("key3"); exists {
		t.Error("key3 should not exist after Clear")
	}

	// Cache should still be functional
	cache.Set("key4", 4)
	val, exists := cache.Get("key4")
	if !exists || val != 4 {
		t.Error("Cache should be functional after Clear")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Recency and Access Order
// ----------------------------------------------------------------------------

func TestLRUCache_RecencyOrder_AfterMultipleGets(t *testing.T) {
	cache := NewLRUCache[int, string](3)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	// Access 1 multiple times (should stay most recent)
	cache.Get(1)
	cache.Get(1)
	cache.Get(1)

	// Add 4, should evict 2 (least recent)
	cache.Set(4, "four")

	if _, exists := cache.Get(1); !exists {
		t.Error("1 should exist (most recently accessed)")
	}
	if _, exists := cache.Get(2); exists {
		t.Error("2 should be evicted (least recent)")
	}
	if _, exists := cache.Get(3); !exists {
		t.Error("3 should exist")
	}
	if _, exists := cache.Get(4); !exists {
		t.Error("4 should exist")
	}
}

func TestLRUCache_MixedSetAndGetOrder(t *testing.T) {
	cache := NewLRUCache[int, string](3)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Get(1) // 1 is now most recent
	cache.Set(3, "three")
	cache.Get(2) // 2 is now most recent

	// Order: 2 (most recent), 3, 1 (least recent)
	// Add 4, should evict 1
	cache.Set(4, "four")

	if _, exists := cache.Get(1); exists {
		t.Error("1 should be evicted")
	}
	if _, exists := cache.Get(2); !exists {
		t.Error("2 should exist")
	}
	if _, exists := cache.Get(3); !exists {
		t.Error("3 should exist")
	}
	if _, exists := cache.Get(4); !exists {
		t.Error("4 should exist")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Delete Edge Cases
// ----------------------------------------------------------------------------

func TestLRUCache_Delete_MostRecentItem(t *testing.T) {
	cache := NewLRUCache[int, string](3)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	// Add new item
	cache.Set(4, "four")

	// 3 is most recent, delete it
	if !cache.Delete(3) {
		t.Error("Delete should return true")
	}

	// 1 should be evicted (least recent)
	if _, exists := cache.Get(1); exists {
		t.Error("1 should be evicted")
	}
	if _, exists := cache.Get(2); !exists {
		t.Error("2 should exist")
	}
	if _, exists := cache.Get(3); exists {
		t.Error("3 should not exist (was deleted)")
	}
	if _, exists := cache.Get(4); !exists {
		t.Error("4 should exist")
	}
}

func TestLRUCache_MultipleDeletesSameKey(t *testing.T) {
	cache := NewLRUCache[string, int](5)

	cache.Set("key1", 100)

	// Delete multiple times
	if !cache.Delete("key1") {
		t.Error("First delete should return true")
	}
	if !cache.Delete("key1") {
		t.Error("Second delete should return true")
	}
	if !cache.Delete("key1") {
		t.Error("Third delete should return true")
	}

	// Should not exist
	_, exists := cache.Get("key1")
	if exists {
		t.Error("key1 should not exist")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Additional Flush Tests
// ----------------------------------------------------------------------------

func TestLRUCache_Refresh_ZeroCapacity(t *testing.T) {
	cache := NewLRUCache[string, int](0)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	// Zero capacity means unlimited, so Flush should not remove items.
	cache.Flush()

	if _, exists := cache.Get("key1"); !exists {
		t.Error("key1 should be kept after Flush with zero capacity")
	}
	if _, exists := cache.Get("key2"); !exists {
		t.Error("key2 should be kept after Flush with zero capacity")
	}
	if _, exists := cache.Get("key3"); !exists {
		t.Error("key3 should be kept after Flush with zero capacity")
	}

	// Cache should still be functional after Flush
	cache.Set("key4", 4)
	val, exists := cache.Get("key4")
	if !exists || val != 4 {
		t.Error("Cache should be functional after Flush")
	}
}

func TestLRUCache_MultipleRefreshCalls(t *testing.T) {
	cache := NewLRUCache[int, int](5)

	cache.Set(1, 1)
	cache.Set(2, 2)
	cache.Set(3, 3)

	// Multiple refresh calls
	cache.Flush()
	cache.Flush()
	cache.Flush()

	// All should still exist
	if _, exists := cache.Get(1); !exists {
		t.Error("1 should exist")
	}
	if _, exists := cache.Get(2); !exists {
		t.Error("2 should exist")
	}
	if _, exists := cache.Get(3); !exists {
		t.Error("3 should exist")
	}
}

func TestLRUCache_RefreshAfterEviction(t *testing.T) {
	cache := NewLRUCache[int, string](3)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")
	cache.Set(4, "four") // Evicts 1

	// Flush should not change anything
	cache.Flush()

	if _, exists := cache.Get(1); exists {
		t.Error("1 should still be evicted after refresh")
	}
	if _, exists := cache.Get(2); !exists {
		t.Error("2 should exist")
	}
	if _, exists := cache.Get(3); !exists {
		t.Error("3 should exist")
	}
	if _, exists := cache.Get(4); !exists {
		t.Error("4 should exist")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Boundary Conditions
// ----------------------------------------------------------------------------

func TestLRUCache_SetGetImmediately(t *testing.T) {
	cache := NewLRUCache[string, string](5)

	for i := 0; i < 10; i++ {
		key := "key"
		cache.Set(key, "value")
		val, exists := cache.Get(key)
		if !exists {
			t.Error("Key should exist immediately after Set")
		}
		if val != "value" {
			t.Errorf("Expected 'value', got '%s'", val)
		}
	}
}

func TestLRUCache_UpdateRecencyOnGet(t *testing.T) {
	cache := NewLRUCache[int, string](2)

	cache.Set(1, "one")
	cache.Set(2, "two")

	// Access 1 to update recency
	cache.Get(1)

	// Add 3, should evict 2
	cache.Set(3, "three")

	if _, exists := cache.Get(1); !exists {
		t.Error("1 should exist (recency updated)")
	}
	if _, exists := cache.Get(2); exists {
		t.Error("2 should be evicted")
	}
	if _, exists := cache.Get(3); !exists {
		t.Error("3 should exist")
	}
}

func TestLRUCache_NoEvictionWhenNotFull(t *testing.T) {
	cache := NewLRUCache[int, int](10)

	for i := 1; i <= 5; i++ {
		cache.Set(i, i*10)
	}

	// All should exist (not at capacity)
	for i := 1; i <= 5; i++ {
		if _, exists := cache.Get(i); !exists {
			t.Errorf("Key %d should exist", i)
		}
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Value Preservation
// ----------------------------------------------------------------------------

func TestLRUCache_ValuePreservation_AfterGet(t *testing.T) {
	cache := NewLRUCache[string, string](5)

	cache.Set("key", "value")

	// Multiple gets should preserve value
	for i := 0; i < 10; i++ {
		val, exists := cache.Get("key")
		if !exists {
			t.Error("key should exist")
		}
		if val != "value" {
			t.Errorf("Value should remain 'value', got '%s'", val)
		}
	}
}

func TestLRUCache_ValuePreservation_DuringEviction(t *testing.T) {
	cache := NewLRUCache[int, string](3)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	// Access 2 and 3 to update recency
	cache.Get(2)
	cache.Get(3)

	// Add more items to trigger evictions
	cache.Set(4, "four")
	cache.Set(5, "five")

	// Verify values are preserved for non-evicted items
	if val, exists := cache.Get(3); !exists || val != "three" {
		t.Error("Value for 3 should be preserved")
	}
	if val, exists := cache.Get(4); !exists || val != "four" {
		t.Error("Value for 4 should be preserved")
	}
	if val, exists := cache.Get(5); !exists || val != "five" {
		t.Error("Value for 5 should be preserved")
	}
}
