package core

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"time"

	"github.com/glycerine/abcgo/bdd/mtr"
)

// DdNode is an edge handle. Complemented edges have their own canonical Go
// handle rather than a tag hidden in a pointer. Inspect children on Regular().
// Every regular nonterminal has an uncomplemented then edge.
type DdNode struct {
	Index            uint32
	Ref              uint32
	T, E             *DdNode
	Value            float64
	ID               uint64
	manager          *DdManager
	regular, inverse *DdNode
	family           uint8
	childrenReleased bool
}
type bddUniqueKey struct {
	index uint32
	t, e  *DdNode
}
type bddITEKey struct{ f, g, h *DdNode }

// DdManager owns canonical ROBDD nodes and the ITE computed table. Like CUDD,
// a manager must not be modified concurrently without external synchronization.
type DdManager struct {
	One, LogicZero                    *DdNode
	Zero, PlusInfinity, MinusInfinity *DdNode
	Background                        *DdNode
	ZddSize                           int
	tableBDDSize, tableZddSize        uint32
	initSlots, slotCount              uint32
	bddTableSlots, zddTableSlots      []uint32
	tableGrowthFactor                 uint32
	minDead                           uint32
	gcFraction                        float64
	cacheSlack                        int32
	PermZ, InvPermZ                   []int
	Univ                              []*DdNode
	Tree, TreeZ                       *mtr.MtrNode
	zddUnique                         map[bddUniqueKey]*DdNode
	addUnique                         map[bddUniqueKey]*DdNode
	addCache                          map[bddITEKey]*DdNode
	constants                         map[uint64]*DdNode
	Vars                              []*DdNode
	Perm, InvPerm                     []int
	unique                            map[bddUniqueKey]*DdNode
	cache                             map[bddITEKey]*DdNode
	computed                          map[computedKey]*DdNode
	interact                          []uint64
	nextID                            uint64
	cacheLookups, cacheHits           uint64
	garbageCollections, gcTime        uint64
	epsilon                           float64
	cacheSlots                        uint32
	ReorderCycle                      int32
	groupcheck                        Cudd_AggregationType
	gcEnabled                         int32
	countDead                         uint32
	recomb, symmviolation             int32
	arcviolation, populationSize       int32
	numberXovers                      int32
	closed                            bool
	linearMatrix                      []uint64
	linearSize                        int
	linearMemory                      uint64
	errorCode                         Cudd_ErrorType
	stdout, stderr                    io.Writer
	allocationLimited                 bool
	allocationCeiling                 int
	reorderMethod                     Cudd_ReorderingType
	autoDyn, autoDynZdd               bool
	autoMethod, autoMethodZdd         Cudd_ReorderingType
	realignBdd, realignZdd            bool
	minHit, maxGrowthAlt              float64
	looseUpTo, maxCache, maxCacheHard uint32
	nextDyn, maxLive                  uint32
	maxMemory                         uint64
	boundVars                         map[int]bool
	variableTypes                     map[int]Cudd_VariableType
	pairIndices                       map[int]int
	lazyGroups                        map[int]Cudd_LazyGroupType
	reorderings                       uint
	maxGrowth                         float64
	siftMaxVar, siftMaxSwap           int
	symmetryGroups                    [][]int
	reorderingTime                    time.Duration
	hooks                             map[Cudd_HookType][]DD_HFP
}

// Cudd_InitLinearMatrix initializes the manager's linear transform matrix to
// the identity matrix. The packed representation uses one uint64 per 64
// columns, matching cuddLinear.c on 64-bit platforms.
func Cudd_InitLinearMatrix(dd *DdManager) bool {
	if dd == nil || dd.closed {
		if dd != nil {
			dd.errorCode = CUDD_MEMORY_OUT
		}
		return false
	}
	nvars := len(dd.Vars)
	wordsPerRow := (nvars + 63) / 64
	maxInt := int(^uint(0) >> 1)
	if wordsPerRow != 0 && nvars > maxInt/wordsPerRow {
		dd.errorCode = CUDD_MEMORY_OUT
		return false
	}
	words := wordsPerRow * nvars
	if uint64(words) > ^uint64(0)/8 {
		dd.errorCode = CUDD_MEMORY_OUT
		return false
	}
	linear := make([]uint64, words)
	for i := 0; i < nvars; i++ {
		linear[i*wordsPerRow+i/64] = uint64(1) << uint(i%64)
	}
	dd.linearMatrix = linear
	dd.linearSize = nvars
	dd.linearMemory = uint64(words) * 8
	return true
}

