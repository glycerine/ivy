package core

import (
	"io"
	"os"

	"github.com/glycerine/abcgo/bdd/mtr"
)

func Cudd_Output(dd *DdManager) io.Writer {
	if dd == nil || dd.stdout == nil {
		return os.Stdout
	}
	return dd.stdout
}
func Cudd_SetOutput(dd *DdManager, stdout, stderr io.Writer) {
	if dd == nil || dd.closed {
		return
	}
	if stdout != nil {
		dd.stdout = stdout
	}
	if stderr != nil {
		dd.stderr = stderr
	}
}
func Cudd_ReadTree(dd *DdManager) *mtr.MtrNode {
	if dd == nil || dd.closed {
		return nil
	}
	return dd.Tree
}
func Cudd_FreeTree(dd *DdManager) {
	if dd != nil {
		mtr.Mtr_FreeTree(dd.Tree)
		dd.Tree = nil
	}
}
func Cudd_ReadZddTree(dd *DdManager) *mtr.MtrNode {
	if dd == nil || dd.closed {
		return nil
	}
	return dd.TreeZ
}
func Cudd_FreeZddTree(dd *DdManager) {
	if dd != nil {
		mtr.Mtr_FreeTree(dd.TreeZ)
		dd.TreeZ = nil
	}
}
func fixVarTree(tree *mtr.MtrNode, permutation []int) {
	for node := tree; node != nil; {
		index := int(node.Low)
		node.Index = node.Low
		if index >= 0 && index < len(permutation) {
			node.Low = uint32(permutation[index])
		}
		if node.Child != nil {
			fixVarTree(node.Child, permutation)
		}
		node = node.Younger
	}
}
func Cudd_SetTree(dd *DdManager, tree *mtr.MtrNode) {
	if dd == nil || dd.closed {
		return
	}
	mtr.Mtr_FreeTree(dd.Tree)
	dd.Tree = tree
	if tree != nil {
		fixVarTree(tree, dd.Perm)
	}
}
func Cudd_SetZddTree(dd *DdManager, tree *mtr.MtrNode) {
	if dd == nil || dd.closed {
		return
	}
	mtr.Mtr_FreeTree(dd.TreeZ)
	dd.TreeZ = tree
	if tree != nil {
		fixVarTree(tree, dd.PermZ)
	}
}

func Cudd_IsNonConstant(node *DdNode) bool {
	return node != nil && !Cudd_IsConstant(node)
}

