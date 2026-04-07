package graph

import (
	"fmt"
	"testing"
)

// cycleBenchSizes drives sub-benchmarks by vertex count.
var cycleBenchSizes = []int{8, 64, 512}

// allCycleImpls returns undirected cycleImpls plus presized variants for n.
func allCycleImpls(n int) []struct {
	name string
	new  func() cycleMatrix
} {
	return append(cycleImpls, []struct {
		name string
		new  func() cycleMatrix
	}{
		{"AdjacencyMatrix/presized", func() cycleMatrix {
			return NewAdjacencyMatrix[int](TopologyParams{Scale: n})
		}},
		{"BitMatrix/presized", func() cycleMatrix {
			return NewBitMatrix[int](TopologyParams{Scale: n})
		}},
		{"CSR/presized", func() cycleMatrix {
			return NewCSR[int](TopologyParams{Scale: n})
		}},
	}...)
}

// allDirectedCycleImpls returns directed cycleImpls plus presized variants for n.
func allDirectedCycleImpls(n int) []struct {
	name string
	new  func() cycleMatrix
} {
	return append(directedCycleImpls, []struct {
		name string
		new  func() cycleMatrix
	}{
		{"AdjacencyMatrix/presized", func() cycleMatrix {
			return NewAdjacencyMatrix[int](TopologyParams{Features: []Feature{Directed}, Scale: n})
		}},
		{"BitMatrix/presized", func() cycleMatrix {
			return NewBitMatrix[int](TopologyParams{Features: []Feature{Directed}, Scale: n})
		}},
		{"CSR/presized", func() cycleMatrix {
			return NewCSR[int](TopologyParams{Features: []Feature{Directed}, Scale: n})
		}},
	}...)
}

// allAcyclicCycleImpls returns acyclic undirected impls plus presized variants for n.
func allAcyclicCycleImpls(n int) []struct {
	name string
	new  func() cycleMatrix
} {
	return append(acyclicCycleImpls, []struct {
		name string
		new  func() cycleMatrix
	}{
		{"AdjacencyMatrix/presized", func() cycleMatrix {
			return NewAdjacencyMatrix[int](TopologyParams{Features: []Feature{Acyclic}, Scale: n})
		}},
		{"BitMatrix/presized", func() cycleMatrix {
			return NewBitMatrix[int](TopologyParams{Features: []Feature{Acyclic}, Scale: n})
		}},
		{"CSR/presized", func() cycleMatrix {
			return NewCSR[int](TopologyParams{Features: []Feature{Acyclic}, Scale: n})
		}},
	}...)
}

// allAcyclicDirectedCycleImpls returns acyclic directed impls plus presized variants for n.
func allAcyclicDirectedCycleImpls(n int) []struct {
	name string
	new  func() cycleMatrix
} {
	return append(acyclicDirectedCycleImpls, []struct {
		name string
		new  func() cycleMatrix
	}{
		{"AdjacencyMatrix/presized", func() cycleMatrix {
			return NewAdjacencyMatrix[int](TopologyParams{Features: []Feature{Directed, Acyclic}, Scale: n})
		}},
		{"BitMatrix/presized", func() cycleMatrix {
			return NewBitMatrix[int](TopologyParams{Features: []Feature{Directed, Acyclic}, Scale: n})
		}},
		{"CSR/presized", func() cycleMatrix {
			return NewCSR[int](TopologyParams{Features: []Feature{Directed, Acyclic}, Scale: n})
		}},
	}...)
}

// buildTree wires 0-1-2-…-(n-1) as an undirected path (no cycle).
func buildTree(g cycleMatrix, n int) {
	for i := range n - 1 {
		g.Set(i, i+1)
	}
}

// buildUndirectedCycle wires 0-1-2-…-(n-1)-0.
func buildUndirectedCycle(g cycleMatrix, n int) {
	for i := range n - 1 {
		g.Set(i, i+1)
	}
	g.Set(n-1, 0) // closes the cycle
}