// Cudd_ResizeLinearMatrix preserves the existing rows and adds identity
// entries for variables introduced since the matrix was initialized.
func Cudd_ResizeLinearMatrix(dd *DdManager) bool {
	if dd == nil || dd.closed {
		if dd != nil {
			dd.errorCode = CUDD_MEMORY_OUT
		}
		return false
	}
	nvars := len(dd.Vars)
	oldNvars := dd.linearSize
	oldWordsPerRow := (oldNvars + 63) / 64
	wordsPerRow := (nvars + 63) / 64
	maxInt := int(^uint(0) >> 1)
	if wordsPerRow != 0 && nvars > maxInt/wordsPerRow {
		dd.errorCode = CUDD_MEMORY_OUT
		return false
	}
	words := wordsPerRow * nvars
	if uint64(words) > ^uint64(0)/8 {
		dd.errorCode = CUDD_MEMORY_OUT
		return false
	}
	linear := make([]uint64, words)
	for row := 0; row < oldNvars; row++ {
		copy(linear[row*wordsPerRow:row*wordsPerRow+oldWordsPerRow], dd.linearMatrix[row*oldWordsPerRow:(row+1)*oldWordsPerRow])
	}
	for row := oldNvars; row < nvars; row++ {
		linear[row*wordsPerRow+row/64] = uint64(1) << uint(row%64)
	}
	dd.linearMatrix = linear
	dd.linearSize = nvars
	dd.linearMemory = uint64(words) * 8
	return true
}

// Cudd_LinearMatrixState returns a copy of the packed matrix and its dimension.
func Cudd_LinearMatrixState(dd *DdManager) ([]uint64, int) {
	if dd == nil {
		return nil, 0
	}
	return append([]uint64(nil), dd.linearMatrix...), dd.linearSize
}

// Cudd_ReadLinearMatrixSize returns the dimension currently represented by
// the manager's linear transformation matrix.
func Cudd_ReadLinearMatrixSize(dd *DdManager) int {
	if dd == nil {
		return 0
	}
	return dd.linearSize
}

// Cudd_XorLinearRows replaces row x with row x XOR row y in the packed
// linear transform matrix.
func Cudd_XorLinearRows(dd *DdManager, x, y int) {
	wordsPerRow := (dd.linearSize + 63) / 64
	xstart, ystart := x*wordsPerRow, y*wordsPerRow
	for i := 0; i < wordsPerRow; i++ {
		dd.linearMatrix[xstart+i] ^= dd.linearMatrix[ystart+i]
	}
}

// Cudd_ReadLinear reads one bit from the packed linear transform matrix.
// Like the C API, callers must provide row and column indices in range.
func Cudd_ReadLinear(dd *DdManager, row, column int) int {
	wordsPerRow := (dd.linearSize + 63) / 64
	word := wordsPerRow*row + column/64
	return int((dd.linearMatrix[word] >> uint(column%64)) & 1)
}

// Cudd_PrintLinear writes the matrix in the bit order and row width used by
// cuddLinear.c on 64-bit platforms. It returns 0 if the writer fails.
func Cudd_PrintLinear(dd *DdManager) int {
	wordsPerRow := (dd.linearSize + 63) / 64
	for row := 0; row < dd.linearSize; row++ {
		for wordIndex := 0; wordIndex < wordsPerRow; wordIndex++ {
			word := dd.linearMatrix[row*wordsPerRow+wordIndex]
			for bit := 0; bit < 64; bit++ {
				if _, err := fmt.Fprintf(dd.stdout, "%d", word&1); err != nil {
					return 0
				}
				word >>= 1
			}
		}
		if _, err := fmt.Fprintln(dd.stdout); err != nil {
			return 0
		}
	}
	return 1
}

// Cudd_ReadErrorCode returns the most recently recorded manager error.
func Cudd_ReadErrorCode(dd *DdManager) Cudd_ErrorType {
	if dd == nil {
		return CUDD_INVALID_ARG
	}
	return dd.errorCode
}

// Cudd_ClearErrorCode clears the manager's recorded error.
func Cudd_ClearErrorCode(dd *DdManager) {
	if dd != nil {
		dd.errorCode = CUDD_NO_ERROR
	}
}

