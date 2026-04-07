package graph

import "container/heap"

// ITopologyTraversal is the callback interface passed to DFS, BFS, Dijkstra,
// and AStar. Returning false from any visiting method stops the traversal early.
type ITopologyTraversal[Key comparable] interface {
	// ShouldVisit is called before each edge; return false to skip it.
	ShouldVisit(from, to Key) bool
	// OnVisit is called when a vertex is first settled. Return false to stop.
	OnVisit(from, to Key) bool
	// OnCycle is called when a back-edge (cycle) is detected. Return false to stop.
	OnCycle(path []Key) bool
	// EdgeWeight returns the cost of the edge from→to (must be non-negative).
	EdgeWeight(from, to Key) Weight
	// Heuristic returns an admissible estimate of the cost from current to goal.
	// Return 0 to make AStar behave like Dijkstra.
	Heuristic(current, goal Key) Weight
	// OnRelax is called whenever a shorter path to a vertex is discovered.
	OnRelax(from, to Key, newCost Weight)
}

// BlankTopologyTraversal is a zero-value base type that satisfies
// ITopologyTraversal. Embed it in your own struct and override only the methods
// you need; unoverridden methods return their neutral zero values.
//
//   - ShouldVisit → false  (no edges visited unless overridden)
//   - OnVisit     → false  (stops traversal unless overridden)
//   - OnCycle     → false  (stops traversal unless overridden)
//   - EdgeWeight  → 0
//   - Heuristic   → 0
//   - OnRelax     → no-op
type BlankTopologyTraversal[Key comparable] struct{}

func (BlankTopologyTraversal[Key]) ShouldVisit(_, _ Key) bool  { return false }
func (BlankTopologyTraversal[Key]) OnVisit(_, _ Key) bool      { return false }
func (BlankTopologyTraversal[Key]) OnCycle(_ []Key) bool       { return false }
func (BlankTopologyTraversal[Key]) EdgeWeight(_, _ Key) Weight { return 0 }
func (BlankTopologyTraversal[Key]) Heuristic(_, _ Key) Weight  { return 0 }
func (BlankTopologyTraversal[Key]) OnRelax(_, _ Key, _ Weight) {}

// DFS traverses the graph depth-first from start. Back edges are reported via
// OnCycle; the parent edge is skipped on undirected graphs to suppress false
// cycle reports.
func DFS[Key comparable](topology ITopology[Key], start Key, traversal ITopologyTraversal[Key]) {
	visited := make(map[Key]bool)
	path := make([]Key, 0)
	inPath := make(map[Key]bool)

	var walk func(v, parent Key, hasParent bool) bool
	walk = func(v, parent Key, hasParent bool) bool {
		visited[v] = true
		path = append(path, v)
		inPath[v] = true

		for _, next := range topology.Neighbors(v) {
			if hasParent && next == parent {
				continue // skip parent edge in undirected graphs
			}
			if inPath[next] {
				i := 0
				for path[i] != next {
					i++
				}
				cycle := make([]Key, len(path)-i+1)
				copy(cycle, path[i:])
				cycle[len(cycle)-1] = next
				if !traversal.OnCycle(cycle) {
					return false
				}
				continue
			}
			if visited[next] {
				continue
			}
			if !traversal.ShouldVisit(v, next) {
				continue
			}
			if !traversal.OnVisit(v, next) {
				return false
			}
			if !walk(next, v, true) {
				return false
			}
		}

		path = path[:len(path)-1]
		inPath[v] = false
		return true
	}

	walk(start, start, false)
}

// BFS traverses the graph breadth-first from start. Already-visited edges are
// reported as OnCycle([from, to]).
func BFS[Key comparable](topology ITopology[Key], start Key, traversal ITopologyTraversal[Key]) {
	type edge struct{ from, to Key }

	visited := make(map[Key]bool)
	parent := make(map[Key]Key)
	visited[start] = true

	queue := make([]edge, 0)
	for _, next := range topology.Neighbors(start) {
		queue = append(queue, edge{start, next})
	}

	for len(queue) > 0 {
		e := queue[0]
		queue = queue[1:]

		if visited[e.to] {
			if p, ok := parent[e.from]; !ok || p != e.to {
				if !traversal.OnCycle([]Key{e.from, e.to}) {
					return
				}
			}
			continue
		}

		if !traversal.ShouldVisit(e.from, e.to) {
			continue
		}
		if !traversal.OnVisit(e.from, e.to) {
			return
		}

		visited[e.to] = true
		parent[e.to] = e.from
		for _, next := range topology.Neighbors(e.to) {
			queue = append(queue, edge{e.to, next})
		}
	}
}

// heapEntry is a (key, priority) pair for the min-heap used by shortest-path algorithms.
type heapEntry[Key comparable] struct {
	key      Key
	priority Weight
}

type minHeap[Key comparable] []heapEntry[Key]

