package linkedlist

import (
	"math/rand/v2"
	"sort"
	"testing"

	"github.com/0x626f/gota/collections"
)

// --- helpers ---

var intCmp collections.Comparator[int] = func(a, b int) int {
	if a < b {
		return collections.LOWER
	}
	if a > b {
		return collections.GREATER
	}
	return collections.EQUAL
}

// toSlice collects all list elements via forward traversal.
func toSlice[D any](list *LinkedList[D]) []D {
	out := make([]D, 0, list.Size())
	list.ForEach(func(_ int, v D) bool { out = append(out, v); return true })
	return out
}

// assertForwardBackwardConsistency verifies that the doubly-linked structure is
// intact: forward traversal and backward traversal from tail yield the same
// elements in opposite order.
func assertStructure(t *testing.T, list *LinkedList[int]) {
	t.Helper()
	forward := toSlice(list)
	if len(forward) != list.Size() {
		t.Errorf("structure: ForEach count %d ≠ Size() %d", len(forward), list.Size())
	}

	// backward traversal
	var backward []int
	node := list.tail
	for node != nil {
		backward = append(backward, node.Data)
		node = node.left
	}
	// single-element list has no tail (tail==nil), handled separately
	if list.Size() > 1 && len(backward) != list.Size() {
		t.Errorf("structure: backward count %d ≠ Size() %d", len(backward), list.Size())
	}
	for i, v := range forward {
		rev := len(forward) - 1 - i
		if list.Size() > 1 && backward[rev] != v {
			t.Errorf("structure: forward[%d]=%d ≠ backward[%d]=%d", i, v, rev, backward[rev])
		}
	}
}

// --- constructors ---

func TestNewLinkedList_IsEmpty(t *testing.T) {
	list := NewLinkedList[int]()
	if !list.IsEmpty() || list.Size() != 0 {
		t.Errorf("NewLinkedList: expected empty list")
	}
}

// --- Push / PushFront / PushAll ---

func TestPush_AppendsToTail(t *testing.T) {
	list := NewLinkedList[int]()
	list.Push(1)
	list.Push(2)
	list.Push(3)
	got := toSlice(list)
	want := []int{1, 2, 3}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("Push: at [%d] expected %d, got %d", i, v, got[i])
		}
	}
}

func TestPushFront_InsertsAtHead(t *testing.T) {
	list := NewLinkedList[int]()
	list.Push(3)
	list.PushFront(2)
	list.PushFront(1)
	got := toSlice(list)
	want := []int{1, 2, 3}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("PushFront: at [%d] expected %d, got %d", i, v, got[i])
		}
	}
}

func TestPushAll_AppendsAllInOrder(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3, 4)
	if list.Size() != 4 || list.First() != 1 || list.Last() != 4 {
		t.Errorf("PushAll: unexpected state")
	}
}

// --- Insert / InsertFront ---

func TestInsert_ReturnsNode(t *testing.T) {
	list := NewLinkedList[int]()
	node := list.Insert(99)
	if node == nil || node.Data != 99 {
		t.Errorf("Insert: returned nil or wrong node")
	}
	if list.Size() != 1 {
		t.Errorf("Insert: expected size 1")
	}
}

func TestInsertFront_ReturnsNodeAtHead(t *testing.T) {
	list := NewLinkedList[int]()
	list.Push(10)
	node := list.InsertFront(5)
	if node == nil || node.Data != 5 {
		t.Errorf("InsertFront: returned nil or wrong node")
	}
	if list.First() != 5 {
		t.Errorf("InsertFront: expected head=5, got %d", list.First())
	}
}

// --- At / Get ---

func TestAt_PositiveIndex(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(10, 20, 30)
	if list.At(0) != 10 || list.At(1) != 20 || list.At(2) != 30 {
		t.Errorf("At: wrong elements at positive indices")
	}
}

func TestAt_NegativeIndex(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(10, 20, 30)
	if list.At(-1) != 30 {
		t.Errorf("At(-1): expected 30, got %d", list.At(-1))
	}
	if list.At(-2) != 20 {
		t.Errorf("At(-2): expected 20, got %d", list.At(-2))
	}
	if list.At(-3) != 10 {
		t.Errorf("At(-3): expected 10, got %d", list.At(-3))
	}
}

