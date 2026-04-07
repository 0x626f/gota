package graph

import (
	"fmt"
	"testing"
)

// ─── bench vertex ─────────────────────────────────────────────────────────────

type intVertex int

func (v intVertex) Key() int { return int(v) }

// ─── factories ────────────────────────────────────────────────────────────────

type graphBenchFactory struct {
	name string
	new  func(n int) IGraph[intVertex, float64, int]
}

var graphBenchFactories = []graphBenchFactory{
	{"AdjacencyMatrix", func(n int) IGraph[intVertex, float64, int] {
		return NewGraph[intVertex, float64](TopologyParams{
			Key:      AdjacencyMatrixTopology,
			Features: Features(Directed),
			Scale:    n,
		})
	}},
	{"BitMatrix", func(n int) IGraph[intVertex, float64, int] {
		return NewGraph[intVertex, float64](TopologyParams{
			Key:      BitMatrixTopology,
			Features: Features(Directed),
			Scale:    n,
		})
	}},
	{"CSR", func(n int) IGraph[intVertex, float64, int] {
		return NewGraph[intVertex, float64](TopologyParams{
			Key:      CRSTopology,
			Features: Features(Directed),
			Scale:    n,
		})
	}},
}

const graphBenchN = 8

// buildGraphChain wires 0→1→…→(n-1) into g and returns vertex slice.
func buildGraphChain(g IGraph[intVertex, float64, int], n int) []intVertex {
	for i := range n - 1 {
		g.Set(intVertex(i), intVertex(i+1), 1.0)
	}
	vs := make([]intVertex, n)
	for i := range n {
		vs[i] = intVertex(i)
	}
	return vs
}

// ─── Add ──────────────────────────────────────────────────────────────────────

func BenchmarkGraph_Add(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		for i := range graphBenchN {
			g.Add(intVertex(i))
		}
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.Add(intVertex(graphBenchN))
				g.Delete(intVertex(graphBenchN))
			}
		})
	}
}

func BenchmarkGraph_Add_Existing(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		g.Add(intVertex(0))
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.Add(intVertex(0))
			}
		})
	}
}

// ─── Contains ─────────────────────────────────────────────────────────────────

func BenchmarkGraph_Contains_Present(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		g.Add(intVertex(0))
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.Contains(intVertex(0))
			}
		})
	}
}

func BenchmarkGraph_Contains_Absent(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.Contains(intVertex(0))
			}
		})
	}
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func BenchmarkGraph_Delete(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		for i := range graphBenchN {
			g.Add(intVertex(i))
		}
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.Delete(intVertex(graphBenchN - 1))
				g.Add(intVertex(graphBenchN - 1))
			}
		})
	}
}

// ─── Set ──────────────────────────────────────────────────────────────────────

func BenchmarkGraph_Set(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		for i := range graphBenchN {
			g.Add(intVertex(i))
		}
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				g.Set(intVertex(i%graphBenchN), intVertex((i+1)%graphBenchN), 1.0)
				i++
			}
		})
	}
}

// ─── Has ──────────────────────────────────────────────────────────────────────

func BenchmarkGraph_Has_Vertex(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		buildGraphChain(g, graphBenchN)
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.Has(intVertex(graphBenchN / 2))
			}
		})
	}
}

func BenchmarkGraph_Has_Edge(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		buildGraphChain(g, graphBenchN)
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.Has(intVertex(graphBenchN/2-1), intVertex(graphBenchN/2))
			}
		})
	}
}

// ─── GetVertex / GetEdge ──────────────────────────────────────────────────────

func BenchmarkGraph_GetVertex(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		for i := range graphBenchN {
			g.Add(intVertex(i))
		}
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.GetVertex(graphBenchN / 2)
			}
		})
	}
}

func BenchmarkGraph_GetEdge(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		buildGraphChain(g, graphBenchN)
		a, b2 := intVertex(graphBenchN/2-1), intVertex(graphBenchN/2)
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.GetEdge(a, b2)
			}
		})
	}
}

// ─── Neighbors ────────────────────────────────────────────────────────────────

func BenchmarkGraph_Neighbors(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		buildGraphChain(g, graphBenchN)
		v := intVertex(0)
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.Neighbors(v)
			}
		})
	}
}

// ─── Paths ────────────────────────────────────────────────────────────────────

var graphPathSizes = []int{8, 32, 64}

func BenchmarkGraph_Paths_DFS_LinearChain(b *testing.B) {
	for _, n := range graphPathSizes {
		for _, f := range graphBenchFactories {
			g := f.new(n)
			for i := range n - 1 {
				g.Set(intVertex(i), intVertex(i+1), 1.0)
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					g.Paths(0, false, DFSSearch)
				}
			})
		}
	}
}

func BenchmarkGraph_Paths_BFS_LinearChain(b *testing.B) {
	for _, n := range graphPathSizes {
		for _, f := range graphBenchFactories {
			g := f.new(n)
			for i := range n - 1 {
				g.Set(intVertex(i), intVertex(i+1), 1.0)
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					g.Paths(0, false, BFSSearch)
				}
			})
		}
	}
}

// ─── DFS / BFS traversal ──────────────────────────────────────────────────────

func BenchmarkGraph_DFS(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		buildGraphChain(g, graphBenchN)
		noop := noopTraversal[int]{}
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.DFS(0, noop)
			}
		})
	}
}

func BenchmarkGraph_BFS(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		buildGraphChain(g, graphBenchN)
		noop := noopTraversal[int]{}
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.BFS(0, noop)
			}
		})
	}
}

// ─── Dijkstra / AStar ─────────────────────────────────────────────────────────

func BenchmarkGraph_Dijkstra(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		buildGraphChain(g, graphBenchN)
		noop := noopTraversal[int]{}
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.Dijkstra(0, noop)
			}
		})
	}
}

func BenchmarkGraph_AStar(b *testing.B) {
	for _, f := range graphBenchFactories {
		g := f.new(graphBenchN)
		buildGraphChain(g, graphBenchN)
		noop := noopTraversal[int]{}
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g.AStar(0, graphBenchN-1, noop)
			}
		})
	}
}
