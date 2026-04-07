package graph

// unionFind implements disjoint-set union with path halving and union by rank.
// It owns the alive-vertex list so callers do not need a separate []Key field.
type unionFind[Key comparable] struct {
	parent   map[Key]Key
	rank     map[Key]int
	vertices []Key
}

func newUnionFind[Key comparable](vertices []Key) *unionFind[Key] {
	uf := &unionFind[Key]{
		parent:   make(map[Key]Key, len(vertices)),
		rank:     make(map[Key]int, len(vertices)),
		vertices: append([]Key(nil), vertices...),
	}
	for _, v := range uf.vertices {
		uf.parent[v] = v
	}
	return uf
}

// addVertex registers x as a singleton and appends it to the vertex list.
// No-op if already present.
func (uf *unionFind[Key]) addVertex(x Key) {
	if _, exists := uf.parent[x]; !exists {
		uf.parent[x] = x
		uf.vertices = append(uf.vertices, x)
	}
}

// removeVertex removes x from the vertex list and its union state.
func (uf *unionFind[Key]) removeVertex(x Key) {
	uf.vertices = removeFromSlice(uf.vertices, x)
	delete(uf.parent, x)
	delete(uf.rank, x)
}

// reset clears all union state and re-seeds the owned vertex list as singletons,
// ready for a fresh union pass.
func (uf *unionFind[Key]) reset() {
	clear(uf.parent)
	clear(uf.rank)
	for _, v := range uf.vertices {
		uf.parent[v] = v
	}
}

func (uf *unionFind[Key]) find(x Key) Key {
	for uf.parent[x] != x {
		uf.parent[x] = uf.parent[uf.parent[x]] // path halving
		x = uf.parent[x]
	}
	return x
}

// union merges the sets containing x and y. Returns false when they are
// already in the same set (adding this edge would create a cycle).
func (uf *unionFind[Key]) union(x, y Key) bool {
	rx, ry := uf.find(x), uf.find(y)
	if rx == ry {
		return false
	}
	if uf.rank[rx] < uf.rank[ry] {
		rx, ry = ry, rx
	}
	uf.parent[ry] = rx
	if uf.rank[rx] == uf.rank[ry] {
		uf.rank[rx]++
	}
	return true
}

// findCycleDirected detects the first cycle in a directed graph using DFS
// with white/gray/black colouring. Returns [start … start] or nil if acyclic.
func findCycleDirected[Key comparable](vertices []Key, neighbors func(Key) []Key) []Key {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[Key]int, len(vertices))
	var stack []Key

	var dfs func(Key) []Key
	dfs = func(v Key) []Key {
		color[v] = gray
		stack = append(stack, v)
		for _, w := range neighbors(v) {
			if color[w] == gray {
				for i, sv := range stack {
					if sv == w {
						cycle := make([]Key, len(stack)-i+1)
						copy(cycle, stack[i:])
						cycle[len(cycle)-1] = w
						return cycle
					}
				}
			}
			if color[w] == white {
				if cycle := dfs(w); cycle != nil {
					return cycle
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[v] = black
		return nil
	}

	for _, v := range vertices {
		if color[v] == white {
			if cycle := dfs(v); cycle != nil {
				return cycle
			}
		}
	}
	return nil
}

// findCycleUndirected detects the first cycle in an undirected graph using DFS.
// Returns [start … start] or nil if acyclic.
func findCycleUndirected[Key comparable](vertices []Key, neighbors func(Key) []Key) []Key {
	visited := make(map[Key]bool, len(vertices))
	parent := make(map[Key]Key, len(vertices))

	var dfs func(v Key, hasParent bool, par Key) []Key
	dfs = func(v Key, hasParent bool, par Key) []Key {
		visited[v] = true
		for _, w := range neighbors(v) {
			if !visited[w] {
				parent[w] = v
				if cycle := dfs(w, true, v); cycle != nil {
					return cycle
				}
			} else if hasParent && w != par {
				var path []Key
				for cur := v; cur != w; cur = parent[cur] {
					path = append(path, cur)
				}
				for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
					path[i], path[j] = path[j], path[i]
				}
				cycle := make([]Key, 0, len(path)+2)
				cycle = append(cycle, w)
				cycle = append(cycle, path...)
				cycle = append(cycle, w)
				return cycle
			}
		}
		return nil
	}

	for _, v := range vertices {
		if !visited[v] {
			if cycle := dfs(v, false, v); cycle != nil {
				return cycle
			}
		}
	}
	return nil
}

// isReachable reports whether to is reachable from from via iterative DFS.
func isReachable[Key comparable](from, to Key, neighbors func(Key) []Key) bool {
	visited := make(map[Key]bool)
	stack := []Key{from}
	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if v == to {
			return true
		}
		if visited[v] {
			continue
		}
		visited[v] = true
		stack = append(stack, neighbors(v)...)
	}
	return false
}
