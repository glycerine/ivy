package core

import "sort"

var zddLinearSwapCount int

func zddLinearSwap(dd *DdManager, x, y int) bool {
	if !swapZddAdjacentInPlace(dd, x, y) {
		return false
	}
	zddLinearSwapCount++
	return true
}

// linearZddAdjacentInPlace applies CUDD's adjacent ZDD linear transform at
// levels x and y.  It is the source algorithm's cofactor re-expression: the
// y cofactors of each x node are crossed, then canonicalized back into the
// two affected levels.  The transform is an involution.
func linearZddAdjacentInPlace(dd *DdManager, x, y int) bool {
	if dd == nil || x < 0 || y != x+1 || y >= dd.ZddSize || !validZddOrder(dd) {
		return false
	}
	xi, yi := dd.InvPermZ[x], dd.InvPermZ[y]
	// Preserve all handles, including a y function reached from above x.
	// Keep the original edges readable until every new cofactor is known.
	type rewrite struct {
		node, oldT, oldE, newT, newE *DdNode
		index                        uint32
		owned                        bool
	}
	var local [32]rewrite
	plan := local[:0]
	if size := dd.zddUnique.levelCount(xi) + dd.zddUnique.levelCount(yi); size > len(local) {
		plan = make([]rewrite, 0, size)
	}
	collect := func(entry uniqueEntry) {
		n := entry.node
		plan = append(plan, rewrite{node: n, oldT: entry.t, oldE: entry.e, owned: !n.childrenReleased})
	}
	dd.zddUnique.forLevel(xi, collect)
	nx := len(plan)
	dd.zddUnique.forLevel(yi, collect)
	dd.zddUnique.clearLevel(xi)
	dd.zddUnique.clearLevel(yi)
	// Special x nodes become canonical y nodes. Install them before looking
	// up any other node's new y cofactors.
	for i := range nx {
		r := &plan[i]
		if r.oldT.Index == uint32(yi) && r.oldT.E == dd.Zero && r.oldE.Index != uint32(yi) {
			r.index, r.newT, r.newE = uint32(yi), r.oldT.T, r.oldE
			dd.zddUnique.set(bddUniqueKey{r.index, r.newT, r.newE}, r.node)
		}
	}
	cofactors := func(n *DdNode) (*DdNode, *DdNode) {
		if n.Index == uint32(yi) {
			return n.T, n.E
		}
		return dd.Zero, n
	}
	for i := range plan {
		r := &plan[i]
		if r.newT != nil {
			continue
		}
		r.index = uint32(xi)
		if i < nx {
			f11, f10 := cofactors(r.oldT)
			f01, f00 := cofactors(r.oldE)
			r.newT = dd.internReorderedZDD(uint32(yi), f01, f10)
			r.newE = dd.internReorderedZDD(uint32(yi), f11, f00)
		} else {
			// Missing x in the old family means new x and y must agree.
			r.newT = dd.internReorderedZDD(uint32(yi), r.oldT, dd.Zero)
			r.newE = r.oldE
		}
		dd.zddUnique.set(bddUniqueKey{r.index, r.newT, r.newE}, r.node)
	}
	for _, r := range plan {
		r.node.Index, r.node.inverse.Index = r.index, r.index
		r.node.T, r.node.E = r.newT, r.newE
	}
	// A shared y node can move to x and remain an external root. Acquire all
	// new arcs before dropping any old arcs so revival follows the new graph.
	for _, r := range plan {
		if r.owned {
			Cudd_Ref(r.newT)
			Cudd_Ref(r.newE)
		}
	}
	for _, r := range plan {
		if r.owned {
			dd.recursiveDeref(r.oldT)
			dd.recursiveDeref(r.oldE)
		}
	}
	if len(dd.Univ) > y && dd.Univ[x] != nil {
		dd.Univ[y] = dd.Univ[x].T
	}
	Cudd_CacheFlush(dd)

	return true
}

type zddLinearMove struct {
	x, y, size int
	kind       uint8
}

const (
	zddLinearSwapMove uint8 = iota
	zddLinearTransformMove
	zddLinearInverseTransformMove
)

func zddLinearStepUp(dd *DdManager, y, low int, moves *[]zddLinearMove) bool {
	limit := zddHeapSize(dd)
	for level := y; level > low; level-- {
		x := level - 1
		if !zddLinearSwap(dd, x, y) {
			return false
		}
		swappedSize := zddHeapSize(dd)
		if !linearZddAdjacentInPlace(dd, x, y) {
			return false
		}
		newSize := zddHeapSize(dd)
		kind := zddLinearTransformMove
		if newSize > swappedSize {
			if !linearZddAdjacentInPlace(dd, x, y) {
				return false
			}
			newSize = swappedSize
			kind = zddLinearSwapMove
		}
		*moves = append([]zddLinearMove{{x, y, newSize, kind}}, *moves...)
		if float64(newSize) > float64(limit)*dd.maxGrowth {
			break
		}
		if newSize < limit {
			limit = newSize
		}
		y = x
	}
	return true
}