// Cudd_SetInteraction and Cudd_TestInteraction maintain the symmetric
// variable interaction relation used by adjacent-swap reordering.
func Cudd_SetInteraction(dd *DdManager, x, y int) {
	if dd == nil || x == y || x < 0 || y < 0 || x >= len(dd.Vars) || y >= len(dd.Vars) {
		return
	}
	if x > y {
		x, y = y, x
	}
	index := ((2*len(dd.Vars)-x-3)*x)/2 + y - 1
	if index < 0 {
		return
	}
	word, bit := index/64, uint(index%64)
	if word >= len(dd.interact) {
		dd.interact = append(dd.interact, make([]uint64, word+1-len(dd.interact))...)
	}
	dd.interact[word] |= uint64(1) << bit
}

func Cudd_TestInteraction(dd *DdManager, x, y int) bool {
	if dd == nil || x == y || x < 0 || y < 0 || x >= len(dd.Vars) || y >= len(dd.Vars) {
		return false
	}
	if x > y {
		x, y = y, x
	}
	index := ((2*len(dd.Vars)-x-3)*x)/2 + y - 1
	word, bit := index/64, uint(index%64)
	return word >= 0 && word < len(dd.interact) && dd.interact[word]&(uint64(1)<<bit) != 0
}

// Cudd_ComputeInteractions rebuilds the interaction relation from the
// support of every canonical BDD node. Including unreferenced canonical
// nodes can add harmless extra interactions, but cannot omit a dependency;
// the reorderer only uses this relation to skip independent swaps.
func Cudd_ComputeInteractions(dd *DdManager) {
	if dd == nil {
		return
	}
	dd.interact = nil
	for _, root := range dd.unique {
		seen := make(map[*DdNode]bool)
		support := make(map[int]bool)
		var visit func(*DdNode)
		visit = func(n *DdNode) {
			n = Cudd_Regular(n)
			if n == nil || Cudd_IsConstant(n) || seen[n] {
				return
			}
			seen[n] = true
			support[int(n.Index)] = true
			visit(n.T)
			visit(n.E)
		}
		visit(root)
		indices := make([]int, 0, len(support))
		for i := range support {
			indices = append(indices, i)
		}
		for i := 0; i < len(indices); i++ {
			for j := i + 1; j < len(indices); j++ {
				Cudd_SetInteraction(dd, indices[i], indices[j])
			}
		}
	}
}

const bddConstantIndex uint32 = CUDD_CONST_INDEX

func NewBDDManager(variables int) *DdManager {
	if variables < 0 || uint64(variables) >= uint64(bddConstantIndex) {
		return nil
	}
	dd := &DdManager{unique: map[bddUniqueKey]*DdNode{}, cache: map[bddITEKey]*DdNode{}, computed: map[computedKey]*DdNode{}, nextID: 2, epsilon: 1e-12, cacheSlots: CUDD_CACHE_SLOTS, maxLive: ^uint32(0), reorderMethod: CUDD_REORDER_SIFT, maxGrowth: 1.1, maxGrowthAlt: 2.2, minHit: 30.0 / 70.0, looseUpTo: 5 * CUDD_UNIQUE_SLOTS, maxCache: 2 * CUDD_CACHE_SLOTS, maxCacheHard: 3 * CUDD_CACHE_SLOTS, siftMaxVar: 1000, siftMaxSwap: 2000000, stdout: os.Stdout, stderr: os.Stderr}
	dd.tableGrowthFactor = 1
	dd.groupcheck = CUDD_GROUP_CHECK7
	dd.gcEnabled = 1
	dd.countDead = ^uint32(0)
	dd.One = &DdNode{Index: bddConstantIndex, Value: 1, ID: 0, manager: dd}
	dd.LogicZero = &DdNode{Index: bddConstantIndex, Value: 0, ID: 1, manager: dd, regular: dd.One, inverse: dd.One}
	dd.One.inverse = dd.LogicZero
	dd.constants = map[uint64]*DdNode{math.Float64bits(1): dd.One}
	dd.zddUnique = map[bddUniqueKey]*DdNode{}
	dd.Zero = dd.constant(0)
	dd.Background = dd.Zero
	dd.PlusInfinity = dd.constant(math.Inf(1))
	dd.MinusInfinity = dd.constant(math.Inf(-1))
	for i := 0; i < variables; i++ {
		Cudd_bddIthVar(dd, i)
	}
	return dd
}

type computedKey struct {
	operation uintptr
	f, g, h   *DdNode
	family    uint8
}

// Cudd_CacheLookup and Cudd_CacheInsert expose the generic computed table
// needed by the translated CUDD operator modules.
func Cudd_CacheLookup(dd *DdManager, operation uintptr, f, g, h *DdNode, family uint8) *DdNode {
	if dd == nil {
		return nil
	}
	key := computedKey{operation: operation, f: f, g: g, h: h, family: family}
	dd.cacheLookups++
	result := dd.computed[key]
	if result != nil {
		dd.cacheHits++
	}
	return result
}

