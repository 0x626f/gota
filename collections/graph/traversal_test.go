package graph

import (
	"fmt"
	"slices"
	"testing"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

type visitEdge[Key comparable] struct{ from, to Key }

type relaxEvent[Key comparable] struct {
	from, to Key
	cost     float64
}

// recorder captures every callback so tests can inspect results.
type recorder[Key comparable] struct {
	visits        []visitEdge[Key]
	cycles        [][]Key
	relaxes       []relaxEvent[Key]
	shouldVisitFn func(Key, Key) bool
	onVisitFn     func(Key, Key) bool
	onCycleFn     func([]Key) bool
	edgeWeightFn  func(Key, Key) float64
	heuristicFn   func(Key, Key) float64
}

func (r *recorder[Key]) ShouldVisit(from, to Key) bool {
	if r.shouldVisitFn != nil {
		return r.shouldVisitFn(from, to)
	}
	return true
}
func (r *recorder[Key]) OnVisit(from, to Key) bool {
	r.visits = append(r.visits, visitEdge[Key]{from, to})
	if r.onVisitFn != nil {
		return r.onVisitFn(from, to)
	}
	return true
}
func (r *recorder[Key]) OnCycle(cycle []Key) bool {
	cp := make([]Key, len(cycle))
	copy(cp, cycle)
	r.cycles = append(r.cycles, cp)
	if r.onCycleFn != nil {
		return r.onCycleFn(cycle)
	}
	return true
}
func (r *recorder[Key]) EdgeWeight(from, to Key) float64 {
	if r.edgeWeightFn != nil {
		return r.edgeWeightFn(from, to)
	}
	return 1
}
func (r *recorder[Key]) Heuristic(current, goal Key) float64 {
	if r.heuristicFn != nil {
		return r.heuristicFn(current, goal)
	}
	return 0
}
func (r *recorder[Key]) OnRelax(from, to Key, cost float64) {
	r.relaxes = append(r.relaxes, relaxEvent[Key]{from, to, cost})
}
func (r *recorder[Key]) hasVisit(from, to Key) bool {
	for _, v := range r.visits {
		if v.from == from && v.to == to {
			return true
		}
	}
	return false
}
func (r *recorder[Key]) visitedDestinations() []Key {
	out := make([]Key, len(r.visits))
	for i, v := range r.visits {
		out[i] = v.to
	}
	return out
}
func (r *recorder[Key]) lastCostTo(to Key) (float64, bool) {
	for i := len(r.relaxes) - 1; i >= 0; i-- {
		if r.relaxes[i].to == to {
			return r.relaxes[i].cost, true
		}
	}
	return 0, false
}

// noopTraversal always continues — used in benchmarks to eliminate callback overhead.
type noopTraversal[Key comparable] struct{}

func (noopTraversal[Key]) ShouldVisit(Key, Key) bool   { return true }
func (noopTraversal[Key]) OnVisit(Key, Key) bool       { return true }
func (noopTraversal[Key]) OnCycle([]Key) bool          { return true }
func (noopTraversal[Key]) EdgeWeight(Key, Key) float64 { return 1 }
func (noopTraversal[Key]) Heuristic(Key, Key) float64  { return 0 }
func (noopTraversal[Key]) OnRelax(Key, Key, float64)   {}

// impls and directedImpls (already defined in bench_test.go) provide the three
// topology constructors. We declare local typed aliases to work with ITopology.

type topoFactory struct {
	name string
	new  func() ITopology[int]
}

var undirectedFactories = []topoFactory{
	{"AdjacencyMatrix", func() ITopology[int] { return NewAdjacencyMatrix[int]() }},
	{"BitMatrix", func() ITopology[int] { return NewBitMatrix[int]() }},
	{"CSR", func() ITopology[int] { return NewCSR[int]() }},
}

var directedFactories = []topoFactory{
	{"AdjacencyMatrix", func() ITopology[int] {
		return NewAdjacencyMatrix[int](TopologyParams{Features: []Feature{Directed}})
	}},
	{"BitMatrix", func() ITopology[int] {
		return NewBitMatrix[int](TopologyParams{Features: []Feature{Directed}})
	}},
	{"CSR", func() ITopology[int] {
		return NewCSR[int](TopologyParams{Features: []Feature{Directed}})
	}},
}

// ─── DFS tests ────────────────────────────────────────────────────────────────

func TestDFS_AbsentStart_NoCallbacks(t *testing.T) {
	for _, f := range undirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			r := &recorder[int]{}
			DFS(g, 99, r)
			if len(r.visits) != 0 || len(r.cycles) != 0 {
				t.Error("expected no callbacks for absent start vertex")
			}
		})
	}
}

