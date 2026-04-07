package graph

import (
	"slices"
	"testing"
)

// ─── test vertex ─────────────────────────────────────────────────────────────

type strVertex string

func (v strVertex) Key() string { return string(v) }

// ─── factories ────────────────────────────────────────────────────────────────

type graphFactory struct {
	name string
	new  func() IGraph[strVertex, float64, string]
}

var graphFactories = []graphFactory{
	{"AdjacencyMatrix", func() IGraph[strVertex, float64, string] {
		return NewGraph[strVertex, float64](TopologyParams{Key: AdjacencyMatrixTopology})
	}},
	{"BitMatrix", func() IGraph[strVertex, float64, string] {
		return NewGraph[strVertex, float64](TopologyParams{Key: BitMatrixTopology})
	}},
	{"CSR", func() IGraph[strVertex, float64, string] {
		return NewGraph[strVertex, float64](TopologyParams{Key: CRSTopology})
	}},
}

var directedGraphFactories = []graphFactory{
	{"AdjacencyMatrix", func() IGraph[strVertex, float64, string] {
		return NewGraph[strVertex, float64](TopologyParams{Key: AdjacencyMatrixTopology, Features: Features(Directed)})
	}},
	{"BitMatrix", func() IGraph[strVertex, float64, string] {
		return NewGraph[strVertex, float64](TopologyParams{Key: BitMatrixTopology, Features: Features(Directed)})
	}},
	{"CSR", func() IGraph[strVertex, float64, string] {
		return NewGraph[strVertex, float64](TopologyParams{Key: CRSTopology, Features: Features(Directed)})
	}},
}

// ─── Add / Contains / Delete ──────────────────────────────────────────────────

func TestGraph_Add_Contains(t *testing.T) {
	for _, f := range graphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a := strVertex("a")
			if g.Contains(a) {
				t.Fatal("should not contain vertex before Add")
			}
			g.Add(a)
			if !g.Contains(a) {
				t.Fatal("should contain vertex after Add")
			}
		})
	}
}

func TestGraph_Add_Idempotent(t *testing.T) {
	for _, f := range graphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a := strVertex("a")
			g.Add(a)
			g.Add(a)
			if !g.Contains(a) {
				t.Fatal("vertex must still be present after double Add")
			}
		})
	}
}

func TestGraph_Delete_Present(t *testing.T) {
	for _, f := range graphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a := strVertex("a")
			g.Add(a)
			if !g.Delete(a) {
				t.Fatal("Delete returned false for present vertex")
			}
			if g.Contains(a) {
				t.Fatal("vertex must be absent after Delete")
			}
		})
	}
}

func TestGraph_Delete_Absent(t *testing.T) {
	for _, f := range graphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			if g.Delete(strVertex("x")) {
				t.Fatal("Delete must return false for absent vertex")
			}
		})
	}
}

func TestGraph_Delete_RemovesEdges(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 1.0)
			g.Set(b, c, 2.0)
			g.Delete(b)
			if g.Has(a, b) {
				t.Error("edge a→b must be gone after deleting b")
			}
			if g.Has(b, c) {
				t.Error("edge b→c must be gone after deleting b")
			}
		})
	}
}

// ─── Set / Has ────────────────────────────────────────────────────────────────

func TestGraph_Set_Has_Edge(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b := strVertex("a"), strVertex("b")
			if !g.Set(a, b, 1.0) {
				t.Fatal("Set returned false for new edge")
			}
			if !g.Has(a, b) {
				t.Fatal("Has must return true after Set")
			}
		})
	}
}

func TestGraph_Set_AutoAddsFrom(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b := strVertex("a"), strVertex("b")
			g.Set(a, b, 0)
			if !g.Contains(a) {
				t.Fatal("Set must implicitly add source vertex")
			}
			if !g.Has(a, b) {
				t.Fatal("Set must make edge traversable immediately")
			}
		})
	}
}

func TestGraph_Has_Vertex(t *testing.T) {
	for _, f := range graphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a := strVertex("a")
			g.Add(a)
			if !g.Has(a) {
				t.Fatal("Has(single vertex) must return true")
			}
		})
	}
}

func TestGraph_Has_Path(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 1.0)
			g.Set(b, c, 1.0)
			if !g.Has(a, b, c) {
				t.Fatal("Has must detect path a→b→c")
			}
			if g.Has(a, c) {
				t.Fatal("Has must not detect non-existent direct edge a→c")
			}
		})
	}
}

