package ivyutils

// TopologicalSort returns items in an order respecting the given partial order.
// order is a list of (before, after) pairs. key maps items to comparable keys.
func TopologicalSort[T any, K comparable](items []T, order [][2]T, key func(T) K) []T {
	// Build adjacency map: key -> list of successors
	adj := make(map[K][]T)
	for _, pair := range order {
		k := key(pair[0])
		adj[k] = append(adj[k], pair[1])
	}

	var result []T
	done := make(map[K]struct{})
	stack := make([]T, len(items))
	copy(stack, items)

	for len(stack) > 0 {
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		k := key(item)
		if _, ok := done[k]; ok {
			continue
		}
		succs, hasSuccs := adj[k]
		if hasSuccs {
			stack = append(stack, item)
			for _, s := range succs {
				sk := key(s)
				if sk != k {
					stack = append(stack, s)
				}
			}
			delete(adj, k)
		} else {
			done[k] = struct{}{}
			result = append(result, item)
		}
	}

	// Reverse
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// Reachable returns descendants of items following successors.
func Reachable[T any, K comparable](items []T, iterSucc func(K) []T, key func(T) K) []T {
	var result []T
	visited := make(map[K]struct{})
	expanded := make(map[K]struct{})
	stack := make([]T, len(items))
	copy(stack, items)

	for len(stack) > 0 {
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		k := key(item)
		if _, ok := visited[k]; ok {
			continue
		}
		if _, ok := expanded[k]; !ok {
			stack = append(stack, item)
			for _, s := range iterSucc(k) {
				sk := key(s)
				if sk != k {
					stack = append(stack, s)
				}
			}
			expanded[k] = struct{}{}
		} else {
			visited[k] = struct{}{}
			result = append(result, item)
		}
	}

	// Reverse
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// Tarjan computes the strongly connected components of a directed graph
// using Tarjan's algorithm. Returns SCCs in reverse topological order.
// Matches Python's tarjan(graph) from the tarjan package.
func Tarjan(graph map[string]map[string]bool) [][]string {
	var (
		index    int
		stack    []string
		onStack  = map[string]bool{}
		indices  = map[string]int{}
		lowlinks = map[string]int{}
		result   [][]string
	)

	var strongConnect func(v string)
	strongConnect = func(v string) {
		indices[v] = index
		lowlinks[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		if succs, ok := graph[v]; ok {
			for w := range succs {
				if _, visited := indices[w]; !visited {
					strongConnect(w)
					if lowlinks[w] < lowlinks[v] {
						lowlinks[v] = lowlinks[w]
					}
				} else if onStack[w] {
					if indices[w] < lowlinks[v] {
						lowlinks[v] = indices[w]
					}
				}
			}
		}

		if lowlinks[v] == indices[v] {
			var scc []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			result = append(result, scc)
		}
	}

	// Collect all vertices (both keys and successors)
	vertices := map[string]bool{}
	for v, succs := range graph {
		vertices[v] = true
		for w := range succs {
			vertices[w] = true
		}
	}

	// Process in sorted order for determinism
	sorted := make([]string, 0, len(vertices))
	for v := range vertices {
		sorted = append(sorted, v)
	}
	// Simple sort for determinism
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	for _, v := range sorted {
		if _, visited := indices[v]; !visited {
			strongConnect(v)
		}
	}

	return result
}

// Arc represents a directed edge.
type Arc[K comparable] struct {
	From, To K
}

// FindCycle finds a cycle in a directed graph or returns nil.
// arcs is a list of directed edges.
func FindCycle[K comparable](arcs []Arc[K]) []Arc[K] {
	adj := make(map[K][]Arc[K])
	for _, a := range arcs {
		adj[a.From] = append(adj[a.From], a)
	}

	heap := make(map[K]struct{})  // permanently visited
	stack := make(map[K]struct{}) // currently on DFS stack
	var path []Arc[K]

	var dfs func(node K) bool
	dfs = func(node K) bool {
		if _, ok := heap[node]; ok {
			return false
		}
		if _, ok := stack[node]; ok {
			return true
		}
		stack[node] = struct{}{}
		// Match Python defaultdict auto-vivification
		if _, ok := adj[node]; !ok {
			adj[node] = nil
		}
		for _, arc := range adj[node] {
			if dfs(arc.To) {
				path = append(path, arc)
				return true
			}
		}
		delete(stack, node)
		heap[node] = struct{}{}
		return false
	}

	for _, a := range arcs {
		if dfs(a.From) {
			if len(path) == 0 {
				return nil
			}
			// Extract the actual cycle from path
			end := path[0].To
			var cycle []Arc[K]
			for _, arc := range path {
				cycle = append(cycle, arc)
				if arc.From == end {
					// Reverse cycle
					for i, j := 0, len(cycle)-1; i < j; i, j = i+1, j-1 {
						cycle[i], cycle[j] = cycle[j], cycle[i]
					}
					return cycle
				}
			}
			return path
		}
	}
	return nil
}
