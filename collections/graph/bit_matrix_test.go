package graph

import "testing"

// --- Add ---

func TestBitMatrix_Add_VertexIsPresent(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	if !m.Contains("A") {
		t.Error("Add: vertex A should be present")
	}
}

func TestBitMatrix_Add_Idempotent(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	m.Add("A")
	if !m.Contains("A") {
		t.Error("Add idempotent: vertex A should still be present")
	}
}

func TestBitMatrix_Add_ManyVertices_AllPresent(t *testing.T) {
	m := NewBitMatrix[int]()
	const n = 64
	for i := range n {
		m.Add(i)
	}
	for i := range n {
		if !m.Contains(i) {
			t.Errorf("Add: vertex %d missing after adding %d vertices", i, n)
		}
	}
}

// --- Contains ---

func TestBitMatrix_Contains_Present(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	if !m.Contains("A") {
		t.Error("Contains: expected true for present vertex")
	}
}

func TestBitMatrix_Contains_Absent(t *testing.T) {
	m := NewBitMatrix[string]()
	if m.Contains("Z") {
		t.Error("Contains: expected false for absent vertex")
	}
}

// --- Delete ---

func TestBitMatrix_Delete_PresentVertex_ReturnsTrue(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	if !m.Delete("A") {
		t.Error("Delete: expected true for present vertex")
	}
}

func TestBitMatrix_Delete_PresentVertex_IsGone(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	m.Delete("A")
	if m.Contains("A") {
		t.Error("Delete: vertex should not be present after deletion")
	}
}

func TestBitMatrix_Delete_AbsentVertex_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	if m.Delete("Z") {
		t.Error("Delete: expected false for absent vertex")
	}
}

func TestBitMatrix_Delete_Idempotent(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	m.Delete("A")
	if m.Delete("A") {
		t.Error("Delete: second call on deleted vertex should return false")
	}
}

func TestBitMatrix_Delete_ClearsIncomingEdges(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Set("A", "B") // undirected: A↔B
	m.Delete("A")
	if m.Has("B", "A") {
		t.Error("Delete: incoming edge B→A should be cleared after deleting A")
	}
}

func TestBitMatrix_Delete_ReleasesIndex_ReusedOnNextAdd(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A") // index 0
	m.Delete("A")
	m.Add("B") // reuses index 0
	if !m.Contains("B") {
		t.Error("Delete+Add: reused index should be accessible via Contains")
	}
	if m.Contains("A") {
		t.Error("Delete+Add: deleted vertex A should not be present")
	}
}

// --- Set (undirected) ---

func TestBitMatrix_Set_CreatesEdgeBothDirections(t *testing.T) {
	m := NewBitMatrix[string]()
	if !m.Set("A", "B") {
		t.Error("Set: expected true")
	}
	if !m.Has("A", "B") {
		t.Error("Set undirected: edge A→B missing")
	}
	if !m.Has("B", "A") {
		t.Error("Set undirected: reverse edge B→A missing")
	}
}

func TestBitMatrix_Set_CreatesVerticesImplicitly(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Set("A", "B")
	if !m.Contains("A") || !m.Contains("B") {
		t.Error("Set: vertices should be implicitly created")
	}
}

func TestBitMatrix_Set_OverwriteEdge_ReturnsTrue(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Set("A", "B")
	if !m.Set("A", "B") {
		t.Error("Set: overwriting an edge should return true")
	}
}

func TestBitMatrix_Set_IndicesSpanningByteBoundary(t *testing.T) {
	m := NewBitMatrix[int]()
	for i := range 10 {
		m.Add(i)
	}
	if !m.Set(0, 9) {
		t.Error("Set: expected true for vertices spanning byte boundary")
	}
	if !m.Has(0, 9) {
		t.Error("Set: edge 0→9 missing across byte boundary")
	}
}

// --- Set (directed) ---

func TestBitMatrix_Set_DirectedEdge_OnlyOneDirection(t *testing.T) {
	m := NewBitMatrix[string](TopologyParams{Features: []Feature{Directed}})
	m.Set("A", "B")
	if !m.Has("A", "B") {
		t.Error("Set directed: edge A→B missing")
	}
	if m.Has("B", "A") {
		t.Error("Set directed: reverse edge B→A should not exist")
	}
}