func TestGraph_Set_Undirected_BothDirections(t *testing.T) {
	for _, f := range graphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b := strVertex("a"), strVertex("b")
			g.Set(a, b, 1.0)
			if !g.Has(a, b) || !g.Has(b, a) {
				t.Fatal("undirected Set must make edge traversable both ways")
			}
		})
	}
}

// ─── Remove ───────────────────────────────────────────────────────────────────

func TestGraph_Remove_Edge(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b := strVertex("a"), strVertex("b")
			g.Add(a)
			g.Add(b)
			g.Set(a, b, 1.0)
			if !g.Remove(a, b) {
				t.Fatal("Remove must return true for existing edge")
			}
			if g.Has(a, b) {
				t.Fatal("edge must be absent after Remove")
			}
			// Vertices still exist
			if !g.Has(a) || !g.Has(b) {
				t.Fatal("Remove(edge) must not delete vertices")
			}
		})
	}
}

func TestGraph_Remove_Vertex(t *testing.T) {
	for _, f := range graphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a := strVertex("a")
			g.Add(a)
			if !g.Remove(a) {
				t.Fatal("Remove must return true for existing vertex")
			}
			if g.Contains(a) {
				t.Fatal("vertex must be absent after Remove")
			}
		})
	}
}

func TestGraph_Remove_AbsentEdge_ReturnsFalse(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b := strVertex("a"), strVertex("b")
			g.Add(a)
			g.Add(b)
			if g.Remove(a, b) {
				t.Fatal("Remove must return false for non-existent edge")
			}
		})
	}
}

func TestGraph_Remove_ClearsEdgeData(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b := strVertex("a"), strVertex("b")
			g.Set(a, b, 42.0)
			g.Remove(a, b)
			if _, ok := g.GetEdge(a, b); ok {
				t.Fatal("GetEdge must return false after Remove")
			}
		})
	}
}

// ─── IsCycled ─────────────────────────────────────────────────────────────────

func TestGraph_IsCycled_Acyclic(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 0)
			g.Set(b, c, 0)
			if g.IsCycled() {
				t.Fatal("DAG must not be cycled")
			}
		})
	}
}

func TestGraph_IsCycled_WithCycle(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 0)
			g.Set(b, c, 0)
			g.Set(c, a, 0)
			if !g.IsCycled() {
				t.Fatal("directed cycle must be detected")
			}
		})
	}
}

// ─── Neighbors ────────────────────────────────────────────────────────────────

func TestGraph_Neighbors(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 0)
			g.Set(a, c, 0)
			nbrs := g.Neighbors(a)
			if len(nbrs) != 2 {
				t.Fatalf("want 2 neighbors, got %v", nbrs)
			}
			if !slices.Contains(nbrs, "b") || !slices.Contains(nbrs, "c") {
				t.Errorf("expected neighbors b and c, got %v", nbrs)
			}
		})
	}
}

func TestGraph_Neighbors_Absent(t *testing.T) {
	for _, f := range graphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			nbrs := g.Neighbors(strVertex("x"))
			if len(nbrs) != 0 {
				t.Fatalf("absent vertex must have no neighbors, got %v", nbrs)
			}
		})
	}
}

// ─── GetVertex ────────────────────────────────────────────────────────────────

func TestGraph_GetVertex_Present(t *testing.T) {
	for _, f := range graphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a := strVertex("a")
			g.Add(a)
			v, ok := g.GetVertex("a")
			if !ok {
				t.Fatal("GetVertex must return ok=true for added vertex")
			}
			if v.Key() != "a" {
				t.Fatalf("GetVertex returned wrong vertex: %v", v)
			}
		})
	}
}

func TestGraph_GetVertex_Absent(t *testing.T) {
	for _, f := range graphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			if _, ok := g.GetVertex("z"); ok {
				t.Fatal("GetVertex must return ok=false for absent vertex")
			}
		})
	}
}

// ─── GetEdge ──────────────────────────────────────────────────────────────────

func TestGraph_GetEdge_Present(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b := strVertex("a"), strVertex("b")
			g.Set(a, b, 3.14)
			edge, ok := g.GetEdge(a, b)
			if !ok {
				t.Fatal("GetEdge must return ok=true for existing edge")
			}
			if edge != 3.14 {
				t.Fatalf("expected edge weight 3.14, got %v", edge)
			}
		})
	}
}

func TestGraph_GetEdge_Absent(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b := strVertex("a"), strVertex("b")
			g.Add(a)
			g.Add(b)
			if _, ok := g.GetEdge(a, b); ok {
				t.Fatal("GetEdge must return ok=false for absent edge")
			}
		})
	}
}

