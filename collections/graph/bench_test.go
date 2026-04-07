package graph

import (
	"fmt"
	"testing"
)

const N = 8

// graphMatrix is the common interface that parameterises every benchmark
// over all matrix implementations using concrete [int] types.
type graphMatrix interface {
	Add(int)
	Contains(int) bool
	Delete(int) bool
	Set(int, int) bool
	Has(...int) bool
	Remove(...int) bool
}

type implDef struct {
	name string
	new  func() graphMatrix
}

var impls = []implDef{
	{"AdjacencyMatrix", func() graphMatrix { return NewAdjacencyMatrix[int]() }},
	{"BitMatrix", func() graphMatrix { return NewBitMatrix[int]() }},
	{"CSR", func() graphMatrix { return NewCSR[int]() }},
}

var directedImpls = []implDef{
	{"AdjacencyMatrix", func() graphMatrix { return NewAdjacencyMatrix[int](TopologyParams{Features: []Feature{Directed}}) }},
	{"BitMatrix", func() graphMatrix { return NewBitMatrix[int](TopologyParams{Features: []Feature{Directed}}) }},
	{"CSR", func() graphMatrix { return NewCSR[int](TopologyParams{Features: []Feature{Directed}}) }},
}

// pathSizes drives the exponential sub-benchmarks for Has/Remove path.
var pathSizes = []int{2, 4, 8} //, 16, 32, 64}

// buildChain wires 0→1→2→…→(n-1) into m and returns the vertex path.
func buildChain(m graphMatrix, n int) []int {
	for i := range n - 1 {
		m.Set(i, i+1)
	}
	path := make([]int, n)
	for i := range n {
		path[i] = i
	}
	return path
}

func rebuildChain(m graphMatrix, path []int) {
	for i := range len(path) - 1 {
		m.Set(path[i], path[i+1])
	}
}

// ============================================================
// Benchmarks
// ============================================================

func BenchmarkAdd(b *testing.B) {
	for _, impl := range impls {
		m := impl.new()
		for v := range N {
			m.Add(v)
		}
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.Add(N)
				m.Delete(N)
			}
		})
	}
}

func BenchmarkAdd_Existing(b *testing.B) {
	for _, impl := range impls {
		m := impl.new()
		m.Add(0)
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.Add(0)
			}
		})
	}
}

func BenchmarkContains_Present(b *testing.B) {
	for _, impl := range impls {
		m := impl.new()
		m.Add(0)
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.Contains(0)
			}
		})
	}
}

func BenchmarkContains_Absent(b *testing.B) {
	for _, impl := range impls {
		m := impl.new()
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.Contains(0)
			}
		})
	}
}

func BenchmarkDelete(b *testing.B) {
	for _, impl := range impls {
		m := impl.new()
		for v := range N {
			m.Add(v)
		}
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.Delete(N - 1)
				m.Add(N - 1)
			}
		})
	}
}

func BenchmarkSet(b *testing.B) {
	// Pre-add a fixed vertex pool so Set only measures edge insertion,
	// not slice growth. Without this, BitMatrix bitRows grow as O(index),
	// making total memory O(N²) and triggering OOM on large benchmark runs.
	for _, impl := range impls {
		m := impl.new()
		for v := range N {
			m.Add(v)
		}
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				m.Set(i%N, (i+1)%N)
				i++
			}
		})
	}
}

func BenchmarkSet_Directed(b *testing.B) {
	for _, impl := range directedImpls {
		m := impl.new()
		for v := range N {
			m.Add(v)
		}
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				m.Set(i%N, (i+1)%N)
				i++
			}
		})
	}
}

func BenchmarkHas_Vertex(b *testing.B) {
	for _, impl := range impls {
		m := impl.new()
		buildChain(m, N)
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.Has(N / 2)
			}
		})
	}
}

func BenchmarkHas_Edge(b *testing.B) {
	for _, impl := range impls {
		m := impl.new()
		buildChain(m, N)
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.Has(N/2-1, N/2)
			}
		})
	}
}

func BenchmarkHas_Path(b *testing.B) {
	for _, n := range pathSizes {
		for _, impl := range impls {
			m := impl.new()
			path := buildChain(m, n)
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					m.Has(path...)
				}
			})
		}
	}
}

func BenchmarkRemove_Vertex(b *testing.B) {
	for _, impl := range impls {
		m := impl.new()
		for v := range N {
			m.Add(v)
		}
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.Remove(N - 1)
				m.Add(N - 1)
			}
		})
	}
}

func BenchmarkRemove_Edge(b *testing.B) {
	for _, impl := range impls {
		m := impl.new()
		buildChain(m, N)
		b.Run(impl.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.Remove(N/2+1, N/2)
				m.Set(N/2+1, N/2)
			}
		})
	}
}

func BenchmarkRemove_Path(b *testing.B) {
	for _, n := range pathSizes {
		for _, impl := range impls {
			m := impl.new()
			path := buildChain(m, n)
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					m.Remove(path...)
					rebuildChain(m, path)
				}
			})
		}
	}
}
