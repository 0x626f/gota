package graph

import (
	"math/bits"

	"github.com/0x626f/gota/collections"
)

// BitMatrix stores edges as compact bit-rows indexed by integer slot.
// Neighbour iteration is cache-friendly; edge lookup and insert are O(1).
// Vertex deletion is O(V) due to bit-clearing across all rows.
type BitMatrix[Key comparable] struct {
	baseTopology[Key]

	indexer *collections.Indexer
	indexes map[Key]int
	keys    []Key // slot-indexed; zero value at released slots
	linkage []bitRow
}

// NewBitMatrix creates a BitMatrix with the given options.
func NewBitMatrix[Key comparable](params ...TopologyParams) *BitMatrix[Key] {
	var p TopologyParams
	if len(params) > 0 {
		p = params[0]
	}
	m := &BitMatrix[Key]{
		indexer: collections.NewIndexer(),
		indexes: make(map[Key]int, p.Scale),
	}
	if p.Scale > 0 {
		m.keys = make([]Key, p.Scale)
		m.linkage = make([]bitRow, p.Scale)
	}
	for _, f := range p.Features {
		m.features.SetFeature(f)
	}
	return m
}

func (matrix *BitMatrix[Key]) Add(vertex Key) {
	if _, exists := matrix.indexes[vertex]; exists {
		return
	}
	index := matrix.indexer.Next()
	matrix.indexes[vertex] = index
	matrix.keys = growTo(matrix.keys, index)
	matrix.keys[index] = vertex
	matrix.linkage = growTo(matrix.linkage, index)
	matrix.linkage[index] = make(bitRow, 0)
	if matrix.uf != nil {
		matrix.uf.addVertex(vertex)
	}
}

func (matrix *BitMatrix[Key]) Contains(vertex Key) (result bool) {
	_, result = matrix.indexes[vertex]
	return
}

func (matrix *BitMatrix[Key]) Delete(vertex Key) bool {
	if !matrix.Contains(vertex) {
		return false
	}
	index := matrix.indexes[vertex]
	matrix.linkage[index] = nil
	for i, row := range matrix.linkage {
		if row != nil {
			matrix.clearBit(i, index)
		}
	}
	delete(matrix.indexes, vertex)
	var zeroKey Key
	matrix.keys[index] = zeroKey
	matrix.indexer.Release(index)
	if matrix.uf != nil {
		matrix.uf.removeVertex(vertex)
	}
	matrix.rebuildUnions()
	return true
}

func (matrix *BitMatrix[Key]) Set(vertex0, vertex1 Key) bool {
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
	matrix.Add(vertex1)
	index0 := matrix.indexes[vertex0]
	index1 := matrix.indexes[vertex1]
	matrix.setBit(index0, index1)
	if !matrix.features.HasFeature(Directed) {
		matrix.setBit(index1, index0)
	}
	if matrix.features.HasFeature(Acyclic) && !matrix.features.HasFeature(Directed) {
		matrix.uf.addVertex(vertex0)
		matrix.uf.addVertex(vertex1)
		matrix.uf.union(vertex0, vertex1)
	}
	return true
}

func (matrix *BitMatrix[Key]) Has(path ...Key) bool {
	if len(path) == 0 {
		return false
	}

	if len(path) == 1 {
		return matrix.Contains(path[0])
	}

	fromIndex, exists := matrix.indexes[path[0]]
	if !exists {
		return false
	}

	for i := 1; i < len(path); i++ {
		toIndex, ok := matrix.indexes[path[i]]
		if !ok || !matrix.getBit(fromIndex, toIndex) {
			return false
		}
		fromIndex = toIndex
	}

	return true
}

func (matrix *BitMatrix[Key]) Remove(path ...Key) bool {
	if len(path) == 0 {
		return false
	}

	if len(path) == 1 {
		return matrix.Delete(path[0])
	}

	for i := 0; i+1 < len(path); i++ {
		if !matrix.Has(path[i], path[i+1]) {
			return false
		}
	}

	for i := 0; i+1 < len(path); i++ {
		matrix.remove(path[i], path[i+1])
	}

	matrix.rebuildUnions()
	return true
}

func (matrix *BitMatrix[Key]) IsCycled() bool {
	vertices := make([]Key, 0, len(matrix.indexes))
	for v := range matrix.indexes {
		vertices = append(vertices, v)
	}
	if matrix.features.HasFeature(Directed) {
		return findCycleDirected(vertices, matrix.neighbors) != nil
	}
	return findCycleUndirected(vertices, matrix.neighbors) != nil
}

func (matrix *BitMatrix[Key]) Neighbors(v Key) []Key { return matrix.neighbors(v) }

func (matrix *BitMatrix[Key]) neighbors(v Key) []Key {
	idx, ok := matrix.indexes[v]
	if !ok || idx >= len(matrix.linkage) || matrix.linkage[idx] == nil {
		return nil
	}
	row := matrix.linkage[idx]
	// Count set bits first so we can pre-allocate with exact capacity,
	// avoiding repeated reallocations from appending to a nil slice.
	n := 0
	for _, b := range row {
		n += bits.OnesCount8(b)
	}
	if n == 0 {
		return nil
	}
	result := make([]Key, 0, n)
	for byteIdx, b := range row {
		for bit := 0; bit < 8; bit++ {
			if b&(1<<bit) != 0 {
				toIdx := byteIdx*8 + bit
				if toIdx < len(matrix.keys) {
					result = append(result, matrix.keys[toIdx])
				}
			}
		}
	}
	return result
}

func (matrix *BitMatrix[Key]) initUF() {
	if matrix.uf != nil {
		return
	}
	vertices := make([]Key, 0, len(matrix.indexes))
	for v := range matrix.indexes {
		vertices = append(vertices, v)
	}
	matrix.uf = newUnionFind(vertices)
	for _, v := range matrix.uf.vertices {
		for _, w := range matrix.neighbors(v) {
			matrix.uf.union(v, w)
		}
	}
}

func (matrix *BitMatrix[Key]) rebuildUnions() {
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

func (matrix *BitMatrix[Key]) remove(from, to Key) {
	fromIndex := matrix.indexes[from]
	toIndex := matrix.indexes[to]
	matrix.clearBit(fromIndex, toIndex)
	if !matrix.features.HasFeature(Directed) {
		matrix.clearBit(toIndex, fromIndex)
	}
}

func (matrix *BitMatrix[Key]) setBit(from, to int) {
	byteIndex := to / 8
	if byteIndex >= len(matrix.linkage[from]) {
		matrix.linkage[from] = append(matrix.linkage[from], make(bitRow, byteIndex-len(matrix.linkage[from])+1)...)
	}
	matrix.linkage[from][byteIndex] |= 1 << (to % 8)
}

func (matrix *BitMatrix[Key]) clearBit(from, to int) {
	byteIndex := to / 8
	if byteIndex >= len(matrix.linkage[from]) {
		return
	}
	matrix.linkage[from][byteIndex] &^= 1 << (to % 8)
}

func (matrix *BitMatrix[Key]) getBit(from, to int) bool {
	byteIndex := to / 8
	if from >= len(matrix.linkage) || matrix.linkage[from] == nil || byteIndex >= len(matrix.linkage[from]) {
		return false
	}
	return matrix.linkage[from][byteIndex]&(1<<(to%8)) != 0
}
