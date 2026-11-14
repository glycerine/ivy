package core

import (
	"sort"
	"time"

	"github.com/glycerine/abcgo/bdd/mtr"
)

var cuddZddRand, cuddZddRand2, cuddZddShuffleSelect int64
var cuddZddShuffleTable [64]int64

// Cudd_Srandom and Cudd_Random port the shared portable generator used by
// CUDD's random reorder heuristics.
func Cudd_Srandom(seed int64) {
	if seed < 0 {
		seed = -seed
	}
	if seed == 0 {
		seed = 1
	}
	cuddZddRand, cuddZddRand2 = seed, seed
	for i := 0; i < 64+11; i++ {
		w := cuddZddRand / 53668
		cuddZddRand = 40014*(cuddZddRand-w*53668) - w*12211
		if cuddZddRand < 0 {
			cuddZddRand += 2147483563
		}
		cuddZddShuffleTable[i%64] = cuddZddRand
	}
	cuddZddShuffleSelect = cuddZddShuffleTable[1]
}

func Cudd_Random() int64 {
	if cuddZddRand == 0 {
		Cudd_Srandom(1)
	}
	w := cuddZddRand / 53668
	cuddZddRand = 40014*(cuddZddRand-w*53668) - w*12211
	if cuddZddRand < 0 {
		cuddZddRand += 2147483563
	}
	w = cuddZddRand2 / 52774
	cuddZddRand2 = 40692*(cuddZddRand2-w*52774) - w*3791
	if cuddZddRand2 < 0 {
		cuddZddRand2 += 2147483399
	}
	i := int(cuddZddShuffleSelect / 33554431)
	cuddZddShuffleSelect = cuddZddShuffleTable[i] - cuddZddRand2
	cuddZddShuffleTable[i] = cuddZddRand
	if cuddZddShuffleSelect < 1 {
		cuddZddShuffleSelect += 2147483562
	}
	return cuddZddShuffleSelect - 1
}

type zddReorderKey struct {
	node  *DdNode
	level int
}
type zddCofactorKey struct {
	node     *DdNode
	variable int
	include  bool
}

