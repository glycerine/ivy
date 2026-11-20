package core

import "math/bits"

type cacheOperation uint8

const (
	cacheExternal cacheOperation = iota
	cacheBDDAnd
	cacheBDDIte
	cacheADDIte
)

type computedKey struct {
	f, g, h *DdNode
	op      uintptr
	kind    cacheOperation
}

type computedEntry struct {
	key    computedKey
	result *DdNode
}

type computedTable struct {
	slots                 map[uint32]computedEntry
	shift                 uint
	valid                 int
	baseHits, baseLookups uint64
	missOffset            float64
}

func (k computedKey) hash() uint32 {
	// CUDD uses the high bits of a multiplicative hash of the node IDs.
	var f, g uint32
	if k.f != nil {
		f = uint32(k.f.ID)
	}
	if k.g != nil {
		g = uint32(k.g.ID)
	} else {
		g = f
	}
	op := uint32(k.op) + 741457*uint32(k.kind)
	if k.h != nil {
		op += 741457 * uint32(k.h.ID+1)
	}
	return ((f+op)*12582917 + g) * 4256249
}

func (dd *DdManager) resetCacheWindow() {
	t := &dd.computed
	t.shift = uint(32 - bits.TrailingZeros32(dd.cacheSlots))
	t.baseHits, t.baseLookups = dd.cacheHits, dd.cacheLookups
	t.missOffset = float64(int64(float64(dd.cacheSlots)*dd.minHit + 1))
}

// Cudd_InitCache initializes the shared, direct-mapped computed table.
func Cudd_InitCache(dd *DdManager, requested, hardLimit uint32) int32 {
	if dd == nil || dd.closed {
		return 0
	}
	requested = max(requested, Cudd_ReadSlots(dd)/2, 1)
	dd.cacheSlots = 1 << (bits.Len32(requested) - 1)
	dd.maxCacheHard = hardLimit
	dd.updateCacheSlack()
	dd.computed = computedTable{}
	dd.cacheHits, dd.cacheLookups = 0, 0
	Cudd_SetMinHit(dd, 30)
	dd.resetCacheWindow()
	return 1
}

// Cudd_ResizeCache doubles the slot count and rehashes all valid entries.
func Cudd_ResizeCache(dd *DdManager) {
	if dd == nil || dd.closed || dd.cacheSlots == 0 || dd.cacheSlots > ^uint32(0)/2 {
		return
	}
	dd.cacheSlots *= 2
	dd.cacheSlack -= int32(dd.cacheSlots)
	dd.resetCacheWindow()
	t := &dd.computed
	old := t.slots
	t.slots = nil
	if t.valid != 0 {
		t.slots = make(map[uint32]computedEntry, t.valid)
		for _, e := range old {
			if e.result != nil {
				t.slots[e.key.hash()>>t.shift] = e
			}
		}
	}
}

func (dd *DdManager) cacheLookup(key computedKey, hash uint32, reclaim bool) *DdNode {
	t := &dd.computed
	dd.cacheLookups++
	entry := t.slots[hash>>t.shift]
	if entry.result != nil && entry.key == key {
		dd.cacheHits++
		if reclaim {
			reclaimNode(Cudd_Regular(entry.result))
		}
		return entry.result
	}
	if dd.cacheSlack >= 0 {
		hits := dd.cacheHits - t.baseHits
		misses := dd.cacheLookups - t.baseLookups - hits
		if float64(hits) > (float64(misses)+t.missOffset)*dd.minHit {
			Cudd_ResizeCache(dd)
		}
	}
	return nil
}

func (dd *DdManager) cacheInsert(key computedKey, result *DdNode, hash uint32) {
	t := &dd.computed
	if t.slots == nil {
		t.slots = make(map[uint32]computedEntry)
	}
	slot := hash >> t.shift
	if t.slots[slot].result == nil {
		t.valid++
	}
	t.slots[slot] = computedEntry{key, result}
}

// The lookup family selects whether to revive children. All callers use the
// same insertion operation, including ZDD and constant-only lookups.
func Cudd_CacheLookup(dd *DdManager, operation uintptr, f, g, h *DdNode, family uint8) *DdNode {
	if dd == nil || dd.closed {
		return nil
	}
	key := computedKey{f, g, h, operation, cacheExternal}
	return dd.cacheLookup(key, key.hash(), family != 2)
}

func Cudd_CacheInsert(dd *DdManager, operation uintptr, f, g, h, result *DdNode, family uint8) {
	if dd == nil || dd.closed || result == nil {
		return
	}
	key := computedKey{f, g, h, operation, cacheExternal}
	dd.cacheInsert(key, result, key.hash())
}

// CUDD's used-slot statistic survives invalidation until the next resize.
// Sparse slot indices retain this statistic without retaining node handles.
func Cudd_CacheFlush(dd *DdManager) {
	if dd != nil && dd.computed.valid != 0 {
		for slot := range dd.computed.slots {
			dd.computed.slots[slot] = computedEntry{}
		}
		dd.computed.valid = 0
	}
}

func Cudd_CacheSize(dd *DdManager) int {
	if dd == nil {
		return 0
	}
	return dd.computed.valid
}
