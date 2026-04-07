package graph

import (
	"fmt"
	"testing"
)

// ─── Paths benchmarks ─────────────────────────────────────────────────────────

var pathsBenchSizes = []int{8, 32, 128}

func BenchmarkPaths_DFS_LinearChain(b *testing.B) {
	for _, n := range pathsBenchSizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n - 1 {
				g.Set(i, i+1)
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Paths(g, 0, false, DFSSearch)
				}
			})
		}
	}
}

func BenchmarkPaths_BFS_LinearChain(b *testing.B) {
	for _, n := range pathsBenchSizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n - 1 {
				g.Set(i, i+1)
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Paths(g, 0, false, BFSSearch)
				}
			})
		}
	}
}

func BenchmarkPaths_DFS_BinaryTree(b *testing.B) {
	for _, n := range pathsBenchSizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n {
				if l := 2*i + 1; l < n {
					g.Set(i, l)
				}
				if r := 2*i + 2; r < n {
					g.Set(i, r)
				}
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Paths(g, 0, false, DFSSearch)
				}
			})
		}
	}
}

func BenchmarkPaths_BFS_BinaryTree(b *testing.B) {
	for _, n := range pathsBenchSizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n {
				if l := 2*i + 1; l < n {
					g.Set(i, l)
				}
				if r := 2*i + 2; r < n {
					g.Set(i, r)
				}
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Paths(g, 0, false, BFSSearch)
				}
			})
		}
	}
}

func BenchmarkPaths_DFS_Dense(b *testing.B) {
	// Dense graphs produce exponentially many paths — keep sizes small.
	sizes := []int{5, 7, 9}
	for _, n := range sizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n {
				for j := range n {
					if i != j {
						g.Set(i, j)
					}
				}
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Paths(g, 0, false, DFSSearch)
				}
			})
		}
	}
}

func BenchmarkPaths_BFS_Dense(b *testing.B) {
	sizes := []int{5, 7, 9}
	for _, n := range sizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n {
				for j := range n {
					if i != j {
						g.Set(i, j)
					}
				}
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Paths(g, 0, false, BFSSearch)
				}
			})
		}
	}
}

func BenchmarkPaths_DFS_Cycled(b *testing.B) {
	// Directed cycle 0→1→…→n-1→0: cycled=true adds cycle-closure paths.
	for _, n := range pathsBenchSizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n - 1 {
				g.Set(i, i+1)
			}
			g.Set(n-1, 0)
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Paths(g, 0, true, DFSSearch)
				}
			})
		}
	}
}

func BenchmarkPaths_BFS_Cycled(b *testing.B) {
	for _, n := range pathsBenchSizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n - 1 {
				g.Set(i, i+1)
			}
			g.Set(n-1, 0)
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Paths(g, 0, true, BFSSearch)
				}
			})
		}
	}
}
