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
	var xnodes []*DdNode
	for _, n := range dd.zddUnique {
		if int(n.Index) == xi {
			xnodes = append(xnodes, n)
		}
	}
	// Drop old x keys before creating y nodes. This mirrors the C routine's
	// removal of the x subtable into its temporary chains.
	for key := range dd.zddUnique {
		if int(key.index) == xi {
			delete(dd.zddUnique, key)
		}
	}
	for _, f := range xnodes {
		f1, f0 := f.T, f.E
		f11, f10 := dd.Zero, f1
		if int(f1.Index) == yi {
			f11, f10 = f1.T, f1.E
		}
		f01, f00 := dd.Zero, f0
		if int(f0.Index) == yi {
			f01, f00 = f0.T, f0.E
		}
		// These are precisely the two y nodes constructed by the C source;
		// zddNode also performs its zero-suppression and unique-table lookup.
		newT := dd.zddNode(uint32(yi), f01, f10)
		newE := dd.zddNode(uint32(yi), f11, f00)
		if newT == dd.Zero {
			// The C special-node path moves this handle from x to y.
			f.Index, f.T, f.E = uint32(yi), newE.T, newE.E
		} else {
			f.T, f.E = newT, newE
		}
	}
	// Rebuild keys for the mutated x layer and retain all other canonical
	// entries. If a transformed handle meets an equivalent y node, either
	// handle denotes the same family; retaining the existing canonical entry
	// matches the source's duplicate-node handling.
	for _, f := range xnodes {
		key := bddUniqueKey{f.Index, f.T, f.E}
		if dd.zddUnique[key] == nil {
			dd.zddUnique[key] = f
		}
	}
	dd.Univ[y] = dd.Univ[x].T
	clear(dd.cache)
	clear(dd.computed)
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
	for _, n := range dd.zddUnique {
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