func TestAt_OutOfBoundsReturnsZero(t *testing.T) {
	list := NewLinkedList[int]()
	list.Push(1)
	if list.At(99) != 0 {
		t.Errorf("At(out-of-bounds): expected 0")
	}
	if list.At(-99) != 0 {
		t.Errorf("At(negative out-of-bounds): expected 0")
	}
}

func TestGet_AliasesAt(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	if list.Get(1) != list.At(1) {
		t.Errorf("Get: should equal At")
	}
}

// --- First / Last ---

func TestFirst_ReturnsHead(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(5, 10, 15)
	if list.First() != 5 {
		t.Errorf("First: expected 5, got %d", list.First())
	}
}

func TestLast_ReturnsTail(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(5, 10, 15)
	if list.Last() != 15 {
		t.Errorf("Last: expected 15, got %d", list.Last())
	}
}

func TestFirstLast_SingleElement(t *testing.T) {
	list := NewLinkedList[int]()
	list.Push(42)
	if list.First() != 42 || list.Last() != 42 {
		t.Errorf("First/Last: single element case failed")
	}
}

func TestFirst_EmptyReturnsZero(t *testing.T) {
	if NewLinkedList[int]().First() != 0 {
		t.Errorf("First on empty: expected 0")
	}
}

func TestLast_EmptyReturnsZero(t *testing.T) {
	if NewLinkedList[int]().Last() != 0 {
		t.Errorf("Last on empty: expected 0")
	}
}

// --- Pop ---

func TestPop_ByPositiveIndex(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3, 4)
	v := list.Pop(1)
	if v != 2 {
		t.Errorf("Pop(1): expected 2, got %d", v)
	}
	if list.Size() != 3 {
		t.Errorf("Pop: expected size 3, got %d", list.Size())
	}
	assertStructure(t, list)
}

func TestPop_ByNegativeIndex(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	v := list.Pop(-1) // pops last
	if v != 3 {
		t.Errorf("Pop(-1): expected 3, got %d", v)
	}
	if list.Size() != 2 {
		t.Errorf("Pop: expected size 2, got %d", list.Size())
	}
}

func TestPop_OutOfBoundsReturnsZero(t *testing.T) {
	list := NewLinkedList[int]()
	list.Push(1)
	if list.Pop(99) != 0 {
		t.Errorf("Pop(out-of-bounds): expected 0")
	}
	if list.Size() != 1 {
		t.Errorf("Pop(out-of-bounds): size should not change")
	}
}

// --- PopLeft / PopRight ---

func TestPopLeft_RemovesHead(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	v := list.PopLeft()
	if v != 1 {
		t.Errorf("PopLeft: expected 1, got %d", v)
	}
	if list.First() != 2 {
		t.Errorf("PopLeft: new head should be 2, got %d", list.First())
	}
	assertStructure(t, list)
}

func TestPopRight_RemovesTail(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	v := list.PopRight()
	if v != 3 {
		t.Errorf("PopRight: expected 3, got %d", v)
	}
	if list.Last() != 2 {
		t.Errorf("PopRight: new tail should be 2, got %d", list.Last())
	}
	assertStructure(t, list)
}

func TestPopLeft_SingleElement(t *testing.T) {
	list := NewLinkedList[int]()
	list.Push(99)
	v := list.PopLeft()
	if v != 99 || !list.IsEmpty() {
		t.Errorf("PopLeft: single element case failed")
	}
}

func TestPopRight_SingleElement(t *testing.T) {
	list := NewLinkedList[int]()
	list.Push(99)
	v := list.PopRight()
	if v != 99 || !list.IsEmpty() {
		t.Errorf("PopRight: single element case failed")
	}
}

func TestPopLeft_EmptyReturnsZero(t *testing.T) {
	if NewLinkedList[int]().PopLeft() != 0 {
		t.Errorf("PopLeft on empty: expected 0")
	}
}