func Cudd_CacheInsert(dd *DdManager, operation uintptr, f, g, h, result *DdNode, family uint8) {
	if dd == nil || result == nil {
		return
	}
	if dd.computed == nil {
		dd.computed = make(map[computedKey]*DdNode)
	}
	if dd.maxCacheHard > 0 && uint32(len(dd.computed)) >= dd.maxCacheHard {
		clear(dd.computed)
	}
	dd.computed[computedKey{operation: operation, f: f, g: g, h: h, family: family}] = result
}

func Cudd_CacheFlush(dd *DdManager) {
	if dd != nil {
		clear(dd.computed)
	}
}

func Cudd_CacheSize(dd *DdManager) int {
	if dd == nil {
		return 0
	}
	return len(dd.computed)
}
func Cudd_Regular(n *DdNode) *DdNode {
	if n == nil {
		return nil
	}
	if n.regular != nil {
		return n.regular
	}
	return n
}
func Cudd_IsComplement(n *DdNode) bool { return n != nil && n.regular != nil }
func Cudd_Not(n *DdNode) *DdNode {
	if n == nil {
		return nil
	}
	return n.inverse
}
func Cudd_NotCond(n *DdNode, condition bool) *DdNode {
	if condition {
		return Cudd_Not(n)
	}
	return n
}
func Cudd_IsConstant(n *DdNode) bool           { return n != nil && Cudd_Regular(n).Index == bddConstantIndex }
func Cudd_T(n *DdNode) *DdNode                 { return Cudd_Regular(n).T }
func Cudd_E(n *DdNode) *DdNode                 { return Cudd_Regular(n).E }
func Cudd_NodeReadIndex(n *DdNode) uint32      { return Cudd_Regular(n).Index }
func Cudd_ReadOne(dd *DdManager) *DdNode       { return dd.One }
func Cudd_ReadLogicZero(dd *DdManager) *DdNode { return dd.LogicZero }
func Cudd_ReadSize(dd *DdManager) int          { return len(dd.Vars) }
func Cudd_ReadPerm(dd *DdManager, index int) int {
	if index < 0 || index >= len(dd.Perm) {
		return -1
	}
	return dd.Perm[index]
}
func Cudd_ReadInvPerm(dd *DdManager, level int) int {
	if level < 0 || level >= len(dd.InvPerm) {
		return -1
	}
	return dd.InvPerm[level]
}
func Cudd_ReadNodeCount(dd *DdManager) int {
	return len(dd.unique) + len(dd.zddUnique) + len(dd.addUnique) + len(dd.constants)
}

func Cudd_ReadEpsilon(dd *DdManager) float64 {
	if dd == nil || dd.closed {
		return 0
	}
	return dd.epsilon
}

func Cudd_SetEpsilon(dd *DdManager, epsilon float64) {
	if dd != nil && !dd.closed {
		dd.epsilon = epsilon
	}
}

// Cudd_ExternalRefCount subtracts references held by live unique-node edges.
// Manager internals use this to identify roots separately from child links.
func Cudd_ExternalRefCount(dd *DdManager, node *DdNode) uint32 {
	if dd == nil || dd.closed || node == nil || node.manager != dd {
		return 0
	}
	node = Cudd_Regular(node)
	if node.Ref == ^uint32(0) {
		return node.Ref
	}
	var incoming uint64
	for _, table := range []map[bddUniqueKey]*DdNode{dd.unique, dd.addUnique, dd.zddUnique} {
		for _, parent := range table {
			if parent.childrenReleased {
				continue
			}
			if Cudd_Regular(parent.T) == node {
				incoming++
			}
			if Cudd_Regular(parent.E) == node {
				incoming++
			}
		}
	}
	if incoming >= uint64(node.Ref) {
		return 0
	}
	return node.Ref - uint32(incoming)
}

