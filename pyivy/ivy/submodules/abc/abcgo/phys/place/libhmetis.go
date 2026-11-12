package place

import (
	"fmt"
	"math/rand"
	"sort"
)

// hypergraphPartition is a standard-library replacement for proprietary hMetis.
// It optimizes weighted cut nets under a per-part weight cap using greedy
// construction, positive-gain moves and pair swaps. Small cases are searched
// exactly. It does not reproduce hMetis's coarsening or seed-dependent answers.
const (
	nativeCutStarts        = 4
	nativeCutPasses        = 10
	nativeCutSwapLimit     = 256
	nativeCutExactStates   = 1 << 20
	nativeCutExactEdgeWork = 40_000_000
)

type hypergraphPartition struct {
	edges, adj           [][]int
	weights, edgeWeights []int64
	k                    int
	cap                  int64
	fixed                []bool
	seed                 []int
}

func readHypergraph(nv, ne int32, vw, eptr, eind, ew []int32, k, imbalance int32, options, initial []int32) (*hypergraphPartition, error) {
	if nv < 0 || ne < 0 || k <= 0 || (nv > 0 && k > nv) || imbalance < 0 || len(eptr) < int(ne)+1 || len(initial) < int(nv) {
		return nil, fmt.Errorf("invalid hypergraph dimensions")
	}
	if eptr[0] != 0 || eptr[ne] < 0 || int(eptr[ne]) > len(eind) {
		return nil, fmt.Errorf("invalid hyperedge offsets")
	}
	if vw != nil && len(vw) < int(nv) || ew != nil && len(ew) < int(ne) {
		return nil, fmt.Errorf("short hypergraph weight array")
	}
	g := &hypergraphPartition{edges: make([][]int, ne), adj: make([][]int, nv), weights: make([]int64, nv), edgeWeights: make([]int64, ne), k: int(k), fixed: make([]bool, nv), seed: make([]int, nv)}
	var total, largest int64
	for v := range g.weights {
		w := int64(1)
		if vw != nil {
			w = int64(vw[v])
		}
		if w < 0 {
			return nil, fmt.Errorf("negative vertex weight")
		}
		g.weights[v] = w
		total += w
		largest = max(largest, w)
		if initial[v] < -1 || initial[v] >= k {
			return nil, fmt.Errorf("initial partition outside range")
		}
		g.seed[v] = int(initial[v])
		if len(options) > 6 && options[6] != 0 && initial[v] >= 0 {
			if initial[v] >= k {
				return nil, fmt.Errorf("fixed partition outside range")
			}
			g.fixed[v] = true
		}
	}
	if total == 0 {
		for v := range g.weights {
			g.weights[v] = 1
		}
		total = int64(nv)
		largest = 1
	}
	if total > 0 && int64(100+int64(imbalance)) > (1<<63-1)/total {
		return nil, fmt.Errorf("partition weight bound overflow")
	}
	denominator := int64(k) * 100
	numerator := total * (100 + int64(imbalance))
	g.cap = numerator / denominator
	if numerator%denominator != 0 {
		g.cap++
	}
	g.cap = max(g.cap, largest)
	for e := range g.edges {
		if eptr[e] < 0 || eptr[e] > eptr[e+1] || int(eptr[e+1]) > len(eind) {
			return nil, fmt.Errorf("nonmonotonic hyperedge offsets")
		}
		seen := map[int]bool{}
		for _, id := range eind[eptr[e]:eptr[e+1]] {
			if id < 0 || id >= nv {
				return nil, fmt.Errorf("hyperedge vertex outside range")
			}
			v := int(id)
			if !seen[v] {
				seen[v] = true
				g.edges[e] = append(g.edges[e], v)
				g.adj[v] = append(g.adj[v], e)
			}
		}
		w := int64(1)
		if ew != nil {
			w = int64(ew[e])
		}
		if w < 0 {
			return nil, fmt.Errorf("negative hyperedge weight")
		}
		g.edgeWeights[e] = w
	}
	return g, nil
}
func (g *hypergraphPartition) edgeCost(part []int, e int) int64 {
	first := -1
	for _, v := range g.edges[e] {
		if part[v] < 0 {
			continue
		}
		if first < 0 {
			first = part[v]
		} else if part[v] != first {
			return g.edgeWeights[e]
		}
	}
	return 0
}
func (g *hypergraphPartition) cost(part []int) int64 {
	var cost int64
	for e := range g.edges {
		cost += g.edgeCost(part, e)
	}
	return cost
}
func (g *hypergraphPartition) incidentCost(part []int, edges []int) int64 {
	var cost int64
	for _, e := range edges {
		cost += g.edgeCost(part, e)
	}
	return cost
}
func (g *hypergraphPartition) sharedIncidentEdges(a, b int) []int {
	result := append([]int(nil), g.adj[a]...)
	seen := make(map[int]bool, len(result))
	for _, e := range result {
		seen[e] = true
	}
	for _, e := range g.adj[b] {
		if !seen[e] {
			result = append(result, e)
		}
	}
	return result
}
func (g *hypergraphPartition) loads(part []int) []int64 {
	loads := make([]int64, g.k)
	for v, p := range part {
		if p >= 0 {
			loads[p] += g.weights[v]
		}
	}
	return loads
}
func (g *hypergraphPartition) feasible(loads []int64) bool {
	for _, w := range loads {
		if w > g.cap {
			return false
		}
	}
	return true
}
func (g *hypergraphPartition) construct(order []int) []int {
	part := make([]int, len(g.weights))
	for v := range part {
		part[v] = -1
		if g.fixed[v] {
			part[v] = g.seed[v]
		}
	}
	loads := g.loads(part)
	if !g.feasible(loads) {
		return nil
	}
	for _, v := range order {
		if g.fixed[v] {
			continue
		}
		best := -1
		var cost int64
		for p := 0; p < g.k; p++ {
			if loads[p]+g.weights[v] > g.cap {
				continue
			}
			part[v] = p
			c := g.incidentCost(part, g.adj[v])
			if best < 0 || c < cost || c == cost && (loads[p] < loads[best] || loads[p] == loads[best] && p == g.seed[v]) {
				best, cost = p, c
			}
		}
		if best < 0 {
			return nil
		}
		part[v] = best
		loads[best] += g.weights[v]
	}
	return part
}
func (g *hypergraphPartition) refine(part []int) {
	loads := g.loads(part)
	for pass := 0; pass < nativeCutPasses; pass++ {
		before := g.cost(part)
		if before == 0 {
			return
		}
		changed := false
		for v, from := range part {
			if g.fixed[v] {
				continue
			}
			best, bestCost := from, g.incidentCost(part, g.adj[v])
			for to := 0; to < g.k; to++ {
				if to == from || loads[to]+g.weights[v] > g.cap {
					continue
				}
				part[v] = to
				c := g.incidentCost(part, g.adj[v])
				if c < bestCost {
					best, bestCost = to, c
				}
			}
			part[v] = best
			if best != from {
				loads[from] -= g.weights[v]
				loads[best] += g.weights[v]
				changed = true
			}
		}
		if len(part) <= nativeCutSwapLimit {
			for a := 0; a < len(part); a++ {
				if g.fixed[a] {
					continue
				}
				for b := a + 1; b < len(part); b++ {
					pa, pb := part[a], part[b]
					if pa == pb || g.fixed[b] || loads[pa]-g.weights[a]+g.weights[b] > g.cap || loads[pb]-g.weights[b]+g.weights[a] > g.cap {
						continue
					}
					incident := g.sharedIncidentEdges(a, b)
					old := g.incidentCost(part, incident)
					part[a], part[b] = pb, pa
					if g.incidentCost(part, incident) < old {
						loads[pa] += g.weights[b] - g.weights[a]
						loads[pb] += g.weights[a] - g.weights[b]
						changed = true
					} else {
						part[a], part[b] = pa, pb
					}
				}
			}
		}
		if !changed || g.cost(part) >= before {
			return
		}
	}
}
func (g *hypergraphPartition) solve(options []int32) ([]int, error) {
	if len(g.weights) == 0 {
		return []int{}, nil
	}
	order := make([]int, len(g.weights))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := order[i], order[j]
		if g.weights[a] != g.weights[b] {
			return g.weights[a] > g.weights[b]
		}
		return len(g.adj[a]) > len(g.adj[b])
	})
	seed := int64(12261980)
	if len(options) > 7 {
		seed = int64(options[7])
	}
	rng := rand.New(rand.NewSource(seed))
	var best []int
	bestCost := int64(1<<63 - 1)
	for trial := 0; trial < nativeCutStarts; trial++ {
		if trial > 0 {
			rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		}
		part := g.construct(order)
		if part == nil {
			continue
		}
		g.refine(part)
		cost := g.cost(part)
		if cost < bestCost {
			best, bestCost = append([]int(nil), part...), cost
		}
		if bestCost == 0 {
			break
		}
	}
	// Exact branch-and-bound for at most one million assignments. The partial
	// cut cost is a lower bound because adding vertices cannot uncut a net.
	states := int64(1)
	var free []int
	part := make([]int, len(g.weights))
	for v := range part {
		part[v] = -1
		if g.fixed[v] {
			part[v] = g.seed[v]
		} else {
			free = append(free, v)
			if states <= nativeCutExactStates {
				states *= int64(g.k)
			}
		}
	}
	if states <= nativeCutExactStates && states*int64(len(g.edges)) <= nativeCutExactEdgeWork {
		loads := g.loads(part)
		var search func(int)
		search = func(i int) {
			if !g.feasible(loads) || g.cost(part) >= bestCost {
				return
			}
			if i == len(free) {
				best, bestCost = append([]int(nil), part...), g.cost(part)
				return
			}
			v := free[i]
			for p := 0; p < g.k; p++ {
				if loads[p]+g.weights[v] > g.cap {
					continue
				}
				part[v] = p
				loads[p] += g.weights[v]
				search(i + 1)
				loads[p] -= g.weights[v]
			}
			part[v] = -1
		}
		search(0)
	}
	if best == nil {
		return nil, fmt.Errorf("no balanced hypergraph partition found under weight cap %d", g.cap)
	}
	return best, nil
}

