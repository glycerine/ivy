package dsd

import (
	"sync"

	bdd "github.com/glycerine/abcgo/bdd/cudd/core"
)

type identityKey [4]*DdNode

// The compatibility cache is synchronized; manager decomposition uses its own
// memo table instead, so independent managers never share mutable node marks.
var checkCache struct {
	sync.Mutex
	entries  map[identityKey]bool
	capacity int
}

func Dsd_CheckCacheAllocate(nEntries int) {
	if nEntries <= 0 {
		panic("dsd: nonpositive cache size")
	}
	checkCache.Lock()
	defer checkCache.Unlock()
	checkCache.capacity = nEntries
	checkCache.entries = make(map[identityKey]bool, nEntries)
}
func Dsd_CheckCacheDeallocate() {
	checkCache.Lock()
	defer checkCache.Unlock()
	checkCache.entries = nil
	checkCache.capacity = 0
}
func Dsd_CheckCacheClear() { checkCache.Lock(); defer checkCache.Unlock(); clear(checkCache.entries) }
func Dsd_CheckRootFunctionIdentity(dd *DdManager, f1, f2, c1, c2 *DdNode) int {
	checkCache.Lock()
	defer checkCache.Unlock()
	cache := checkCache.entries
	if cache == nil {
		cache = make(map[identityKey]bool)
	} else if len(cache) >= checkCache.capacity {
		clear(cache)
	}
	if rootIdentity(dd, identityKey{f1, f2, c1, c2}, cache) {
		return 1
	}
	return 0
}
func rootIdentity(dd *DdManager, k identityKey, cache map[identityKey]bool) bool {
	f1, f2, c1, c2 := k[0], k[1], k[2], k[3]
	if c1 == dd.LogicZero || c2 == dd.LogicZero {
		panic("dsd: identity condition is false")
	}
	if c1 == dd.One && c2 == dd.One {
		return f1 == f2
	}
	if f1 == dd.LogicZero {
		return bdd.Cudd_bddLeq(dd, c2, bdd.Cudd_Not(f2))
	}
	if f1 == dd.One {
		return bdd.Cudd_bddLeq(dd, c2, f2)
	}
	if f2 == dd.LogicZero {
		return bdd.Cudd_bddLeq(dd, c1, bdd.Cudd_Not(f1))
	}
	if f2 == dd.One {
		return bdd.Cudd_bddLeq(dd, c1, f1)
	}
	if value, ok := cache[k]; ok {
		return value
	}
	lev := level(dd, k[0])
	for _, f := range k[1:] {
		lev = min(lev, level(dd, f))
	}
	var lo, hi identityKey
	for i, f := range k {
		lo[i], hi[i] = split(dd, f, lev)
	}
	choose := func(i int) (*DdNode, *DdNode) {
		if lo[i] != dd.LogicZero {
			return lo[i-2], lo[i]
		}
		return hi[i-2], hi[i]
	}
	has1, has2 := level(dd, c1) == lev, level(dd, c2) == lev
	var result bool
	switch {
	case has1 && has2:
		a, c := choose(2)
		b, d := choose(3)
		result = rootIdentity(dd, identityKey{a, b, c, d}, cache)
	case has1:
		a, c := choose(2)
		result = rootIdentity(dd, identityKey{a, lo[1], c, lo[3]}, cache) && rootIdentity(dd, identityKey{a, hi[1], c, hi[3]}, cache)
	case has2:
		b, d := choose(3)
		result = rootIdentity(dd, identityKey{lo[0], b, lo[2], d}, cache) && rootIdentity(dd, identityKey{hi[0], b, hi[2], d}, cache)
	default:
		result = rootIdentity(dd, lo, cache) && rootIdentity(dd, hi, cache)
	}
	cache[k] = result
	return result
}
func (m *Dsd_Manager_t) ident(f1, f2, c1, c2 *DdNode) bool {
	if len(m.identity) >= 5000 {
		clear(m.identity)
	}
	return rootIdentity(m.dd, identityKey{f1, f2, c1, c2}, m.identity)
}
