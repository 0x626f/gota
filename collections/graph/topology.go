package graph

// TopologyKey selects the underlying storage implementation for NewTopology.
type TopologyKey uint8

const (
	AdjacencyMatrixTopology TopologyKey = iota // map-of-maps; O(1) edge lookup
	BitMatrixTopology                          // packed bit-rows; cache-friendly traversal
	CRSTopology                                // sorted adjacency lists; compact memory
)

// String returns a human-readable name of the topology implementation.
func (key TopologyKey) String() string {
	switch key {
	case BitMatrixTopology:
		return "Bit Matrix"
	case CRSTopology:
		return "CSR"
	case AdjacencyMatrixTopology:
		return "Adjacency Matrix"
	default:
		return "Unknown"
	}
}

// NewTopology creates an ITopology backed by the implementation named in
// params.Key. Returns nil when params.Key is unrecognised.
func NewTopology[Key comparable](params ...TopologyParams) ITopology[Key] {
	var p TopologyParams
	if len(params) > 0 {
		p = params[0]
	}
	switch p.Key {
	case AdjacencyMatrixTopology:
		return NewAdjacencyMatrix[Key](params...)
	case BitMatrixTopology:
		return NewBitMatrix[Key](params...)
	case CRSTopology:
		return NewCSR[Key](params...)
	default:
		return nil
	}
}

// ITopology is the common interface satisfied by all graph storage backends.
type ITopology[Key comparable] interface {
	// Add registers a vertex. No-op if the vertex already exists.
	Add(Key)
	// Contains reports whether the vertex is present.
	Contains(Key) bool
	// Delete removes a vertex and all edges incident to it. Returns false if absent.
	Delete(Key) bool
	// Set creates an edge from vertex0 to vertex1, adding both vertices as needed.
	// Returns false when the edge is rejected (e.g. would create a cycle in an
	// Acyclic topology, or the reverse already exists in a Directed one).
	Set(Key, Key) bool
	// Has reports whether a vertex (single arg) or a chain of edges (multiple args) exists.
	Has(path ...Key) bool
	// Remove deletes a vertex (single arg) or all edges along a path (multiple args).
	// Returns false if any vertex or edge in the path is absent.
	Remove(path ...Key) bool
	// IsCycled reports whether the graph contains at least one cycle.
	IsCycled() bool
	// Neighbors returns the direct successors of v.
	Neighbors(Key) []Key
}

// TopologyParams configures a topology at construction time.
type TopologyParams struct {
	Key      TopologyKey // storage backend (default: AdjacencyMatrixTopology)
	Features []Feature   // Directed, Acyclic, etc.
	Scale    int         // expected vertex count; pre-allocates internal maps/slices
}

// baseTopology holds shared state embedded by all topology implementations.
type baseTopology[Key comparable] struct {
	features FeatureStorage
	uf       *unionFind[Key] // non-nil only when Acyclic + undirected is first used
}