// reorderZdd rebuilds the zero-suppressed graph under a new level order while
// preserving all existing node handles. The source table is left untouched
// until the complete replacement graph has been constructed.
func reorderZdd(dd *DdManager, size int, permutation, inverse []int) bool {
	if dd == nil || size < dd.ZddSize || len(permutation) != size || len(inverse) != size {
		return false
	}
	scratch := NewBDDManager(0)
	scratch.ZddSize = size
	scratch.PermZ = append([]int(nil), permutation...)
	scratch.InvPermZ = append([]int(nil), inverse...)
	if !validZddOrder(scratch) {
		return false
	}
	cofactorMemo := make(map[zddCofactorKey]*DdNode)
	var cofactor func(*DdNode, int, bool) *DdNode
	cofactor = func(node *DdNode, variable int, include bool) *DdNode {
		if node == dd.Zero {
			return dd.Zero
		}
		if node == dd.One {
			if include {
				return dd.Zero
			}
			return dd.One
		}
		if variable >= len(dd.PermZ) {
			if include {
				return dd.Zero
			}
			return node
		}
		key := zddCofactorKey{node: node, variable: variable, include: include}
		if result, ok := cofactorMemo[key]; ok {
			return result
		}
		oldLevel, variableLevel := dd.PermZ[node.Index], dd.PermZ[variable]
		var result *DdNode
		switch {
		case oldLevel > variableLevel:
			if include {
				result = dd.Zero
			} else {
				result = node
			}
		case node.Index == uint32(variable):
			if include {
				result = node.T
			} else {
				result = node.E
			}
		default:
			t := cofactor(node.T, variable, include)
			e := cofactor(node.E, variable, include)
			result = dd.zddNode(node.Index, t, e)
		}
		cofactorMemo[key] = result
		return result
	}
	reorderMemo := make(map[zddReorderKey]*DdNode)
	var visit func(*DdNode, int) *DdNode
	visit = func(node *DdNode, level int) *DdNode {
		if node == dd.Zero {
			return scratch.Zero
		}
		if node == dd.One {
			return scratch.One
		}
		key := zddReorderKey{node: node, level: level}
		if result := reorderMemo[key]; result != nil {
			return result
		}
		if level >= size {
			return nil
		}
		variable := inverse[level]
		include := cofactor(node, variable, true)
		exclude := cofactor(node, variable, false)
		t := visit(include, level+1)
		e := visit(exclude, level+1)
		if t == nil || e == nil {
			return nil
		}
		result := scratch.zddNode(uint32(variable), t, e)
		reorderMemo[key] = result
		return result
	}
	oldNodes := make([]*DdNode, 0, len(dd.zddUnique)+len(dd.Univ))
	seen := map[*DdNode]bool{}
	for _, node := range dd.zddUnique {
		if !seen[node] {
			seen[node] = true
			oldNodes = append(oldNodes, node)
		}
	}
	for _, node := range dd.Univ {
		if node != nil && !seen[node] {
			seen[node] = true
			oldNodes = append(oldNodes, node)
		}
	}
	remap := make(map[*DdNode]*DdNode, len(oldNodes))
	preserve := make(map[*DdNode]*DdNode, len(oldNodes))
	for _, node := range oldNodes {
		fresh := visit(node, 0)
		if fresh == nil {
			return false
		}
		remap[node] = fresh
		if !Cudd_IsConstant(fresh) {
			if previous := preserve[fresh]; previous == nil || node.Ref > previous.Ref {
				preserve[fresh] = node
			}
		}
	}
	newNodes := make([]*DdNode, 0, len(scratch.zddUnique))
	for _, node := range scratch.zddUnique {
		newNodes = append(newNodes, node)
	}
	mapped := map[*DdNode]*DdNode{scratch.One: dd.One, scratch.Zero: dd.Zero}
	for _, fresh := range newNodes {
		if old := preserve[fresh]; old != nil {
			mapped[fresh] = old
		} else {
			mapped[fresh] = &DdNode{ID: dd.nextID, manager: dd, family: 1}
			dd.nextID++
		}
	}
	for _, fresh := range newNodes {
		node := mapped[fresh]
		node.Index = fresh.Index
		node.T = mapped[fresh.T]
		node.E = mapped[fresh.E]
		node.manager = dd
		node.family = 1
	}
	nextTable := make(map[bddUniqueKey]*DdNode, len(newNodes))
	for _, fresh := range newNodes {
		node := mapped[fresh]
		key := bddUniqueKey{index: node.Index, t: node.T, e: node.E}
		if other := nextTable[key]; other != nil && other != node {
			return false
		}
		nextTable[key] = node
	}
	newUniverse := make([]*DdNode, len(dd.Univ))
	for i, node := range dd.Univ {
		if node != nil {
			fresh := remap[node]
			if fresh == nil {
				return false
			}
			newUniverse[i] = mapped[Cudd_Regular(fresh)]
		}
	}
	dd.zddUnique = nextTable
	dd.ZddSize = size
	dd.PermZ = append(dd.PermZ[:0], permutation...)
	dd.InvPermZ = append(dd.InvPermZ[:0], inverse...)
	dd.Univ = newUniverse
	return cuddZddInitUniv(dd) == 1
}

func validZddOrder(dd *DdManager) bool {
	if dd == nil || dd.ZddSize < 0 || len(dd.PermZ) != dd.ZddSize || len(dd.InvPermZ) != dd.ZddSize {
		return false
	}
	seen := make([]bool, dd.ZddSize)
	for level, index := range dd.InvPermZ {
		if index < 0 || index >= dd.ZddSize || seen[index] || dd.PermZ[index] != level {
			return false
		}
		seen[index] = true
	}
	return true
}