func TestDFS_IsolatedVertex_NoVisits(t *testing.T) {
	for _, f := range undirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Add(0)
			r := &recorder[int]{}
			DFS(g, 0, r)
			if len(r.visits) != 0 {
				t.Errorf("expected 0 visits, got %d", len(r.visits))
			}
		})
	}
}

func TestDFS_LinearChain_VisitOrder(t *testing.T) {
	// directed chain 0→1→2→3: each node has exactly one outgoing edge,
	// so DFS order is fully deterministic across all implementations.
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			for i := range 3 {
				g.Set(i, i+1)
			}
			r := &recorder[int]{}
			DFS(g, 0, r)
			want := []visitEdge[int]{{0, 1}, {1, 2}, {2, 3}}
			if len(r.visits) != len(want) {
				t.Fatalf("want %d visits, got %d", len(want), len(r.visits))
			}
			for i, w := range want {
				if r.visits[i] != w {
					t.Errorf("visit[%d]: want %v, got %v", i, w, r.visits[i])
				}
			}
			if len(r.cycles) != 0 {
				t.Errorf("expected no cycles, got %d", len(r.cycles))
			}
		})
	}
}

func TestDFS_UndirectedEdge_ParentNotReportedAsCycle(t *testing.T) {
	// A-B undirected: DFS from A should visit (A,B) but the reverse edge
	// B→A is the parent edge and must NOT trigger OnCycle.
	for _, f := range undirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			r := &recorder[int]{}
			DFS(g, 0, r)
			if len(r.visits) != 1 || !r.hasVisit(0, 1) {
				t.Errorf("expected exactly visit (0,1), got %v", r.visits)
			}
			if len(r.cycles) != 0 {
				t.Errorf("parent edge must not be reported as cycle, got %v", r.cycles)
			}
		})
	}
}

func TestDFS_UndirectedTriangle_CycleReported(t *testing.T) {
	for _, f := range undirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0)
			r := &recorder[int]{}
			DFS(g, 0, r)
			if len(r.cycles) != 1 {
				t.Fatalf("expected 1 cycle, got %d: %v", len(r.cycles), r.cycles)
			}
			cycle := r.cycles[0]
			for _, v := range []int{0, 1, 2} {
				if !slices.Contains(cycle, v) {
					t.Errorf("cycle missing vertex %d: %v", v, cycle)
				}
			}
		})
	}
}

func TestDFS_DirectedCycle_CycleReported(t *testing.T) {
	// 0→1→2→0: a directed 3-cycle (anti-parallel guard only blocks direct
	// reverse edges, not longer cycles).
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0)
			r := &recorder[int]{}
			DFS(g, 0, r)
			if len(r.cycles) == 0 {
				t.Fatal("expected a cycle to be reported")
			}
			cycle := r.cycles[0]
			if cycle[0] != cycle[len(cycle)-1] {
				t.Errorf("cycle path must start and end with the same vertex: %v", cycle)
			}
		})
	}
}

func TestDFS_DAG_NoCycleReported(t *testing.T) {
	// Diamond DAG: 0→1, 0→2, 1→3, 2→3.
	// Vertex 3 is reachable via two paths but that is not a cycle.
	// DFS visits 3 only once (as a tree edge); the second path to 3 is a
	// cross edge that is skipped because 3 is already fully explored.
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(0, 2)
			g.Set(1, 3)
			g.Set(2, 3)
			r := &recorder[int]{}
			DFS(g, 0, r)
			if len(r.cycles) != 0 {
				t.Errorf("DAG must not produce cycles, got %v", r.cycles)
			}
			// All four vertices must be reachable (appear as destinations).
			for _, v := range []int{1, 2, 3} {
				if !slices.Contains(r.visitedDestinations(), v) {
					t.Errorf("vertex %d unreachable from 0", v)
				}
			}
		})
	}
}