func TestPopRight_EmptyReturnsZero(t *testing.T) {
	if NewLinkedList[int]().PopRight() != 0 {
		t.Errorf("PopRight on empty: expected 0")
	}
}

// --- Delete ---

func TestDelete_ByIndex(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3, 4)
	list.Delete(2)
	if list.Size() != 3 {
		t.Errorf("Delete: expected size 3, got %d", list.Size())
	}
	if list.Some(func(v int) bool { return v == 3 }) {
		t.Errorf("Delete: value 3 still present")
	}
	assertStructure(t, list)
}

func TestDelete_Head(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	list.Delete(0)
	if list.First() != 2 || list.Size() != 2 {
		t.Errorf("Delete head: expected first=2, size=2")
	}
	assertStructure(t, list)
}

func TestDelete_Tail(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	list.Delete(-1)
	if list.Last() != 2 || list.Size() != 2 {
		t.Errorf("Delete tail: expected last=2, size=2")
	}
	assertStructure(t, list)
}

// --- DeleteBy ---

func TestDeleteBy_RemovesMatchingElements(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3, 4, 5)
	list.DeleteBy(func(v int) bool { return v%2 == 0 })
	if list.Size() != 3 {
		t.Errorf("DeleteBy: expected size 3, got %d", list.Size())
	}
	if list.Some(func(v int) bool { return v%2 == 0 }) {
		t.Errorf("DeleteBy: even numbers still present")
	}
}

func TestDeleteBy_RemovesAll(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(2, 4, 6)
	list.DeleteBy(func(v int) bool { return v%2 == 0 })
	if !list.IsEmpty() {
		t.Errorf("DeleteBy: expected empty list after removing all")
	}
}

// --- DeleteAll ---

func TestDeleteAll_ClearsList(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	list.DeleteAll()
	if !list.IsEmpty() || list.Size() != 0 {
		t.Errorf("DeleteAll: expected empty list")
	}
}

// --- Some / Find ---

func TestSome_ReturnsTrueOnMatch(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	if !list.Some(func(v int) bool { return v == 2 }) {
		t.Errorf("Some: expected true")
	}
}

func TestSome_ReturnsFalseWhenNoMatch(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 3, 5)
	if list.Some(func(v int) bool { return v == 2 }) {
		t.Errorf("Some: expected false")
	}
}

func TestFind_ReturnsFirstMatch(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	v, ok := list.Find(func(v int) bool { return v > 1 })
	if !ok || v != 2 {
		t.Errorf("Find: expected (2, true), got (%d, %v)", v, ok)
	}
}

func TestFind_ReturnsFalseWhenNotFound(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	_, ok := list.Find(func(v int) bool { return v > 99 })
	if ok {
		t.Errorf("Find: expected false for no match")
	}
}

// --- Filter ---

func TestFilter_ReturnsNewList(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3, 4, 5)
	f := list.Filter(func(v int) bool { return v%2 == 0 })
	if f.Size() != 2 {
		t.Errorf("Filter: expected 2 elements, got %d", f.Size())
	}
	if list.Size() != 5 {
		t.Errorf("Filter: original list was modified")
	}
}

func TestFilter_EmptyOnNoMatch(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 3, 5)
	f := list.Filter(func(v int) bool { return v%2 == 0 })
	if !f.IsEmpty() {
		t.Errorf("Filter: expected empty result")
	}
}

// --- ForEach ---

func TestForEach_IteratesAll(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	sum := 0
	list.ForEach(func(_ int, v int) bool { sum += v; return true })
	if sum != 6 {
		t.Errorf("ForEach: expected sum 6, got %d", sum)
	}
}

func TestForEach_EarlyStop(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3, 4, 5)
	count := 0
	list.ForEach(func(_ int, _ int) bool { count++; return count < 3 })
	if count != 3 {
		t.Errorf("ForEach: expected 3 iterations, got %d", count)
	}
}

// --- IndexOf ---

func TestIndexOf_ReturnsIndex(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(10, 20, 30)
	idx, ok := list.IndexOf(func(v int) bool { return v == 20 })
	if !ok || idx != 1 {
		t.Errorf("IndexOf: expected (1, true), got (%d, %v)", idx, ok)
	}
}

