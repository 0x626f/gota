package array

import (
	"testing"

	"github.com/0x626f/gota/collections"
)

// --- helpers ---

func intCmp(a, b int) int {
	if a < b {
		return collections.LOWER
	}
	if a > b {
		return collections.GREATER
	}
	return collections.EQUAL
}

func collect(a collections.Collection[int, int]) []int {
	out := make([]int, 0, a.Size())
	a.ForEach(func(_ int, v int) bool { out = append(out, v); return true })
	return out
}

// --- constructors ---

func TestNew_IsEmpty(t *testing.T) {
	a := New[int]()
	if !a.IsEmpty() || a.Size() != 0 {
		t.Errorf("expected empty array")
	}
}

func TestNew_WithPreSize_IsEmptyButReservesCapacity(t *testing.T) {
	a := New[int](10)
	if !a.IsEmpty() || a.Size() != 0 {
		t.Errorf("New(preSize): expected empty array, got size %d", a.Size())
	}
	if cap(a.items) < 10 {
		t.Errorf("New(preSize): expected capacity >= 10, got %d", cap(a.items))
	}
}

func TestNew_WithZeroPreSize_NoBehaviourChange(t *testing.T) {
	a := New[int](0)
	if !a.IsEmpty() {
		t.Errorf("New(0): expected empty array")
	}
}

func TestNew_WithNegativePreSize_NoBehaviourChange(t *testing.T) {
	a := New[int](-1)
	if !a.IsEmpty() {
		t.Errorf("New(-1): expected empty array")
	}
}

func TestNew_WithPreSize_FunctionalBeyondCapacity(t *testing.T) {
	a := New[int](3)
	for i := range 6 { // push twice the reserved capacity
		a.Push(i)
	}
	if a.Size() != 6 {
		t.Errorf("New(preSize): expected size 6, got %d", a.Size())
	}
	if a.At(5) != 5 {
		t.Errorf("New(preSize): unexpected value at index 5")
	}
}

func TestWrap_ContainsItems(t *testing.T) {
	a := Wrap([]int{1, 2, 3})
	if a.Size() != 3 || a.At(0) != 1 || a.At(2) != 3 {
		t.Errorf("Wrap: unexpected contents")
	}
}

func TestFrom_ContainsItems(t *testing.T) {
	a := From(4, 5, 6)
	if a.Size() != 3 || a.At(0) != 4 || a.At(2) != 6 {
		t.Errorf("From: unexpected contents")
	}
}

// --- At / Get ---

func TestAt_ReturnsElement(t *testing.T) {
	a := From(10, 20, 30)
	if a.At(0) != 10 || a.At(1) != 20 || a.At(2) != 30 {
		t.Errorf("At: wrong elements")
	}
}

func TestGet_NegativeIndex(t *testing.T) {
	a := From(1, 2, 3)
	if a.Get(-1) != 3 {
		t.Errorf("Get(-1): expected 3, got %d", a.Get(-1))
	}
	if a.Get(-2) != 2 {
		t.Errorf("Get(-2): expected 2, got %d", a.Get(-2))
	}
}

func TestGet_PositiveWrapping(t *testing.T) {
	a := From(1, 2, 3)
	if a.Get(3) != 1 { // wraps to index 0
		t.Errorf("Get(3): expected 1 (wrap), got %d", a.Get(3))
	}
	if a.Get(4) != 2 { // wraps to index 1
		t.Errorf("Get(4): expected 2 (wrap), got %d", a.Get(4))
	}
}

// --- Push / PushAll ---

func TestPush_AppendsToEnd(t *testing.T) {
	a := New[int]()
	a.Push(1)
	a.Push(2)
	if a.Size() != 2 || a.At(0) != 1 || a.At(1) != 2 {
		t.Errorf("Push: wrong state")
	}
}

func TestPushAll_AppendsAll(t *testing.T) {
	a := New[int]()
	a.PushAll(1, 2, 3)
	if a.Size() != 3 || a.At(2) != 3 {
		t.Errorf("PushAll: wrong state")
	}
}

// --- First / Last ---

func TestFirst_ReturnsHead(t *testing.T) {
	a := From(10, 20, 30)
	if a.First() != 10 {
		t.Errorf("First: expected 10, got %d", a.First())
	}
}

func TestLast_ReturnsTail(t *testing.T) {
	a := From(10, 20, 30)
	if a.Last() != 30 {
		t.Errorf("Last: expected 30, got %d", a.Last())
	}
}

// --- Swap ---

func TestSwap_ExchangesElements(t *testing.T) {
	a := From(1, 2, 3)
	a.Swap(0, 2)
	if a.At(0) != 3 || a.At(2) != 1 {
		t.Errorf("Swap: expected [3,2,1], got [%d,%d,%d]", a.At(0), a.At(1), a.At(2))
	}
}

// --- Slice ---

func TestSlice_ReturnsSubArray(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	s := a.Slice(1, 4)
	if s.Size() != 3 || s.At(0) != 2 || s.At(2) != 4 {
		t.Errorf("Slice: unexpected result")
	}
}

// --- Delete ---