func TestDFS_SelfLoop_CycleReported(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 0)
			r := &recorder[int]{}
			DFS(g, 0, r)
			if len(r.cycles) != 1 {
				t.Fatalf("expected 1 cycle for self-loop, got %d", len(r.cycles))
			}
			if r.cycles[0][0] != 0 || r.cycles[0][len(r.cycles[0])-1] != 0 {
				t.Errorf("self-loop cycle should be [0 0], got %v", r.cycles[0])
			}
		})
	}
}

func TestDFS_Disconnected_OnlyReachableVerticesVisited(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2) // component A: 0→1→2
			g.Set(3, 4) // component B: 3→4 (unreachable from 0)
			r := &recorder[int]{}
			DFS(g, 0, r)
			destinations := r.visitedDestinations()
			if slices.Contains(destinations, 3) || slices.Contains(destinations, 4) {
				t.Error("DFS from 0 must not visit disconnected component 3→4")
			}
			if !r.hasVisit(0, 1) || !r.hasVisit(1, 2) {
				t.Error("DFS must visit the reachable component 0→1→2")
			}
		})
	}
}

func TestDFS_ShouldVisit_FiltersEdge(t *testing.T) {
	// 0→1→2 and 0→3: filter out the edge (0,3).
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(0, 3)
			r := &recorder[int]{
				shouldVisitFn: func(from, to int) bool { return !(from == 0 && to == 3) },
			}
			DFS(g, 0, r)
			if r.hasVisit(0, 3) {
				t.Error("filtered edge (0,3) must not appear in visits")
			}
			if !r.hasVisit(0, 1) || !r.hasVisit(1, 2) {
				t.Error("non-filtered edges must be visited")
			}
		})
	}
}

func TestDFS_OnVisit_EarlyStop(t *testing.T) {
	// OnVisit returns false after the first edge → traversal must stop immediately.
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			for i := range 4 {
				g.Set(i, i+1)
			}
			r := &recorder[int]{
				onVisitFn: func(int, int) bool { return false },
			}
			DFS(g, 0, r)
			if len(r.visits) != 1 {
				t.Errorf("expected stop after first visit, got %d visits", len(r.visits))
			}
		})
	}
}

func TestDFS_OnCycle_EarlyStop(t *testing.T) {
	// OnCycle returns false → traversal must stop when the first cycle is found.
	for _, f := range undirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0) // closes the triangle
			stopped := false
			r := &recorder[int]{
				onCycleFn: func([]int) bool {
					stopped = true
					return false
				},
			}
			DFS(g, 0, r)
			if !stopped {
				t.Error("OnCycle must have been called")
			}
			if len(r.cycles) != 1 {
				t.Errorf("expected exactly 1 cycle before stop, got %d", len(r.cycles))
			}
		})
	}
}

// ─── BFS tests ────────────────────────────────────────────────────────────────

func TestBFS_AbsentStart_NoCallbacks(t *testing.T) {
	for _, f := range undirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			r := &recorder[int]{}
			BFS(g, 99, r)
			if len(r.visits) != 0 || len(r.cycles) != 0 {
				t.Error("expected no callbacks for absent start vertex")
			}
		})
	}
}

func TestBFS_IsolatedVertex_NoVisits(t *testing.T) {
	for _, f := range undirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Add(0)
			r := &recorder[int]{}
			BFS(g, 0, r)
			if len(r.visits) != 0 {
				t.Errorf("expected 0 visits, got %d", len(r.visits))
			}
		})
	}
}

