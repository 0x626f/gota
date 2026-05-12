package cache

import (
	"testing"
)

// ============================================================================
// COMPREHENSIVE TEST SUITE FOR LFU CACHE
// ============================================================================

// ----------------------------------------------------------------------------
// Edge Cases: Basic Operations
// ----------------------------------------------------------------------------

func TestLFUCache_NewCache(t *testing.T) {
	cache := NewLFUCache[string, int](5)
	if cache == nil {
		t.Fatal("NewLFUCache returned nil")
	}
	if cache.capacity != 5 {
		t.Errorf("Expected capacity 5, got %d", cache.capacity)
	}
}

func TestLFUCache_NewCache_ZeroCapacity(t *testing.T) {
	cache := NewLFUCache[string, int](0)
	if cache == nil {
		t.Fatal("NewLFUCache returned nil")
	}
	if cache.capacity != 0 {
		t.Errorf("Expected capacity 0, got %d", cache.capacity)
	}
}

func TestLFUCache_NewCache_NegativeCapacity(t *testing.T) {
	cache := NewLFUCache[string, int](-1)
	if cache.capacity != 0 {
		t.Errorf("Expected negative capacity to become 0, got %d", cache.capacity)
	}

	cache.Set("key1", 1)
	cache.Flush()

	if val, exists := cache.Get("key1"); !exists || val != 1 {
		t.Error("negative capacity should behave as unlimited")
	}
}

func TestLFUCache_SetAndGet_SingleItem(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("key1", 100)

	val, exists := cache.Get("key1")
	if !exists {
		t.Error("Expected key1 to exist")
	}
	if val != 100 {
		t.Errorf("Expected value 100, got %d", val)
	}
}

func TestLFUCache_Get_NonExistentKey(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	val, exists := cache.Get("nonexistent")
	if exists {
		t.Error("Expected key to not exist")
	}
	if val != 0 {
		t.Errorf("Expected zero value, got %d", val)
	}
}

