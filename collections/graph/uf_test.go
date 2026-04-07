package graph

import "testing"

// cycleMatrix extends graphMatrix with IsCycled for cycle-specific tests.
type cycleMatrix interface {
	graphMatrix
	IsCycled() bool
}

var cycleImpls = []struct {
	name string
	new  func() cycleMatrix
}{
	{"AdjacencyMatrix", func() cycleMatrix { return NewAdjacencyMatrix[int]() }},
	{"BitMatrix", func() cycleMatrix { return NewBitMatrix[int]() }},
	{"CSR", func() cycleMatrix { return NewCSR[int]() }},
}

var directedCycleImpls = []struct {
	name string
	new  func() cycleMatrix
}{
	{"AdjacencyMatrix", func() cycleMatrix {
		return NewAdjacencyMatrix[int](TopologyParams{Features: []Feature{Directed}})
	}},
	{"BitMatrix", func() cycleMatrix {
		return NewBitMatrix[int](TopologyParams{Features: []Feature{Directed}})
	}},
	{"CSR", func() cycleMatrix {
		return NewCSR[int](TopologyParams{Features: []Feature{Directed}})
	}},
}

var acyclicCycleImpls = []struct {
	name string
	new  func() cycleMatrix
}{
	{"AdjacencyMatrix", func() cycleMatrix {
		return NewAdjacencyMatrix[int](TopologyParams{Features: []Feature{Acyclic}})
	}},
	{"BitMatrix", func() cycleMatrix {
		return NewBitMatrix[int](TopologyParams{Features: []Feature{Acyclic}})
	}},
	{"CSR", func() cycleMatrix {
		return NewCSR[int](TopologyParams{Features: []Feature{Acyclic}})
	}},
}

var acyclicDirectedCycleImpls = []struct {
	name string
	new  func() cycleMatrix
}{
	{"AdjacencyMatrix", func() cycleMatrix {
		return NewAdjacencyMatrix[int](TopologyParams{Features: []Feature{Directed, Acyclic}})
	}},
	{"BitMatrix", func() cycleMatrix {
		return NewBitMatrix[int](TopologyParams{Features: []Feature{Directed, Acyclic}})
	}},
	{"CSR", func() cycleMatrix {
		return NewCSR[int](TopologyParams{Features: []Feature{Directed, Acyclic}})
	}},
}

// ── IsCycled — undirected ────────────────────────────────────────────────────

func TestIsCycled_Undirected_NoCycle(t *testing.T) {
	for _, impl := range cycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			// tree: 0-1-2-3
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 3)
			if g.IsCycled() {
				t.Error("expected false, got true")
			}
		})
	}
}

func TestIsCycled_Undirected_Triangle(t *testing.T) {
	for _, impl := range cycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0)
			if !g.IsCycled() {
				t.Fatal("expected true, got false")
			}
		})
	}
}

func TestIsCycled_Undirected_LongerCycle(t *testing.T) {
	for _, impl := range cycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			// 0-1-2-3-4-0
			for i := 0; i < 4; i++ {
				g.Set(i, i+1)
			}
			g.Set(4, 0)
			if !g.IsCycled() {
				t.Fatal("expected true, got false")
			}
		})
	}
}

func TestIsCycled_Undirected_Empty(t *testing.T) {
	for _, impl := range cycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			if g.IsCycled() {
				t.Error("expected false on empty graph, got true")
			}
		})
	}
}

// ── IsCycled — directed ──────────────────────────────────────────────────────

func TestIsCycled_Directed_NoCycle_DAG(t *testing.T) {
	for _, impl := range directedCycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			// DAG: 0→1, 0→2, 1→3, 2→3
			g.Set(0, 1)
			g.Set(0, 2)
			g.Set(1, 3)
			g.Set(2, 3)
			if g.IsCycled() {
				t.Error("expected false for DAG, got true")
			}
		})
	}
}

func TestIsCycled_Directed_SimpleCycle(t *testing.T) {
	for _, impl := range directedCycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			// Bypass the anti-parallel guard by building the cycle manually via Add.
			// 0→1→2→0
			g.Add(0)
			g.Add(1)
			g.Add(2)
			g.Set(0, 1)
			g.Set(1, 2)
			g.Set(2, 0)
			if !g.IsCycled() {
				t.Fatal("expected true, got false")
			}
		})
	}
}

func TestIsCycled_Directed_SelfLoop(t *testing.T) {
	for _, impl := range directedCycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			g.Add(0)
			g.Set(0, 0) // self-loop
			if !g.IsCycled() {
				t.Fatal("expected true for self-loop, got false")
			}
		})
	}
}

// ── Acyclic feature — undirected ─────────────────────────────────────────────

func TestAcyclic_Undirected_AllowsTree(t *testing.T) {
	for _, impl := range acyclicCycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			if !g.Set(0, 1) {
				t.Error("expected Set(0,1) to succeed")
			}
			if !g.Set(1, 2) {
				t.Error("expected Set(1,2) to succeed")
			}
			if !g.Set(2, 3) {
				t.Error("expected Set(2,3) to succeed")
			}
		})
	}
}

func TestAcyclic_Undirected_RejectsClosingEdge(t *testing.T) {
	for _, impl := range acyclicCycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			g.Set(0, 1)
			g.Set(1, 2)
			// closing 0-2 would form triangle 0-1-2-0
			if g.Set(2, 0) {
				t.Error("expected Set(2,0) to be rejected as cycle-forming")
			}
			if g.IsCycled() {
				t.Error("graph should remain acyclic after rejection")
			}
		})
	}
}

func TestAcyclic_Undirected_AllowsNewVertex(t *testing.T) {
	for _, impl := range acyclicCycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			g.Set(0, 1)
			// vertex 2 is new — should be allowed
			if !g.Set(1, 2) {
				t.Error("expected Set(1,2) to succeed for new vertex")
			}
		})
	}
}

// ── Acyclic feature — directed ───────────────────────────────────────────────

func TestAcyclic_Directed_AllowsDAGEdge(t *testing.T) {
	for _, impl := range acyclicDirectedCycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			// 0→1, 0→2, 1→3, 2→3 is a valid DAG
			if !g.Set(0, 1) {
				t.Error("0→1 should be allowed")
			}
			if !g.Set(0, 2) {
				t.Error("0→2 should be allowed")
			}
			if !g.Set(1, 3) {
				t.Error("1→3 should be allowed")
			}
			if !g.Set(2, 3) {
				t.Error("2→3 should be allowed")
			}
		})
	}
}

func TestAcyclic_Directed_RejectsBackEdge(t *testing.T) {
	for _, impl := range acyclicDirectedCycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			g.Set(0, 1)
			g.Set(1, 2)
			// 2→0 would close the cycle 0→1→2→0
			if g.Set(2, 0) {
				t.Error("expected Set(2,0) to be rejected as back edge")
			}
			if g.IsCycled() {
				t.Error("graph should remain acyclic after rejection")
			}
		})
	}
}

func TestAcyclic_Directed_AllowsCrossEdge(t *testing.T) {
	for _, impl := range acyclicDirectedCycleImpls {
		t.Run(impl.name, func(t *testing.T) {
			g := impl.new()
			g.Set(0, 1)
			g.Set(0, 2)
			// 1→2 is a cross edge, no cycle
			if !g.Set(1, 2) {
				t.Error("expected cross edge 1→2 to be allowed")
			}
		})
	}
}