func TestIndexOf_ReturnsFalseWhenNotFound(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(10, 20, 30)
	_, ok := list.IndexOf(func(v int) bool { return v == 99 })
	if ok {
		t.Errorf("IndexOf: expected false for missing element")
	}
}

// --- Join / Merge ---

func TestJoin_AddsFromOtherCollection(t *testing.T) {
	a := NewLinkedList[int]()
	a.PushAll(1, 2)
	b := NewLinkedList[int]()
	b.PushAll(3, 4)
	a.Join(b)
	if a.Size() != 4 || a.Last() != 4 {
		t.Errorf("Join: expected size 4 and last=4")
	}
	assertStructure(t, a)
}

func TestMerge_ReturnsNewList(t *testing.T) {
	a := NewLinkedList[int]()
	a.PushAll(1, 2)
	b := NewLinkedList[int]()
	b.PushAll(3, 4)
	m := a.Merge(b)
	if a.Size() != 2 {
		t.Errorf("Merge: original modified")
	}
	if m.Size() != 4 {
		t.Errorf("Merge: expected size 4, got %d", m.Size())
	}
}

// --- Swap ---

func TestSwap_NonAdjacentNodes(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3, 4, 5)
	list.Swap(0, 4)
	if list.First() != 5 || list.Last() != 1 {
		t.Errorf("Swap(0,4): expected first=5 last=1, got first=%d last=%d", list.First(), list.Last())
	}
	assertStructure(t, list)
}

func TestSwap_AdjacentNodes(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	list.Swap(0, 1)
	got := toSlice(list)
	if got[0] != 2 || got[1] != 1 || got[2] != 3 {
		t.Errorf("Swap(0,1): expected [2,1,3], got %v", got)
	}
	assertStructure(t, list)
}

func TestSwap_SameIndexIsNoOp(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	before := toSlice(list)
	list.Swap(1, 1)
	after := toSlice(list)
	for i, v := range before {
		if after[i] != v {
			t.Errorf("Swap(i,i): list changed at [%d]", i)
		}
	}
}

func TestSwap_HeadAndTail(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(10, 20, 30)
	list.Swap(0, 2)
	if list.First() != 30 || list.Last() != 10 {
		t.Errorf("Swap head/tail: expected first=30 last=10")
	}
	assertStructure(t, list)
}

// --- Move ---

func TestMove_ForwardShift(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3, 4, 5)
	list.Move(0, 2) // move first element to position 2
	got := toSlice(list)
	// element 1 should no longer be at position 0
	if got[0] == 1 {
		t.Errorf("Move(0,2): element 1 still at head")
	}
	assertStructure(t, list)
}

// --- MoveToFront ---

func TestMoveToFront_MovesNodeToHead(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	node := list.Insert(99)
	list.MoveToFront(node)
	if list.First() != 99 {
		t.Errorf("MoveToFront: expected 99 at head, got %d", list.First())
	}
	assertStructure(t, list)
}

// --- Remove ---

func TestRemove_DirectNodeRemoval(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	node := list.Insert(99)
	list.Remove(node)
	if list.Some(func(v int) bool { return v == 99 }) {
		t.Errorf("Remove: node still in list")
	}
	if list.Size() != 3 {
		t.Errorf("Remove: expected size 3, got %d", list.Size())
	}
	assertStructure(t, list)
}

func TestRemove_NilIsNoOp(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	list.Remove(nil)
	if list.Size() != 3 {
		t.Errorf("Remove(nil): size changed unexpectedly")
	}
}

// --- Shrink ---

func TestShrink_ReducesToCapacity(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3, 4, 5)
	list.Shrink(3)
	if list.Size() != 3 {
		t.Errorf("Shrink(3): expected size 3, got %d", list.Size())
	}
	if list.First() != 1 {
		t.Errorf("Shrink: expected head=1, got %d", list.First())
	}
	assertStructure(t, list)
}