// HMETIS_PartRecursive partitions a compressed hyperedge list. eptr has ne+1
// offsets into eind; duplicate pins are ignored. Nil weights mean unit weights.
// part supplies initial labels (-1 means unassigned). options[6] locks supplied
// nonnegative labels, and options[7] selects a deterministic shuffle seed. Other
// proprietary tuning options do not apply to this native algorithm. imbalance
// is a percentage above average allowed part weight; one large vertex always
// fits. On error, part and cut are unchanged.
func HMETIS_PartRecursive(nv, ne int32, vw, eptr, eind, ew []int32, k, imbalance int32, options, part []int32, cut *int32) error {
	g, err := readHypergraph(nv, ne, vw, eptr, eind, ew, k, imbalance, options, part)
	if err != nil {
		return err
	}
	result, err := g.solve(options)
	if err != nil {
		return err
	}
	cost := g.cost(result)
	if cost > 1<<31-1 {
		return fmt.Errorf("partition cut cost exceeds int32")
	}
	for v, p := range result {
		part[v] = int32(p)
	}
	if cut != nil {
		*cut = int32(cost)
	}
	return nil
}

// HMETIS_PartKway uses the same native direct-k-way optimizer; the two legacy
// entry points differ in hMetis's implementation, not this replacement's search.
func HMETIS_PartKway(nv, ne int32, vw, eptr, eind, ew []int32, k, imbalance int32, options, part []int32, cut *int32) error {
	return HMETIS_PartRecursive(nv, ne, vw, eptr, eind, ew, k, imbalance, options, part, cut)
}