func TestLFUCache_Set_UpdatesExistingKey(t *testing.T) {
	cache := NewLFUCache[string, int](10)

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

// ----------------------------------------------------------------------------
// Edge Cases: Frequency Tracking
// ----------------------------------------------------------------------------

func TestLFUCache_Get_IncreasesFrequency(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("key1", 100)

	// Get multiple times to increase frequency
	cache.Get("key1")
	cache.Get("key1")
	cache.Get("key1")

	// Should still exist and return correct value
	val, exists := cache.Get("key1")
	if !exists {
		t.Error("key1 should exist")
	}
	if val != 100 {
		t.Errorf("Expected value 100, got %d", val)
	}
}

func TestLFUCache_MultipleKeys_DifferentFrequencies(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	// Access key1 three times
	cache.Get("key1")
	cache.Get("key1")
	cache.Get("key1")

	// Access key2 once
	cache.Get("key2")

	// key3 not accessed after Set

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

func TestLFUCache_FrequencyProgression(t *testing.T) {
	cache := NewLFUCache[int, string](10)

	cache.Set(1, "one")

	// Start at frequency 1
	val1, _ := cache.Get(1)
	if val1 != "one" {
		t.Error("Value should be 'one'")
	}

	// Now at frequency 2
	val2, _ := cache.Get(1)
	if val2 != "one" {
		t.Error("Value should still be 'one'")
	}

	// Now at frequency 3
	val3, _ := cache.Get(1)
	if val3 != "one" {
		t.Error("Value should still be 'one'")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Delete Operations
// ----------------------------------------------------------------------------

func TestLFUCache_Delete_ExistingKey(t *testing.T) {
	cache := NewLFUCache[string, int](10)

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

func TestLFUCache_Delete_NonExistentKey(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	deleted := cache.Delete("nonexistent")
	if !deleted {
		t.Error("Delete should return true even for non-existent key")
	}
}

func TestLFUCache_Delete_HighFrequencyKey(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("key1", 100)

	// Access many times
	for i := 0; i < 10; i++ {
		cache.Get("key1")
	}

	deleted := cache.Delete("key1")
	if !deleted {
		t.Error("Should be able to delete high frequency key")
	}

	_, exists := cache.Get("key1")
	if exists {
		t.Error("key1 should not exist after deletion")
	}
}

func TestLFUCache_Delete_ThenReAdd(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("key1", 100)
	cache.Get("key1")
	cache.Get("key1")

	cache.Delete("key1")
	cache.Set("key1", 200)

	// Should reset to frequency 1
	val, exists := cache.Get("key1")
	if !exists {
		t.Error("key1 should exist after re-adding")
	}
	if val != 200 {
		t.Errorf("Expected value 200, got %d", val)
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Different Data Types
// ----------------------------------------------------------------------------

func TestLFUCache_IntKey_StringValue(t *testing.T) {
	cache := NewLFUCache[int, string](10)

	cache.Set(1, "one")
	cache.Set(2, "two")

	val, exists := cache.Get(1)
	if !exists || val != "one" {
		t.Error("Failed with int key, string value")
	}
}

func TestLFUCache_StringKey_StructValue(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	cache := NewLFUCache[string, User](10)

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

func TestLFUCache_PointerValues(t *testing.T) {
	type Data struct {
		Value int
	}

	cache := NewLFUCache[string, *Data](10)

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

func TestLFUCache_NilValues(t *testing.T) {
	cache := NewLFUCache[string, *int](10)

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
// Edge Cases: Complex Access Patterns
// ----------------------------------------------------------------------------

func TestLFUCache_MixedAccessPattern(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	// Access 'a' 5 times
	for i := 0; i < 5; i++ {
		cache.Get("a")
	}

	// Access 'b' 2 times
	cache.Get("b")
	cache.Get("b")

	// 'c' accessed only once (from Set)

	// All should exist
	if _, exists := cache.Get("a"); !exists {
		t.Error("a should exist")
	}
	if _, exists := cache.Get("b"); !exists {
		t.Error("b should exist")
	}
	if _, exists := cache.Get("c"); !exists {
		t.Error("c should exist")
	}
}

func TestLFUCache_AlternatingAccess(t *testing.T) {
	cache := NewLFUCache[int, string](10)

	cache.Set(1, "one")
	cache.Set(2, "two")

	// Alternate between them
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			cache.Get(1)
		} else {
			cache.Get(2)
		}
	}

	// Both should exist
	if _, exists := cache.Get(1); !exists {
		t.Error("1 should exist")
	}
	if _, exists := cache.Get(2); !exists {
		t.Error("2 should exist")
	}
}

func TestLFUCache_SingleKeyHighFrequency(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("hot", 999)

	// Access many times
	for i := 0; i < 100; i++ {
		val, exists := cache.Get("hot")
		if !exists || val != 999 {
			t.Errorf("hot key should always exist and return 999")
		}
	}
}

func TestLFUCache_Refresh_WithinCapacity(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	cache.Flush()

	// With capacity 10 and only 3 frequency buckets, refresh shouldn't affect anything
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

func TestLFUCache_Refresh_ExceedsCapacity(t *testing.T) {
	cache := NewLFUCache[string, int](2)

	// Create 3 different frequency buckets
	cache.Set("key1", 1) // freq 1
	cache.Set("key2", 2) // freq 1
	cache.Set("key3", 3) // freq 1

	// Access key1 once (freq 2)
	cache.Get("key1")

	// Access key2 twice (freq 3)
	cache.Get("key2")
	cache.Get("key2")

	// Now we have 3 frequency buckets: freq 1 (key3), freq 2 (key1), freq 3 (key2)
	// Capacity is 2, so refresh keeps 2 HIGHEST frequency buckets
	// Sort descending: freq 3, freq 2, freq 1
	// Keep first 2: freq 3 (key2), freq 2 (key1)
	// Remove: freq 1 (key3)

	cache.Flush()

	// High frequency items should be kept
	if _, exists := cache.Get("key2"); !exists {
		t.Error("key2 (freq 3) should be kept")
	}
	if _, exists := cache.Get("key1"); !exists {
		t.Error("key1 (freq 2) should be kept")
	}

	// Low frequency item should be removed
	if _, exists := cache.Get("key3"); exists {
		t.Error("key3 (freq 1) should be removed")
	}

	// Verify cache is still functional
	cache.Set("key4", 4)
	if val, exists := cache.Get("key4"); !exists || val != 4 {
		t.Error("Cache should be functional after refresh")
	}
}

func TestLFUCache_Refresh_UsesItemCapacity(t *testing.T) {
	cache := NewLFUCache[string, int](2)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	cache.Flush()

	kept := 0
	for _, key := range []string{"key1", "key2", "key3"} {
		if _, exists := cache.Get(key); exists {
			kept++
		}
	}

	if kept != 2 {
		t.Errorf("expected item capacity to keep 2 items, got %d", kept)
	}
}

func TestLFUCache_Refresh_IgnoresEmptyFrequencyBuckets(t *testing.T) {
	cache := NewLFUCache[string, int](1)

	cache.Set("key1", 1)
	cache.Get("key1")
	cache.Delete("key1")
	cache.Set("key2", 2)

	cache.Flush()

	if val, exists := cache.Get("key2"); !exists || val != 2 {
		t.Error("Flush should keep live items instead of empty frequency buckets")
	}
}

func TestLFUCache_Refresh_SortsAndShrinks(t *testing.T) {
	cache := NewLFUCache[int, string](3)

	// Create multiple frequency levels
	cache.Set(1, "one")   // freq 1
	cache.Set(2, "two")   // freq 1
	cache.Set(3, "three") // freq 1
	cache.Set(4, "four")  // freq 1
	cache.Set(5, "five")  // freq 1

	// Access to create different frequencies
	cache.Get(1) // freq 2
	cache.Get(1) // freq 3

	cache.Get(2) // freq 2
	cache.Get(2) // freq 3
	cache.Get(2) // freq 4

	cache.Get(3) // freq 2

	// Frequency buckets: freq 1 (keys 4,5), freq 2 (key 3), freq 3 (key 1), freq 4 (key 2)
	// Sort descending: freq 4, freq 3, freq 2, freq 1
	// Keep first 3: freq 4 (key 2), freq 3 (key 1), freq 2 (key 3)
	// Remove: freq 1 (keys 4, 5)

	cache.Flush()

	// High frequency items should be kept
	if _, exists := cache.Get(2); !exists {
		t.Error("Key 2 (freq 4) should be kept")
	}
	if _, exists := cache.Get(1); !exists {
		t.Error("Key 1 (freq 3) should be kept")
	}
	if _, exists := cache.Get(3); !exists {
		t.Error("Key 3 (freq 2) should be kept")
	}

	// Low frequency items should be removed
	if _, exists := cache.Get(4); exists {
		t.Error("Key 4 (freq 1) should be removed")
	}
	if _, exists := cache.Get(5); exists {
		t.Error("Key 5 (freq 1) should be removed")
	}

	// Verify cache is still functional after refresh
	cache.Set(10, "ten")
	if val, exists := cache.Get(10); !exists || val != "ten" {
		t.Error("Cache should be functional after refresh")
	}
}

func TestLFUCache_Refresh_ZeroCapacity(t *testing.T) {
	cache := NewLFUCache[string, int](0)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	cache.Get("key1")
	cache.Get("key2")
	cache.Get("key2")

	// Zero capacity means unlimited, so Flush should not remove items.
	cache.Flush()

	if _, exists := cache.Get("key1"); !exists {
		t.Error("key1 should be kept with zero capacity")
	}
	if _, exists := cache.Get("key2"); !exists {
		t.Error("key2 should be kept with zero capacity")
	}
	if _, exists := cache.Get("key3"); !exists {
		t.Error("key3 should be kept with zero capacity")
	}
}

func TestLFUCache_Refresh_CapacityOne(t *testing.T) {
	cache := NewLFUCache[string, int](1)

	cache.Set("key1", 1) // freq 1
	cache.Set("key2", 2) // freq 1

	cache.Get("key1") // freq 2
	cache.Get("key2") // freq 2

	cache.Get("key1") // freq 3

	// Frequency buckets: freq 2 (key2), freq 3 (key1)
	// Sort descending: freq 3, freq 2
	// Keep first 1: freq 3 (key1)
	// Remove: freq 2 (key2)

	cache.Flush()

	// Highest frequency item should be kept
	if _, exists := cache.Get("key1"); !exists {
		t.Error("key1 (freq 3) should be kept")
	}

	// Lower frequency item should be removed
	if _, exists := cache.Get("key2"); exists {
		t.Error("key2 (freq 2) should be removed")
	}

	// Verify cache is still functional
	cache.Set("key3", 3)
	if val, exists := cache.Get("key3"); !exists || val != 3 {
		t.Error("Cache should be functional after refresh")
	}
}

func TestLFUCache_Refresh_MultipleKeysPerFrequency(t *testing.T) {
	cache := NewLFUCache[int, string](2)

	// Create many keys at same frequencies
	for i := 1; i <= 10; i++ {
		cache.Set(i, "value")
	}

	// All are at frequency 1
	// Access some to frequency 2
	for i := 1; i <= 5; i++ {
		cache.Get(i)
	}

	// Access some to frequency 3
	for i := 1; i <= 3; i++ {
		cache.Get(i)
	}

	cache.Flush()

	// Capacity is item-based, so only two of the highest-frequency items remain.
	kept := 0
	for i := 1; i <= 3; i++ {
		if _, exists := cache.Get(i); !exists {
			continue
		}
		kept++
	}

	if kept != 2 {
		t.Errorf("expected 2 highest-frequency keys to be kept, got %d", kept)
	}

	for i := 4; i <= 10; i++ {
		if _, exists := cache.Get(i); exists {
			t.Errorf("Key %d should be removed", i)
		}
	}

	// Verify cache is still functional with multiple keys
	cache.Set(20, "twenty")
	if val, exists := cache.Get(20); !exists || val != "twenty" {
		t.Error("Cache should be functional after refresh")
	}
}

func TestLFUCache_Refresh_AfterDeletes(t *testing.T) {
	cache := NewLFUCache[string, int](3)

	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	cache.Get("key1")
	cache.Get("key2")
	cache.Get("key2")

	// Delete a key
	cache.Delete("key2")

	// Flush should still work
	cache.Flush()

	// key1 should still exist
	if _, exists := cache.Get("key1"); !exists {
		t.Error("key1 should exist")
	}
}

func TestLFUCache_Refresh_EmptyFrequencyBuckets(t *testing.T) {
	cache := NewLFUCache[string, int](5)

	// Create and delete keys to potentially create empty frequency buckets
	cache.Set("key1", 1)
	cache.Get("key1")
	cache.Delete("key1")

	cache.Set("key2", 2)
	cache.Get("key2")
	cache.Get("key2")
	cache.Delete("key2")

	cache.Set("key3", 3)

	// Flush should handle empty frequency buckets gracefully
	cache.Flush()

	// key3 should still exist
	if _, exists := cache.Get("key3"); !exists {
		t.Error("key3 should exist")
	}
}

func TestLFUCache_Refresh_PreservesData(t *testing.T) {
	cache := NewLFUCache[string, string](5)

	cache.Set("a", "apple")
	cache.Set("b", "banana")
	cache.Set("c", "cherry")

	cache.Get("a")
	cache.Get("b")
	cache.Get("b")

	cache.Flush()

	// Verify data is preserved
	val, exists := cache.Get("a")
	if !exists || val != "apple" {
		t.Error("Data should be preserved after refresh")
	}

	val, exists = cache.Get("b")
	if !exists || val != "banana" {
		t.Error("Data should be preserved after refresh")
	}

	val, exists = cache.Get("c")
	if !exists || val != "cherry" {
		t.Error("Data should be preserved after refresh")
	}
}

func TestLFUCache_Refresh_LargeCapacity(t *testing.T) {
	cache := NewLFUCache[int, int](1000)

	// Add many keys with different frequencies
	for i := 0; i < 100; i++ {
		cache.Set(i, i)
		for j := 0; j <= i%10; j++ {
			cache.Get(i)
		}
	}

	// Flush with large capacity should not affect much
	cache.Flush()

	// Verify some keys still exist
	for i := 0; i < 10; i++ {
		if _, exists := cache.Get(i); !exists {
			t.Errorf("Key %d should exist after refresh", i)
		}
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Empty Cache Operations
// ----------------------------------------------------------------------------

func TestLFUCache_EmptyCache_Get(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	_, exists := cache.Get("anything")
	if exists {
		t.Error("Get on empty cache should return false")
	}
}

func TestLFUCache_EmptyCache_Delete(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	deleted := cache.Delete("anything")
	if !deleted {
		t.Error("Delete on empty cache should return true")
	}
}

func TestLFUCache_EmptyCache_Refresh(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	// Should not panic
	cache.Flush()
}

// ----------------------------------------------------------------------------
// Edge Cases: Stress Tests
// ----------------------------------------------------------------------------

func TestLFUCache_ManyKeys(t *testing.T) {
	cache := NewLFUCache[int, int](10)

	// Add many keys
	for i := 0; i < 100; i++ {
		cache.Set(i, i*2)
	}

	// All should exist
	for i := 0; i < 100; i++ {
		val, exists := cache.Get(i)
		if !exists {
			t.Errorf("Key %d should exist", i)
		}
		if val != i*2 {
			t.Errorf("Expected %d, got %d", i*2, val)
		}
	}
}

func TestLFUCache_ManyOperations(t *testing.T) {
	cache := NewLFUCache[int, int](10)

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

func TestLFUCache_RepeatedSetGet(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	for i := 0; i < 50; i++ {
		key := "key"
		cache.Set(key, i)
		cache.Get(key)
	}

	// Should have the latest value.
	val, exists := cache.Get("key")
	if !exists {
		t.Error("key should exist")
	}
	if val != 49 {
		t.Errorf("Expected latest value 49, got %d", val)
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Sequential Operations
// ----------------------------------------------------------------------------

func TestLFUCache_SetGetDeleteSequence(t *testing.T) {
	cache := NewLFUCache[int, string](10)

	// Set
	cache.Set(1, "one")

	// Get
	val, exists := cache.Get(1)
	if !exists || val != "one" {
		t.Error("Should be able to get after set")
	}

	// Delete
	cache.Delete(1)

	// Get again (should not exist)
	_, exists = cache.Get(1)
	if exists {
		t.Error("Should not exist after delete")
	}
}

func TestLFUCache_MultipleDeletesSameKey(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("key1", 100)

	// Delete multiple times
	cache.Delete("key1")
	cache.Delete("key1")
	cache.Delete("key1")

	// Should handle gracefully
	_, exists := cache.Get("key1")
	if exists {
		t.Error("key1 should not exist")
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Frequency Bucket Management
// ----------------------------------------------------------------------------

func TestLFUCache_SameFrequencyMultipleKeys(t *testing.T) {
	cache := NewLFUCache[int, string](10)

	// Add multiple keys with same frequency
	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	// All start at frequency 1
	// Access all once more (frequency 2)
	cache.Get(1)
	cache.Get(2)
	cache.Get(3)

	// All should exist
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

func TestLFUCache_TransitionBetweenFrequencies(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("key", 100)

	// Transition through multiple frequency levels
	for i := 0; i < 10; i++ {
		val, exists := cache.Get("key")
		if !exists {
			t.Errorf("key should exist at iteration %d", i)
		}
		if val != 100 {
			t.Errorf("Expected value 100, got %d at iteration %d", val, i)
		}
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Value Updates After Get
// ----------------------------------------------------------------------------

func TestLFUCache_ValuePersistsAcrossFrequencies(t *testing.T) {
	cache := NewLFUCache[string, string](10)

	cache.Set("key", "original")

	// Move through several frequency levels
	for i := 0; i < 5; i++ {
		val, exists := cache.Get("key")
		if !exists {
			t.Error("key should exist")
		}
		if val != "original" {
			t.Errorf("Value should remain 'original', got '%s'", val)
		}
	}
}

// ----------------------------------------------------------------------------
// Edge Cases: Zero and Negative Values
// ----------------------------------------------------------------------------

func TestLFUCache_ZeroValue(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("zero", 0)

	val, exists := cache.Get("zero")
	if !exists {
		t.Error("zero should exist")
	}
	if val != 0 {
		t.Errorf("Expected 0, got %d", val)
	}
}

func TestLFUCache_NegativeValue(t *testing.T) {
	cache := NewLFUCache[string, int](10)

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

func TestLFUCache_EmptyStringKey(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	cache.Set("", 100)

	val, exists := cache.Get("")
	if !exists {
		t.Error("Empty string key should exist")
	}
	if val != 100 {
		t.Errorf("Expected 100, got %d", val)
	}
}

func TestLFUCache_SpecialCharacterKeys(t *testing.T) {
	cache := NewLFUCache[string, string](10)

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

func TestLFUCache_Clear_EmptyCache(t *testing.T) {
	cache := NewLFUCache[string, int](10)

	// Clear on empty cache should not panic
	cache.Clear()

	// Verify cache is empty
	if _, exists := cache.Get("anything"); exists {
		t.Error("Cache should be empty after Clear")
	}
}

func TestLFUCache_Clear_WithItems(t *testing.T) {
	cache := NewLFUCache[string, int](10)

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

func TestLFUCache_Clear_DifferentFrequencies(t *testing.T) {
	cache := NewLFUCache[int, string](10)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	// Create different frequencies
	cache.Get(1)
	cache.Get(1)
	cache.Get(2)

	// Clear the cache
	cache.Clear()

	// All items should be removed regardless of frequency
	if _, exists := cache.Get(1); exists {
		t.Error("1 should not exist after Clear")
	}
	if _, exists := cache.Get(2); exists {
		t.Error("2 should not exist after Clear")
	}
	if _, exists := cache.Get(3); exists {
		t.Error("3 should not exist after Clear")
	}
}

func TestLFUCache_Clear_ThenReuse(t *testing.T) {
	cache := NewLFUCache[string, int](5)

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

func TestLFUCache_Clear_MultipleTimes(t *testing.T) {
	cache := NewLFUCache[int, int](5)

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

func TestLFUCache_Clear_WithLargeDataset(t *testing.T) {
	cache := NewLFUCache[int, int](100)

	// Add many items
	for i := 0; i < 100; i++ {
		cache.Set(i, i*2)
		cache.Get(i)
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