func Cudd_ReadPlusInfinity(dd *DdManager) *DdNode {
	if dd == nil || dd.closed {
		return nil
	}
	return dd.PlusInfinity
}
func Cudd_ReadMinusInfinity(dd *DdManager) *DdNode {
	if dd == nil || dd.closed {
		return nil
	}
	return dd.MinusInfinity
}
func Cudd_ReadBackground(dd *DdManager) *DdNode {
	if dd == nil || dd.closed {
		return nil
	}
	return dd.Background
}
func Cudd_SetBackground(dd *DdManager, background *DdNode) {
	if dd != nil && !dd.closed && background != nil && background.manager == dd && background.family == 2 {
		dd.Background = background
	}
}
func Cudd_ReadCacheSlots(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	return CUDD_CACHE_SLOTS
}
func Cudd_ReadCacheUsedSlots(dd *DdManager) float64 {
	if dd == nil || dd.closed {
		return 0
	}
	return float64(len(dd.cache)+len(dd.addCache)) / float64(CUDD_CACHE_SLOTS)
}
func Cudd_ReadCacheLookUps(dd *DdManager) float64 {
	if dd == nil || dd.closed {
		return 0
	}
	return float64(dd.cacheLookups)
}
func Cudd_ReadCacheHits(dd *DdManager) float64 {
	if dd == nil || dd.closed {
		return 0
	}
	return float64(dd.cacheHits)
}
func Cudd_ReadMinHit(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	return uint32(0.5 + 100*dd.minHit/(1+dd.minHit))
}
func Cudd_SetMinHit(dd *DdManager, hitRate uint32) {
	if dd == nil || dd.closed {
		return
	}
	if hitRate >= 100 {
		hitRate = 99
	}
	dd.minHit = float64(hitRate) / float64(100-hitRate)
}
func Cudd_ReadLooseUpTo(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	return dd.looseUpTo
}
func Cudd_SetLooseUpTo(dd *DdManager, looseUpTo uint32) {
	if dd != nil && !dd.closed {
		if looseUpTo == 0 {
			looseUpTo = Cudd_ReadSlots(dd) * DD_MAX_LOOSE_FRACTION
		}
		dd.looseUpTo = looseUpTo
	}
}
func Cudd_ReadMaxCache(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	return dd.maxCache
}
func Cudd_ReadMaxCacheHard(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	return dd.maxCacheHard
}
func Cudd_SetMaxCacheHard(dd *DdManager, size uint32) {
	if dd != nil && !dd.closed {
		if size == 0 {
			size = Cudd_ReadSlots(dd) * DD_MAX_CACHE_FRACTION
		}
		dd.maxCacheHard = size
	}
}
func Cudd_SetMaxCache(dd *DdManager, size uint32) {
	if dd != nil && !dd.closed {
		dd.maxCache = size
	}
}
func Cudd_ReadMaxGrowthAlternate(dd *DdManager) float64 {
	if dd == nil || dd.closed {
		return 0
	}
	return dd.maxGrowthAlt
}
func Cudd_SetMaxGrowthAlternate(dd *DdManager, growth float64) {
	if dd != nil && !dd.closed && growth >= 1 {
		dd.maxGrowthAlt = growth
	}
}
func Cudd_ReadSlots(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	if dd.slotCount != 0 {
		return dd.slotCount
	}
	return uint32((len(dd.Vars) + dd.ZddSize + 1) * CUDD_UNIQUE_SLOTS)
}
func Cudd_ReadKeys(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	return uint32(len(dd.unique) + len(dd.addUnique) + len(dd.zddUnique) + len(dd.constants))
}
func Cudd_ReadUsedSlots(dd *DdManager) float64 {
	slots := Cudd_ReadSlots(dd)
	if slots == 0 {
		return 0
	}
	return float64(Cudd_ReadKeys(dd)) / float64(slots)
}
func Cudd_ExpectedUsedSlots(dd *DdManager) float64 {
	return Cudd_ReadUsedSlots(dd)
}
func Cudd_ReadDead(dd *DdManager) uint32 { return 0 }
func Cudd_ReadMinDead(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	return uint32(len(dd.unique) / 2)
}
func Cudd_ReadGarbageCollections(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	return uint32(dd.garbageCollections)
}
func Cudd_ReadGarbageCollectionTime(dd *DdManager) int64 {
	if dd == nil || dd.closed {
		return 0
	}
	return int64(dd.gcTime / 1e6)
}
func Cudd_ReadNodesFreed(dd *DdManager) float64    { return -1 }
func Cudd_ReadNodesDropped(dd *DdManager) float64  { return -1 }
func Cudd_ReadUniqueLookUps(dd *DdManager) float64 { return -1 }
func Cudd_ReadUniqueLinks(dd *DdManager) float64   { return -1 }
func Cudd_ReadMemoryInUse(dd *DdManager) uint64 {
	if dd == nil || dd.closed {
		return 0
	}
	// Go cannot report per-manager allocations; estimate this graph inventory.
	return uint64(Cudd_ReadNodeCount(dd))*96 + uint64(len(dd.cache)+len(dd.addCache))*48 + dd.linearMemory
}
func Cudd_ReadPeakNodeCount(dd *DdManager) uint64  { return uint64(Cudd_ReadNodeCount(dd)) }
func Cudd_ReadPeakLiveNodeCount(dd *DdManager) int { return Cudd_ReadNodeCount(dd) }
func Cudd_zddReadNodeCount(dd *DdManager) int {
	if dd == nil || dd.closed {
		return 0
	}
	return len(dd.zddUnique)
}
func Cudd_ReadNextReordering(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	return dd.nextDyn
}
func Cudd_SetNextReordering(dd *DdManager, next uint32) {
	if dd != nil && !dd.closed {
		dd.nextDyn = next
	}
}
func Cudd_ReadSwapSteps(dd *DdManager) float64 { return -1 }
func Cudd_ReadMaxLive(dd *DdManager) uint32 {
	if dd == nil || dd.closed {
		return 0
	}
	return dd.maxLive
}
func Cudd_SetMaxLive(dd *DdManager, max uint32) {
	if dd != nil && !dd.closed {
		dd.maxLive = max
	}
}
func Cudd_ReadMaxMemory(dd *DdManager) uint64 {
	if dd == nil || dd.closed {
		return 0
	}
	return dd.maxMemory
}
func Cudd_SetMaxMemory(dd *DdManager, max uint64) {
	if dd != nil && !dd.closed {
		dd.maxMemory = max
	}
}
func validVarIndex(dd *DdManager, index int) bool {
	return dd != nil && !dd.closed && index >= 0 && index < len(dd.Vars)
}
func Cudd_bddBindVar(dd *DdManager, index int) int {
	if !validVarIndex(dd, index) {
		return 0
	}
	if dd.boundVars == nil {
		dd.boundVars = map[int]bool{}
	}
	dd.boundVars[index] = true
	return 1
}
func Cudd_bddUnbindVar(dd *DdManager, index int) int {
	if !validVarIndex(dd, index) {
		return 0
	}
	delete(dd.boundVars, index)
	return 1
}
func Cudd_bddVarIsBound(dd *DdManager, index int) int {
	if !validVarIndex(dd, index) {
		return 0
	}
	if dd.boundVars[index] {
		return 1
	}
	return 0
}
func Cudd_bddSetVariableType(dd *DdManager, index int, kind Cudd_VariableType) int {
	if !validVarIndex(dd, index) || kind < CUDD_VAR_PRIMARY_INPUT || kind > CUDD_VAR_NEXT_STATE {
		return 0
	}
	if dd.variableTypes == nil {
		dd.variableTypes = map[int]Cudd_VariableType{}
	}
	dd.variableTypes[index] = kind
	return 1
}
func Cudd_bddVariableType(dd *DdManager, index int) int {
	if !validVarIndex(dd, index) {
		return -1
	}
	return int(dd.variableTypes[index])
}
func Cudd_bddSetPairIndex(dd *DdManager, index, pair int) int {
	if !validVarIndex(dd, index) {
		return 0
	}
	if dd.pairIndices == nil {
		dd.pairIndices = map[int]int{}
	}
	dd.pairIndices[index] = pair
	return 1
}
func Cudd_bddReadPairIndex(dd *DdManager, index int) int {
	if !validVarIndex(dd, index) {
		return -1
	}
	if pair, ok := dd.pairIndices[index]; ok {
		return pair
	}
	return -1
}
func Cudd_bddSetLazyGroup(dd *DdManager, index int, kind Cudd_LazyGroupType) int {
	if !validVarIndex(dd, index) {
		return 0
	}
	if dd.lazyGroups == nil {
		dd.lazyGroups = map[int]Cudd_LazyGroupType{}
	}
	dd.lazyGroups[index] = kind
	return 1
}
func Cudd_bddLazyGroup(dd *DdManager, index int) int {
	if !validVarIndex(dd, index) {
		return -1
	}
	return int(dd.lazyGroups[index])
}