// buildDAG wires 0→1→2→…→(n-1) as a directed path (no cycle).
// Assumes g was created with the Directed feature.
func buildDAG(g cycleMatrix, n int) {
	for i := range n - 1 {
		g.Set(i, i+1)
	}
}

// buildDirectedCycle wires 0→1→…→(n-1)→0.
// Uses Add to pre-register all vertices so Set(n-1, 0) is not blocked
// by the anti-parallel guard (Has(0, n-1) is false).
// Assumes g was created with the Directed feature.
func buildDirectedCycle(g cycleMatrix, n int) {
	for i := range n {
		g.Add(i)
	}
	for i := range n - 1 {
		g.Set(i, i+1)
	}
	g.Set(n-1, 0)
}

// ── IsCycled ─────────────────────────────────────────────────────────────────

func BenchmarkCycle_IsCycled_Undirected_Acyclic(b *testing.B) {
	for _, n := range cycleBenchSizes {
		for _, impl := range allCycleImpls(n) {
			g := impl.new()
			buildTree(g, n)
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					g.IsCycled()
				}
			})
		}
	}
}

func BenchmarkCycle_IsCycled_Undirected_Cyclic(b *testing.B) {
	for _, n := range cycleBenchSizes {
		for _, impl := range allCycleImpls(n) {
			g := impl.new()
			buildUndirectedCycle(g, n)
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					g.IsCycled()
				}
			})
		}
	}
}

func BenchmarkCycle_IsCycled_Directed_Acyclic(b *testing.B) {
	for _, n := range cycleBenchSizes {
		for _, impl := range allDirectedCycleImpls(n) {
			g := impl.new()
			buildDAG(g, n)
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					g.IsCycled()
				}
			})
		}
	}
}

func BenchmarkCycle_IsCycled_Directed_Cyclic(b *testing.B) {
	for _, n := range cycleBenchSizes {
		for _, impl := range allDirectedCycleImpls(n) {
			g := impl.new()
			buildDirectedCycle(g, n)
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					g.IsCycled()
				}
			})
		}
	}
}

// ── Set with Acyclic feature ──────────────────────────────────────────────────

func BenchmarkCycle_Set_Acyclic_Undirected(b *testing.B) {
	for _, n := range cycleBenchSizes {
		for _, impl := range allAcyclicCycleImpls(n) {
			g := impl.new()
			// build a path 0-1-…-(n-2); leaf (n-1) is held in reserve
			for i := range n - 2 {
				g.Set(i, i+1)
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					// attach leaf, then detach so the graph stays a tree
					g.Set(n-2, n-1)
					g.Remove(n-2, n-1)
				}
			})
		}
	}
}

func BenchmarkCycle_Set_Acyclic_Directed(b *testing.B) {
	for _, n := range cycleBenchSizes {
		for _, impl := range allAcyclicDirectedCycleImpls(n) {
			g := impl.new()
			// build a directed path 0→1→…→(n-2)
			for i := range n - 2 {
				g.Set(i, i+1)
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					g.Set(n-2, n-1)
					g.Remove(n-2, n-1)
				}
			})
		}
	}
}

func BenchmarkCycle_Set_Acyclic_Reject_Undirected(b *testing.B) {
	for _, n := range cycleBenchSizes {
		for _, impl := range allAcyclicCycleImpls(n) {
			g := impl.new()
			buildTree(g, n) // 0-1-…-(n-1) already connected
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					g.Set(0, n-1) // would close cycle — always rejected
				}
			})
		}
	}
}

func BenchmarkCycle_Set_Acyclic_Reject_Directed(b *testing.B) {
	for _, n := range cycleBenchSizes {
		for _, impl := range allAcyclicDirectedCycleImpls(n) {
			g := impl.new()
			buildDAG(g, n) // 0→1→…→(n-1) already connected
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					g.Set(n-1, 0) // would close cycle — always rejected
				}
			})
		}
	}
}
