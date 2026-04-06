package array

import (
	"math/rand/v2"
	"sort"
	"testing"
)

// --- IsSorted ---

func TestIsSorted_EmptyAndSingleElement(t *testing.T) {
	if !New[int]().IsSorted(intCmp) {
		t.Errorf("IsSorted: empty array should be sorted")
	}
	if !From(42).IsSorted(intCmp) {
		t.Errorf("IsSorted: single-element array should be sorted")
	}
}

func TestIsSorted_Ascending(t *testing.T) {
	if !From(1, 2, 3, 4, 5).IsSorted(intCmp) {
		t.Errorf("IsSorted: ascending array not detected as sorted")
	}
}

func TestIsSorted_Descending(t *testing.T) {
	if !From(5, 4, 3, 2, 1).IsSorted(intCmp) {
		t.Errorf("IsSorted: descending array not detected as sorted")
	}
}

func TestIsSorted_Unsorted(t *testing.T) {
	if From(1, 3, 2, 4).IsSorted(intCmp) {
		t.Errorf("IsSorted: unsorted array incorrectly detected as sorted")
	}
}

func TestIsSorted_WithTrailingEqual(t *testing.T) {
	// [1, 2, 2] is ascending with a duplicate at the end — should be sorted
	if !From(1, 2, 2).IsSorted(intCmp) {
		t.Errorf("IsSorted: [1,2,2] should be considered sorted")
	}
}

// --- InsertionSort ---

func TestInsertionSort_RandomInput(t *testing.T) {
	a := From(5, 3, 1, 4, 2)
	a.InsertionSort(intCmp)
	for i := 1; i < a.Size(); i++ {
		if a.At(i) < a.At(i-1) {
			t.Errorf("InsertionSort: not sorted at index %d", i)
		}
	}
}

func TestInsertionSort_AlreadySorted(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	a.InsertionSort(intCmp)
	want := []int{1, 2, 3, 4, 5}
	for i, v := range want {
		if a.At(i) != v {
			t.Errorf("InsertionSort: already-sorted input corrupted at [%d]", i)
		}
	}
}

func TestInsertionSort_ReverseSorted(t *testing.T) {
	a := From(5, 4, 3, 2, 1)
	a.InsertionSort(intCmp)
	for i := 1; i < a.Size(); i++ {
		if a.At(i) < a.At(i-1) {
			t.Errorf("InsertionSort: reverse input not sorted at index %d", i)
		}
	}
}

func TestInsertionSort_Duplicates(t *testing.T) {
	a := From(3, 1, 2, 1, 3)
	a.InsertionSort(intCmp)
	for i := 1; i < a.Size(); i++ {
		if a.At(i) < a.At(i-1) {
			t.Errorf("InsertionSort: duplicates not handled at index %d", i)
		}
	}
}

func TestInsertionSort_SingleElement(t *testing.T) {
	a := From(42)
	a.InsertionSort(intCmp)
	if a.At(0) != 42 {
		t.Errorf("InsertionSort: single element corrupted")
	}
}

// --- HeapSort ---

func TestHeapSort_RandomInput(t *testing.T) {
	a := From(5, 3, 1, 4, 2)
	a.HeapSort(intCmp)
	for i := 1; i < a.Size(); i++ {
		if a.At(i) < a.At(i-1) {
			t.Errorf("HeapSort: not sorted at index %d", i)
		}
	}
}

func TestHeapSort_AlreadySorted(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	a.HeapSort(intCmp)
	for i := 1; i < a.Size(); i++ {
		if a.At(i) < a.At(i-1) {
			t.Errorf("HeapSort: already-sorted input corrupted at [%d]", i)
		}
	}
}

func TestHeapSort_ReverseSorted(t *testing.T) {
	a := From(5, 4, 3, 2, 1)
	a.HeapSort(intCmp)
	for i := 1; i < a.Size(); i++ {
		if a.At(i) < a.At(i-1) {
			t.Errorf("HeapSort: reverse input not sorted at index %d", i)
		}
	}
}

func TestHeapSort_Duplicates(t *testing.T) {
	a := From(3, 1, 2, 1, 3)
	a.HeapSort(intCmp)
	for i := 1; i < a.Size(); i++ {
		if a.At(i) < a.At(i-1) {
			t.Errorf("HeapSort: duplicates not handled at index %d", i)
		}
	}
}

func TestHeapSort_SingleElement(t *testing.T) {
	a := From(42)
	a.HeapSort(intCmp)
	if a.At(0) != 42 {
		t.Errorf("HeapSort: single element corrupted")
	}
}

func TestHeapSort_PreservesAllElements(t *testing.T) {
	original := []int{5, 3, 1, 4, 2, 3}
	a := Wrap(append([]int{}, original...))
	a.HeapSort(intCmp)
	sum := 0
	for _, v := range original {
		sum += v
	}
	sortedSum := 0
	a.ForEach(func(_ int, v int) bool { sortedSum += v; return true })
	if sum != sortedSum {
		t.Errorf("HeapSort: elements lost during sort (sum %d → %d)", sum, sortedSum)
	}
}

// --- BinarySearch ---

