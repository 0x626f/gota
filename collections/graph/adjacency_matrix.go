package graph

// AdjacencyMatrix stores edges as a map-of-maps (vertex → set of neighbours).
// Edge lookup, insert, and delete are all O(1). Vertex deletion is O(V) because
// it must scan every row to remove incoming edges.
type AdjacencyMatrix[Key comparable] struct {
	baseTopology[Key]
	data map[Key]map[Key]struct{}
}

// NewAdjacencyMatrix creates an AdjacencyMatrix with the given options.
func NewAdjacencyMatrix[Key comparable](params ...TopologyParams) *AdjacencyMatrix[Key] {
	var p TopologyParams
	if len(params) > 0 {
		p = params[0]
	}
	m := &AdjacencyMatrix[Key]{
		data: make(map[Key]map[Key]struct{}, p.Scale),
	}
	for _, f := range p.Features {
		m.features.SetFeature(f)
	}
	return m
}

func (matrix *AdjacencyMatrix[Key]) Add(vertex Key) {
	if _, exists := matrix.data[vertex]; !exists {
		matrix.data[vertex] = make(map[Key]struct{})
		if matrix.uf != nil {
			matrix.uf.addVertex(vertex)
		}
	}
}

func (matrix *AdjacencyMatrix[Key]) Contains(vertex Key) (result bool) {
	_, result = matrix.data[vertex]
	return
}

func (matrix *AdjacencyMatrix[Key]) Delete(vertex Key) (result bool) {
	if matrix.Contains(vertex) {
		delete(matrix.data, vertex)
		for _, edges := range matrix.data {
			delete(edges, vertex)
		}
		if matrix.uf != nil {
			matrix.uf.removeVertex(vertex)
		}
		matrix.rebuildUnions()
		result = true
	}
	return
}

func (matrix *AdjacencyMatrix[Key]) Set(vertex0, vertex1 Key) bool {
	if matrix.features.HasFeature(Directed) && matrix.Has(vertex1, vertex0) {
		return false
	}
	if matrix.features.HasFeature(Acyclic) {
		if matrix.features.HasFeature(Directed) {
			if isReachable(vertex1, vertex0, matrix.neighbors) {
				return false
			}
		} else {
			matrix.initUF()
			if matrix.Contains(vertex0) && matrix.Contains(vertex1) {
				if matrix.uf.find(vertex0) == matrix.uf.find(vertex1) {
					return false
				}
			}
		}
	}

	matrix.Add(vertex0)
	matrix.data[vertex0][vertex1] = struct{}{}

	if !matrix.features.HasFeature(Directed) {
		matrix.Add(vertex1)
		matrix.data[vertex1][vertex0] = struct{}{}
	}

	if matrix.features.HasFeature(Acyclic) && !matrix.features.HasFeature(Directed) {
		matrix.uf.addVertex(vertex0)
		matrix.uf.addVertex(vertex1)
		matrix.uf.union(vertex0, vertex1)
	}

	return true
}

func (matrix *AdjacencyMatrix[Key]) IsCycled() bool {
	vertices := make([]Key, 0, len(matrix.data))
	for v := range matrix.data {
		vertices = append(vertices, v)
	}
	if matrix.features.HasFeature(Directed) {
		return findCycleDirected(vertices, matrix.neighbors) != nil
	}
	return findCycleUndirected(vertices, matrix.neighbors) != nil
}

func (matrix *AdjacencyMatrix[Key]) Has(path ...Key) bool {
	if len(path) == 0 {
		return false
	}

	if len(path) == 1 {
		return matrix.Contains(path[0])
	}

	for cursor, from := range path {
		if _, exists := matrix.data[from]; exists {
			if cursor+1 == len(path) {
				break
			}
			to := path[cursor+1]

			if _, exists = matrix.data[from][to]; !exists {
				return false
			}
		} else if cursor+1 != len(path) {
			return false
		}
	}

	return true
}

func (matrix *AdjacencyMatrix[Key]) Remove(path ...Key) bool {
	if len(path) == 0 {
		return false
	}

	if len(path) == 1 {
		return matrix.Delete(path[0])
	}

	for cursor, from := range path {
		if cursor+1 == len(path) {
			break
		}

		if _, ok := matrix.data[from]; !ok {
			return false
		}

		to := path[cursor+1]
		if _, ok := matrix.data[from][to]; !ok {
			return false
		}
	}

	for cursor, from := range path {
		if cursor+1 == len(path) {
			break
		}
		if _, exists := matrix.data[from]; exists {
			to := path[cursor+1]
			matrix.remove(from, to)
		}
	}
	matrix.rebuildUnions()
	return true
}

func (matrix *AdjacencyMatrix[Key]) remove(from, to Key) {
	delete(matrix.data[from], to)
	if !matrix.features.HasFeature(Directed) {
		delete(matrix.data[to], from)
	}
}

func (matrix *AdjacencyMatrix[Key]) Neighbors(v Key) []Key { return matrix.neighbors(v) }

func (matrix *AdjacencyMatrix[Key]) neighbors(v Key) []Key {
	if edges, ok := matrix.data[v]; ok {
		result := make([]Key, 0, len(edges))
		for k := range edges {
			result = append(result, k)
		}
		return result
	}
	return nil
}

func (matrix *AdjacencyMatrix[Key]) initUF() {
	if matrix.uf != nil {
		return
	}
	vertices := make([]Key, 0, len(matrix.data))
	for v := range matrix.data {
		vertices = append(vertices, v)
	}
	matrix.uf = newUnionFind(vertices)
	for _, v := range matrix.uf.vertices {
		for _, w := range matrix.neighbors(v) {
			matrix.uf.union(v, w)
		}
	}
}

func (matrix *AdjacencyMatrix[Key]) rebuildUnions() {
	if matrix.uf == nil {
		return
	}
	matrix.uf.reset()
	for _, v := range matrix.uf.vertices {
		for _, w := range matrix.neighbors(v) {
			matrix.uf.union(v, w)
		}
	}
}