// Cudd_CheckZeroRef reports canonical nodes that still have references from
// outside manager-owned edges and terminals. This is the Go graph equivalent
// of the CUDD leak check used by testcudd after it releases its temporary roots.
func Cudd_CheckZeroRef(dd *DdManager) int32 {
	if dd == nil || dd.closed {
		return 0
	}
	seen := make(map[*DdNode]struct{})
	count := int32(0)
	check := func(table map[bddUniqueKey]*DdNode) {
		for _, node := range table {
			node = Cudd_Regular(node)
			if _, ok := seen[node]; ok {
				continue
			}
			seen[node] = struct{}{}
			if Cudd_ExternalRefCount(dd, node) != 0 {
				count++
			}
		}
	}
	check(dd.unique)
	check(dd.addUnique)
	check(dd.zddUnique)
	for _, node := range dd.constants {
		node = Cudd_Regular(node)
		if _, ok := seen[node]; ok {
			continue
		}
		seen[node] = struct{}{}
		if Cudd_ExternalRefCount(dd, node) != 0 {
			count++
		}
	}
	return count
}

// Cudd_InitTable creates unique-table state before variables are added. Go
// maps replace the C subtables and slabs; capacities and slot thresholds are
// retained because they affect public statistics and later growth policy.
func Cudd_InitTable(numVars, numVarsZ, numSlots, looseUpTo uint32) *DdManager {
	if numVars >= CUDD_MAXINDEX || numVarsZ >= CUDD_MAXINDEX {
		return nil
	}
	slots := uint32(8)
	for slots < numSlots {
		if slots > ^uint32(0)/2 {
			return nil
		}
		slots <<= 1
	}
	totalSlots := (uint64(numVars) + uint64(numVarsZ) + 1) * uint64(slots)
	if totalSlots > uint64(^uint32(0)) {
		return nil
	}
	dd := NewBDDManager(0)
	if dd == nil {
		return nil
	}
	dd.ZddSize = int(numVarsZ)
	dd.PermZ, dd.InvPermZ = make([]int, numVarsZ), make([]int, numVarsZ)
	for i := range dd.PermZ {
		dd.PermZ[i], dd.InvPermZ[i] = i, i
	}
	dd.tableBDDSize, dd.tableZddSize = numVars, numVarsZ
	dd.initSlots, dd.slotCount = slots, uint32(totalSlots)
	dd.gcFraction = DD_GC_FRAC_HI
	dd.minDead = uint32(dd.gcFraction * float64(dd.slotCount))
	dd.bddTableSlots = make([]uint32, numVars)
	dd.zddTableSlots = make([]uint32, numVarsZ)
	for i := range dd.bddTableSlots {
		dd.bddTableSlots[i] = slots
	}
	for i := range dd.zddTableSlots {
		dd.zddTableSlots[i] = slots
	}
	dd.looseUpTo = looseUpTo
	return dd
}

// Cudd_AllocNode allocates a manager-owned, initially empty node handle.
// Go's heap replaces CUDD's platform-specific aligned slab; this operation
// retains the manager's node-limit error behavior and ID accounting.
func Cudd_AllocNode(dd *DdManager) *DdNode {
	if dd == nil || dd.closed {
		return nil
	}
	if dd.maxLive != 0 && uint32(Cudd_ReadNodeCount(dd)) > dd.maxLive {
		dd.errorCode = CUDD_TOO_MANY_NODES
		return nil
	}
	n := &DdNode{ID: dd.nextID, manager: dd}
	dd.nextID++
	return n
}
func (dd *DdManager) valid(nodes ...*DdNode) bool {
	if dd == nil || dd.closed {
		return false
	}
	for _, n := range nodes {
		if n == nil || n.manager != dd || Cudd_Regular(n).family != 0 || (Cudd_IsConstant(n) && Cudd_Regular(n) != dd.One) {
			return false
		}
	}
	return true
}
func (dd *DdManager) node(index uint32, t, e *DdNode) *DdNode {
	if t == e {
		return t
	}
	complemented := Cudd_IsComplement(t)
	if complemented {
		t = Cudd_Not(t)
		e = Cudd_Not(e)
	}
	key := bddUniqueKey{index, t, e}
	n := dd.unique[key]
	if n == nil {
		if dd.allocationLimited && len(dd.unique) >= dd.allocationCeiling {
			panic(bddAllocationLimit{})
		}
		n = &DdNode{Index: index, T: t, E: e, ID: dd.nextID, manager: dd}
		n.inverse = &DdNode{Index: index, ID: dd.nextID + 1, manager: dd, regular: n, inverse: n}
		dd.nextID += 2
		Cudd_Ref(t)
		Cudd_Ref(e)
		dd.unique[key] = n
	} else {
		reclaimNode(n)
	}
	return Cudd_NotCond(n, complemented)
}

// Cudd_BddNode returns the canonical BDD node for an index and children.
func Cudd_BddNode(dd *DdManager, index int, t, e *DdNode) *DdNode {
	if !dd.valid(t, e) || index < 0 || index >= len(dd.Vars) || dd.Perm[index] >= dd.level(t) || dd.Perm[index] >= dd.level(e) {
		return nil
	}
	return dd.node(uint32(index), t, e)
}

