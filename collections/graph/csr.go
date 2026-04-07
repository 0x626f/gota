package graph

import "github.com/0x626f/gota/collections"

// CSR (Compressed Sparse Row) stores each vertex's neighbours as a sorted
// []int slice. Edge lookup is O(log degree) via binary search; insert and
// delete are O(degree) due to in-slice shifting. More memory-compact than
// AdjacencyMatrix and faster on neighbour iteration than BitMatrix for sparse
// graphs.
type CSR[Key comparable] struct {
	baseTopology[Key]

	indexer *collections.Indexer
	indexes map[Key]int
	keys    []Key   // slot-indexed; zero value at released slots
	rows    [][]int // rows[i] = sorted neighbour indices for vertex i
}

// NewCSR creates a CSR topology with the given options.
func NewCSR[Key comparable](params ...TopologyParams) *CSR[Key] {
	var p TopologyParams
	if len(params) > 0 {
		p = params[0]
	}
	c := &CSR[Key]{
		indexer: collections.NewIndexer(),
		indexes: make(map[Key]int, p.Scale),
	}
	if p.Scale > 0 {
		c.keys = make([]Key, p.Scale)
		c.rows = make([][]int, p.Scale)
	}
	for _, f := range p.Features {
		c.features.SetFeature(f)
	}
	return c
}

func (csr *CSR[Key]) Add(vertex Key) {
	if _, exists := csr.indexes[vertex]; exists {
		return
	}
	index := csr.indexer.Next()
	csr.indexes[vertex] = index
	csr.keys = growTo(csr.keys, index)
	csr.keys[index] = vertex
	csr.rows = growTo(csr.rows, index)
	csr.rows[index] = nil
	if csr.uf != nil {
		csr.uf.addVertex(vertex)
	}
}

func (csr *CSR[Key]) Contains(vertex Key) (result bool) {
	_, result = csr.indexes[vertex]
	return
}

func (csr *CSR[Key]) Delete(vertex Key) bool {
	if !csr.Contains(vertex) {
		return false
	}
	index := csr.indexes[vertex]
	csr.rows[index] = nil

	for i, row := range csr.rows {
		if row != nil {
			csr.rows[i] = removeSorted(row, index)
		}
	}

	delete(csr.indexes, vertex)
	var zeroKey Key
	csr.keys[index] = zeroKey

	csr.indexer.Release(index)
	if csr.uf != nil {
		csr.uf.removeVertex(vertex)
	}
	csr.rebuildUnions()
	return true
}

func (csr *CSR[Key]) Set(vertex0, vertex1 Key) bool {
	if csr.features.HasFeature(Directed) && csr.Has(vertex1, vertex0) {
		return false
	}

	if csr.features.HasFeature(Acyclic) {
		if csr.features.HasFeature(Directed) {
			if isReachable(vertex1, vertex0, csr.neighbors) {
				return false
			}
		} else {
			csr.initUF()
			if csr.Contains(vertex0) && csr.Contains(vertex1) {
				if csr.uf.find(vertex0) == csr.uf.find(vertex1) {
					return false
				}
			}
		}
	}
	csr.Add(vertex0)
	csr.Add(vertex1)
	index0 := csr.indexes[vertex0]
	index1 := csr.indexes[vertex1]
	insertSorted(&csr.rows[index0], index1)

	if !csr.features.HasFeature(Directed) {
		insertSorted(&csr.rows[index1], index0)
	}

	if csr.features.HasFeature(Acyclic) && !csr.features.HasFeature(Directed) {
		csr.uf.addVertex(vertex0)
		csr.uf.addVertex(vertex1)
		csr.uf.union(vertex0, vertex1)
	}

	return true
}

func (csr *CSR[Key]) Has(path ...Key) bool {
	if len(path) == 0 {
		return false
	}

	if len(path) == 1 {
		return csr.Contains(path[0])
	}

	fromIndex, exists := csr.indexes[path[0]]
	if !exists {
		return false
	}

	for i := 1; i < len(path); i++ {
		toIndex, ok := csr.indexes[path[i]]
		if !ok || !searchSorted(csr.rows[fromIndex], toIndex) {
			return false
		}
		fromIndex = toIndex
	}

	return true
}

func (csr *CSR[Key]) Remove(path ...Key) bool {
	if len(path) == 0 {
		return false
	}

	if len(path) == 1 {
		return csr.Delete(path[0])
	}

	for i := 0; i+1 < len(path); i++ {
		if !csr.Has(path[i], path[i+1]) {
			return false
		}
	}

	for i := 0; i+1 < len(path); i++ {
		csr.remove(path[i], path[i+1])
	}

	csr.rebuildUnions()
	return true
}

func (csr *CSR[Key]) IsCycled() bool {
	vertices := make([]Key, 0, len(csr.indexes))
	for v := range csr.indexes {
		vertices = append(vertices, v)
	}
	if csr.features.HasFeature(Directed) {
		return findCycleDirected(vertices, csr.neighbors) != nil
	}
	return findCycleUndirected(vertices, csr.neighbors) != nil
}

func (csr *CSR[Key]) Neighbors(v Key) []Key { return csr.neighbors(v) }

func (csr *CSR[Key]) neighbors(v Key) []Key {
	index, ok := csr.indexes[v]
	if !ok {
		return nil
	}

	result := make([]Key, 0, len(csr.rows[index]))
	for _, toIndex := range csr.rows[index] {
		result = append(result, csr.keys[toIndex])
	}

	return result
}

func (csr *CSR[Key]) initUF() {
	if csr.uf != nil {
		return
	}

	vertices := make([]Key, 0, len(csr.indexes))
	for v := range csr.indexes {
		vertices = append(vertices, v)
	}
	csr.uf = newUnionFind(vertices)
	for _, v := range csr.uf.vertices {
		for _, w := range csr.neighbors(v) {
			csr.uf.union(v, w)
		}
	}
}

func (csr *CSR[Key]) rebuildUnions() {
	if csr.uf == nil {
		return
	}

	csr.uf.reset()
	for _, v := range csr.uf.vertices {
		for _, w := range csr.neighbors(v) {
			csr.uf.union(v, w)
		}
	}
}

func (csr *CSR[Key]) remove(from, to Key) {
	fromIndex := csr.indexes[from]
	toIndex := csr.indexes[to]
	csr.rows[fromIndex] = removeSorted(csr.rows[fromIndex], toIndex)
	if !csr.features.HasFeature(Directed) {
		csr.rows[toIndex] = removeSorted(csr.rows[toIndex], fromIndex)
	}
}