func TestBinarySearch_FindsExistingElement(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	v, ok := a.BinarySearch(3, intCmp)
	if !ok || v != 3 {
		t.Errorf("BinarySearch: expected (3, true), got (%d, %v)", v, ok)
	}
}

func TestBinarySearch_FindsFirstElement(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	v, ok := a.BinarySearch(1, intCmp)
	if !ok || v != 1 {
		t.Errorf("BinarySearch: expected (1, true), got (%d, %v)", v, ok)
	}
}

func TestBinarySearch_FindsLastElement(t *testing.T) {
	a := From(1, 2, 3, 4, 5)
	v, ok := a.BinarySearch(5, intCmp)
	if !ok || v != 5 {
		t.Errorf("BinarySearch: expected (5, true), got (%d, %v)", v, ok)
	}
}

func TestBinarySearch_ReturnsFalseWhenNotFound(t *testing.T) {
	a := From(1, 2, 4, 5)
	_, ok := a.BinarySearch(3, intCmp)
	if ok {
		t.Errorf("BinarySearch: expected false for missing element")
	}
}

func TestBinarySearch_EmptyArray(t *testing.T) {
	_, ok := New[int]().BinarySearch(1, intCmp)
	if ok {
		t.Errorf("BinarySearch: expected false on empty array")
	}
}

func TestBinarySearch_SingleElement_Found(t *testing.T) {
	a := From(42)
	v, ok := a.BinarySearch(42, intCmp)
	if !ok || v != 42 {
		t.Errorf("BinarySearch: single element not found")
	}
}

func TestBinarySearch_SingleElement_NotFound(t *testing.T) {
	a := From(42)
	_, ok := a.BinarySearch(1, intCmp)
	if ok {
		t.Errorf("BinarySearch: found non-existent element in single-element array")
	}
}

// --- Min / Max ---

func TestMin_ReturnsSmallest(t *testing.T) {
	a := From(3, 1, 4, 1, 5, 9)
	v, ok := a.Min(intCmp)
	if !ok || v != 1 {
		t.Errorf("Min: expected (1, true), got (%d, %v)", v, ok)
	}
}

func TestMin_SingleElement(t *testing.T) {
	v, ok := From(7).Min(intCmp)
	if !ok || v != 7 {
		t.Errorf("Min: single element case failed")
	}
}

func TestMin_EmptyReturnsFalse(t *testing.T) {
	_, ok := New[int]().Min(intCmp)
	if ok {
		t.Errorf("Min: expected false on empty array")
	}
}

func TestMax_ReturnsLargest(t *testing.T) {
	a := From(3, 1, 4, 1, 5, 9)
	v, ok := a.Max(intCmp)
	if !ok || v != 9 {
		t.Errorf("Max: expected (9, true), got (%d, %v)", v, ok)
	}
}

func TestMax_SingleElement(t *testing.T) {
	v, ok := From(7).Max(intCmp)
	if !ok || v != 7 {
		t.Errorf("Max: single element case failed")
	}
}

func TestMax_EmptyReturnsFalse(t *testing.T) {
	_, ok := New[int]().Max(intCmp)
	if ok {
		t.Errorf("Max: expected false on empty array")
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

func nearlySortedInts(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	// swap 5% of elements
	swaps := n / 20
	for range swaps {
		i, j := rand.IntN(n), rand.IntN(n)
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func BenchmarkInsertionSort(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		data := randomInts(n)
		b.Run("random", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				a := Wrap(append([]int{}, data...))
				a.InsertionSort(intCmp)
			}
		})
		nearly := nearlySortedInts(n)
		b.Run("nearly_sorted", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				a := Wrap(append([]int{}, nearly...))
				a.InsertionSort(intCmp)
			}
		})
	}
}

func BenchmarkHeapSort(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		data := randomInts(n)
		b.Run("random", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				a := Wrap(append([]int{}, data...))
				a.HeapSort(intCmp)
			}
		})
	}
}

func BenchmarkSortSlice(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		data := randomInts(n)
		b.Run("random", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				s := append([]int{}, data...)
				sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
			}
		})
	}
}

func BenchmarkBinarySearch(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000, 100_000} {
		data := make([]int, n)
		for i := range data {
			data[i] = i * 2 // even numbers 0..2n
		}
		a := Wrap(data)
		target := data[n/2]

		b.Run("BinarySearch", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				a.BinarySearch(target, intCmp)
			}
		})
		b.Run("LinearSearch", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				a.Find(func(v int) bool { return v == target })
			}
		})
	}
}

func BenchmarkNew_PreSize(b *testing.B) {
	const n = 10_000
	b.Run("WithPreSize", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			a := New[int](n)
			for i := range n {
				a.Push(i)
			}
		}
	})
	b.Run("WithoutPreSize", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			a := New[int]()
			for i := range n {
				a.Push(i)
			}
		}
	})
}

func BenchmarkMinMax(b *testing.B) {
	data := randomInts(10_000)
	a := Wrap(data)

	b.Run("Min", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			a.Min(intCmp)
		}
	})
	b.Run("Max", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			a.Max(intCmp)
		}
	})
	b.Run("SortSliceThenFirst", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			s := append([]int{}, data...)
			sort.Ints(s)
			_ = s[0]
		}
	})
}