func Cudd_bddIthVar(dd *DdManager, index int) *DdNode {
	if !dd.valid() || index < 0 || uint64(index) >= uint64(bddConstantIndex) {
		return nil
	}
	for len(dd.Vars) <= index {
		i := len(dd.Vars)
		dd.Perm = append(dd.Perm, i)
		dd.InvPerm = append(dd.InvPerm, i)
		dd.Vars = append(dd.Vars, dd.node(uint32(i), dd.One, dd.LogicZero))
	}
	return dd.Vars[index]
}
func Cudd_bddNewVar(dd *DdManager) *DdNode { return Cudd_bddIthVar(dd, len(dd.Vars)) }
func (dd *DdManager) level(n *DdNode) int {
	r := Cudd_Regular(n)
	if r.Index == bddConstantIndex {
		return math.MaxInt
	}
	return dd.Perm[r.Index]
}
func (dd *DdManager) cofactors(n *DdNode, level int) (*DdNode, *DdNode) {
	if dd.level(n) != level {
		return n, n
	}
	r := Cudd_Regular(n)
	return Cudd_NotCond(r.T, Cudd_IsComplement(n)), Cudd_NotCond(r.E, Cudd_IsComplement(n))
}
func (dd *DdManager) ite(f, g, h *DdNode) *DdNode {
	return dd.iteChecked(f, g, h, nil)
}

func (dd *DdManager) iteChecked(f, g, h *DdNode, check func() bool) *DdNode {
	if check != nil && !check() {
		return nil
	}
	if f == dd.One {
		return g
	}
	if f == dd.LogicZero {
		return h
	}
	if g == h {
		return g
	}
	if g == dd.One && h == dd.LogicZero {
		return f
	}
	if g == dd.LogicZero && h == dd.One {
		return Cudd_Not(f)
	}
	if Cudd_IsComplement(f) {
		f = Cudd_Not(f)
		g, h = h, g
	}
	invert := Cudd_IsComplement(g)
	if invert {
		g = Cudd_Not(g)
		h = Cudd_Not(h)
	}
	key := bddITEKey{f, g, h}
	dd.cacheLookups++
	if n := dd.cache[key]; n != nil {
		dd.cacheHits++
		return Cudd_NotCond(n, invert)
	}
	level := min(dd.level(f), min(dd.level(g), dd.level(h)))
	ft, fe := dd.cofactors(f, level)
	gt, ge := dd.cofactors(g, level)
	ht, he := dd.cofactors(h, level)
	t := dd.iteChecked(ft, gt, ht, check)
	if t == nil {
		return nil
	}
	e := dd.iteChecked(fe, ge, he, check)
	if e == nil {
		return nil
	}
	n := dd.node(uint32(dd.InvPerm[level]), t, e)
	dd.cache[key] = n
	return Cudd_NotCond(n, invert)
}

func Cudd_Ref(n *DdNode) {
	n = Cudd_Regular(n)
	if n != nil {
		reclaimNode(n)
		if n.Ref < math.MaxUint32 {
			n.Ref++
		}
	}
}

func reclaimNode(n *DdNode) {
	if n == nil || !n.childrenReleased {
		return
	}
	n.childrenReleased = false
	Cudd_Ref(n.T)
	Cudd_Ref(n.E)
}
func Cudd_Deref(n *DdNode) {
	n = Cudd_Regular(n)
	if n != nil && n.Ref > 0 && n.Ref < math.MaxUint32 {
		n.Ref--
	}
}
func Cudd_RecursiveDeref(dd *DdManager, n *DdNode) {
	if dd == nil || dd.closed || n == nil || n.manager != dd {
		return
	}
	n = Cudd_Regular(n)
	if n.Ref == 0 || n.Ref == math.MaxUint32 {
		return
	}
	n.Ref--
	if n.Ref == 0 && !n.childrenReleased {
		n.childrenReleased = true
		Cudd_RecursiveDeref(dd, n.T)
		Cudd_RecursiveDeref(dd, n.E)
	}
}