func TestBFS_LinearChain_LevelOrder(t *testing.T) {
	// directed chain 0→1→2→3: BFS order is fully deterministic
	// (each level has exactly one node).
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			for i := range 3 {
				g.Set(i, i+1)
			}
			r := &recorder[int]{}
			BFS(g, 0, r)
			want := []visitEdge[int]{{0, 1}, {1, 2}, {2, 3}}
			if len(r.visits) != len(want) {
				t.Fatalf("want %d visits, got %d", len(want), len(r.visits))
			}
			for i, w := range want {
				if r.visits[i] != w {
					t.Errorf("visit[%d]: want %v, got %v", i, w, r.visits[i])
				}
			}
			if len(r.cycles) != 0 {
				t.Errorf("expected no cycles, got %d", len(r.cycles))
			}
		})
	}
}

func TestBFS_UndirectedEdge_ParentNotReportedAsCycle(t *testing.T) {
	for _, f := range undirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			r := &recorder[int]{}
			BFS(g, 0, r)
			if len(r.visits) != 1 || !r.hasVisit(0, 1) {
				t.Errorf("expected exactly visit (0,1), got %v", r.visits)
			}
			if len(r.cycles) != 0 {
				t.Errorf("parent edge must not be reported as cycle, got %v", r.cycles)
			}
		})
	}
}

func TestBFS_UndirectedTriangle_CycleReported(t *testing.T) {
	for _, f := range undirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0)
			r := &recorder[int]{}
			BFS(g, 0, r)
			if len(r.cycles) == 0 {
				t.Fatal("expected at least one cycle to be reported")
			}
			// each reported cycle entry is [from, to] where to is already visited
			for _, c := range r.cycles {
				if len(c) != 2 {
					t.Errorf("BFS cycle report must be [from, to], got %v", c)
				}
			}
		})
	}
}

func TestBFS_DirectedCycle_CycleReported(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0)
			r := &recorder[int]{}
			BFS(g, 0, r)
			if len(r.cycles) == 0 {
				t.Fatal("expected a cycle to be reported")
			}
		})
	}
}

func TestBFS_DirectedTree_NoCycleReported(t *testing.T) {
	// A directed tree has no cross edges, so BFS reports no cycles.
	// (BFS reports OnCycle for any already-visited destination, which
	// includes cross edges in directed DAGs — use a tree to avoid that.)
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			// tree: 0→1, 0→2, 1→3, 1→4
			g.Set(0, 1)
			g.Set(0, 2)
			g.Set(1, 3)
			g.Set(1, 4)
			r := &recorder[int]{}
			BFS(g, 0, r)
			if len(r.cycles) != 0 {
				t.Errorf("directed tree must not produce cycle reports, got %v", r.cycles)
			}
			for _, v := range []int{1, 2, 3, 4} {
				if !slices.Contains(r.visitedDestinations(), v) {
					t.Errorf("vertex %d unreachable from 0", v)
				}
			}
		})
	}
}

func TestBFS_SelfLoop_CycleReported(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 0)
			r := &recorder[int]{}
			BFS(g, 0, r)
			if len(r.cycles) != 1 {
				t.Fatalf("expected 1 cycle for self-loop, got %d", len(r.cycles))
			}
			c := r.cycles[0]
			if c[0] != 0 || c[1] != 0 {
				t.Errorf("self-loop cycle should be [0 0], got %v", c)
			}
		})
	}
}

func TestBFS_Disconnected_OnlyReachableVerticesVisited(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(3, 4)
			r := &recorder[int]{}
			BFS(g, 0, r)
			destinations := r.visitedDestinations()
			if slices.Contains(destinations, 3) || slices.Contains(destinations, 4) {
				t.Error("BFS from 0 must not visit disconnected component 3→4")
			}
			if !r.hasVisit(0, 1) || !r.hasVisit(1, 2) {
				t.Error("BFS must visit the reachable component 0→1→2")
			}
		})
	}
}

func TestBFS_ShouldVisit_FiltersEdge(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(0, 3)
			r := &recorder[int]{
				shouldVisitFn: func(from, to int) bool { return !(from == 0 && to == 3) },
			}
			BFS(g, 0, r)
			if r.hasVisit(0, 3) {
				t.Error("filtered edge (0,3) must not appear in visits")
			}
			if !r.hasVisit(0, 1) || !r.hasVisit(1, 2) {
				t.Error("non-filtered edges must be visited")
			}
		})
	}
}

