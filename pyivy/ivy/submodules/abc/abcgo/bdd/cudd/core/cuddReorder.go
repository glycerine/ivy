package core

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// Dynamic ordering uses CUDD's sifting and symmetric grouping definitions from
// cuddReorder.c/cuddSymmetry.c. Scratch canonical tables replace raw in-place
// unique-subtable mutation. Referenced roots and projection handles keep identity;
// unreachable zero-reference nodes are collected during preprocessing, as in C.
type reorderRoots struct {
	active []*DdNode
	pins   []*DdNode
}
type reorderHeap struct {
	dd           *DdManager
	active, pins []*DdNode
	memo         map[*DdNode]*DdNode
	nodes        map[*DdNode]bool
	external     map[*DdNode]uint32
	incoming     map[*DdNode]int
	levels       []int
	cost         int
}

func (dd *DdManager) liveReorderRoots() reorderRoots {
	roots := reorderRoots{pins: append([]*DdNode(nil), dd.Vars...)}
	for _, table := range []map[bddUniqueKey]*DdNode{dd.unique, dd.addUnique} {
		for _, node := range table {
			if Cudd_ExternalRefCount(dd, node) != 0 {
				roots.active = append(roots.active, node)
			}
		}
	}
	for _, node := range dd.constants {
		if Cudd_ExternalRefCount(dd, node) != 0 {
			roots.active = append(roots.active, node)
		}
	}
	sort.Slice(roots.active, func(i, j int) bool { return roots.active[i].ID < roots.active[j].ID })
	return roots
}
func reorderBuild(dd *DdManager, roots reorderRoots, order []int) *reorderHeap {
	scratch := NewBDDManager(len(dd.Vars))
	copy(scratch.InvPerm, order)
	for level, index := range order {
		scratch.Perm[index] = level
	}
	h := &reorderHeap{dd: scratch, memo: map[*DdNode]*DdNode{}, nodes: map[*DdNode]bool{}, external: map[*DdNode]uint32{}, incoming: map[*DdNode]int{}, levels: make([]int, len(order))}
	var transfer func(*DdNode) *DdNode
	transfer = func(edge *DdNode) *DdNode {
		old := Cudd_Regular(edge)
		fresh := h.memo[old]
		if fresh == nil {
			if Cudd_IsConstant(old) {
				fresh = scratch.constant(old.Value)
			} else {
				t, e := transfer(old.T), transfer(old.E)
				if old.family == 2 {
					fresh = Cudd_addIte(scratch, Cudd_addIthVar(scratch, int(old.Index)), t, e)
				} else {
					fresh = scratch.ite(scratch.Vars[old.Index], t, e)
				}
			}
			// Canonical BDD polarity equals truth on the all-ones assignment, so pure
			// variable reordering never complements a previously regular function.
			if Cudd_IsComplement(fresh) {
				panic("cudd: reordering changed canonical polarity")
			}
			h.memo[old] = fresh
		}
		return Cudd_NotCond(fresh, Cudd_IsComplement(edge))
	}
	for _, root := range roots.active {
		fresh := transfer(root)
		h.active = append(h.active, fresh)
		external := Cudd_ExternalRefCount(dd, root)
		h.external[fresh] += external
	}
	for _, pin := range roots.pins {
		h.pins = append(h.pins, transfer(pin))
	}
	var collect func(*DdNode, map[*DdNode]bool)
	collect = func(edge *DdNode, seen map[*DdNode]bool) {
		node := Cudd_Regular(edge)
		if seen[node] {
			return
		}
		seen[node] = true
		if !Cudd_IsConstant(node) {
			collect(node.T, seen)
			collect(node.E, seen)
		}
	}
	for _, root := range h.active {
		collect(root, h.nodes)
	}
	// Cudd_Init explicitly pins One, Zero, both infinities and the background.
	// They contribute a fixed heap offset to growth bounds even when roots do not
	// reach them; ignoring this changes the source's maxGrowth pruning threshold.
	for _, constant := range []*DdNode{dd.One, dd.Zero, dd.PlusInfinity, dd.MinusInfinity, dd.Background} {
		if constant != nil {
			collect(transfer(constant), h.nodes)
		}
	}
	for _, pin := range h.pins {
		collect(pin, h.nodes)
	}
	for node := range h.nodes {
		if Cudd_IsConstant(node) {
			continue
		}
		h.incoming[Cudd_Regular(node.T)]++
		h.incoming[Cudd_Regular(node.E)]++
	}
	for node := range h.nodes {
		if Cudd_IsConstant(node) {
			continue
		}
		h.levels[scratch.Perm[node.Index]]++
		if !reorderIsIsolatedProjection(h, node) {
			h.cost++
		}
	}
	return h
}

