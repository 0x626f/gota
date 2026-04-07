package graph

import (
	"slices"
	"testing"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

// containsPath reports whether got contains a path equal to want.
func containsPath[Key comparable](got []Path[Key], want []Key) bool {
	for _, p := range got {
		if slices.Equal(p, want) {
			return true
		}
	}
	return false
}

// pathModes runs a sub-test for both DFSSearch and BFSSearch.
func pathModes(t *testing.T, name string, fn func(t *testing.T, mode SearchMode)) {
	t.Helper()
	t.Run(name+"/DFS", func(t *testing.T) { fn(t, DFSSearch) })
	t.Run(name+"/BFS", func(t *testing.T) { fn(t, BFSSearch) })
}

// pathsFactories is the same three implementations used everywhere.
var pathsFactories = []topoFactory{
	{"AdjacencyMatrix", func() ITopology[int] { return NewAdjacencyMatrix[int]() }},
	{"BitMatrix", func() ITopology[int] { return NewBitMatrix[int]() }},
	{"CSR", func() ITopology[int] { return NewCSR[int]() }},
}

var pathsDirectedFactories = []topoFactory{
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

// ─── tests ────────────────────────────────────────────────────────────────────

func TestPaths_AbsentStart_Nil(t *testing.T) {
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				if got := Paths(g, 99, false, mode); got != nil {
					t.Errorf("absent start: want nil, got %v", got)
				}
			})
		})
	}
}

func TestPaths_IsolatedVertex_Nil(t *testing.T) {
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Add(0)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				if got := Paths(g, 0, false, mode); got != nil {
					t.Errorf("isolated vertex: want nil, got %v", got)
				}
			})
		})
	}
}

func TestPaths_ExcludedStart_Nil(t *testing.T) {
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				if got := Paths(g, 0, false, mode, 0); got != nil {
					t.Errorf("excluded start: want nil, got %v", got)
				}
			})
		})
	}
}

func TestPaths_LinearChain_SinglePath(t *testing.T) {
	// 0→1→2→3: only one simple path from 0.
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			for i := range 3 {
				g.Set(i, i+1)
			}
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, false, mode)
				if len(got) != 1 {
					t.Fatalf("want 1 path, got %d: %v", len(got), got)
				}
				if !containsPath(got, []int{0, 1, 2, 3}) {
					t.Errorf("expected path [0 1 2 3], got %v", got)
				}
			})
		})
	}
}

func TestPaths_BranchingTree_AllLeafPaths(t *testing.T) {
	// 0→1, 0→2, 1→3, 1→4: two leaf paths from 0.
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(0, 2)
			g.Set(1, 3)
			g.Set(1, 4)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, false, mode)
				if len(got) != 3 {
					t.Fatalf("want 3 paths, got %d: %v", len(got), got)
				}
				for _, want := range [][]int{{0, 1, 3}, {0, 1, 4}, {0, 2}} {
					if !containsPath(got, want) {
						t.Errorf("missing path %v in %v", want, got)
					}
				}
			})
		})
	}
}

func TestPaths_DiamondDAG_TwoPaths(t *testing.T) {
	// 0→1→3 and 0→2→3: two distinct simple paths to the same sink.
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(0, 2)
			g.Set(1, 3)
			g.Set(2, 3)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, false, mode)
				if len(got) != 2 {
					t.Fatalf("want 2 paths, got %d: %v", len(got), got)
				}
				for _, want := range [][]int{{0, 1, 3}, {0, 2, 3}} {
					if !containsPath(got, want) {
						t.Errorf("missing path %v in %v", want, got)
					}
				}
			})
		})
	}
}

func TestPaths_Undirected_BothDirections(t *testing.T) {
	// A-B undirected: paths from A are [A,B]; B's only remaining
	// neighbour is A (in path) so B is a leaf.
	for _, f := range pathsFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, false, mode)
				if len(got) != 1 || !containsPath(got, []int{0, 1}) {
					t.Errorf("want [[0 1]], got %v", got)
				}
			})
		})
	}
}