func TestBFS_OnVisit_EarlyStop(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			for i := range 4 {
				g.Set(i, i+1)
			}
			r := &recorder[int]{
				onVisitFn: func(int, int) bool { return false },
			}
			BFS(g, 0, r)
			if len(r.visits) != 1 {
				t.Errorf("expected stop after first visit, got %d visits", len(r.visits))
			}
		})
	}
}

func TestBFS_OnCycle_EarlyStop(t *testing.T) {
	for _, f := range undirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0)
			stopped := false
			r := &recorder[int]{
				onCycleFn: func([]int) bool {
					stopped = true
					return false
				},
			}
			BFS(g, 0, r)
			if !stopped {
				t.Error("OnCycle must have been called")
			}
			if len(r.cycles) != 1 {
				t.Errorf("expected exactly 1 cycle before stop, got %d", len(r.cycles))
			}
		})
	}
}

// ─── benchmarks ───────────────────────────────────────────────────────────────

var traversalBenchSizes = []int{64, 512, 4096}

type benchFactory struct {
	name string
	new  func(...TopologyParams) ITopology[int]
}

var benchFactories = []benchFactory{
	{"AdjacencyMatrix", func(p ...TopologyParams) ITopology[int] { return NewAdjacencyMatrix[int](p...) }},
	{"BitMatrix", func(p ...TopologyParams) ITopology[int] { return NewBitMatrix[int](p...) }},
	{"CSR", func(p ...TopologyParams) ITopology[int] { return NewCSR[int](p...) }},
}

func BenchmarkDFS_LinearChain(b *testing.B) {
	for _, n := range traversalBenchSizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n - 1 {
				g.Set(i, i+1)
			}
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					DFS(g, 0, tr)
				}
			})
		}
	}
}

func BenchmarkDFS_Tree(b *testing.B) {
	// binary tree: parent i has children 2i+1 and 2i+2
	for _, n := range traversalBenchSizes {
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
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					DFS(g, 0, tr)
				}
			})
		}
	}
}

func BenchmarkDFS_Dense(b *testing.B) {
	sizes := []int{32, 128, 512}
	for _, n := range sizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Scale: n}) // undirected
			for i := range n {
				for j := i + 1; j < n; j++ {
					g.Set(i, j)
				}
			}
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					DFS(g, 0, tr)
				}
			})
		}
	}
}

func BenchmarkBFS_LinearChain(b *testing.B) {
	for _, n := range traversalBenchSizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n - 1 {
				g.Set(i, i+1)
			}
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					BFS(g, 0, tr)
				}
			})
		}
	}
}

func BenchmarkBFS_Tree(b *testing.B) {
	for _, n := range traversalBenchSizes {
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
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					BFS(g, 0, tr)
				}
			})
		}
	}
}

func BenchmarkBFS_Dense(b *testing.B) {
	sizes := []int{32, 128, 512}
	for _, n := range sizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Scale: n}) // undirected
			for i := range n {
				for j := i + 1; j < n; j++ {
					g.Set(i, j)
				}
			}
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					BFS(g, 0, tr)
				}
			})
		}
	}
}

// ─── Dijkstra tests ───────────────────────────────────────────────────────────

// weights returns an edgeWeightFn that looks up w[from][to]; missing edges default to 1.
func weights[Key comparable](w map[Key]map[Key]float64) func(Key, Key) float64 {
	return func(from, to Key) float64 {
		if row, ok := w[from]; ok {
			if c, ok := row[to]; ok {
				return c
			}
		}
		return 1
	}
}

func TestDijkstra_LinearChain_Costs(t *testing.T) {
	// 0 -2-> 1 -3-> 2: shortest costs should be 2 and 5.
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			r := &recorder[int]{
				edgeWeightFn: weights(map[int]map[int]float64{
					0: {1: 2},
					1: {2: 3},
				}),
			}
			Dijkstra(g, 0, r)
			if cost, ok := r.lastCostTo(1); !ok || cost != 2 {
				t.Errorf("dist[1]: want 2, got %v (ok=%v)", cost, ok)
			}
			if cost, ok := r.lastCostTo(2); !ok || cost != 5 {
				t.Errorf("dist[2]: want 5, got %v (ok=%v)", cost, ok)
			}
			if !r.hasVisit(0, 1) || !r.hasVisit(1, 2) {
				t.Errorf("expected OnVisit(0,1) and OnVisit(1,2), got %v", r.visits)
			}
		})
	}
}