// CUDD's sifting objective is keys-isolated: constants are outside the BDD
// unique-table key count, and a projection with no external or parent refs is
// counted as an isolated key and removed from the objective.
func reorderIsIsolatedProjection(h *reorderHeap, node *DdNode) bool {
	index := int(node.Index)
	return index >= 0 && index < len(h.dd.Vars) && h.dd.Vars[index] == node &&
		h.external[node] == 0 && h.incoming[node] == 0
}
func reorderValidOrder(order []int, n int) bool {
	if len(order) != n {
		return false
	}
	seen := make([]bool, n)
	for _, v := range order {
		if v < 0 || v >= n || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}

// reorderCommit prepares all replacement edges before mutating any existing
// handle. Constants, referenced nodes and BDD projection variables retain their
// canonical pointers. Old zero-reference nodes outside the live heap are dead.
func reorderCommit(dd *DdManager, roots reorderRoots, h *reorderHeap, order []int) {
	mapped := map[*DdNode]*DdNode{}
	usedOld := map[*DdNode]bool{}
	for node := range h.nodes {
		if Cudd_IsConstant(node) {
			mapped[node] = dd.constant(node.Value)
		}
	}
	preserve := func(old *DdNode) {
		fresh := h.memo[old]
		if fresh != nil {
			mapped[fresh] = old
			usedOld[old] = true
		}
	}
	for _, old := range roots.active {
		preserve(old)
	}
	for _, old := range roots.pins {
		preserve(old)
	}
	for old, fresh := range h.memo {
		if h.nodes[fresh] && mapped[fresh] == nil {
			mapped[fresh] = old
			usedOld[old] = true
		}
	}
	// Assign handles first, so child wiring is independent of traversal order.
	for fresh := range h.nodes {
		if mapped[fresh] != nil {
			continue
		}
		node := &DdNode{manager: dd, ID: dd.nextID, family: fresh.family}
		dd.nextID++
		if fresh.family == 0 && !Cudd_IsConstant(fresh) {
			node.inverse = &DdNode{manager: dd, ID: dd.nextID, regular: node, inverse: node}
			dd.nextID++
		}
		mapped[fresh] = node
	}
	edge := func(fresh *DdNode) *DdNode {
		return Cudd_NotCond(mapped[Cudd_Regular(fresh)], Cudd_IsComplement(fresh))
	}
	// Save the old tables for dead-handle invalidation before replacing them.
	oldBDD, oldADD, oldConstants := dd.unique, dd.addUnique, dd.constants
	nextBDD, nextADD := map[bddUniqueKey]*DdNode{}, map[bddUniqueKey]*DdNode{}
	nextConstants := map[uint64]*DdNode{}
	for fresh, node := range mapped {
		if Cudd_IsConstant(fresh) {
			nextConstants[math.Float64bits(fresh.Value)] = node
			continue
		}
		node.Index = fresh.Index
		node.T = edge(fresh.T)
		node.E = edge(fresh.E)
		node.family = fresh.family
		node.manager = dd
		if node.inverse != nil {
			node.inverse.Index = node.Index
			node.inverse.manager = dd
		}
		key := bddUniqueKey{node.Index, node.T, node.E}
		if node.family == 2 {
			nextADD[key] = node
		} else {
			nextBDD[key] = node
		}
	}
	// Scratch rebuilding changes child edges and may reuse an old handle for a
	// different internal node. Rebuild its reference count from retained owners
	// and the new incoming edges instead of keeping counts from the old DAG.
	// ZDD edges are outside this reorder and still own their constant children.
	zddIncoming := map[*DdNode]uint64{}
	for _, parent := range dd.zddUnique {
		if !parent.childrenReleased {
			zddIncoming[Cudd_Regular(parent.T)]++
			zddIncoming[Cudd_Regular(parent.E)]++
		}
	}
	for fresh, node := range mapped {
		refs := uint64(h.external[fresh]) + uint64(h.incoming[fresh]) + zddIncoming[node]
		node.Ref = uint32(min(refs, uint64(math.MaxUint32)))
		node.childrenReleased = false
	}
	for _, table := range []map[bddUniqueKey]*DdNode{oldBDD, oldADD} {
		for _, old := range table {
			if !usedOld[old] {
				old.manager = nil
				if old.inverse != nil {
					old.inverse.manager = nil
				}
			}
		}
	}
	for _, old := range oldConstants {
		if nextConstants[math.Float64bits(old.Value)] != old {
			old.manager = nil
			if old.inverse != nil {
				old.inverse.manager = nil
			}
		}
	}
	dd.constants = nextConstants
	dd.unique = nextBDD
	dd.addUnique = nextADD
	clear(dd.cache)
	clear(dd.addCache)
	copy(dd.InvPerm, order)
	for level, index := range order {
		dd.Perm[index] = level
	}
}

// Cudd_ShuffleHeap takes an index at each new level, not an index-to-level map.
// It preserves functions and referenced root handles while changing Perm/InvPerm.
func Cudd_ShuffleHeap(dd *DdManager, order []int) int {
	if dd == nil || dd.closed || !reorderValidOrder(order, len(dd.Vars)) {
		return 0
	}
	identical := true
	for i, v := range order {
		if dd.InvPerm[i] != v {
			identical = false
			break
		}
	}
	if identical {
		return 1
	}
	roots := dd.liveReorderRoots()
	heap := reorderBuild(dd, roots, order)
	reorderCommit(dd, roots, heap, order)
	dd.symmetryGroups = nil
	return 1
}

// adjacentSymmetry implements CUDD's positive/negative cofactor equalities plus
// its incoming-arc condition. The latter prevents grouping when a y-level node
// is referenced independently of the x level. Isolated projection pins do not
// create false positives or prohibit otherwise valid groups.
func (h *reorderHeap) adjacentSymmetry(level int) bool {
	if level < 0 || level+1 >= len(h.dd.InvPerm) || h.levels[level] <= 1 {
		return false
	}
	x, y := h.dd.InvPerm[level], h.dd.InvPerm[level+1]
	yvar := h.dd.Vars[y]
	if h.levels[level+1] == 1 && h.external[yvar] == 0 && h.incoming[yvar] == 0 {
		return false
	}
	positive, negative := true, true
	arcs, total := 0, -1
	cofactor := func(edge *DdNode, high bool) *DdNode {
		node := Cudd_Regular(edge)
		if Cudd_IsConstant(node) || int(node.Index) != y {
			return edge
		}
		child := node.E
		if high {
			child = node.T
		}
		return Cudd_NotCond(child, Cudd_IsComplement(edge))
	}
	for node := range h.nodes {
		if Cudd_IsConstant(node) {
			continue
		}
		if int(node.Index) == y {
			total += h.incoming[node] + int(h.external[node])
			if node == yvar {
				total++
			}
		}
		if int(node.Index) != x {
			continue
		}
		if !Cudd_IsConstant(node.T) && int(Cudd_Regular(node.T).Index) == y {
			arcs++
		}
		if !Cudd_IsConstant(node.E) && int(Cudd_Regular(node.E).Index) == y {
			arcs++
		}
		if node == h.dd.Vars[x] && h.incoming[node] == 0 && h.external[node] == 0 {
			continue
		}
		positive = positive && cofactor(node.E, true) == cofactor(node.T, false)
		negative = negative && cofactor(node.T, true) == cofactor(node.E, false)
		if !positive && !negative {
			return false
		}
	}
	return arcs == total && (positive || negative)
}
func cloneGroups(groups [][]int) [][]int {
	result := make([][]int, len(groups))
	for i, g := range groups {
		result[i] = append([]int(nil), g...)
	}
	return result
}
func groupOrder(groups [][]int) []int {
	var result []int
	for _, g := range groups {
		result = append(result, g...)
	}
	return result
}
func mergeSymmetric(groups [][]int, heap *reorderHeap) [][]int {
	result := [][]int{}
	level := 0
	for _, group := range groups {
		if len(result) > 0 && heap.adjacentSymmetry(level-1) {
			result[len(result)-1] = append(result[len(result)-1], group...)
		} else {
			result = append(result, append([]int(nil), group...))
		}
		level += len(group)
	}
	return result
}

// Cudd_ReduceHeap implements sifting and symmetric sifting, optionally to
// convergence. Other heuristic enum values return failure instead of silently
// selecting a different algorithm. Public success is 1 (the source tree
// dispatch normalizes the internal symmetric helper's group count).
func Cudd_ReduceHeap(dd *DdManager, method Cudd_ReorderingType, minsize int) int {
	if dd == nil || dd.closed || minsize < 0 {
		return 0
	}
	if method == CUDD_REORDER_SAME {
		method = dd.reorderMethod
	}
	if method == CUDD_REORDER_NONE {
		return 1
	}
	if method == CUDD_REORDER_ANNEALING {
		if len(dd.Vars) < 2 {
			return 1
		}
		return Cudd_AnnealRange(dd, 0, len(dd.Vars)-1, minsize)
	}
	switch method {
	case CUDD_REORDER_WINDOW2, CUDD_REORDER_WINDOW3, CUDD_REORDER_WINDOW4,
		CUDD_REORDER_WINDOW2_CONV, CUDD_REORDER_WINDOW3_CONV, CUDD_REORDER_WINDOW4_CONV:
		if len(dd.Vars) == 0 {
			return 0
		}
		if windowPreprocessSize(dd) < minsize {
			return 1
		}
		started, proceed := Cudd_BeginBDDReordering(dd, method)
		if !proceed {
			return 0
		}
		clear(dd.cache)
		clear(dd.addCache)
		clear(dd.computed)
		return boolToInt(Cudd_EndBDDReordering(dd, started, Cudd_WindowReorder(dd, 0, len(dd.Vars)-1, method) != 0))
	}
	symmetric := method == CUDD_REORDER_SYMM_SIFT || method == CUDD_REORDER_SYMM_SIFT_CONV
	converge := method == CUDD_REORDER_SIFT_CONVERGE || method == CUDD_REORDER_SYMM_SIFT_CONV
	if method != CUDD_REORDER_SIFT && method != CUDD_REORDER_SIFT_CONVERGE && !symmetric {
		return 0
	}
	roots := dd.liveReorderRoots()
	order := append([]int(nil), dd.InvPerm...)
	heap := reorderBuild(dd, roots, order)
	if len(heap.nodes) < minsize {
		return 1
	}
	started, proceed := Cudd_BeginBDDReordering(dd, method)
	if !proceed {
		return 0
	}
	dd.symmetryGroups = nil
	if len(roots.active) == 0 {
		reorderCommit(dd, roots, heap, order)
		return boolToInt(Cudd_EndBDDReordering(dd, started, true))
	}
	groups := make([][]int, len(order))
	for i, v := range order {
		groups[i] = []int{v}
	}
	if symmetric && dd.siftMaxVar > 0 && dd.siftMaxSwap > 0 {
		groups = mergeSymmetric(groups, heap)
	}
	swaps := 0
	for {
		before, beforeGroups := heap.cost, len(groups)
		// CUDD sorts variables once per pass by unique-subtable population.
		variables := append([]int(nil), order...)
		levelOf := make([]int, len(order))
		for level, index := range order {
			levelOf[index] = level
		}
		sort.SliceStable(variables, func(i, j int) bool { return heap.levels[levelOf[variables[i]]] > heap.levels[levelOf[variables[j]]] })
		processed := map[int]bool{}
		for visit, variable := range variables {
			if visit >= dd.siftMaxVar || swaps >= dd.siftMaxSwap {
				break
			}
			if processed[variable] {
				continue
			}
			selected := groupIndex(groups, variable)
			for _, v := range groups[selected] {
				processed[v] = true
			}
			base := cloneGroups(groups)
			best := cloneGroups(groups)
			bestHeap := heap
			bestCost := heap.cost
			directions := []int{-1, 1}
			if selected > len(groups)-1-selected {
				directions = []int{1, -1}
			}
			for _, direction := range directions {
				trial := cloneGroups(base)
				for pos := selected + direction; pos >= 0 && pos < len(trial); pos += direction {
					previous := pos - direction
					swaps += len(trial[pos]) * len(trial[previous])
					trial[pos], trial[previous] = trial[previous], trial[pos]
					candidate := reorderBuild(dd, roots, groupOrder(trial))
					if candidate.cost <= bestCost {
						bestCost = candidate.cost
						best = cloneGroups(trial)
						bestHeap = candidate
					}
					if float64(candidate.cost) > dd.maxGrowth*float64(bestCost) || swaps >= dd.siftMaxSwap {
						break
					}
				}
			}
			groups = best
			heap = bestHeap
			if symmetric {
				groups = mergeSymmetric(groups, heap)
			}
			order = groupOrder(groups)
		}
		if !converge || (heap.cost >= before && len(groups) == beforeGroups) || swaps >= dd.siftMaxSwap {
			break
		}
	}
	reorderCommit(dd, roots, heap, order)
	dd.symmetryGroups = nil
	if symmetric {
		dd.symmetryGroups = cloneGroups(groups)
	}
	return boolToInt(Cudd_EndBDDReordering(dd, started, true))
}

func boolToInt(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

// Cudd_AnnealRange runs simulated annealing over variable orders in [lower,
// upper]. Candidate graphs are built off-manager; only the best order is
// committed, so live node handles remain valid throughout the search.
func Cudd_AnnealRange(dd *DdManager, lower, upper, minsize int) int {
	if dd == nil || dd.closed || lower < 0 || upper >= len(dd.Vars) || lower > upper || minsize < 0 {
		return 0
	}
	started := time.Now()
	defer func() { dd.reorderingTime += time.Since(started) }()
	roots := dd.liveReorderRoots()
	order := append([]int(nil), dd.InvPerm...)
	current := reorderBuild(dd, roots, order)
	if len(current.nodes) < minsize || upper-lower < 1 {
		return 1
	}
	bestOrder := append([]int(nil), order...)
	bestCost := current.cost
	rng := rand.New(rand.NewSource(int64(dd.nextID + uint64(lower*31+upper))))
	startTemp := math.Max(1, float64(current.cost)*0.05)
	for temp, generation := startTemp, 0; temp > 0.05 && generation < 40; temp, generation = temp*0.82, generation+1 {
		steps := max(12, (upper-lower+1)*6)
		for step := 0; step < steps; step++ {
			a := lower + rng.Intn(upper-lower+1)
			b := lower + rng.Intn(upper-lower+1)
			if a == b {
				b = lower + (b-lower+1)%(upper-lower+1)
			}
			candidateOrder := append([]int(nil), order...)
			candidateOrder[a], candidateOrder[b] = candidateOrder[b], candidateOrder[a]
			candidate := reorderBuild(dd, roots, candidateOrder)
			delta := candidate.cost - current.cost
			if delta <= 0 || rng.Float64() < math.Exp(-float64(delta)/temp) {
				order, current = candidateOrder, candidate
			}
			if current.cost < bestCost {
				bestCost = current.cost
				bestOrder = append(bestOrder[:0], order...)
			}
		}
	}
	reorderCommit(dd, roots, reorderBuild(dd, roots, bestOrder), bestOrder)
	dd.reorderings++
	dd.symmetryGroups = nil
	return 1
}
func groupIndex(groups [][]int, variable int) int {
	for i, g := range groups {
		for _, v := range g {
			if v == variable {
				return i
			}
		}
	}
	panic("cudd: missing reorder group")
}
func Cudd_ReadReorderings(dd *DdManager) uint {
	if dd == nil {
		return 0
	}
	return dd.reorderings
}
func Cudd_ReadReorderingTime(dd *DdManager) int64 {
	if dd == nil {
		return 0
	}
	return dd.reorderingTime.Milliseconds()
}
func Cudd_ReadSiftMaxVar(dd *DdManager) int { return dd.siftMaxVar }
func Cudd_SetSiftMaxVar(dd *DdManager, n int) {
	if n >= 0 {
		dd.siftMaxVar = n
	}
}
func Cudd_ReadSiftMaxSwap(dd *DdManager) int { return dd.siftMaxSwap }
func Cudd_SetSiftMaxSwap(dd *DdManager, n int) {
	if n >= 0 {
		dd.siftMaxSwap = n
	}
}
func Cudd_ReadMaxGrowth(dd *DdManager) float64 { return dd.maxGrowth }
func Cudd_SetMaxGrowth(dd *DdManager, growth float64) {
	if growth >= 1 && !math.IsInf(growth, 0) && !math.IsNaN(growth) {
		dd.maxGrowth = growth
	}
}

// SymmetryGroups returns discovered nonsingleton groups without exposing the
// manager's mutable grouping metadata.
func (dd *DdManager) SymmetryGroups() [][]int {
	var groups [][]int
	for _, g := range dd.symmetryGroups {
		if len(g) > 1 {
			groups = append(groups, append([]int(nil), g...))
		}
	}
	return groups
}
func Cudd_SymmProfile(dd *DdManager, lower, upper int) {
	if dd == nil || lower < 0 || upper >= len(dd.Vars) || lower > upper {
		return
	}
	count, groups := 0, 0
	for _, g := range dd.SymmetryGroups() {
		if dd.Perm[g[0]] < lower || dd.Perm[g[0]] > upper {
			continue
		}
		fmt.Print("Group:")
		for _, index := range g {
			fmt.Printf("  %d", index)
			count++
		}
		fmt.Println()
		groups++
	}
	fmt.Printf("Total Symmetric = %d\nTotal Groups = %d\n", count, groups)
}