func (h minHeap[Key]) Len() int           { return len(h) }
func (h minHeap[Key]) Less(i, j int) bool { return h[i].priority < h[j].priority }
func (h minHeap[Key]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap[Key]) Push(x any)        { *h = append(*h, x.(heapEntry[Key])) }
func (h *minHeap[Key]) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// shortestPath is the shared core of Dijkstra and AStar. Passing goal == nil
// settles all reachable vertices; a non-nil goal stops on first settlement.
func shortestPath[Key comparable](topology ITopology[Key], start Key, goal *Key, traversal ITopologyTraversal[Key]) {
	dist := make(map[Key]Weight)
	dist[start] = 0
	parent := make(map[Key]Key)
	hasParent := make(map[Key]bool)
	settled := make(map[Key]bool)

	var startPriority Weight
	if goal != nil {
		startPriority = traversal.Heuristic(start, *goal)
	}

	h := &minHeap[Key]{{start, startPriority}}
	heap.Init(h)

	for h.Len() > 0 {
		item := heap.Pop(h).(heapEntry[Key])
		v := item.key
		if settled[v] {
			continue
		}
		settled[v] = true

		if hasParent[v] {
			if !traversal.OnVisit(parent[v], v) {
				return
			}
		}

		if goal != nil && v == *goal {
			return
		}

		for _, next := range topology.Neighbors(v) {
			if settled[next] || !traversal.ShouldVisit(v, next) {
				continue
			}
			newDist := dist[v] + traversal.EdgeWeight(v, next)
			if d, ok := dist[next]; !ok || newDist < d {
				dist[next] = newDist
				parent[next] = v
				hasParent[next] = true
				traversal.OnRelax(v, next, newDist)
				priority := newDist
				if goal != nil {
					priority += traversal.Heuristic(next, *goal)
				}
				heap.Push(h, heapEntry[Key]{next, priority})
			}
		}
	}
}

// Dijkstra finds shortest paths from start to all reachable vertices using a
// lazy-deletion min-heap. EdgeWeight must return non-negative values.
func Dijkstra[Key comparable](topology ITopology[Key], start Key, traversal ITopologyTraversal[Key]) {
	shortestPath(topology, start, nil, traversal)
}

// AStar finds the shortest path from start to goal. Heuristic must be admissible;
// a zero heuristic degrades to Dijkstra.
func AStar[Key comparable](topology ITopology[Key], start, goal Key, traversal ITopologyTraversal[Key]) {
	shortestPath(topology, start, &goal, traversal)
}

// Path is an ordered sequence of vertex keys representing a graph path.
type Path[Key comparable] = []Key

// SearchMode selects the traversal strategy used by Paths.
type SearchMode uint8

const (
	DFSSearch SearchMode = iota // depth-first backtracking
	BFSSearch                   // breadth-first; shortest paths appear first
)

// Paths returns all simple paths from start. cycled includes cycle-closing
// paths (e.g. [A B C A]). exclude lists vertices to skip; mode selects
// DFSSearch or BFSSearch.
func Paths[Key comparable](topology ITopology[Key], start Key, cycled bool, mode SearchMode, exclude ...Key) []Path[Key] {
	excluded := make(map[Key]struct{}, len(exclude))
	for _, k := range exclude {
		excluded[k] = struct{}{}
	}
	if _, ok := excluded[start]; ok {
		return nil
	}
	if mode == BFSSearch {
		return pathsBFS(topology, start, cycled, excluded)
	}
	return pathsDFS(topology, start, cycled, excluded)
}

// pathsDFS collects all simple paths using recursive depth-first backtracking.
func pathsDFS[Key comparable](topology ITopology[Key], start Key, cycled bool, excluded map[Key]struct{}) []Path[Key] {
	var result []Path[Key]
	path := []Key{start}
	inPath := make(map[Key]struct{})
	inPath[start] = struct{}{}

	var walk func(v Key)
	walk = func(v Key) {
		hasUnvisited := false
		for _, next := range topology.Neighbors(v) {
			if _, ok := excluded[next]; ok {
				continue
			}
			if _, ok := inPath[next]; ok {
				if cycled {
					i := 0
					for path[i] != next {
						i++
					}
					p := make([]Key, len(path)-i+1)
					copy(p, path[i:])
					p[len(p)-1] = next
					result = append(result, p)
				}
				continue
			}
			hasUnvisited = true
			path = append(path, next)
			inPath[next] = struct{}{}
			walk(next)
			path = path[:len(path)-1]
			delete(inPath, next)
		}
		if !hasUnvisited && len(path) > 1 {
			p := make([]Key, len(path))
			copy(p, path)
			result = append(result, p)
		}
	}

	walk(start)
	return result
}

// pathState carries an in-progress path and its membership set for BFS.
type pathState[Key comparable] struct {
	path   []Key
	inPath map[Key]struct{}
}

// pathsBFS collects all simple paths using iterative breadth-first backtracking.
// Each queue item owns the full path from start to the current frontier vertex,
// so paths are extended and recorded in breadth (shortest-first) order.
func pathsBFS[Key comparable](topology ITopology[Key], start Key, cycled bool, excluded map[Key]struct{}) []Path[Key] {
	var result []Path[Key]

	initial := pathState[Key]{
		path:   []Key{start},
		inPath: map[Key]struct{}{start: {}},
	}
	queue := []pathState[Key]{initial}

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]

		v := state.path[len(state.path)-1]
		hasUnvisited := false

		for _, next := range topology.Neighbors(v) {
			if _, ok := excluded[next]; ok {
				continue
			}
			if _, ok := state.inPath[next]; ok {
				if cycled {
					i := 0
					for state.path[i] != next {
						i++
					}
					p := make([]Key, len(state.path)-i+1)
					copy(p, state.path[i:])
					p[len(p)-1] = next
					result = append(result, p)
				}
				continue
			}
			hasUnvisited = true
			newPath := make([]Key, len(state.path)+1)
			copy(newPath, state.path)
			newPath[len(newPath)-1] = next
			newInPath := make(map[Key]struct{}, len(state.inPath)+1)
			for k := range state.inPath {
				newInPath[k] = struct{}{}
			}
			newInPath[next] = struct{}{}
			queue = append(queue, pathState[Key]{path: newPath, inPath: newInPath})
		}

		if !hasUnvisited && len(state.path) > 1 {
			result = append(result, state.path)
		}
	}

	return result
}
