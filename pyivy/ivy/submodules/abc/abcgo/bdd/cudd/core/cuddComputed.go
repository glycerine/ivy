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
	cacheBDDExist
	cacheBDDAndAbstract
	cacheZddIsopBDD // cuddZddIsop's BDD result
	cacheZddIsopZDD // Cudd_zddIsop's ZDD cover
	cacheBddIsop    // cuddBddIsop
	cacheZddSubset0 // zdd_subset0_aux (op: the variable)
	cacheZddSubset1 // zdd_subset1_aux (op: the variable)
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
	sparse    []uint64
	firstSlot [4]uint32
	index     []uint32
	first     [4]computedEntry
	blocks    []*computedBlock
	used      int
}

// Each sparse word stores the logical slot above the one-based record position.
func probeCacheSlot(table []uint64, slot uint32) int {
	mask := len(table) - 1
	i := int((slot * 2654435761) >> uint(bits.LeadingZeros32(uint32(mask))))
	for table[i] != 0 && uint32(table[i]>>32) != slot {
		i = (i + 1) & mask
	}
	return i
}

func (s *computedSlots) position(slot uint32) uint32 {
	if s.index != nil {
		return s.index[slot]
	}
	if s.sparse != nil {
		return uint32(s.sparse[probeCacheSlot(s.sparse, slot)])
	}
	for i := 0; i < s.used; i++ {
		if s.firstSlot[i] == slot {
			return uint32(i + 1)
		}
	}
	return 0
}

func (s *computedSlots) record(pos uint32) *computedEntry {
	i := pos - 1
	if i < uint32(len(s.first)) {
		return &s.first[i]
	}
	i -= uint32(len(s.first))
	return &s.blocks[i/cacheBlockSize][i%cacheBlockSize]
}

func (s *computedSlots) prepare(size uint32, expected int) {
	if s.index != nil || expected <= len(s.first) {
		return
	}
	dense := expected >= 64 && uint64(expected)*16 >= uint64(size)
	if !dense && expected <= len(s.sparse)*3/4 {
		return
	}
	old := s.sparse
	if dense {
		s.index = make([]uint32, size)
		s.sparse = nil
	} else {
		capacity := 16
		for expected > capacity*3/4 {
			capacity *= 2
		}
		s.sparse = make([]uint64, capacity)
	}
	if old == nil {
		for i := 0; i < s.used; i++ {
			s.setPosition(s.firstSlot[i], uint32(i+1))
		}
	} else {
		for _, word := range old {
			if word != 0 {
				s.setPosition(uint32(word>>32), uint32(word))
			}
		}
	}
}

func (s *computedSlots) setPosition(slot, pos uint32) {
	if s.index != nil {
		s.index[slot] = pos
		return
	}
	if s.sparse != nil {
		s.sparse[probeCacheSlot(s.sparse, slot)] = uint64(slot)<<32 | uint64(pos)
		return
	}
	s.firstSlot[pos-1] = slot
}

// write returns true if the slot already held a valid result.
func (s *computedSlots) write(slot uint32, e computedEntry) bool {
	pos := s.position(slot)
	if pos == 0 {
		pos = uint32(s.used + 1)
		s.setPosition(slot, pos)
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
	if pos := t.position(hash >> t.shift); pos != 0 {
		entry := t.record(pos)
		if entry.result != nil && entry.key == key {
			dd.cacheHits++
			if reclaim {
				reclaimNode(Cudd_Regular(entry.result))
			}
			return entry.result
		}
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
