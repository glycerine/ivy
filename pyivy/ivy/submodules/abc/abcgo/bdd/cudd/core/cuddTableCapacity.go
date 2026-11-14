package core

func tableInitialSlots(dd *DdManager) uint32 {
	if dd.initSlots != 0 {
		return dd.initSlots
	}
	return CUDD_UNIQUE_SLOTS
}

func (dd *DdManager) ensureTableSlots(zdd bool, size int) bool {
	if size < 0 || uint64(size) > uint64(CUDD_MAXINDEX) {
		return false
	}
	entries := &dd.bddTableSlots
	if zdd {
		entries = &dd.zddTableSlots
	}
	if len(*entries) >= size {
		return true
	}
	initial := tableInitialSlots(dd)
	add := uint64(size-len(*entries)) * uint64(initial)
	if uint64(dd.slotCount)+add > uint64(^uint32(0)) {
		dd.errorCode = CUDD_MEMORY_OUT
		return false
	}
	for len(*entries) < size {
		*entries = append(*entries, initial)
	}
	dd.slotCount += uint32(add)
	if zdd {
		dd.tableZddSize = uint32(size)
	} else {
		dd.tableBDDSize = uint32(size)
	}
	return true
}

func (dd *DdManager) rehashTable(zdd bool, level int, grow bool) bool {
	if dd == nil || dd.closed {
		return false
	}
	entries := dd.bddTableSlots
	if zdd {
		entries = dd.zddTableSlots
	}
	if level < 0 || level >= len(entries) {
		return false
	}
	old := entries[level]
	var next uint32
	if grow {
		if old > ^uint32(0)/2 || uint64(dd.slotCount)+uint64(old) > uint64(^uint32(0)) {
			dd.errorCode = CUDD_MEMORY_OUT
			return false
		}
		next = old * 2
		dd.slotCount += old
	} else {
		if old <= 8 {
			return false
		}
		next = old / 2
		dd.slotCount -= old - next
	}
	if zdd {
		dd.zddTableSlots[level] = next
	} else {
		dd.bddTableSlots[level] = next
	}
	Cudd_FixTableLimits(dd)
	return true
}

// Cudd_RehashBDDSubtable doubles the selected BDD subtable's slot threshold.
func Cudd_RehashBDDSubtable(dd *DdManager, level int) bool { return dd.rehashTable(false, level, true) }

// Cudd_ShrinkBDDSubtable halves a BDD subtable while retaining the minimum.
func Cudd_ShrinkBDDSubtable(dd *DdManager, level int) bool {
	return dd.rehashTable(false, level, false)
}

// Cudd_RehashZddSubtable doubles the selected ZDD subtable's slot threshold.
func Cudd_RehashZddSubtable(dd *DdManager, level int) bool { return dd.rehashTable(true, level, true) }

// Cudd_ShrinkZddSubtable halves a ZDD subtable while retaining the minimum.
func Cudd_ShrinkZddSubtable(dd *DdManager, level int) bool { return dd.rehashTable(true, level, false) }

// Cudd_InsertBDDSubtables grows the allocated BDD subtable range.
func Cudd_InsertBDDSubtables(dd *DdManager, count, level int) bool {
	if dd == nil || dd.closed || count <= 0 || level < 0 || level >= len(dd.bddTableSlots) {
		return false
	}
	if !dd.ensureTableSlots(false, len(dd.bddTableSlots)+count) {
		return false
	}
	if dd.slotCount > dd.looseUpTo {
		dd.gcFraction = DD_GC_FRAC_LO
	}
	Cudd_FixTableLimits(dd)
	return true
}

// Cudd_DestroyBDDSubtables removes allocated BDD subtables from the end.
func Cudd_DestroyBDDSubtables(dd *DdManager, count int) bool {
	if dd == nil || dd.closed || count < 0 || count > len(dd.bddTableSlots)-len(dd.Vars) {
		return false
	}
	for _, slots := range dd.bddTableSlots[len(dd.bddTableSlots)-count:] {
		dd.slotCount -= slots
	}
	dd.bddTableSlots = dd.bddTableSlots[:len(dd.bddTableSlots)-count]
	dd.tableBDDSize = uint32(len(dd.bddTableSlots))
	Cudd_FixTableLimits(dd)
	return true
}

// Cudd_ResizeZddTable grows capacity up to index while retaining active size.
func Cudd_ResizeZddTable(dd *DdManager, index int) bool {
	if dd == nil || dd.closed || index < 0 || index >= int(CUDD_MAXINDEX) {
		return false
	}
	if !dd.ensureZddVar(index) || !dd.ensureTableSlots(true, index+1) {
		return false
	}
	Cudd_FixTableLimits(dd)
	return true
}

// Cudd_ResizeBDDTable grows capacity up to index while retaining active size.
func Cudd_ResizeBDDTable(dd *DdManager, index int) bool {
	if dd == nil || dd.closed || index < 0 || index >= int(CUDD_MAXINDEX) {
		return false
	}
	if Cudd_bddIthVar(dd, index) == nil || !dd.ensureTableSlots(false, index+1) {
		return false
	}
	Cudd_FixTableLimits(dd)
	return true
}

// Cudd_SlowTableGrowth raises density thresholds after memory pressure so the
// Go map-backed manager performs fewer resize/rebuild cycles.
func Cudd_SlowTableGrowth(dd *DdManager) {
	if dd == nil || dd.closed {
		return
	}
	if dd.tableGrowthFactor == 0 {
		dd.tableGrowthFactor = 1
	}
	if dd.tableGrowthFactor <= ^uint32(0)/4 {
		dd.tableGrowthFactor *= 4
	}
	if dd.cacheSlots > 0 {
		dd.maxCacheHard = dd.cacheSlots - 1
	}
	dd.cacheSlack = -int32(dd.cacheSlots + 1)
	dd.gcFraction = DD_GC_FRAC_MIN
	dd.minDead = uint32(dd.gcFraction * float64(dd.slotCount))
}

// Cudd_FindParent reports whether an earlier-level BDD node has both
// children equal to node, matching the parent check used before table removal.
func Cudd_FindParent(dd *DdManager, node *DdNode) bool {
	if dd == nil || dd.closed || node == nil || node.manager != dd || Cudd_IsConstant(node) {
		return false
	}
	level := dd.level(node)
	for key, parent := range dd.unique {
		if int(key.index) >= len(dd.Perm) || dd.Perm[key.index] >= level {
			continue
		}
		if parent.T == node && Cudd_Regular(parent.E) == node {
			return true
		}
	}
	return false
}

// Cudd_FixTableLimits recomputes the garbage-collection and computed-cache
// thresholds after unique-table capacity changes.
func Cudd_FixTableLimits(dd *DdManager) {
	if dd == nil || dd.closed {
		return
	}
	if dd.gcFraction == 0 {
		dd.gcFraction = DD_GC_FRAC_HI
	}
	dd.minDead = uint32(dd.gcFraction * float64(dd.slotCount))
	cacheLimit := uint64(DD_MAX_CACHE_TO_SLOTS_RATIO) * uint64(dd.slotCount)
	if cacheLimit > uint64(dd.maxCacheHard) {
		cacheLimit = uint64(dd.maxCacheHard)
	}
	dd.cacheSlack = int32(cacheLimit) - 2*int32(dd.cacheSlots)
	if dd.cacheSlots < dd.slotCount/2 && dd.cacheSlack >= 0 {
		dd.cacheSlots = uint32(cacheLimit)
	}
}