// Cudd_zddShuffleHeap reorders ZDD variables. order[level] is the variable
// index that must occupy that level, matching the C API's permutation format.
func Cudd_zddShuffleHeap(dd *DdManager, order []int) int {
	if dd == nil || dd.closed || dd.ZddSize < 0 || len(order) < dd.ZddSize || !reorderValidOrder(order[:dd.ZddSize], dd.ZddSize) {
		return 0
	}
	order = order[:dd.ZddSize]
	order = order[:dd.ZddSize]
	// Match zddShuffle: bring the requested variable at each level upward
	// through adjacent swaps, preserving the source algorithm's swap sequence.
	for level, index := range order {
		position := dd.PermZ[index]
		for position > level {
			if !swapZddAdjacentInPlace(dd, position-1, position) {
				return 0
			}
			position--
		}
	}
	return 1
}

// Cudd_zddAlignToBdd restores the CUDD block alignment between BDD variables
// and their corresponding groups of ZDD variables.
func Cudd_zddAlignToBdd(dd *DdManager) int {
	if dd == nil || dd.closed {
		return 0
	}
	if dd.ZddSize == 0 {
		return 1
	}
	if len(dd.Vars) == 0 {
		return 0
	}
	multiplicity := dd.ZddSize / len(dd.Vars)
	if multiplicity*len(dd.Vars) != dd.ZddSize {
		return 0
	}
	order := make([]int, dd.ZddSize)
	for bddLevel := range dd.InvPerm {
		index := dd.InvPerm[bddLevel]
		indexZ := index * multiplicity
		levelZ := dd.PermZ[indexZ]
		levelZ = (levelZ / multiplicity) * multiplicity
		for j := 0; j < multiplicity; j++ {
			order[multiplicity*bddLevel+j] = dd.InvPermZ[levelZ+j]
		}
	}
	if Cudd_zddShuffleHeap(dd, order) == 0 {
		return 0
	}
	zddFixTree(dd, dd.TreeZ)
	return 1
}

// Cudd_bddAlignToZdd restores the BDD order from the first ZDD variable in
// each multiplicity-sized block, matching the post-ZDD-reorder CUDD path.
func Cudd_bddAlignToZdd(dd *DdManager) int {
	if dd == nil || dd.closed {
		return 0
	}
	if len(dd.Vars) == 0 {
		return 1
	}
	multiplicity := dd.ZddSize / len(dd.Vars)
	if multiplicity*len(dd.Vars) != dd.ZddSize {
		return 0
	}
	if multiplicity == 0 {
		return 1
	}
	order := make([]int, len(dd.Vars))
	for level := range order {
		indexZ := dd.InvPermZ[level*multiplicity]
		order[level] = indexZ / multiplicity
	}
	if Cudd_ShuffleHeap(dd, order) == 0 {
		return 0
	}
	bddFixTree(dd, dd.Tree)
	return 1
}

func bddFixTree(dd *DdManager, node *mtr.MtrNode) {
	if node == nil {
		return
	}
	if int(node.Index) < len(dd.Vars) {
		node.Low = uint32(dd.Perm[node.Index])
	} else {
		node.Low = node.Index
	}
	for child := node.Child; child != nil; child = child.Younger {
		bddFixTree(dd, child)
	}
	if node.Parent != nil && node.Low < node.Parent.Low {
		node.Parent.Low = node.Low
		node.Parent.Index = node.Index
	}
}

func zddFixTree(dd *DdManager, node *mtr.MtrNode) {
	if node == nil {
		return
	}
	if int(node.Index) < dd.ZddSize {
		node.Low = uint32(dd.PermZ[node.Index])
	} else {
		node.Low = node.Index
	}
	for child := node.Child; child != nil; child = child.Younger {
		zddFixTree(dd, child)
	}
	if node.Parent != nil && node.Low < node.Parent.Low {
		node.Parent.Low = node.Low
		node.Parent.Index = node.Index
	}
}

