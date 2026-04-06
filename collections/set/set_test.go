package set

import (
	"testing"
)

// --- helpers ---

// intItem is a test type implementing Keyable[int].
type intItem struct{ v int }

func (i *intItem) Key() int { return i.v }

func item(v int) *intItem { return &intItem{v} }

// collectValues gathers all values from a set via ForEach into a map for
// order-independent assertions (sets have no defined iteration order).
func collectValues[T comparable](s *Set[T, *KeyableWrapper[T]]) map[T]struct{} {
	out := make(map[T]struct{})
	s.ForEach(func(k T, v *KeyableWrapper[T]) bool {
		out[v.Wrapped] = struct{}{}
		return true
	})
	return out
}

// --- constructors ---

func TestNew_IsEmpty(t *testing.T) {
	s := New[int, *intItem]()
	if !s.IsEmpty() || s.Size() != 0 {
		t.Errorf("New: expected empty set")
	}
}

func TestNew_WithPreSize_IsEmpty(t *testing.T) {
	s := New[int, *intItem](10)
	if !s.IsEmpty() || s.Size() != 0 {
		t.Errorf("New(preSize): expected empty set, got size %d", s.Size())
	}
}

func TestNew_WithZeroPreSize_NoBehaviourChange(t *testing.T) {
	s := New[int, *intItem](0)
	if !s.IsEmpty() {
		t.Errorf("New(0): expected empty set")
	}
}

func TestNew_WithPreSize_FunctionalBeyondCapacity(t *testing.T) {
	s := New[int, *intItem](3)
	for i := range 6 { // insert twice the hinted capacity
		s.Push(item(i))
	}
	if s.Size() != 6 {
		t.Errorf("New(preSize): expected size 6, got %d", s.Size())
	}
	if !s.Has(item(5)) {
		t.Errorf("New(preSize): element 5 missing after overflow")
	}
}

func TestNewPrimitiveSet_IsEmpty(t *testing.T) {
	s := NewPrimitiveSet[string]()
	if !s.IsEmpty() {
		t.Errorf("NewPrimitiveSet: expected empty set")
	}
}

func TestWrap_ContainsItems(t *testing.T) {
	s := Wrap([]*intItem{item(1), item(2), item(3)})
	if s.Size() != 3 {
		t.Errorf("Wrap: expected size 3, got %d", s.Size())
	}
}

func TestWrapPrimitives_ContainsItems(t *testing.T) {
	s := WrapPrimitives([]int{1, 2, 3})
	if s.Size() != 3 {
		t.Errorf("WrapPrimitives: expected size 3, got %d", s.Size())
	}
}

func TestFrom_ContainsItems(t *testing.T) {
	s := From(item(10), item(20))
	if s.Size() != 2 {
		t.Errorf("From: expected size 2, got %d", s.Size())
	}
}

func TestFromPrimitives_ContainsItems(t *testing.T) {
	s := FromPrimitives("a", "b", "c")
	if s.Size() != 3 {
		t.Errorf("FromPrimitives: expected size 3, got %d", s.Size())
	}
}

// --- uniqueness ---

func TestPush_IgnoresDuplicates(t *testing.T) {
	s := New[int, *intItem]()
	s.Push(item(1))
	s.Push(item(1))
	s.Push(item(1))
	if s.Size() != 1 {
		t.Errorf("Push: expected size 1 after 3 duplicate pushes, got %d", s.Size())
	}
}

func TestWrapPrimitives_IgnoresDuplicates(t *testing.T) {
	s := WrapPrimitives([]int{1, 1, 2, 2, 3})
	if s.Size() != 3 {
		t.Errorf("WrapPrimitives: expected 3 unique elements, got %d", s.Size())
	}
}

func TestFromPrimitives_IgnoresDuplicates(t *testing.T) {
	s := FromPrimitives(5, 5, 5)
	if s.Size() != 1 {
		t.Errorf("FromPrimitives: expected 1 unique element, got %d", s.Size())
	}
}

// --- Has ---

func TestHas_PresentAndAbsent(t *testing.T) {
	s := New[int, *intItem]()
	s.Push(item(42))
	if !s.Has(item(42)) {
		t.Errorf("Has: expected true for present item")
	}
	if s.Has(item(99)) {
		t.Errorf("Has: expected false for absent item")
	}
}

func TestHas_EmptySet(t *testing.T) {
	s := New[int, *intItem]()
	if s.Has(item(1)) {
		t.Errorf("Has: expected false on empty set")
	}
}

// --- At / Get ---

func TestAt_ReturnsItemByKey(t *testing.T) {
	s := New[int, *intItem]()
	s.Push(item(7))
	got := s.At(7)
	if got == nil || got.v != 7 {
		t.Errorf("At: expected item with v=7")
	}
}

