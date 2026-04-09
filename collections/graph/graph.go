package graph

// IGraph is a typed graph that pairs a topology with per-vertex and per-edge
// data stores. Vertex must implement IVertex[Key]; Key is inferred and does not
// need to be supplied explicitly to NewGraph.
type IGraph[Vertex IVertex[Key], Edge any, Key comparable] interface {
	// Add registers a vertex.
	Add(Vertex)
	// Contains reports whether the vertex is present.
	Contains(Vertex) bool
	// Delete removes a vertex and all its edges. Returns false if absent.
	Delete(Vertex) bool
	// Set creates an edge with the given data; delegates to the underlying topology.
	Set(Vertex, Vertex, Edge) bool
	// Has reports whether a vertex or chain of edges exists.
	Has(path ...Vertex) bool
	// Remove deletes a vertex or path of edges. Returns false if any step is absent.
	Remove(path ...Vertex) bool
	// IsCycled reports whether the graph contains a cycle.
	IsCycled() bool
	// Density returns the ratio of present edges to the maximum possible edges.
	Density() float64
	// Neighbors returns the keys of direct successors of v.
	Neighbors(Vertex) []Key
	// GetVertex retrieves the vertex stored under the given key.
	GetVertex(Key) (IVertex[Key], bool)
	// GetEdge retrieves the edge data between two vertices.
	GetEdge(Vertex, Vertex) (Edge, bool)
	// Paths returns all simple paths from start; see the package-level Paths function.
	Paths(Key, bool, SearchMode, ...Key) []Path[Key]
	// DFS runs a depth-first traversal from the given start key.
	DFS(Key, ITopologyTraversal[Key])
	// BFS runs a breadth-first traversal from the given start key.
	BFS(Key, ITopologyTraversal[Key])
	// Dijkstra finds shortest paths from start using non-negative edge weights.
	Dijkstra(Key, ITopologyTraversal[Key])
	// AStar finds the shortest path from start to goal using a heuristic.
	AStar(Key, Key, ITopologyTraversal[Key])
}

// NewGraph creates a Graph backed by the topology selected in params.
// The Key type parameter is inferred from Vertex and need not be supplied.
func NewGraph[Vertex IVertex[Key], Edge any, Key comparable](params ...TopologyParams) IGraph[Vertex, Edge, Key] {
	var p TopologyParams
	if len(params) > 0 {
		p = params[0]
	}
	return &Graph[Vertex, Edge, Key]{
		base: base[Vertex, Edge, Key]{
			topology: NewTopology[Key](params...),
			vertices: make(map[Key]Vertex, p.Scale),
			edges:    make(map[Key]map[Key]Edge, p.Scale),
		},
	}
}

type base[Vertex IVertex[Key], Edge any, Key comparable] struct {
	topology ITopology[Key]
	vertices map[Key]Vertex
	edges    map[Key]map[Key]Edge
}

type Graph[Vertex IVertex[Key], Edge any, Key comparable] struct {
	base[Vertex, Edge, Key]
}

func (graph *Graph[Vertex, Edge, Key]) Add(v Vertex) {
	graph.topology.Add(v.Key())
	graph.vertices[v.Key()] = v
}

func (graph *Graph[Vertex, Edge, Key]) Contains(v Vertex) bool {
	return graph.topology.Contains(v.Key())
}

func (graph *Graph[Vertex, Edge, Key]) Delete(v Vertex) bool {
	key := v.Key()
	if !graph.topology.Delete(key) {
		return false
	}
	delete(graph.vertices, key)
	delete(graph.edges, key)
	for _, row := range graph.edges {
		delete(row, key)
	}
	return true
}

func (graph *Graph[Vertex, Edge, Key]) Set(from, to Vertex, edge Edge) bool {
	if !graph.topology.Set(from.Key(), to.Key()) {
		return false
	}
	if graph.edges[from.Key()] == nil {
		graph.edges[from.Key()] = make(map[Key]Edge)
	}
	graph.edges[from.Key()][to.Key()] = edge
	return true
}

func (graph *Graph[Vertex, Edge, Key]) Has(path ...Vertex) bool {
	keys := make([]Key, len(path))
	for i, v := range path {
		keys[i] = v.Key()
	}
	return graph.topology.Has(keys...)
}

func (graph *Graph[Vertex, Edge, Key]) Remove(path ...Vertex) bool {
	keys := make([]Key, len(path))
	for i, v := range path {
		keys[i] = v.Key()
	}
	if !graph.topology.Remove(keys...) {
		return false
	}
	for i := 0; i+1 < len(keys); i++ {
		if row, ok := graph.edges[keys[i]]; ok {
			delete(row, keys[i+1])
		}
	}
	return true
}

func (graph *Graph[Vertex, Edge, Key]) IsCycled() bool {
	return graph.topology.IsCycled()
}

func (graph *Graph[Vertex, Edge, Key]) Density() float64 {
	return graph.topology.Density()
}

func (graph *Graph[Vertex, Edge, Key]) Neighbors(v Vertex) []Key {
	return graph.topology.Neighbors(v.Key())
}

func (graph *Graph[Vertex, Edge, Key]) GetVertex(key Key) (vertex IVertex[Key], ok bool) {
	vertex, ok = graph.vertices[key]
	return
}

func (graph *Graph[Vertex, Edge, Key]) GetEdge(from, to Vertex) (edge Edge, ok bool) {
	if _, ok = graph.edges[from.Key()]; ok {
		edge, ok = graph.edges[from.Key()][to.Key()]
	}
	return
}

func (graph *Graph[Vertex, Edge, Key]) Paths(start Key, cycled bool, mode SearchMode, exclude ...Key) []Path[Key] {
	return Paths(graph.topology, start, cycled, mode, exclude...)
}

func (graph *Graph[Vertex, Edge, Key]) DFS(start Key, traversal ITopologyTraversal[Key]) {
	DFS(graph.topology, start, traversal)
}

func (graph *Graph[Vertex, Edge, Key]) BFS(start Key, traversal ITopologyTraversal[Key]) {
	BFS(graph.topology, start, traversal)
}

func (graph *Graph[Vertex, Edge, Key]) Dijkstra(start Key, traversal ITopologyTraversal[Key]) {
	Dijkstra(graph.topology, start, traversal)
}

func (graph *Graph[Vertex, Edge, Key]) AStar(start, goal Key, traversal ITopologyTraversal[Key]) {
	AStar(graph.topology, start, goal, traversal)
}