func TestGraph_GetEdge_UpdatedOnReSet(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b := strVertex("a"), strVertex("b")
			g.Set(a, b, 1.0)
			g.Set(a, b, 9.9) // topology Set is idempotent; edge map updated
			edge, ok := g.GetEdge(a, b)
			if !ok {
				t.Fatal("edge must still exist")
			}
			if edge != 9.9 {
				t.Fatalf("expected updated edge 9.9, got %v", edge)
			}
		})
	}
}

// ─── Paths ────────────────────────────────────────────────────────────────────

func TestGraph_Paths_LinearChain(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 0)
			g.Set(b, c, 0)
			for _, mode := range []SearchMode{DFSSearch, BFSSearch} {
				got := g.Paths("a", false, mode)
				if len(got) != 1 {
					t.Fatalf("mode=%d want 1 path, got %d: %v", mode, len(got), got)
				}
				if !slices.Equal(got[0], []string{"a", "b", "c"}) {
					t.Errorf("mode=%d expected [a b c], got %v", mode, got[0])
				}
			}
		})
	}
}

func TestGraph_Paths_ExcludeVertex(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 0)
			g.Set(a, c, 0)
			for _, mode := range []SearchMode{DFSSearch, BFSSearch} {
				got := g.Paths("a", false, mode, "b")
				if len(got) != 1 {
					t.Fatalf("mode=%d want 1 path after exclude, got %v", mode, got)
				}
				if !slices.Equal(got[0], []string{"a", "c"}) {
					t.Errorf("mode=%d expected [a c], got %v", mode, got[0])
				}
			}
		})
	}
}

func TestGraph_Paths_Cycled(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 0)
			g.Set(b, c, 0)
			g.Set(c, a, 0)
			for _, mode := range []SearchMode{DFSSearch, BFSSearch} {
				got := g.Paths("a", true, mode)
				hasCycle := false
				for _, p := range got {
					if len(p) >= 2 && p[0] == p[len(p)-1] {
						hasCycle = true
					}
				}
				if !hasCycle {
					t.Errorf("mode=%d: cycled=true must include cycle-closure path, got %v", mode, got)
				}
			}
		})
	}
}

// ─── DFS / BFS traversal ──────────────────────────────────────────────────────

func TestGraph_DFS_VisitsAllVertices(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 0)
			g.Set(b, c, 0)
			rec := &recorder[string]{}
			g.DFS("a", rec)
			visited := make(map[string]bool)
			for _, e := range rec.visits {
				visited[e.to] = true
			}
			for _, key := range []string{"b", "c"} {
				if !visited[key] {
					t.Errorf("DFS did not visit %q", key)
				}
			}
		})
	}
}

func TestGraph_BFS_VisitsAllVertices(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 0)
			g.Set(b, c, 0)
			rec := &recorder[string]{}
			g.BFS("a", rec)
			visited := make(map[string]bool)
			for _, e := range rec.visits {
				visited[e.to] = true
			}
			for _, key := range []string{"b", "c"} {
				if !visited[key] {
					t.Errorf("BFS did not visit %q", key)
				}
			}
		})
	}
}

// ─── Dijkstra ─────────────────────────────────────────────────────────────────

func TestGraph_Dijkstra_ShortestPath(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			// a→b(1), a→c(10), b→c(1): shortest a→c via b costs 2
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 1.0)
			g.Set(a, c, 10.0)
			g.Set(b, c, 1.0)
			weights := map[string]map[string]float64{
				"a": {"b": 1, "c": 10},
				"b": {"c": 1},
			}
			rec := &recorder[string]{
				edgeWeightFn: func(from, to string) float64 {
					if row, ok := weights[from]; ok {
						if w, ok2 := row[to]; ok2 {
							return w
						}
					}
					return 1
				},
			}
			g.Dijkstra("a", rec)
			cost, ok := rec.lastCostTo("c")
			if !ok {
				t.Fatal("Dijkstra did not relax vertex c")
			}
			if cost != 2.0 {
				t.Errorf("expected cost 2.0 to c, got %v", cost)
			}
		})
	}
}

// ─── AStar ────────────────────────────────────────────────────────────────────

func TestGraph_AStar_ReachesGoal(t *testing.T) {
	for _, f := range directedGraphFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			a, b, c := strVertex("a"), strVertex("b"), strVertex("c")
			g.Set(a, b, 1.0)
			g.Set(b, c, 1.0)
			rec := &recorder[string]{}
			g.AStar("a", "c", rec)
			visited := make(map[string]bool)
			for _, e := range rec.visits {
				visited[e.to] = true
			}
			if !visited["c"] {
				t.Error("AStar must visit goal vertex c")
			}
		})
	}
}
