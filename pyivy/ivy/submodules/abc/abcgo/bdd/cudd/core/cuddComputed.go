package core

import "math/bits"

type cacheOperation uint8

const (
	cacheExternal cacheOperation = iota
	cacheBDDAnd
	cacheBDDIte
	cacheADDIte
	cacheBDDVarMap
	cacheADDTimes
	cacheADDPlus
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
	computedSlots
	shift                 uint
	valid                 int
	baseHits, baseLookups uint64
	missOffset            float64
}

const cacheBlockSize = 64

type computedBlock [cacheBlockSize]computedEntry

// Entries stay packed as the number of occupied slots grows. Small caches
// use a sparse slot index; fuller caches can address the same entries directly.
type computedSlots struct {
	sparse map[uint32]uint32
	index  []uint32
	first  [4]computedEntry
	blocks []*computedBlock
	used   int
}

func (s *computedSlots) position(slot uint32) uint32 {
	if s.index != nil {
		return s.index[slot]
	}
	return s.sparse[slot]
}

func (s *computedSlots) record(pos uint32) *computedEntry {
	i := pos - 1
	if i < uint32(len(s.first)) {
		return &s.first[i]
	}
	i -= uint32(len(s.first))
	return &s.blocks[i/cacheBlockSize][i%cacheBlockSize]
}

func (s *computedSlots) read(slot uint32) computedEntry {
	if p := s.position(slot); p != 0 {
		return *s.record(p)
	}
	return computedEntry{}
}

func (s *computedSlots) prepare(size uint32, expected int) {
	if s.index != nil {
		return
	}
	if expected >= 64 && uint64(expected)*16 >= uint64(size) {
		s.index = make([]uint32, size)
		for slot, pos := range s.sparse {
			s.index[slot] = pos
		}
		s.sparse = nil
	} else if s.sparse == nil {
		s.sparse = make(map[uint32]uint32, expected)
	}
}

// write returns true if the slot already held a valid result.
func (s *computedSlots) write(slot uint32, e computedEntry) bool {
	pos := s.position(slot)
	if pos == 0 {
		pos = uint32(s.used + 1)
		if s.index != nil {
			s.index[slot] = pos
		} else {
			s.sparse[slot] = pos
		}
		if i := s.used - len(s.first); i >= 0 && i%cacheBlockSize == 0 {
			s.blocks = append(s.blocks, new(computedBlock))
		}
		s.used++
	}
	r := s.record(pos)
	valid := r.result != nil
	*r = e
	return valid
}

func (s *computedSlots) records(yield func(*computedEntry) bool) {
	for i := 0; i < min(s.used, len(s.first)); i++ {
		if !yield(&s.first[i]) {
			return
		}
	}
	for _, block := range s.blocks {
		for i := range block {
			if !yield(&block[i]) {
				return
			}
		}
	}
}

func (s *computedSlots) invalidate(drop func(computedEntry) bool) int {
	removed := 0
	for e := range s.records {
		if e.result != nil && (drop == nil || drop(*e)) {
			*e = computedEntry{}
			removed++
		}
	}
	return removed
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
	old := t.computedSlots
	t.computedSlots = computedSlots{}
	if t.valid != 0 {
		t.prepare(dd.cacheSlots, t.valid)
		for e := range old.records {
			if e.result != nil {
				t.write(e.key.hash()>>t.shift, *e)
			}
		}
	}
}

func (dd *DdManager) cacheLookup(key computedKey, hash uint32, reclaim bool) *DdNode {
	t := &dd.computed
	dd.cacheLookups++
	entry := t.read(hash >> t.shift)
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
	t.prepare(dd.cacheSlots, t.used+1)
	if !t.write(hash>>t.shift, computedEntry{key, result}) {
		t.valid++
	}
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
		dd.computed.invalidate(nil)
		dd.computed.valid = 0
	}
}

func Cudd_CacheSize(dd *DdManager) int {
	if dd == nil {
		return 0
	}
	return dd.computed.valid
}
