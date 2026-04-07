package collections

import "testing"

func TestIndexer_Next_StartsAtZero(t *testing.T) {
	idx := NewIndexer()
	if got := idx.Next(); got != 0 {
		t.Errorf("Next: expected 0, got %d", got)
	}
}

func TestIndexer_Next_Sequential(t *testing.T) {
	idx := NewIndexer()
	for i := range 5 {
		if got := idx.Next(); got != i {
			t.Errorf("Next: expected %d, got %d", i, got)
		}
	}
}

func TestIndexer_Release_IndexReused(t *testing.T) {
	idx := NewIndexer()
	idx.Next() // 0
	idx.Next() // 1
	idx.Release(0)
	if got := idx.Next(); got != 0 {
		t.Errorf("Next after Release(0): expected 0, got %d", got)
	}
}

func TestIndexer_Release_LIFO(t *testing.T) {
	idx := NewIndexer()
	idx.Next() // 0
	idx.Next() // 1
	idx.Next() // 2
	idx.Release(1)
	idx.Release(2)
	if got := idx.Next(); got != 2 {
		t.Errorf("Next after Release(1,2): expected 2 (LIFO), got %d", got)
	}
	if got := idx.Next(); got != 1 {
		t.Errorf("Next after Release(1,2) second call: expected 1 (LIFO), got %d", got)
	}
}

func TestIndexer_Next_AfterAllReleasesConsumed_ContinuesSequence(t *testing.T) {
	idx := NewIndexer()
	idx.Next() // 0
	idx.Next() // 1
	idx.Release(0)
	idx.Next() // reuses 0
	// staged is empty; next fresh index should be 2
	if got := idx.Next(); got != 2 {
		t.Errorf("Next after staged exhausted: expected 2, got %d", got)
	}
}

func TestIndexer_PreSize_DoesNotAlterBehavior(t *testing.T) {
	idx := NewIndexer(10)
	for i := range 5 {
		if got := idx.Next(); got != i {
			t.Errorf("PreSize Next: expected %d, got %d", i, got)
		}
	}
}

func TestIndexer_Release_BeyondPreSize_DoesNotPanic(t *testing.T) {
	idx := NewIndexer(2)
	for range 5 {
		idx.Next()
	}
	for i := range 5 {
		idx.Release(i)
	}
	for range 5 {
		idx.Next()
	}
}

func BenchmarkIndexer_Next(b *testing.B) {
	idx := NewIndexer()
	b.ReportAllocs()
	for b.Loop() {
		idx.Next()
	}
}

func BenchmarkIndexer_ReleaseAndNext(b *testing.B) {
	idx := NewIndexer()
	for i := range 1000 {
		idx.Next()
		idx.Release(i)
	}
	b.ReportAllocs()
	for b.Loop() {
		n := idx.Next()
		idx.Release(n)
	}
}
