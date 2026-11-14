package core

import (
	"sort"

	"github.com/glycerine/abcgo/bdd/mtr"
)

func tableInitialSlots(dd *DdManager) uint32 {
	if dd.initSlots != 0 {
		return dd.initSlots
	}
	return CUDD_UNIQUE_SLOTS
}

func initializeTableMetadata(dd *DdManager) bool {
	if dd.slotCount != 0 {
		return true
	}
	initial := tableInitialSlots(dd)
	total := (uint64(len(dd.Vars)) + uint64(dd.ZddSize) + 1) * uint64(initial)
	if total > uint64(^uint32(0)) {
		dd.errorCode = CUDD_MEMORY_OUT
		return false
	}
	dd.initSlots = initial
	dd.slotCount = uint32(total)
	if len(dd.bddTableSlots) == 0 {
		dd.bddTableSlots = make([]uint32, len(dd.Vars))
		for i := range dd.bddTableSlots {
			dd.bddTableSlots[i] = initial
		}
	}
	if len(dd.zddTableSlots) == 0 {
		dd.zddTableSlots = make([]uint32, dd.ZddSize)
		for i := range dd.zddTableSlots {
			dd.zddTableSlots[i] = initial
		}
	}
	dd.tableBDDSize = uint32(len(dd.bddTableSlots))
	dd.tableZddSize = uint32(len(dd.zddTableSlots))
	return true
}

func (dd *DdManager) ensureTableSlots(zdd bool, size int) bool {
	if size < 0 || uint64(size) > uint64(CUDD_MAXINDEX) {
		return false
	}
	if !initializeTableMetadata(dd) {
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
	if dd == nil || dd.closed || count <= 0 || level < 0 || level > len(dd.Vars) || !initializeTableMetadata(dd) {
		return false
	}
	oldSize := len(dd.Vars)
	if !dd.ensureTableSlots(false, oldSize) || uint64(oldSize)+uint64(count) >= uint64(CUDD_MAXINDEX) {
		return false
	}
	oldOrder := append([]int(nil), dd.InvPerm...)
	newIndices := make([]int, count)
	for i := 0; i < count; i++ {
		newIndex := len(dd.Vars)
		if Cudd_bddNewVar(dd) == nil {
			return false
		}
		newIndices[i] = newIndex
	}
	order := make([]int, 0, oldSize+count)
	order = append(order, oldOrder[:level]...)
	order = append(order, newIndices...)
	order = append(order, oldOrder[level:]...)
	dd.InvPerm = order
	dd.Perm = make([]int, len(order))
	for newLevel, index := range order {
		dd.Perm[index] = newLevel
	}
	newSlots := make([]uint32, 0, len(dd.bddTableSlots)+count)
	newSlots = append(newSlots, dd.bddTableSlots[:level]...)
	for i := 0; i < count; i++ {
		newSlots = append(newSlots, dd.initSlots)
	}
	newSlots = append(newSlots, dd.bddTableSlots[level:]...)
	dd.bddTableSlots = newSlots
	dd.tableBDDSize = uint32(len(newSlots))
	dd.slotCount += uint32(count) * dd.initSlots
	if dd.Tree != nil {
		dd.Tree.Size += uint32(count)
		if len(dd.InvPerm) != 0 {
			dd.Tree.Index = uint32(dd.InvPerm[0])
		}
		patchManagerTree(dd, dd.Tree)
	}
	if dd.slotCount > dd.looseUpTo {
		dd.gcFraction = DD_GC_FRAC_LO
	}
	Cudd_FixTableLimits(dd)
	return true
}

// Cudd_DestroyBDDSubtables removes allocated BDD subtables from the end.
func Cudd_DestroyBDDSubtables(dd *DdManager, count int) bool {
	if dd == nil || dd.closed || count <= 0 || !initializeTableMetadata(dd) {
		return false
	}
	if count > len(dd.Vars) {
		count = len(dd.Vars)
	}
	if count == 0 {
		return false
	}
	first := len(dd.Vars) - count
	removed := make(map[*DdNode]bool, count)
	for index := first; index < len(dd.Vars); index++ {
		removed[Cudd_Regular(dd.Vars[index])] = true
	}
	for _, node := range dd.unique {
		if int(node.Index) >= first {
			if !removed[node] {
				return false
			}
			continue
		}
		if removed[Cudd_Regular(node.T)] || removed[Cudd_Regular(node.E)] {
			return false
		}
	}
	for node := range removed {
		if Cudd_ExternalRefCount(dd, node) != 0 {
			return false
		}
	}
	removedLevels := make([]int, 0, count)
	for index := first; index < len(dd.Vars); index++ {
		node := Cudd_Regular(dd.Vars[index])
		removedLevels = append(removedLevels, dd.Perm[index])
		if !node.childrenReleased {
			node.childrenReleased = true
			Cudd_RecursiveDeref(dd, node.T)
			Cudd_RecursiveDeref(dd, node.E)
		}
		delete(dd.unique, bddUniqueKey{node.Index, node.T, node.E})
		node.manager = nil
		if node.inverse != nil {
			node.inverse.manager = nil
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(removedLevels)))
	for _, level := range removedLevels {
		if level < 0 || level >= len(dd.bddTableSlots) {
			continue
		}
		dd.slotCount -= dd.bddTableSlots[level]
		dd.bddTableSlots = append(dd.bddTableSlots[:level], dd.bddTableSlots[level+1:]...)
	}
	dd.Vars = dd.Vars[:first]
	order := make([]int, 0, first)
	for _, index := range dd.InvPerm {
		if index < first {
			order = append(order, index)
		}
	}
	dd.InvPerm = order
	dd.Perm = make([]int, first)
	for level, index := range order {
		dd.Perm[index] = level
	}
	dd.cache = make(map[bddITEKey]*DdNode)
	clear(dd.computed)
	clear(dd.addCache)
	dd.tableBDDSize = uint32(len(dd.bddTableSlots))
	Cudd_FixTableLimits(dd)
	Cudd_GarbageCollect(dd, true)
	return true
}

func patchManagerTree(dd *DdManager, node *mtr.MtrNode) {
	for current := node; current != nil; current = current.Younger {
		current.Low = uint32(Cudd_ReadPerm(dd, int(current.Index)))
		if current.Child != nil {
			patchManagerTree(dd, current.Child)
		}
	}
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