// Evaluate follows complemented branches without creating nodes.
func (dd *DdManager) Evaluate(n *DdNode, assignment []bool) (bool, error) {
	if !dd.valid(n) {
		return false, fmt.Errorf("cudd: foreign, nil or closed-manager edge")
	}
	for !Cudd_IsConstant(n) {
		r := Cudd_Regular(n)
		if int(r.Index) >= len(assignment) {
			return false, fmt.Errorf("cudd: missing variable %d", r.Index)
		}
		next := r.E
		if assignment[r.Index] {
			next = r.T
		}
		n = Cudd_NotCond(next, Cudd_IsComplement(n))
	}
	return n == dd.One, nil
}

// Support returns variable indices in manager level order.
func (dd *DdManager) Support(n *DdNode) []int {
	if !(dd.valid(n) || dd.validAdd(n)) {
		return nil
	}
	seen := map[*DdNode]bool{}
	indices := map[int]bool{}
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		n = Cudd_Regular(n)
		if seen[n] || Cudd_IsConstant(n) {
			return
		}
		seen[n] = true
		indices[int(n.Index)] = true
		visit(n.T)
		visit(n.E)
	}
	visit(n)
	result := make([]int, 0, len(indices))
	for i := range indices {
		result = append(result, i)
	}
	sort.Slice(result, func(i, j int) bool { return dd.Perm[result[i]] < dd.Perm[result[j]] })
	return result
}
func Cudd_Support(dd *DdManager, f *DdNode) *DdNode {
	if !(dd.valid(f) || dd.validAdd(f)) {
		return nil
	}
	cube := dd.One
	support := dd.Support(f)
	for i := len(support) - 1; i >= 0; i-- {
		cube = Cudd_bddAnd(dd, dd.Vars[support[i]], cube)
	}
	return cube
}
func Cudd_SupportSize(dd *DdManager, f *DdNode) int { return len(dd.Support(f)) }

// Cudd_bddPermute performs simultaneous substitution. ITE reconstructs the
// canonical order even when the supplied permutation reverses variable levels.
func Cudd_bddPermute(dd *DdManager, f *DdNode, permutation []int) *DdNode {
	if !dd.valid(f) {
		return nil
	}
	highest := -1
	for _, i := range dd.Support(f) {
		if i >= len(permutation) || permutation[i] < 0 || uint64(permutation[i]) >= uint64(bddConstantIndex) {
			return nil
		}
		highest = max(highest, permutation[i])
	}
	if highest >= 0 && Cudd_bddIthVar(dd, highest) == nil {
		return nil
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if Cudd_IsConstant(n) {
			return n
		}
		if v := memo[n]; v != nil {
			return v
		}
		r := Cudd_Regular(n)
		t, e := visit(r.T), visit(r.E)
		v := dd.ite(dd.Vars[permutation[r.Index]], t, e)
		v = Cudd_NotCond(v, Cudd_IsComplement(n))
		memo[n] = v
		return v
	}
	return visit(f)
}
func Cudd_bddCompose(dd *DdManager, f, g *DdNode, index int) *DdNode {
	if !dd.valid(f, g) || index < 0 {
		return nil
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if Cudd_IsConstant(n) {
			return n
		}
		if v := memo[n]; v != nil {
			return v
		}
		r := Cudd_Regular(n)
		t, e := visit(r.T), visit(r.E)
		variable := dd.Vars[r.Index]
		if int(r.Index) == index {
			variable = g
		}
		v := Cudd_NotCond(dd.ite(variable, t, e), Cudd_IsComplement(n))
		memo[n] = v
		return v
	}
	return visit(f)
}
func Cudd_Cofactor(dd *DdManager, f, cube *DdNode) *DdNode {
	if !dd.valid(f, cube) || cube == dd.LogicZero {
		return nil
	}
	assignments := map[uint32]bool{}
	for cube != dd.One {
		if Cudd_IsConstant(cube) {
			return nil
		}
		r := Cudd_Regular(cube)
		t, e := dd.cofactors(cube, dd.level(cube))
		if e == dd.LogicZero {
			assignments[r.Index] = true
			cube = t
		} else if t == dd.LogicZero {
			assignments[r.Index] = false
			cube = e
		} else {
			return nil
		}
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if Cudd_IsConstant(n) {
			return n
		}
		if v := memo[n]; v != nil {
			return v
		}
		r := Cudd_Regular(n)
		var v *DdNode
		if value, ok := assignments[r.Index]; ok {
			next := r.E
			if value {
				next = r.T
			}
			v = visit(next)
		} else {
			v = dd.node(r.Index, visit(r.T), visit(r.E))
		}
		v = Cudd_NotCond(v, Cudd_IsComplement(n))
		memo[n] = v
		return v
	}
	return visit(f)
}
func (dd *DdManager) abstract(f, cube *DdNode, universal bool) *DdNode {
	if !dd.valid(f, cube) {
		return nil
	}
	if !dd.positiveCube(cube) {
		return nil
	}
	quantified := map[uint32]bool{}
	for _, i := range dd.Support(cube) {
		quantified[uint32(i)] = true
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if Cudd_IsConstant(n) {
			return n
		}
		if v := memo[n]; v != nil {
			return v
		}
		r := Cudd_Regular(n)
		t, e := dd.cofactors(n, dd.level(n))
		t, e = visit(t), visit(e)
		var v *DdNode
		if quantified[r.Index] {
			if universal {
				v = Cudd_bddAnd(dd, t, e)
			} else {
				v = Cudd_bddOr(dd, t, e)
			}
		} else {
			v = dd.node(r.Index, t, e)
		}
		memo[n] = v
		return v
	}
	return visit(f)
}