func Cudd_AutodynEnable(dd *DdManager, method Cudd_ReorderingType) {
	if dd == nil || dd.closed {
		return
	}
	dd.autoDyn = true
	if method != CUDD_REORDER_SAME {
		dd.autoMethod = method
	}
}
func Cudd_AutodynDisable(dd *DdManager) {
	if dd != nil {
		dd.autoDyn = false
	}
}
func Cudd_ReorderingStatus(dd *DdManager) (Cudd_ReorderingType, bool) {
	if dd == nil || dd.closed {
		return CUDD_REORDER_NONE, false
	}
	return dd.autoMethod, dd.autoDyn
}
func Cudd_AutodynEnableZdd(dd *DdManager, method Cudd_ReorderingType) {
	if dd == nil || dd.closed {
		return
	}
	dd.autoDynZdd = true
	if method != CUDD_REORDER_SAME {
		dd.autoMethodZdd = method
	}
}
func Cudd_AutodynDisableZdd(dd *DdManager) {
	if dd != nil {
		dd.autoDynZdd = false
	}
}
func Cudd_ReorderingStatusZdd(dd *DdManager) (Cudd_ReorderingType, bool) {
	if dd == nil || dd.closed {
		return CUDD_REORDER_NONE, false
	}
	return dd.autoMethodZdd, dd.autoDynZdd
}
func Cudd_zddRealignmentEnabled(dd *DdManager) bool { return dd != nil && dd.realignZdd }
func Cudd_zddRealignEnable(dd *DdManager) {
	if dd != nil {
		dd.realignZdd = true
	}
}
func Cudd_zddRealignDisable(dd *DdManager) {
	if dd != nil {
		dd.realignZdd = false
	}
}
func Cudd_bddRealignmentEnabled(dd *DdManager) bool { return dd != nil && dd.realignBdd }
func Cudd_bddRealignEnable(dd *DdManager) {
	if dd != nil {
		dd.realignBdd = true
	}
}
func Cudd_bddRealignDisable(dd *DdManager) {
	if dd != nil {
		dd.realignBdd = false
	}
}