func TestDelete_UnorderedSwapsWithLast(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	a.Delete(1) // removes index 1, swaps 2 with 5
	if a.Size() != 4 {
		t.Errorf("Delete: expected size 4, got %d", a.Size())
	}
	// element 2 must be gone
	for _, v := range collect(a) {
		if v == 2 {
			t.Errorf("Delete: value 2 still present")
		}
	}
}

func TestDeleteKeepOrdering_PreservesOrder(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	a.DeleteKeepOrdering(1, true) // removes index 1
	got := collect(a)
	want := []int{1, 3, 4, 5}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("DeleteKeepOrdering: at [%d] expected %d, got %d", i, v, got[i])
		}
	}
}

func TestDeleteAll_ClearsArray(t *testing.T) {
	a := From(1, 2, 3)
	a.DeleteAll()
	if !a.IsEmpty() || a.Size() != 0 {
		t.Errorf("DeleteAll: expected empty array")
	}
}

func TestDeleteBy_RemovesMatchingElements(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	a.DeleteBy(func(v int) bool { return v%2 == 0 })
	for _, v := range collect(a) {
		if v%2 == 0 {
			t.Errorf("DeleteBy: even value %d still present", v)
		}
	}
}

func TestDeleteByKeepOrdering_PreservesOrder(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	a.DeleteByKeepOrdering(func(v int) bool { return v%2 == 0 }, true)
	got := collect(a)
	want := []int{1, 3, 5}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("DeleteByKeepOrdering: at [%d] expected %d, got %d", i, v, got[i])
		}
	}
}

// --- Some / Find / Filter ---

func TestSome_ReturnsTrueWhenMatch(t *testing.T) {
	a := From(1, 2, 3)
	if !a.Some(func(v int) bool { return v == 2 }) {
		t.Errorf("Some: expected true")
	}
	if a.Some(func(v int) bool { return v == 99 }) {
		t.Errorf("Some: expected false for missing element")
	}
}

func TestSome_EmptyArrayReturnsFalse(t *testing.T) {
	if New[int]().Some(func(int) bool { return true }) {
		t.Errorf("Some on empty: expected false")
	}
}

func TestFind_ReturnsFirstMatch(t *testing.T) {
	a := From(1, 2, 3, 2)
	v, ok := a.Find(func(v int) bool { return v == 2 })
	if !ok || v != 2 {
		t.Errorf("Find: expected (2, true), got (%d, %v)", v, ok)
	}
}

func TestFind_ReturnsFalseWhenNotFound(t *testing.T) {
	a := From(1, 2, 3)
	_, ok := a.Find(func(v int) bool { return v == 99 })
	if ok {
		t.Errorf("Find: expected false for missing element")
	}
}

func TestFilter_ReturnsMatchingElements(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	f := a.Filter(func(v int) bool { return v%2 == 0 })
	if f.Size() != 2 {
		t.Errorf("Filter: expected 2 elements, got %d", f.Size())
	}
}

func TestFilter_EmptyOnNoMatch(t *testing.T) {
	a := From(1, 3, 5)
	f := a.Filter(func(v int) bool { return v%2 == 0 })
	if !f.IsEmpty() {
		t.Errorf("Filter: expected empty result")
	}
}

// --- ForEach ---

func TestForEach_IteratesAllElements(t *testing.T) {
	a := From(1, 2, 3)
	sum := 0
	a.ForEach(func(_ int, v int) bool { sum += v; return true })
	if sum != 6 {
		t.Errorf("ForEach: expected sum 6, got %d", sum)
	}
}

func TestForEach_EarlyStop(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	count := 0
	a.ForEach(func(_ int, _ int) bool { count++; return count < 3 })
	if count != 3 {
		t.Errorf("ForEach: expected 3 iterations, got %d", count)
	}
}

// --- IndexOf ---

func TestIndexOf_ReturnsFirstMatch(t *testing.T) {
	a := From(10, 20, 30)
	idx := a.IndexOf(func(v int) bool { return v == 20 })
	if idx != 1 {
		t.Errorf("IndexOf: expected 1, got %d", idx)
	}
}

func TestIndexOf_ReturnsMinusOneWhenNotFound(t *testing.T) {
	a := From(10, 20, 30)
	idx := a.IndexOf(func(v int) bool { return v == 99 })
	if idx != -1 {
		t.Errorf("IndexOf: expected -1, got %d", idx)
	}
}

// --- Join / Merge ---

func TestJoin_AddsFromOtherCollection(t *testing.T) {
	a := From(1, 2, 3)
	b := From(4, 5)
	a.Join(b)
	if a.Size() != 5 || a.Last() != 5 {
		t.Errorf("Join: expected size 5 and last=5")
	}
}

func TestMerge_ReturnsNewCollectionLeavingOriginalUnchanged(t *testing.T) {
	a := From(1, 2)
	b := From(3, 4)
	m := a.Merge(b)
	if a.Size() != 2 {
		t.Errorf("Merge: original was modified")
	}
	if m.Size() != 4 {
		t.Errorf("Merge: expected size 4, got %d", m.Size())
	}
}