func TestDijkstra_TwoRoutes_ShorterChosen(t *testing.T) {
	// 0 -10-> 2 (direct, expensive)
	// 0 -1-> 1 -1-> 2 (via 1, cost 2)
	// Dijkstra must settle 2 via 1 with cost 2, not via 0 with cost 10.
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 2)
			g.Set(0, 1)
			g.Set(1, 2)
			r := &recorder[int]{
				edgeWeightFn: weights(map[int]map[int]float64{
					0: {2: 10, 1: 1},
					1: {2: 1},
				}),
			}
			Dijkstra(g, 0, r)
			cost, ok := r.lastCostTo(2)
			if !ok || cost != 2 {
				t.Errorf("dist[2]: want 2 (via 1), got %v (ok=%v)", cost, ok)
			}
			// OnVisit for 2 must come from parent 1
			if !r.hasVisit(1, 2) {
				t.Errorf("expected OnVisit(1,2) for shortest path, got %v", r.visits)
			}
		})
	}
}

func TestDijkstra_Disconnected_UnreachableNotVisited(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(2, 3) // isolated component
			r := &recorder[int]{}
			Dijkstra(g, 0, r)
			for _, e := range r.visits {
				if e.to == 2 || e.to == 3 {
					t.Errorf("unreachable vertex %d must not be visited", e.to)
				}
			}
			if !r.hasVisit(0, 1) {
				t.Error("reachable vertex 1 must be visited")
			}
		})
	}
}

func TestDijkstra_ShouldVisit_FiltersEdge(t *testing.T) {
	// 0 -1-> 1 -1-> 2, also 0 -1-> 2 (filtered).
	// Without filtering, 2 is reached via 0 directly (cost 1).
	// With filtering of edge (0,2), it must be reached via 1 (cost 2).
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(0, 2)
			r := &recorder[int]{
				shouldVisitFn: func(from, to int) bool { return !(from == 0 && to == 2) },
			}
			Dijkstra(g, 0, r)
			cost, ok := r.lastCostTo(2)
			if !ok || cost != 2 {
				t.Errorf("dist[2] should be 2 (via 1 after filtering direct edge), got %v", cost)
			}
		})
	}
}

func TestDijkstra_OnVisit_EarlyStop(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			for i := range 4 {
				g.Set(i, i+1)
			}
			r := &recorder[int]{
				onVisitFn: func(int, int) bool { return false },
			}
			Dijkstra(g, 0, r)
			if len(r.visits) != 1 {
				t.Errorf("expected stop after first settled vertex, got %d visits", len(r.visits))
			}
		})
	}
}

func TestDijkstra_AbsentStart_NoCallbacks(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			r := &recorder[int]{}
			Dijkstra(g, 99, r)
			if len(r.visits) != 0 || len(r.relaxes) != 0 {
				t.Error("absent start must produce no callbacks")
			}
		})
	}
}

// ─── A* tests ─────────────────────────────────────────────────────────────────

func TestAStar_FindsGoal(t *testing.T) {
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 3)
			r := &recorder[int]{}
			AStar(g, 0, 3, r)
			if !slices.Contains(r.visitedDestinations(), 3) {
				t.Errorf("goal 3 must be settled, visits: %v", r.visits)
			}
			cost, ok := r.lastCostTo(3)
			if !ok || cost != 3 {
				t.Errorf("dist[3]: want 3, got %v (ok=%v)", cost, ok)
			}
		})
	}
}