func TestAt_ReturnsZeroForMissingKey(t *testing.T) {
	s := New[int, *intItem]()
	if s.At(99) != nil {
		t.Errorf("At: expected nil for missing key")
	}
}

func TestGet_EquivalentToAt(t *testing.T) {
	s := New[int, *intItem]()
	s.Push(item(5))
	if s.Get(5) != s.At(5) {
		t.Errorf("Get: should return same value as At")
	}
}

// --- PushAll ---

func TestPushAll_AddsUniqueElements(t *testing.T) {
	s := New[int, *intItem]()
	s.PushAll(item(1), item(2), item(2), item(3))
	if s.Size() != 3 {
		t.Errorf("PushAll: expected 3 unique elements, got %d", s.Size())
	}
}

// --- Delete ---

func TestDelete_RemovesElement(t *testing.T) {
	s := New[int, *intItem]()
	s.Push(item(1))
	s.Push(item(2))
	s.Delete(1)
	if s.Has(item(1)) {
		t.Errorf("Delete: item 1 still present after deletion")
	}
	if s.Size() != 1 {
		t.Errorf("Delete: expected size 1, got %d", s.Size())
	}
}

func TestDelete_MissingKeyIsNoOp(t *testing.T) {
	s := New[int, *intItem]()
	s.Push(item(1))
	s.Delete(99) // non-existent key
	if s.Size() != 1 {
		t.Errorf("Delete: size changed on no-op deletion")
	}
}

func TestDeleteBy_RemovesMatchingElements(t *testing.T) {
	s := New[int, *intItem]()
	s.PushAll(item(1), item(2), item(3), item(4))
	s.DeleteBy(func(i *intItem) bool { return i.v%2 == 0 })
	if s.Has(item(2)) || s.Has(item(4)) {
		t.Errorf("DeleteBy: even items still present")
	}
	if s.Size() != 2 {
		t.Errorf("DeleteBy: expected size 2, got %d", s.Size())
	}
}

func TestDeleteAll_ClearsSet(t *testing.T) {
	s := New[int, *intItem]()
	s.PushAll(item(1), item(2), item(3))
	s.DeleteAll()
	if !s.IsEmpty() {
		t.Errorf("DeleteAll: set not empty after clear")
	}
}

// --- Some / Find / Filter ---

func TestSome_ReturnsTrueOnMatch(t *testing.T) {
	s := New[int, *intItem]()
	s.PushAll(item(1), item(2), item(3))
	if !s.Some(func(i *intItem) bool { return i.v == 2 }) {
		t.Errorf("Some: expected true")
	}
}

func TestSome_ReturnsFalseWhenNoMatch(t *testing.T) {
	s := New[int, *intItem]()
	s.PushAll(item(1), item(3))
	if s.Some(func(i *intItem) bool { return i.v == 2 }) {
		t.Errorf("Some: expected false")
	}
}

func TestSome_EmptySetReturnsFalse(t *testing.T) {
	s := New[int, *intItem]()
	if s.Some(func(*intItem) bool { return true }) {
		t.Errorf("Some: expected false on empty set")
	}
}

func TestFind_ReturnsMatchingItem(t *testing.T) {
	s := New[int, *intItem]()
	s.PushAll(item(10), item(20))
	v, ok := s.Find(func(i *intItem) bool { return i.v == 10 })
	if !ok || v.v != 10 {
		t.Errorf("Find: expected (item(10), true)")
	}
}

func TestFind_ReturnsFalseWhenNotFound(t *testing.T) {
	s := New[int, *intItem]()
	s.Push(item(1))
	_, ok := s.Find(func(i *intItem) bool { return i.v == 99 })
	if ok {
		t.Errorf("Find: expected false for missing element")
	}
}

func TestFilter_ReturnsMatchingSubset(t *testing.T) {
	s := New[int, *intItem]()
	s.PushAll(item(1), item(2), item(3), item(4))
	f := s.Filter(func(i *intItem) bool { return i.v%2 == 0 })
	if f.Size() != 2 {
		t.Errorf("Filter: expected 2 elements, got %d", f.Size())
	}
}

func TestFilter_EmptyWhenNoMatch(t *testing.T) {
	s := New[int, *intItem]()
	s.PushAll(item(1), item(3))
	f := s.Filter(func(i *intItem) bool { return i.v%2 == 0 })
	if !f.IsEmpty() {
		t.Errorf("Filter: expected empty result")
	}
}

// --- ForEach ---

