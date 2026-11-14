package core

import (
	"math"
	"sort"
)

func averageDistanceBucket(f, g uint64, slots uint32) uint32 {
	if slots == 0 {
		slots = CUDD_UNIQUE_SLOTS
	}
	if slots&(slots-1) == 0 {
		shift := uint32(32) - uint32(bitsForAverageDistance(slots))
		h := (uint32(f)*12582917 + uint32(g)) * 4256249
		return h >> shift
	}
	return uint32((uint64(uint32(f))*12582917 + uint64(uint32(g))) * 4256249 % uint64(slots))
}

func bitsForAverageDistance(slots uint32) int {
	bits := 0
	for slots > 1 {
		slots >>= 1
		bits++
	}
	return bits
}

// Cudd_AverageDistanceLinks returns the directed node pairs measured by the
// CUDD AverageDistance diagnostic. BDD nodes contribute T, regular(E), and
// collision-chain links; constants contribute collision-chain links only.
//
// The manager's canonical tables are Go maps, so CUDD's physical buckets are
// not available. We construct deterministic per-subtable buckets from stable
// node identities and link colliding entries in newest-first order. This
// preserves the source algorithm's all-table coverage and collision-link
// contribution while making its value dependent on Go's node allocation
// layout, just as the C value depends on the allocator's layout.
func Cudd_AverageDistanceLinks(dd *DdManager) [][2]*DdNode {
	if dd == nil || dd.closed {
		return nil
	}
	links := make([][2]*DdNode, 0, 2*len(dd.unique)+len(dd.constants))
	byLevel := make(map[uint32][]*DdNode)
	for _, n := range dd.unique {
		if n != nil && n.manager == dd && n.family == 0 {
			byLevel[n.Index] = append(byLevel[n.Index], n)
		}
	}
	for level, nodes := range byLevel {
		slots := uint32(CUDD_UNIQUE_SLOTS)
		if int(level) < len(dd.bddTableSlots) && dd.bddTableSlots[level] != 0 {
			slots = dd.bddTableSlots[level]
		}
		buckets := make(map[uint32][]*DdNode)
		for _, n := range nodes {
			// Use the stable Go node IDs as the pointer tokens that C's hash
			// function obtains from child addresses.
			eToken := n.E.ID
			if Cudd_IsComplement(n.E) {
				eToken |= 1
			}
			bucket := averageDistanceBucket(n.T.ID, eToken, slots)
			buckets[bucket] = append(buckets[bucket], n)
			links = append(links, [2]*DdNode{n, Cudd_T(n)})
			links = append(links, [2]*DdNode{n, Cudd_Regular(Cudd_E(n))})
		}
		for _, chain := range buckets {
			sort.Slice(chain, func(i, j int) bool { return chain[i].ID > chain[j].ID })
			for i := 0; i+1 < len(chain); i++ {
				links = append(links, [2]*DdNode{chain[i], chain[i+1]})
			}
		}
	}
	constantSlots := dd.initSlots
	if constantSlots == 0 {
		constantSlots = CUDD_UNIQUE_SLOTS
	}
	constantBuckets := make(map[uint32][]*DdNode)
	for _, n := range dd.constants {
		if n != nil && n.manager == dd {
			value := math.Float64bits(n.Value)
			bucket := averageDistanceBucket(value, value>>32, constantSlots)
			constantBuckets[bucket] = append(constantBuckets[bucket], n)
		}
	}
	for _, chain := range constantBuckets {
		sort.Slice(chain, func(i, j int) bool { return chain[i].ID > chain[j].ID })
		for i := 0; i+1 < len(chain); i++ {
			links = append(links, [2]*DdNode{chain[i], chain[i+1]})
		}
	}
	return links
}