func TestAStar_StopsAtGoal_NodesAfterNotSettled(t *testing.T) {
	// Chain 0→1→2→3→4: goal is 2. Nodes 3 and 4 must not be settled.
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			for i := range 4 {
				g.Set(i, i+1)
			}
			r := &recorder[int]{}
			AStar(g, 0, 2, r)
			for _, e := range r.visits {
				if e.to == 3 || e.to == 4 {
					t.Errorf("vertex %d beyond goal must not be settled", e.to)
				}
			}
			if !slices.Contains(r.visitedDestinations(), 2) {
				t.Error("goal 2 must be settled")
			}
		})
	}
}

func TestAStar_ZeroHeuristic_SameResultAsDijkstra(t *testing.T) {
	// With a zero heuristic A* must settle the same vertices as Dijkstra
	// and produce the same final distances.
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(0, 2)
			g.Set(1, 3)
			g.Set(2, 3)
			w := weights[int](map[int]map[int]float64{
				0: {1: 1, 2: 4},
				1: {3: 1},
				2: {3: 1},
			})
			dijk := &recorder[int]{edgeWeightFn: w}
			Dijkstra(g, 0, dijk)

			astar := &recorder[int]{edgeWeightFn: w} // heuristicFn nil → 0
			AStar(g, 0, 3, astar)

			// Both must agree on the cost to reach goal 3
			dCost, _ := dijk.lastCostTo(3)
			aCost, _ := astar.lastCostTo(3)
			if dCost != aCost {
				t.Errorf("Dijkstra dist[3]=%v, A* dist[3]=%v — must agree", dCost, aCost)
			}
		})
	}
}

func TestAStar_Heuristic_PrioritisesGoalDirection(t *testing.T) {
	// Linear chain 0→1→2→3→4, goal=4.
	// A monotone heuristic h(n)=goal-n pushes the search straight toward the goal.
	// Verify the goal is found with correct cost.
	for _, f := range directedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			for i := range 4 {
				g.Set(i, i+1)
			}
			r := &recorder[int]{
				heuristicFn: func(current, goal int) float64 {
					if goal >= current {
						return float64(goal - current)
					}
					return 0
				},
			}
			AStar(g, 0, 4, r)
			cost, ok := r.lastCostTo(4)
			if !ok || cost != 4 {
				t.Errorf("dist[4]: want 4, got %v (ok=%v)", cost, ok)
			}
		})
	}
}

// ─── weighted benchmarks ──────────────────────────────────────────────────────

func BenchmarkDijkstra_LinearChain(b *testing.B) {
	for _, n := range traversalBenchSizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n - 1 {
				g.Set(i, i+1)
			}
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Dijkstra(g, 0, tr)
				}
			})
		}
	}
}

func BenchmarkDijkstra_Tree(b *testing.B) {
	for _, n := range traversalBenchSizes {
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
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Dijkstra(g, 0, tr)
				}
			})
		}
	}
}

func BenchmarkAStar_LinearChain(b *testing.B) {
	for _, n := range traversalBenchSizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Features: []Feature{Directed}, Scale: n})
			for i := range n - 1 {
				g.Set(i, i+1)
			}
			goal := n - 1
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					AStar(g, 0, goal, tr)
				}
			})
		}
	}
}

func BenchmarkAStar_Tree(b *testing.B) {
	for _, n := range traversalBenchSizes {
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
			goal := n - 1
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					AStar(g, 0, goal, tr)
				}
			})
		}
	}
}

func BenchmarkDijkstra_Dense(b *testing.B) {
	sizes := []int{32, 128, 512}
	for _, n := range sizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Scale: n}) // undirected complete graph
			for i := range n {
				for j := i + 1; j < n; j++ {
					g.Set(i, j)
				}
			}
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					Dijkstra(g, 0, tr)
				}
			})
		}
	}
}

func BenchmarkAStar_Dense(b *testing.B) {
	sizes := []int{32, 128, 512}
	for _, n := range sizes {
		for _, f := range benchFactories {
			g := f.new(TopologyParams{Scale: n}) // undirected complete graph
			for i := range n {
				for j := i + 1; j < n; j++ {
					g.Set(i, j)
				}
			}
			goal := n - 1
			tr := noopTraversal[int]{}
			b.Run(fmt.Sprintf("n=%d/%s", n, f.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					AStar(g, 0, goal, tr)
				}
			})
		}
	}
}