func TestShrink_ZeroCapacityClearsAll(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	list.Shrink(0)
	if !list.IsEmpty() {
		t.Errorf("Shrink(0): expected empty list")
	}
}

func TestShrink_CapacityGeqSizeIsNoOp(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3)
	list.Shrink(5)
	if list.Size() != 3 {
		t.Errorf("Shrink(>=size): size changed unexpectedly")
	}
}

// --- Sort ---

func TestSort_RandomInput(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(5, 3, 1, 4, 2)
	list.Sort(intCmp)
	got := toSlice(list)
	for i := 1; i < len(got); i++ {
		if got[i] < got[i-1] {
			t.Errorf("Sort: not sorted at index %d (%d < %d)", i, got[i], got[i-1])
		}
	}
	assertStructure(t, list)
}

func TestSort_AlreadySorted(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(1, 2, 3, 4, 5)
	list.Sort(intCmp)
	got := toSlice(list)
	for i := 1; i < len(got); i++ {
		if got[i] < got[i-1] {
			t.Errorf("Sort: already-sorted input corrupted at [%d]", i)
		}
	}
	assertStructure(t, list)
}

func TestSort_ReverseSorted(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(5, 4, 3, 2, 1)
	list.Sort(intCmp)
	got := toSlice(list)
	for i := 1; i < len(got); i++ {
		if got[i] < got[i-1] {
			t.Errorf("Sort: reverse input not sorted at [%d]", i)
		}
	}
	assertStructure(t, list)
}

func TestSort_Duplicates(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(3, 1, 2, 1, 3)
	list.Sort(intCmp)
	got := toSlice(list)
	for i := 1; i < len(got); i++ {
		if got[i] < got[i-1] {
			t.Errorf("Sort: duplicates not handled at [%d]", i)
		}
	}
}

func TestSort_SingleElement(t *testing.T) {
	list := NewLinkedList[int]()
	list.Push(42)
	list.Sort(intCmp)
	if list.First() != 42 || list.Size() != 1 {
		t.Errorf("Sort: single-element case corrupted")
	}
}

func TestSort_PreservesAllElements(t *testing.T) {
	input := []int{5, 3, 1, 4, 2, 3, 5}
	list := NewLinkedList[int]()
	list.PushAll(input...)

	sum := 0
	for _, v := range input {
		sum += v
	}

	list.Sort(intCmp)

	sortedSum := 0
	list.ForEach(func(_ int, v int) bool { sortedSum += v; return true })
	if sum != sortedSum {
		t.Errorf("Sort: elements lost during sort (sum %d → %d)", sum, sortedSum)
	}
}

func TestSort_SizeUnchangedAfterSort(t *testing.T) {
	list := NewLinkedList[int]()
	list.PushAll(9, 1, 5, 3, 7)
	list.Sort(intCmp)
	if list.Size() != 5 {
		t.Errorf("Sort: size changed from 5 to %d", list.Size())
	}
}

// --- benchmarks ---

func randomInts(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = rand.Int()
	}
	return s
}

func BenchmarkLinkedList_Sort(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		data := randomInts(n)

		b.Run("LinkedList_Sort", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				list := NewLinkedList[int]()
				list.PushAll(data...)
				list.Sort(intCmp)
			}
		})
		b.Run("sort_Slice", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				s := append([]int{}, data...)
				sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
			}
		})
	}
}

func BenchmarkLinkedList_Push(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		b.Run("LinkedList_Push", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				list := NewLinkedList[int]()
				for i := range n {
					list.Push(i)
				}
			}
		})
		b.Run("Slice_Append", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				s := make([]int, 0, n)
				for i := range n {
					s = append(s, i)
				}
			}
		})
	}
}

func BenchmarkLinkedList_PopLeft(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		b.Run("LinkedList_PopLeft", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				list := NewLinkedList[int]()
				for i := range n {
					list.Push(i)
				}
				for range n {
					list.PopLeft()
				}
			}
		})
		b.Run("Slice_ShiftLeft", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				s := make([]int, n)
				for i := range n {
					s[i] = i
				}
				for len(s) > 0 {
					s = s[1:]
				}
			}
		})
	}
}