// zddHeapSize is the number of live canonical nonterminal ZDD nodes.
func zddHeapSize(dd *DdManager) int { return len(dd.zddUnique) }

// zddSymmCheck ports cuddZddSymmCheck for adjacent levels. DdNode.Ref already
// includes both root references and incoming arcs, as in CUDD's node refcount.
func zddSymmCheck(dd *DdManager, x, y int) bool {
	if dd == nil || x < 0 || y != x+1 || y >= dd.ZddSize {
		return false
	}
	xIndex, yIndex := dd.InvPermZ[x], dd.InvPermZ[y]
	xSym, xSymPrime := true, true
	arcCount, totalRefCount := 0, 0
	for _, f := range dd.zddUnique {
		if int(f.Index) != xIndex {
			continue
		}
		f1, f0 := f.T, f.E
		f11, f10 := dd.Zero, f1
		if int(f1.Index) == yIndex {
			f11, f10 = f1.T, f1.E
			if f10 != dd.Zero {
				arcCount++
			}
		} else if int(f0.Index) != yIndex {
			return false
		}
		f01, f00 := dd.Zero, f0
		if int(f0.Index) == yIndex {
			f01, f00 = f0.T, f0.E
			if f00 != dd.Zero {
				arcCount++
			}
		}
		xSym = xSym && f01 == f10
		xSymPrime = xSymPrime && f11 == f00
		if !xSym && !xSymPrime {
			return false
		}
	}
	for _, f := range dd.zddUnique {
		if int(f.Index) != yIndex || f.E == dd.Zero {
			continue
		}
		totalRefCount += int(f.Ref)
	}
	return arcCount == totalRefCount && (xSym || xSymPrime)
}

type zddSiftMove struct{ low, high, size int }

// zddSiftingPath follows CUDD's upward or downward adjacent-swap path and
// records the resulting heap size after each swap.
func zddSiftingPath(dd *DdManager, variable, direction, initialSize, maxSwap, swaps int) ([]zddSiftMove, int, bool) {
	limit := initialSize
	moves := make([]zddSiftMove, 0, dd.ZddSize)
	position := dd.PermZ[variable]
	for position+direction >= 0 && position+direction < dd.ZddSize && swaps < maxSwap {
		next := position + direction
		low, high := position, next
		if low > high {
			low, high = high, low
		}
		if !swapZddAdjacentInPlace(dd, low, high) {
			return moves, swaps, false
		}
		swaps++
		position = next
		size := zddHeapSize(dd)
		moves = append(moves, zddSiftMove{low: low, high: high, size: size})
		if float64(size) > float64(limit)*dd.maxGrowth {
			break
		}
		if size < limit {
			limit = size
		}
	}
	return moves, swaps, true
}

// zddSiftingBackward undoes the tail of a recorded path and leaves the
// variable at the closest position attaining a strictly smaller minimum.
func zddSiftingBackward(dd *DdManager, moves []zddSiftMove, initialSize int) bool {
	best := -1
	size := initialSize
	for i := len(moves) - 1; i >= 0; i-- {
		if moves[i].size < size {
			best, size = i, moves[i].size
		}
	}
	for i := len(moves) - 1; i > best; i-- {
		if !swapZddAdjacentInPlace(dd, moves[i].low, moves[i].high) {
			return false
		}
	}
	return true
}