func TestForEach_VisitsAllElements(t *testing.T) {
	s := New[int, *intItem]()
	s.PushAll(item(1), item(2), item(3))
	count := 0
	s.ForEach(func(_ int, _ *intItem) bool { count++; return true })
	if count != 3 {
		t.Errorf("ForEach: expected 3 visits, got %d", count)
	}
}

func TestForEach_EarlyStop(t *testing.T) {
	s := New[int, *intItem]()
	s.PushAll(item(1), item(2), item(3), item(4), item(5))
	count := 0
	s.ForEach(func(_ int, _ *intItem) bool { count++; return count < 2 })
	if count != 2 {
		t.Errorf("ForEach: expected early stop at 2, got %d", count)
	}
}

// --- Join / Merge ---

func TestJoin_AddsFromOtherSet(t *testing.T) {
	a := New[int, *intItem]()
	a.PushAll(item(1), item(2))
	b := New[int, *intItem]()
	b.PushAll(item(3), item(4))
	a.Join(b)
	if a.Size() != 4 {
		t.Errorf("Join: expected size 4, got %d", a.Size())
	}
}

func TestJoin_DoesNotAddDuplicates(t *testing.T) {
	a := New[int, *intItem]()
	a.PushAll(item(1), item(2))
	b := New[int, *intItem]()
	b.PushAll(item(2), item(3))
	a.Join(b)
	if a.Size() != 3 {
		t.Errorf("Join: expected 3 unique elements, got %d", a.Size())
	}
}

func TestMerge_ReturnsNewSetLeavingOriginalsUnchanged(t *testing.T) {
	a := New[int, *intItem]()
	a.PushAll(item(1), item(2))
	b := New[int, *intItem]()
	b.PushAll(item(3), item(4))
	m := a.Merge(b)
	if a.Size() != 2 || b.Size() != 2 {
		t.Errorf("Merge: originals were modified")
	}
	if m.Size() != 4 {
		t.Errorf("Merge: expected size 4, got %d", m.Size())
	}
}

func TestMerge_DeduplicatesAcrossSets(t *testing.T) {
	a := New[int, *intItem]()
	a.PushAll(item(1), item(2))
	b := New[int, *intItem]()
	b.PushAll(item(2), item(3))
	m := a.Merge(b)
	if m.Size() != 3 {
		t.Errorf("Merge: expected 3 unique elements, got %d", m.Size())
	}
}

// --- KeyableWrapper ---

func TestKeyableWrapper_KeyAndValue(t *testing.T) {
	w := WrapPrimitive(42)
	if w.Key() != 42 {
		t.Errorf("KeyableWrapper.Key: expected 42")
	}
	if w.Value() != 42 {
		t.Errorf("KeyableWrapper.Value: expected 42")
	}
}

// --- benchmarks ---

func BenchmarkSet_Has(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		s := NewPrimitiveSet[int]()
		for i := range n {
			s.Push(WrapPrimitive(i))
		}
		target := WrapPrimitive(n / 2)

		b.Run("Set_Has", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				s.Has(target)
			}
		})

		m := make(map[int]struct{}, n)
		for i := range n {
			m[i] = struct{}{}
		}
		b.Run("Map_Lookup", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, _ = m[n/2]
			}
		})
	}
}

func BenchmarkNew_PreSize(b *testing.B) {
	const n = 10_000
	b.Run("WithPreSize", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			s := New[int, *intItem](n)
			for i := range n {
				s.Push(item(i))
			}
		}
	})
	b.Run("WithoutPreSize", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			s := New[int, *intItem]()
			for i := range n {
				s.Push(item(i))
			}
		}
	})
}

func BenchmarkSet_Push(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := make([]*KeyableWrapper[int], n)
		for i := range n {
			items[i] = WrapPrimitive(i)
		}

		b.Run("Set_Push", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				s := NewPrimitiveSet[int]()
				for _, item := range items {
					s.Push(item)
				}
			}
		})
		b.Run("Map_Insert", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m := make(map[int]struct{}, n)
				for i := range n {
					m[i] = struct{}{}
				}
			}
		})
	}
}

func BenchmarkSet_Filter(b *testing.B) {
	n := 10_000
	s := NewPrimitiveSet[int]()
	m := make(map[int]struct{}, n)
	for i := range n {
		s.Push(WrapPrimitive(i))
		m[i] = struct{}{}
	}
	isEven := func(i *KeyableWrapper[int]) bool { return i.Wrapped%2 == 0 }

	b.Run("Set_Filter", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			s.Filter(isEven)
		}
	})
	b.Run("Map_Filter", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			out := make(map[int]struct{})
			for k := range m {
				if k%2 == 0 {
					out[k] = struct{}{}
				}
			}
		}
	})
}