func Cudd_bddAndAbstract(dd *DdManager, f, g, cube *DdNode) *DdNode {
	if dd == nil || !dd.valid(f, g, cube) {
		return nil
	}
	for c := cube; c != dd.One; c = c.T {
		if Cudd_IsConstant(c) || Cudd_IsComplement(c) || c.E != dd.LogicZero {
			return nil
		}
	}
	return dd.andAbstractChecked(f, g, cube, nil)
}
func Cudd_bddAndAbstractLimit(dd *DdManager, f, g, cube *DdNode, limit uint32) (result *DdNode) {
	if !dd.valid(f, g, cube) {
		return nil
	}
	oldLimited, oldCeiling := dd.allocationLimited, dd.allocationCeiling
	dd.allocationLimited = true
	dd.allocationCeiling = len(dd.unique) + int(limit)
	defer func() {
		dd.allocationLimited, dd.allocationCeiling = oldLimited, oldCeiling
		if recover() != nil {
			result = nil
		}
	}()
	return Cudd_bddAndAbstract(dd, f, g, cube)
}
func Cudd_DagSize(f *DdNode) int {
	if f == nil {
		return 0
	}
	seen := map[*DdNode]bool{}
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		n = Cudd_Regular(n)
		if seen[n] {
			return
		}
		seen[n] = true
		if !Cudd_IsConstant(n) {
			visit(n.T)
			visit(n.E)
		}
	}
	visit(f)
	return len(seen)
}
func Cudd_CountMinterm(dd *DdManager, f *DdNode, nvars int) float64 {
	if !(dd.valid(f) || dd.validAdd(f)) || nvars < 0 {
		return math.NaN()
	}
	memo := map[*DdNode]float64{}
	var probability func(*DdNode) float64
	probability = func(n *DdNode) float64 {
		if p, ok := memo[n]; ok {
			return p
		}
		r := Cudd_Regular(n)
		var p float64
		if Cudd_IsConstant(r) {
			if r != dd.Zero && r != dd.Background {
				p = 1
			}
		} else {
			p = (probability(r.T) + probability(r.E)) / 2
		}
		if Cudd_IsComplement(n) {
			p = 1 - p
		}
		memo[n] = p
		return p
	}
	return math.Ldexp(probability(f), nvars)
}

func (dd *DdManager) constant(value float64) *DdNode {
	if value == 0 {
		value = 0
	} // Canonicalize IEEE negative zero.
	key := math.Float64bits(value)
	if n := dd.constants[key]; n != nil {
		return n
	}
	for _, n := range dd.constants {
		if n.Value == value || math.Abs(n.Value-value) < dd.epsilon {
			reclaimNode(n)
			return n
		}
	}
	n := &DdNode{Index: bddConstantIndex, Value: value, ID: dd.nextID, manager: dd}
	n.inverse = &DdNode{Index: bddConstantIndex, ID: dd.nextID + 1, manager: dd, regular: n, inverse: n}
	dd.nextID += 2
	dd.constants[key] = n
	return n
}
func (dd *DdManager) zddNode(index uint32, t, e *DdNode) *DdNode {
	if t == dd.Zero {
		return e
	}
	key := bddUniqueKey{index, t, e}
	if n := dd.zddUnique[key]; n != nil {
		reclaimNode(n)
		return n
	}
	n := &DdNode{Index: index, T: t, E: e, ID: dd.nextID, manager: dd, family: 1}
	n.inverse = &DdNode{Index: index, ID: dd.nextID + 1, manager: dd, regular: n, inverse: n}
	dd.nextID += 2
	Cudd_Ref(t)
	Cudd_Ref(e)
	dd.zddUnique[key] = n
	return n
}