func zddSwapAny(dd *DdManager, x, y, initialSize, maxSwap, swaps int) ([]zddSiftMove, int, bool) {
	if x > y {
		x, y = y, x
	}
	xRef, yRef := x, y
	xNext, yNext := x+1, y-1
	limit, lastSize := initialSize, initialSize
	moves := make([]zddSiftMove, 0, 2*(y-x))
	swap := func(a, b int) bool {
		if swaps >= maxSwap || b != a+1 || !swapZddAdjacentInPlace(dd, a, b) {
			return false
		}
		swaps++
		lastSize = zddHeapSize(dd)
		moves = append(moves, zddSiftMove{low: a, high: b, size: lastSize})
		return true
	}
	for {
		if xNext == yNext {
			if !swap(x, xNext) || !swap(yNext, y) || !swap(x, xNext) {
				return moves, swaps, false
			}
			x, y = y, x
		} else if x == yNext {
			if !swap(x, xNext) {
				return moves, swaps, false
			}
			x, y = y, x
		} else {
			if !swap(x, xNext) || !swap(yNext, y) {
				return moves, swaps, false
			}
			x, y = xNext, yNext
		}
		xNext, yNext = x+1, y-1
		if xNext > yRef || swaps >= maxSwap {
			break
		}
		if float64(lastSize) > float64(limit)*dd.maxGrowth {
			break
		}
		if lastSize < limit {
			limit = lastSize
		}
	}
	if yNext >= xRef && swaps < maxSwap {
		if !swap(yNext, y) {
			return moves, swaps, false
		}
	}
	return moves, swaps, true
}

func zddRandomSwapping(dd *DdManager, lower, upper, maxVar, maxSwap int, pivot bool) bool {
	nvars := upper - lower + 1
	swaps := 0
	for iteration := 0; iteration < min(nvars, maxVar) && swaps < maxSwap; iteration++ {
		var x, y int
		if pivot {
			maxKeys, p := -1, lower
			for level := lower; level <= upper; level++ {
				keys := 0
				index := dd.InvPermZ[level]
				for _, node := range dd.zddUnique {
					if int(node.Index) == index {
						keys++
					}
				}
				if keys > maxKeys {
					maxKeys, p = keys, level
				}
			}
			modulo := upper - p
			if modulo == 0 {
				y = p
			} else {
				y = p + 1 + int(Cudd_Random()%int64(modulo))
			}
			modulo = p - lower - 1
			if modulo < 1 {
				x = lower
			} else {
				x = int(Cudd_Random() % int64(modulo)) // Preserves the C source's lower-bound behavior.
			}
		} else {
			x = int(Cudd_Random()%int64(nvars)) + lower
			for {
				y = int(Cudd_Random()%int64(nvars)) + lower
				if x != y {
					break
				}
			}
		}
		initial := zddHeapSize(dd)
		moves, used, ok := zddSwapAny(dd, x, y, initial, maxSwap, swaps)
		swaps = used
		if !ok || !zddSiftingBackward(dd, moves, initial) {
			return false
		}
	}
	return true
}