func zddLinearStepDown(dd *DdManager, x, high int, moves *[]zddLinearMove) bool {
	limit := zddHeapSize(dd)
	for level := x; level < high; level++ {
		y := level + 1
		if !zddLinearSwap(dd, x, y) {
			return false
		}
		swappedSize := zddHeapSize(dd)
		if !linearZddAdjacentInPlace(dd, x, y) {
			return false
		}
		newSize := zddHeapSize(dd)
		kind := zddLinearTransformMove
		if newSize > swappedSize {
			if !linearZddAdjacentInPlace(dd, x, y) {
				return false
			}
			newSize = swappedSize
			kind = zddLinearSwapMove
		}
		*moves = append([]zddLinearMove{{x, y, newSize, kind}}, *moves...)
		if float64(newSize) > float64(limit)*dd.maxGrowth {
			break
		}
		if newSize < limit {
			limit = newSize
		}
		x = y
	}
	return true
}

func zddLinearUndo(dd *DdManager, moves []zddLinearMove) ([]zddLinearMove, bool) {
	var inverse []zddLinearMove
	for _, m := range moves {
		kind := m.kind
		switch m.kind {
		case zddLinearTransformMove:
			kind = zddLinearInverseTransformMove
			if !linearZddAdjacentInPlace(dd, m.x, m.y) || !zddLinearSwap(dd, m.x, m.y) {
				return nil, false
			}
		case zddLinearInverseTransformMove:
			kind = zddLinearTransformMove
			if !zddLinearSwap(dd, m.x, m.y) || !linearZddAdjacentInPlace(dd, m.x, m.y) {
				return nil, false
			}
		default:
			if !zddLinearSwap(dd, m.x, m.y) {
				return nil, false
			}
		}
		inverse = append([]zddLinearMove{{m.x, m.y, zddHeapSize(dd), kind}}, inverse...)
	}
	return inverse, true
}

func zddLinearBackward(dd *DdManager, initial int, moves []zddLinearMove) bool {
	best := initial
	for _, m := range moves {
		if m.size < best {
			best = m.size
		}
	}
	for _, m := range moves {
		if m.size == best {
			return true
		}
		switch m.kind {
		case zddLinearTransformMove:
			if !linearZddAdjacentInPlace(dd, m.x, m.y) || !zddLinearSwap(dd, m.x, m.y) {
				return false
			}
		case zddLinearInverseTransformMove:
			if !zddLinearSwap(dd, m.x, m.y) || !linearZddAdjacentInPlace(dd, m.x, m.y) {
				return false
			}
		default:
			if !zddLinearSwap(dd, m.x, m.y) {
				return false
			}
		}
	}
	return true
}

func zddLinearAux(dd *DdManager, variable, low, high int) bool {
	initial := zddHeapSize(dd)
	pos := dd.PermZ[variable]
	var moves []zddLinearMove
	// CUDD traverses the shorter side first, records the inverse path, then
	// completes the opposite side before selecting the closest minimum.
	if pos-low > high-pos {
		if !zddLinearStepDown(dd, pos, high, &moves) {
			return false
		}
		var ok bool
		moves, ok = zddLinearUndo(dd, moves)
		if !ok || !zddLinearStepUp(dd, pos, low, &moves) {
			return false
		}
	} else {
		if !zddLinearStepUp(dd, pos, low, &moves) {
			return false
		}
		var ok bool
		moves, ok = zddLinearUndo(dd, moves)
		if !ok || !zddLinearStepDown(dd, pos, high, &moves) {
			return false
		}
	}
	return zddLinearBackward(dd, initial, moves)
}

// Cudd_zddLinearSifting ports cuddZddLinearSifting. Variables are considered
// by descending unique-table population and sifted within the inclusive level
// interval. It returns one on success and zero on failure.
func Cudd_zddLinearSifting(dd *DdManager, lower, upper int) int {
	if dd == nil || dd.closed || !validZddOrder(dd) || lower < 0 || upper >= dd.ZddSize || lower > upper {
		return 0
	}
	zddLinearSwapCount = 0
	type item struct{ index, keys int }
	items := make([]item, dd.ZddSize)
	for index := range items {
		items[index].index = index
	}
	for _, n := range dd.zddUnique.entries {
		if int(n.Index) < dd.ZddSize {
			items[dd.PermZ[n.Index]].keys++
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].keys > items[j].keys })
	limit := min(dd.ZddSize, dd.siftMaxVar)
	for i := 0; i < limit; i++ {
		if zddLinearSwapCount >= dd.siftMaxSwap {
			break
		}
		variable := dd.InvPermZ[items[i].index]
		if dd.PermZ[variable] < lower || dd.PermZ[variable] > upper {
			continue
		}
		if !zddLinearAux(dd, variable, lower, upper) {
			return 0
		}
	}
	return 1
}
