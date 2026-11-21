package core

import (
	"time"

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
	width := &entries[level]
	next := uint64(*width) / 2
	if grow {
		dd.adjustGrowthFraction(zdd)
		next = 2 * uint64(*width)
		if zdd && level < len(dd.InvPermZ) {
			keys := dd.zddUnique.levelCount(dd.InvPermZ[level])
			for 4*next < uint64(keys) {
				next *= 2
			}
		}
	} else if *width <= 8 {
		return false
	}
	ok := dd.setSubtableWidth(width, next)
	if ok && !zdd && dd.bddGrowthFactors != nil {
		index := level
		if level < len(dd.InvPerm) {
			index = dd.InvPerm[level]
		}
		delete(dd.bddGrowthFactors, uint32(index))
	}
	return ok
}

func (dd *DdManager) setSubtableWidth(width *uint32, next uint64) bool {
	total := uint64(dd.slotCount) - uint64(*width) + next
	if total > uint64(^uint32(0)) {
		dd.errorCode = CUDD_MEMORY_OUT
		return false
	}
	*width = uint32(next)
	dd.slotCount = uint32(total)
	Cudd_FixTableLimits(dd)
	return true
}

func (dd *DdManager) constantTableWidth() uint32 {
	if dd.constantSlots != 0 {
		return dd.constantSlots
	}
	return tableInitialSlots(dd)
}

func exceedsTableDensity(keys int, width, factor uint32) bool {
	// Rearrange keys > 4*width*factor so even maximum widths cannot overflow.
	return keys > 0 && uint64(keys-1)/4 >= uint64(width)*uint64(factor)
}

// The native unique tables check density before inserting a fresh BDD/ADD
// key, and before looking up a ZDD key or constant. The Go probe arrays grow
// independently of these logical widths.
func (dd *DdManager) checkNodeDensity(key bddUniqueKey, zdd bool) bool {
	index := int(key.index)
	perm, widths := dd.Perm, dd.bddTableSlots
	if zdd {
		perm, widths = dd.PermZ, dd.zddTableSlots
	}
	if index >= len(perm) || perm[index] >= len(widths) {
		return false
	}
	width := &widths[perm[index]]
	var keys int
	if zdd {
		keys = dd.zddUnique.levelCount(index)
	} else {
		keys = dd.unique.levelCount(index) + dd.addUnique.levelCount(index)
	}
	if uint64(keys) <= 4*uint64(*width) {
		return false
	}
	if !zdd && !exceedsTableDensity(keys, *width, dd.bddDensityFactor(key.index)) {
		return false
	}
	if dd.allocationAttempt != nil || dd.pressureGC {
		// A bounded operation may still roll back to its starting set of keys.
		return false
	}
	if dd.gcEnabled != 0 && dd.shouldCollectLevel(index, keys, zdd) {
		dd.collectReleasedKeys(key.t, key.e)
		// Collection and its hooks can invalidate a saved probe position.
		return true
	}
	dd.rehashTable(zdd, perm[index], true)
	return false
}

func (dd *DdManager) bddDensityFactor(index uint32) uint32 {
	return max(1, dd.bddGrowthFactors[index])
}

func (dd *DdManager) adjustGrowthFraction(zdd bool) {
	if dd.slotCount > dd.looseUpTo && (dd.gcFraction == DD_GC_FRAC_HI || zdd) {
		dd.gcFraction = DD_GC_FRAC_LO
	}
}

func (dd *DdManager) shouldCollectLevel(index, keys int, zdd bool) bool {
	released := 0
	count := func(r uniqueEntry) {
		if r.node.childrenReleased {
			released++
		}
	}
	if zdd {
		dd.zddUnique.forLevel(index, count)
		return 10*uint64(released) > 9*uint64(keys) ||
			dd.deadZ > int(dd.minDead)
	}
	if dd.dead > int(dd.minDead) {
		return true
	}
	if dd.dead <= int(dd.minDead)/2 {
		return false
	}
	dd.unique.forLevel(index, count)
	dd.addUnique.forLevel(index, count)
	return float64(released) > 0.95*float64(keys)
}