func TestPaths_Disconnected_OnlyReachable(t *testing.T) {
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1) // reachable component
			g.Set(2, 3) // unreachable component
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, false, mode)
				for _, p := range got {
					for _, v := range p {
						if v == 2 || v == 3 {
							t.Errorf("unreachable vertex %d found in path %v", v, p)
						}
					}
				}
				if !containsPath(got, []int{0, 1}) {
					t.Errorf("missing reachable path [0 1] in %v", got)
				}
			})
		})
	}
}

func TestPaths_ExcludeVertex_PathAvoidsIt(t *testing.T) {
	// 0→1→2 and 0→3: exclude 1; only path via 3 should remain.
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(0, 3)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, false, mode, 1)
				for _, p := range got {
					if slices.Contains(p, 1) {
						t.Errorf("excluded vertex 1 found in path %v", p)
					}
				}
				if !containsPath(got, []int{0, 3}) {
					t.Errorf("expected path [0 3] after excluding 1, got %v", got)
				}
			})
		})
	}
}

func TestPaths_ExcludeMultiple_AllAvoided(t *testing.T) {
	// 0→1, 0→2, 0→3: exclude 1 and 2; only [0,3] survives.
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(0, 2)
			g.Set(0, 3)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, false, mode, 1, 2)
				if len(got) != 1 || !containsPath(got, []int{0, 3}) {
					t.Errorf("want [[0 3]], got %v", got)
				}
			})
		})
	}
}

// ─── cycled=true tests ────────────────────────────────────────────────────────

func TestPaths_Cycled_DirectedCycle_IncludesCyclePath(t *testing.T) {
	// 0→1→2→0: the DFS leaf [0,1,2] (all of 2's neighbours are in-path)
	// and the cycle closure [0,1,2,0] must both appear.
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, true, mode)
				if !containsPath(got, []int{0, 1, 2, 0}) {
					t.Errorf("cycle path [0 1 2 0] missing from %v", got)
				}
			})
		})
	}
}

func TestPaths_Cycled_SelfLoop(t *testing.T) {
	// 0→0: self-loop should produce cycle path [0,0].
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 0)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, true, mode)
				if !containsPath(got, []int{0, 0}) {
					t.Errorf("self-loop cycle [0 0] missing from %v", got)
				}
			})
		})
	}
}

func TestPaths_NotCycled_CyclePathAbsent(t *testing.T) {
	// With cycled=false, cycle-closing paths must never appear.
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, false, mode)
				for _, p := range got {
					if p[0] == p[len(p)-1] {
						t.Errorf("cycled=false must not return a cycle path, got %v", p)
					}
				}
			})
		})
	}
}

func TestPaths_Cycled_UndirectedTriangle(t *testing.T) {
	// Triangle 0-1-2: with cycled=true cycle closure paths appear.
	for _, f := range pathsFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0)
			pathModes(t, "", func(t *testing.T, mode SearchMode) {
				got := Paths(g, 0, true, mode)
				hasCycle := false
				for _, p := range got {
					if len(p) >= 2 && p[0] == p[len(p)-1] {
						hasCycle = true
						break
					}
				}
				if !hasCycle {
					t.Errorf("expected at least one cycle-closure path, got %v", got)
				}
			})
		})
	}
}

// ─── DFS vs BFS ordering ──────────────────────────────────────────────────────

func TestPaths_BFS_ShortestPathsFirst(t *testing.T) {
	// 0→1→2→3 with 0→3 shortcut: BFS records [0,3] before [0,1,2,3].
	for _, f := range pathsDirectedFactories {
		t.Run(f.name, func(t *testing.T) {
			g := f.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 3)
			g.Set(0, 3)
			got := Paths(g, 0, false, BFSSearch)
			// Both paths must be present.
			if !containsPath(got, []int{0, 3}) || !containsPath(got, []int{0, 1, 2, 3}) {
				t.Fatalf("expected both paths, got %v", got)
			}
			// [0,3] (length 2) must appear before [0,1,2,3] (length 4).
			idx := func(want []int) int {
				for i, p := range got {
					if slices.Equal(p, want) {
						return i
					}
				}
				return -1
			}
			if idx([]int{0, 3}) > idx([]int{0, 1, 2, 3}) {
				t.Errorf("BFS must record shorter path [0 3] before longer [0 1 2 3], got %v", got)
			}
		})
	}
}