func TestBitMatrix_Set_DirectedFeature_BlocksReverseEdge(t *testing.T) {
	m := NewBitMatrix[string](TopologyParams{Features: []Feature{Directed}})
	m.Set("A", "B")
	if m.Set("B", "A") {
		t.Error("Set directed: adding reverse edge B→A should return false")
	}
}

// --- Has ---

func TestBitMatrix_Has_EmptyArgs_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	if m.Has() {
		t.Error("Has(): expected false for empty args")
	}
}

func TestBitMatrix_Has_PresentVertex_ReturnsTrue(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	if !m.Has("A") {
		t.Error("Has: expected true for present vertex")
	}
}

func TestBitMatrix_Has_AbsentVertex_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	if m.Has("X") {
		t.Error("Has: expected false for absent vertex")
	}
}

func TestBitMatrix_Has_PresentEdge_ReturnsTrue(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Set("A", "B")
	if !m.Has("A", "B") {
		t.Error("Has: edge A→B should be present")
	}
}

func TestBitMatrix_Has_AbsentEdge_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	m.Add("B")
	if m.Has("A", "B") {
		t.Error("Has: edge A→B should not exist without Set")
	}
}

func TestBitMatrix_Has_Path_AllEdgesPresent_ReturnsTrue(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Set("A", "B")
	m.Set("B", "C")
	if !m.Has("A", "B", "C") {
		t.Error("Has path A→B→C: expected true")
	}
}

func TestBitMatrix_Has_Path_MissingEdge_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Set("A", "B")
	m.Add("C") // B→C not set
	if m.Has("A", "B", "C") {
		t.Error("Has path A→B→C: expected false, B→C edge missing")
	}
}

func TestBitMatrix_Has_Path_AbsentVertex_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Set("A", "B")
	if m.Has("A", "B", "Z") {
		t.Error("Has path A→B→Z: expected false, Z is not a vertex")
	}
}

// --- Remove ---

func TestBitMatrix_Remove_EmptyArgs_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	if m.Remove() {
		t.Error("Remove(): expected false for empty args")
	}
}

func TestBitMatrix_Remove_AbsentVertex_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	if m.Remove("Z") {
		t.Error("Remove: expected false for absent vertex")
	}
}

func TestBitMatrix_Remove_Vertex_ReturnsTrue(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	if !m.Remove("A") {
		t.Error("Remove vertex: expected true")
	}
}

func TestBitMatrix_Remove_Vertex_IsGone(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	m.Remove("A")
	if m.Contains("A") {
		t.Error("Remove vertex: A should not be present after removal")
	}
}

func TestBitMatrix_Remove_Edge_UndirectedRemovesBothDirections(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Set("A", "B")
	if !m.Remove("A", "B") {
		t.Error("Remove edge: expected true")
	}
	if m.Has("A", "B") {
		t.Error("Remove undirected edge: A→B should be gone")
	}
	if m.Has("B", "A") {
		t.Error("Remove undirected edge: B→A should be gone")
	}
}

func TestBitMatrix_Remove_Edge_DirectedRemovesOneDirection(t *testing.T) {
	m := NewBitMatrix[string](TopologyParams{Features: []Feature{Directed}})
	m.Set("A", "B")
	m.Remove("A", "B")
	if m.Has("A", "B") {
		t.Error("Remove directed edge: A→B should be gone")
	}
}

func TestBitMatrix_Remove_Edge_NoEdge_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	m.Add("B")
	if m.Remove("A", "B") {
		t.Error("Remove: expected false when no edge exists between present vertices")
	}
}

func TestBitMatrix_Remove_Path_RemovesAllEdges(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Set("A", "B")
	m.Set("B", "C")
	if !m.Remove("A", "B", "C") {
		t.Error("Remove path: expected true")
	}
	if m.Has("A", "B") {
		t.Error("Remove path: A→B should be gone")
	}
	if m.Has("B", "C") {
		t.Error("Remove path: B→C should be gone")
	}
}

func TestBitMatrix_Remove_Path_MissingEdge_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Set("A", "B")
	m.Add("C") // B→C not set
	if m.Remove("A", "B", "C") {
		t.Error("Remove path A→B→C: expected false, B→C edge missing")
	}
}

func TestBitMatrix_Remove_Path_AbsentVertex_ReturnsFalse(t *testing.T) {
	m := NewBitMatrix[string]()
	m.Add("A")
	if m.Remove("A", "Z") {
		t.Error("Remove path with absent vertex: expected false")
	}
}