// Cudd_zddReduceHeap applies CUDD's ZDD sifting policy to the ZDD order.
// Variables are considered by decreasing current subtable population; each
// variable is moved both directions until the configured growth bound is hit,
// then restored to the best position reached.
func Cudd_zddReduceHeap(dd *DdManager, method Cudd_ReorderingType, minsize int) int {
	if dd == nil || dd.closed {
		return 0
	}
	// The C routine compares against an unsigned conversion of minsize. A
	// negative C int therefore becomes a very large threshold and skips work.
	if minsize < 0 {
		return 1
	}
	if method == CUDD_REORDER_SAME {
		method = dd.autoMethodZdd
		if method == CUDD_REORDER_SAME {
			method = CUDD_REORDER_SIFT
		}
	}
	if method == CUDD_REORDER_NONE {
		return 1
	}
	if zddHeapSize(dd) < minsize {
		return 1
	}
	started := time.Now()
	dd.reorderings++
	if !dd.runHooks(CUDD_PRE_REORDERING_HOOK, "ZDD", method) {
		return 0
	}
	clear(dd.cache)
	clear(dd.computed)
	finish := func(result int) int {
		if result == 0 {
			return 0
		}
		if dd.realignZdd && Cudd_bddAlignToZdd(dd) == 0 {
			return 0
		}
		nextDyn := uint32(zddHeapSize(dd) * 2)
		if dd.reorderings < 20 || nextDyn > dd.nextDyn {
			dd.nextDyn = nextDyn
		} else {
			dd.nextDyn += 20
		}
		if !dd.runHooks(CUDD_POST_REORDERING_HOOK, "ZDD", started.UnixNano()) {
			return 0
		}
		dd.reorderingTime += time.Since(started)
		return result
	}
	if method != CUDD_REORDER_RANDOM && method != CUDD_REORDER_RANDOM_PIVOT && method != CUDD_REORDER_SIFT && method != CUDD_REORDER_SIFT_CONVERGE && method != CUDD_REORDER_SYMM_SIFT && method != CUDD_REORDER_SYMM_SIFT_CONV && method != CUDD_REORDER_LINEAR && method != CUDD_REORDER_LINEAR_CONVERGE {
		return finish(0)
	}
	n := dd.ZddSize
	if n < 2 || zddHeapSize(dd) == 0 {
		return finish(1)
	}
	maxVar, maxSwap := dd.siftMaxVar, dd.siftMaxSwap
	if method == CUDD_REORDER_RANDOM || method == CUDD_REORDER_RANDOM_PIVOT {
		if zddRandomSwapping(dd, 0, n-1, maxVar, maxSwap, method == CUDD_REORDER_RANDOM_PIVOT) {
			return finish(1)
		}
		return finish(0)
	}
	if maxVar == 0 || maxSwap == 0 {
		return finish(1)
	}
	if method == CUDD_REORDER_SYMM_SIFT || method == CUDD_REORDER_SYMM_SIFT_CONV {
		if Cudd_zddTreeSifting(dd, method) == 0 {
			return finish(0)
		}
		return finish(1)
	}
	if method == CUDD_REORDER_LINEAR || method == CUDD_REORDER_LINEAR_CONVERGE {
		// The linear sifter maintains the ZDD-specific linear-combination
		// invariant while it moves variables. Its implementation is in
		// cuddZddLin.go, alongside the linear adjacent operation.
		for {
			initialSize := zddHeapSize(dd)
			if Cudd_zddLinearSifting(dd, 0, n-1) == 0 {
				return finish(0)
			}
			if method != CUDD_REORDER_LINEAR_CONVERGE || zddHeapSize(dd) >= initialSize {
				break
			}
		}
		return finish(1)
	}
	levels := make([]int, n)
	for _, node := range dd.zddUnique {
		if node != nil && int(node.Index) < n {
			levels[dd.PermZ[node.Index]]++
		}
	}
	positions := make([]int, n)
	for i := range positions {
		positions[i] = i
	}
	sort.SliceStable(positions, func(i, j int) bool { return levels[positions[i]] > levels[positions[j]] })
	swaps := 0
	for pass := 0; ; pass++ {
		startSize := zddHeapSize(dd)
		for variableIndex, startLevel := range positions {
			if variableIndex >= maxVar {
				break
			}
			if swaps >= maxSwap {
				break
			}
			variable := dd.InvPermZ[startLevel]
			initialSize := zddHeapSize(dd)
			start := dd.PermZ[variable]
			directions := []int{-1, 1}
			if start == 0 {
				directions = []int{1}
			} else if start == n-1 {
				directions = []int{-1}
			} else if start > n-1-start {
				directions = []int{1, -1}
			} else {
				directions = []int{-1, 1}
			}
			var path []zddSiftMove
			for _, direction := range directions {
				var ok bool
				path, swaps, ok = zddSiftingPath(dd, variable, direction, initialSize, maxSwap, swaps)
				if !ok {
					return finish(0)
				}
			}
			if !zddSiftingBackward(dd, path, initialSize) {
				return finish(0)
			}
		}
		if method != CUDD_REORDER_SIFT_CONVERGE || zddHeapSize(dd) >= startSize || pass >= n {
			break
		}
	}
	return finish(1)
}