func (dd *DdManager) checkConstantDensity() {
	width := dd.constantTableWidth()
	if dd.slotCount == 0 || uint64(len(dd.constants)) <= 4*uint64(width) || dd.pressureGC || dd.allocationAttempt != nil {
		return
	}
	if dd.gcEnabled != 0 && dd.dead != 0 {
		dead := 0
		if dd.dead <= int(dd.minDead) {
			for _, n := range dd.constants {
				if n.childrenReleased {
					dead++
				}
			}
		}
		if dd.dead > int(dd.minDead) || 10*uint64(dead) > 9*uint64(len(dd.constants)) {
			dd.collectReleasedKeys(nil, nil)
			return
		}
	}
	dd.adjustGrowthFraction(false)
	dd.constantSlots = width
	dd.setSubtableWidth(&dd.constantSlots, 2*uint64(width))
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
	if uint64(oldSize)+uint64(count) >= uint64(CUDD_MAXINDEX) || !dd.ensureTableSlots(false, oldSize+count) {
		return false
	}
	if Cudd_bddIthVar(dd, oldSize+count-1) == nil {
		return false
	}
	// Move the old suffix down as a unit, preserving its per-level capacities.
	copy(dd.InvPerm[level+count:], dd.InvPerm[level:oldSize])
	copy(dd.bddTableSlots[level+count:], dd.bddTableSlots[level:oldSize])
	for i := range count {
		dd.InvPerm[level+i] = oldSize + i
		dd.bddTableSlots[level+i] = dd.initSlots
	}
	for l := level; l < len(dd.InvPerm); l++ {
		dd.Perm[dd.InvPerm[l]] = l
	}
	dd.symmetryGroups = nil
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
	// Validate all selected levels before releasing their projections. A Go
	// projection has an implicit manager pin outside its numeric refcount.
	for index := first; index < len(dd.Vars); index++ {
		node := dd.Vars[index]
		if node.Ref != 0 {
			if node.Ref != ^uint32(0) || dd.hasLiveParent(node) {
				return false
			}
		}
		if dd.unique.levelCount(index) > 1 && dd.hasLiveNonProjection(&dd.unique, index, node) ||
			dd.addUnique.levelCount(index) != 0 && dd.hasLiveNonProjection(&dd.addUnique, index, nil) {
			return false
		}
	}
	started := time.Now()
	if !dd.runHooks(CUDD_PRE_GC_HOOK, "DD", nil) {
		return false
	}
	dd.peakLive = max(dd.peakLive, dd.liveBDDKeys())
	dd.purgeReleasedKeys()
	// Every removed level now contains exactly its isolated projection.
	for _, node := range dd.Vars[first:] {
		dd.retireNode(node)
	}
	clear(dd.unique.levels[first:len(dd.Vars)])
	dd.unique.count -= count
	if len(dd.boundVars)+len(dd.variableTypes)+len(dd.pairIndices)+len(dd.lazyGroups)+len(dd.bddGrowthFactors) != 0 {
		for index := first; index < len(dd.Vars); index++ {
			delete(dd.boundVars, index)
			delete(dd.variableTypes, index)
			delete(dd.pairIndices, index)
			delete(dd.lazyGroups, index)
			delete(dd.bddGrowthFactors, uint32(index))
		}
	}
	write := 0
	for level, index := range dd.InvPerm {
		if index >= first {
			dd.slotCount -= dd.bddTableSlots[level]
			continue
		}
		dd.bddTableSlots[write] = dd.bddTableSlots[level]
		dd.InvPerm[write], dd.Perm[index] = index, write
		write++
	}
	// InitTable may have reserved additional slots beyond materialized vars.
	reserved := copy(dd.bddTableSlots[write:], dd.bddTableSlots[len(dd.InvPerm):])
	dd.bddTableSlots = dd.bddTableSlots[:write+reserved]
	clear(dd.Vars[first:])
	dd.Vars, dd.Perm, dd.InvPerm = dd.Vars[:first], dd.Perm[:first], dd.InvPerm[:first]
	dd.discardInteractionsAfterSize(first)
	dd.varMap = nil
	dd.symmetryGroups = nil
	dd.tableBDDSize = uint32(len(dd.bddTableSlots))
	Cudd_CacheFlush(dd)
	Cudd_FixTableLimits(dd)
	dd.garbageCollections++
	dd.gcTime += uint64(time.Since(started).Nanoseconds())
	dd.runHooks(CUDD_POST_GC_HOOK, "DD", nil)
	return true
}

func (dd *DdManager) hasLiveNonProjection(table *uniqueTable, index int, projection *DdNode) bool {
	live := false
	table.forLevel(index, func(entry uniqueEntry) {
		if entry.node != projection && (!entry.node.childrenReleased || entry.node == dd.Background) {
			live = true
		}
	})
	return live
}

// Saturated reference counts cannot tell whether a projection is isolated.
func (dd *DdManager) hasLiveParent(node *DdNode) bool {
	for _, table := range []*uniqueTable{&dd.unique, &dd.addUnique} {
		for _, p := range table.entries {
			if !p.childrenReleased && (Cudd_Regular(p.T) == node || Cudd_Regular(p.E) == node) {
				return true
			}
		}
	}
	return false
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
	return dd.ensureZddVar(index)
}

// Cudd_ResizeBDDTable grows capacity up to index while retaining active size.
func Cudd_ResizeBDDTable(dd *DdManager, index int) bool {
	if dd == nil || dd.closed || index < 0 || index >= int(CUDD_MAXINDEX) {
		return false
	}
	return Cudd_bddIthVar(dd, index) != nil
}

// Cudd_SlowTableGrowth raises density thresholds after memory pressure.
func Cudd_SlowTableGrowth(dd *DdManager) {
	if dd == nil || dd.closed {
		return
	}
	if dd.bddGrowthFactors == nil {
		dd.bddGrowthFactors = make(map[uint32]uint32, len(dd.bddTableSlots))
	}
	for i := range dd.bddTableSlots {
		index := uint32(i)
		factor := dd.bddDensityFactor(index)
		if factor <= ^uint32(0)/4 {
			dd.bddGrowthFactors[index] = 4 * factor
		}
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
	for key, parent := range dd.unique.entries {
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
	dd.updateCacheSlack()
	if dd.cacheSlots < dd.slotCount/2 && dd.cacheSlack >= 0 {
		Cudd_ResizeCache(dd)
	}
}

func (dd *DdManager) updateCacheSlack() {
	limit := min(uint64(DD_MAX_CACHE_TO_SLOTS_RATIO)*uint64(dd.slotCount), uint64(dd.maxCacheHard))
	dd.cacheSlack = int32(uint32(limit) - 2*dd.cacheSlots)
}
