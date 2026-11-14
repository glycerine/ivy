package core

import (
	"sort"

	"github.com/glycerine/abcgo/bdd/mtr"
)

type zddGroupMoveRecord struct{ x, y, size int }
type zddGroupState struct {
	next  []int
	swaps int
}

// Cudd_MakeZddTreeNode creates a ZDD group in the current level tree.
func Cudd_MakeZddTreeNode(dd *DdManager, low, size uint32, flags uint32) *mtr.MtrNode {
	if dd == nil || dd.closed || size == 0 || uint64(low)+uint64(size) > uint64(mtr.MTR_MAXHIGH) {
		return nil
	}
	level := low
	if int(low) < len(dd.PermZ) {
		level = uint32(dd.PermZ[low])
	}
	if dd.TreeZ == nil {
		dd.TreeZ = mtr.Mtr_InitGroupTree(0, dd.ZddSize)
		if dd.TreeZ == nil {
			return nil
		}
		if len(dd.InvPermZ) > 0 {
			dd.TreeZ.Index = uint32(dd.InvPermZ[0])
		}
	}
	if uint64(level)+uint64(size) > uint64(mtr.MTR_MAXHIGH) {
		return nil
	}
	if uint64(level)+uint64(size) > uint64(dd.TreeZ.Low)+uint64(dd.TreeZ.Size) {
		dd.TreeZ.Size = level + size
	}
	g := mtr.Mtr_MakeGroup(dd.TreeZ, level, size, flags)
	if g != nil {
		g.Index = low
	}
	return g
}

// Cudd_zddTreeSifting applies CUDD's postorder group-tree reorder traversal.
func Cudd_zddTreeSifting(dd *DdManager, method Cudd_ReorderingType) int {
	if dd == nil || dd.closed {
		return 0
	}
	temp := dd.TreeZ == nil
	if temp {
		dd.TreeZ = mtr.Mtr_InitGroupTree(0, dd.ZddSize)
		if dd.TreeZ == nil {
			return 0
		}
		if len(dd.InvPermZ) > 0 {
			dd.TreeZ.Index = uint32(dd.InvPermZ[0])
		}
	}
	state := &zddGroupState{next: make([]int, dd.ZddSize)}
	for i := range state.next {
		state.next[i] = i
	}
	ok := zddTreeSiftingAux(dd, dd.TreeZ, method, state)
	if method == CUDD_REORDER_SYMM_SIFT || method == CUDD_REORDER_SYMM_SIFT_CONV {
		zddSymmSaveGroups(dd, state, 0, dd.ZddSize-1)
	}
	if temp {
		mtr.Mtr_FreeTree(dd.TreeZ)
		dd.TreeZ = nil
	}
	if ok {
		return 1
	}
	return 0
}

func zddTreeSiftingAux(dd *DdManager, n *mtr.MtrNode, method Cudd_ReorderingType, s *zddGroupState) bool {
	for c := n; c != nil; c = c.Younger {
		if c.Child != nil {
			if !zddTreeSiftingAux(dd, c.Child, method, s) {
				return false
			}
			lo, hi := zddFindNodeHiLo(dd, c)
			if hi >= lo && !zddReorderChildren(dd, c, lo, hi, method, s) {
				return false
			}
		} else if c.Size > 1 {
			lo, hi := zddFindNodeHiLo(dd, c)
			if hi >= lo && !zddReorderChildren(dd, c, lo, hi, method, s) {
				return false
			}
		}
	}
	return true
}

func zddReorderChildren(dd *DdManager, n *mtr.MtrNode, lo, hi int, method Cudd_ReorderingType, s *zddGroupState) bool {
	if n.Flags&mtr.MTR_FIXED == 0 {
		switch method {
		case CUDD_REORDER_GROUP_SIFT, CUDD_REORDER_GROUP_SIFT_CONV:
			before := zddHeapSize(dd)
			for {
				if !zddGroupSifting(dd, lo, hi, s) {
					return false
				}
				if method != CUDD_REORDER_GROUP_SIFT_CONV || zddHeapSize(dd) >= before {
					break
				}
			}
		case CUDD_REORDER_SIFT, CUDD_REORDER_SIFT_CONVERGE:
			for {
				before := zddHeapSize(dd)
				if !zddReduceHeapRange(dd, lo, hi, s) {
					return false
				}
				if method != CUDD_REORDER_SIFT_CONVERGE || zddHeapSize(dd) >= before {
					break
				}
			}
		case CUDD_REORDER_SYMM_SIFT, CUDD_REORDER_SYMM_SIFT_CONV:
			if zddSymmSiftingWithState(dd, s, lo, hi, method == CUDD_REORDER_SYMM_SIFT_CONV) == 0 {
				return false
			}
		case CUDD_REORDER_LINEAR, CUDD_REORDER_LINEAR_CONVERGE:
			for {
				before := zddHeapSize(dd)
				if Cudd_zddLinearSifting(dd, lo, hi) == 0 {
					return false
				}
				if method != CUDD_REORDER_LINEAR_CONVERGE || zddHeapSize(dd) >= before {
					break
				}
			}
		case CUDD_REORDER_RANDOM, CUDD_REORDER_RANDOM_PIVOT:
			return false // handled by the shared random reorder path
		default:
			return false
		}
	}
	zddMergeGroups(dd, n, lo, hi, s)
	return true
}

// Cudd_zddReduceHeapRange performs variable sifting restricted to a tree node.
func Cudd_zddReduceHeapRange(dd *DdManager, lower, upper int) bool {
	return zddReduceHeapRange(dd, lower, upper, &zddGroupState{})
}

func zddReduceHeapRange(dd *DdManager, lower, upper int, state *zddGroupState) bool {
	if dd == nil || lower < 0 || upper >= dd.ZddSize || lower >= upper {
		return true
	}
	type item struct{ variable, keys int }
	order := make([]item, 0, upper-lower+1)
	for level := lower; level <= upper; level++ {
		idx := dd.InvPermZ[level]
		keys := 0
		for _, node := range dd.zddUnique {
			if int(node.Index) == idx {
				keys++
			}
		}
		order = append(order, item{idx, keys})
	}
	sort.SliceStable(order, func(i, j int) bool { return order[i].keys > order[j].keys })
	for i := 0; i < len(order) && i < dd.siftMaxVar && state.swaps < dd.siftMaxSwap; i++ {
		variable := order[i].variable
		initial := zddHeapSize(dd)
		moves := []zddSiftMove{}
		limit := initial
		position := dd.PermZ[variable]
		directions := []int{-1, 1}
		if position == lower {
			directions = []int{1}
		} else if position == upper {
			directions = []int{-1}
		} else if position-lower > upper-position {
			directions = []int{1, -1}
		}
		for _, dir := range directions {
			for position+dir >= lower && position+dir <= upper && state.swaps < dd.siftMaxSwap {
				next := position + dir
				a, b := position, next
				if a > b {
					a, b = b, a
				}
				if !swapZddAdjacentInPlace(dd, a, b) {
					return false
				}
				state.swaps++
				position = next
				size := zddHeapSize(dd)
				moves = append(moves, zddSiftMove{a, b, size})
				if float64(size) > float64(limit)*dd.maxGrowth {
					break
				}
				if size < limit {
					limit = size
				}
			}
		}
		if !zddSiftingBackward(dd, moves, initial) {
			return false
		}
	}
	return true
}

func zddFindNodeHiLo(dd *DdManager, n *mtr.MtrNode) (int, int) {
	if n == nil || int(n.Low) >= dd.ZddSize {
		return dd.ZddSize, -1
	}
	if int(n.Index) >= len(dd.PermZ) {
		return dd.ZddSize, -1
	}
	lo := dd.PermZ[n.Index]
	hi := lo + int(n.Size) - 1
	if hi >= dd.ZddSize {
		if n.Child == nil {
			hi = dd.ZddSize - 1
		} else {
			hi = dd.ZddSize - 1
			for c := n.Child; c != nil; c = c.Younger {
				cl := dd.PermZ[c.Index]
				ch := cl + int(c.Size) - 1
				if ch >= dd.ZddSize && cl < dd.ZddSize {
					hi = cl - 1
				}
			}
		}
	}
	return lo, hi
}

func zddGroupSifting(dd *DdManager, lower, upper int, s *zddGroupState) bool {
	type candidate struct{ level, keys int }
	vars := make([]candidate, 0, dd.ZddSize)
	for variable := 0; variable < dd.ZddSize; variable++ {
		level := dd.PermZ[variable]
		if level >= s.next[level] {
			keys := 0
			for _, n := range dd.zddUnique {
				if int(n.Index) == variable {
					keys++
				}
			}
			vars = append(vars, candidate{variable, keys})
		}
	}
	sort.SliceStable(vars, func(i, j int) bool { return vars[i].keys > vars[j].keys })
	sifted := make([]bool, dd.ZddSize)
	limit := min(dd.siftMaxVar, len(vars))
	for _, v := range vars[:limit] {
		if s.swaps >= dd.siftMaxSwap {
			break
		}
		xindex := v.level
		if sifted[xindex] {
			continue
		}
		x := dd.PermZ[xindex]
		if x < lower || x > upper {
			continue
		}
		if !zddGroupSiftingAux(dd, x, lower, upper, s) {
			return false
		}
		x = dd.PermZ[xindex]
		if x != s.next[x] {
			start := x
			for {
				j := dd.InvPermZ[x]
				sifted[j] = true
				x = s.next[x]
				if x == start {
					break
				}
			}
		}
	}
	return true
}

func zddGroupSiftingAux(dd *DdManager, x, low, high int, s *zddGroupState) bool {
	initial := zddHeapSize(dd)
	moves := []zddGroupMoveRecord{}
	if x == low {
		if x == high {
			return true
		}
		if !zddGroupDown(dd, x, high, &moves, s) {
			return false
		}
		return zddGroupBackward(dd, moves, initial, s)
	} else if zddNextHigh(dd, x) > high {
		x = s.next[x]
		if !zddGroupUp(dd, x, low, &moves, s) {
			return false
		}
		return zddGroupBackward(dd, moves, initial, s)
	} else if x-low > high-x {
		if !zddGroupDown(dd, x, high, &moves, s) {
			return false
		}
		if len(moves) > 0 {
			x = moves[0].y
		}
		for x < s.next[x] {
			x = s.next[x]
		}
		x = s.next[x]
		if !zddGroupUp(dd, x, low, &moves, s) {
			return false
		}
		return zddGroupBackward(dd, moves, initial, s)
	}
	x = s.next[x]
	if !zddGroupUp(dd, x, low, &moves, s) {
		return false
	}
	if len(moves) > 0 {
		x = moves[0].x
	}
	for x < s.next[x] {
		x = s.next[x]
	}
	if !zddGroupDown(dd, x, high, &moves, s) {
		return false
	}
	return zddGroupBackward(dd, moves, initial, s)
}
func zddNextHigh(dd *DdManager, x int) int { return x + 1 }
func zddNextLow(dd *DdManager, x int) int  { return x - 1 }
func zddSwapGroupLevels(dd *DdManager, x, y int, s *zddGroupState) int {
	if x > y {
		x, y = y, x
	}
	if y != x+1 || s.swaps >= dd.siftMaxSwap || !swapZddAdjacentInPlace(dd, x, y) {
		return 0
	}
	s.swaps++
	return zddHeapSize(dd)
}
func zddSwapGroupLevelsBackward(dd *DdManager, x, y int, s *zddGroupState) int {
	if x > y {
		x, y = y, x
	}
	if y != x+1 || !swapZddAdjacentInPlace(dd, x, y) {
		return 0
	}
	s.swaps++
	return zddHeapSize(dd)
}
func zddGroupUp(dd *DdManager, y, low int, moves *[]zddGroupMoveRecord, s *zddGroupState) bool {
	limit := zddHeapSize(dd)
	x := zddNextLow(dd, y)
	for x >= low {
		top := s.next[x]
		if s.next[x] == x && s.next[y] == y {
			size := zddSwapGroupLevels(dd, x, y, s)
			if size == 0 {
				return false
			}
			*moves = append([]zddGroupMoveRecord{{x, y, size}}, *moves...)
			if float64(size) > float64(limit)*dd.maxGrowth {
				return true
			}
			if size < limit {
				limit = size
			}
		} else {
			size := zddGroupMove(dd, x, y, moves, s)
			if size == 0 {
				return false
			}
			if float64(size) > float64(limit)*dd.maxGrowth {
				return true
			}
			if size < limit {
				limit = size
			}
		}
		y = top
		x = zddNextLow(dd, y)
	}
	return true
}
func zddGroupDown(dd *DdManager, x, high int, moves *[]zddGroupMoveRecord, s *zddGroupState) bool {
	limit := zddHeapSize(dd)
	y := zddNextHigh(dd, x)
	for y <= high {
		bottom := s.next[y]
		for s.next[bottom] != y {
			bottom = s.next[bottom]
		}
		if s.next[x] == x && s.next[y] == y {
			size := zddSwapGroupLevels(dd, x, y, s)
			if size == 0 {
				return false
			}
			*moves = append([]zddGroupMoveRecord{{x, y, size}}, *moves...)
			if float64(size) > float64(limit)*dd.maxGrowth {
				return true
			}
			if size < limit {
				limit = size
			}
			x = y
			y = zddNextHigh(dd, x)
		} else {
			size := zddGroupMove(dd, x, y, moves, s)
			if size == 0 {
				return false
			}
			if float64(size) > float64(limit)*dd.maxGrowth {
				return true
			}
			if size < limit {
				limit = size
			}
		}
		x = bottom
		y = zddNextHigh(dd, x)
	}
	return true
}
func zddGroupMove(dd *DdManager, x, y int, moves *[]zddGroupMoveRecord, s *zddGroupState) int {
	xtop := s.next[x]
	xsize := x - xtop + 1
	ytop := y
	ybot := y
	for ybot < s.next[ybot] {
		ybot = s.next[ybot]
	}
	ysize := ybot - ytop + 1
	swapx, swapy := -1, -1
	size := 0
	for i := 1; i <= ysize; i++ {
		for j := 1; j <= xsize; j++ {
			size = zddSwapGroupLevels(dd, x, y, s)
			if size == 0 {
				return 0
			}
			swapx, swapy = x, y
			y = x
			x = zddNextLow(dd, y)
		}
		y = ytop + i
		x = zddNextLow(dd, y)
	}
	y = xtop
	for i := 0; i < ysize-1; i++ {
		s.next[y] = zddNextHigh(dd, y)
		y = zddNextHigh(dd, y)
	}
	s.next[y] = xtop
	x = zddNextHigh(dd, y)
	newxtop := x
	for i := 0; i < xsize-1; i++ {
		s.next[x] = zddNextHigh(dd, x)
		x = zddNextHigh(dd, x)
	}
	s.next[x] = newxtop
	*moves = append([]zddGroupMoveRecord{{swapx, swapy, zddHeapSize(dd)}}, *moves...)
	return zddHeapSize(dd)
}
func zddGroupMoveBackward(dd *DdManager, x, y int, s *zddGroupState) bool {
	xtop := s.next[x]
	xsize := x - xtop + 1
	ytop := y
	ybot := y
	for ybot < s.next[ybot] {
		ybot = s.next[ybot]
	}
	ysize := ybot - ytop + 1
	for i := 1; i <= ysize; i++ {
		for j := 1; j <= xsize; j++ {
			if zddSwapGroupLevelsBackward(dd, x, y, s) == 0 {
				return false
			}
			y = x
			x = zddNextLow(dd, y)
		}
		y = ytop + i
		x = zddNextLow(dd, y)
	}
	y = xtop
	for i := 0; i < ysize-1; i++ {
		s.next[y] = zddNextHigh(dd, y)
		y = zddNextHigh(dd, y)
	}
	s.next[y] = xtop
	x = zddNextHigh(dd, y)
	top := x
	for i := 0; i < xsize-1; i++ {
		s.next[x] = zddNextHigh(dd, x)
		x = zddNextHigh(dd, x)
	}
	s.next[x] = top
	return true
}
func zddGroupBackward(dd *DdManager, moves []zddGroupMoveRecord, initial int, s *zddGroupState) bool {
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
		if s.next[m.x] == m.x && s.next[m.y] == m.y {
			if zddSwapGroupLevelsBackward(dd, m.x, m.y, s) == 0 {
				return false
			}
		} else if !zddGroupMoveBackward(dd, m.x, m.y, s) {
			return false
		}
	}
	return true
}
func zddMergeGroups(dd *DdManager, n *mtr.MtrNode, low, high int, s *zddGroupState) {
	if n != dd.TreeZ {
		for i := low; i < high; i++ {
			s.next[i] = i + 1
		}
		s.next[high] = low
	}
	save := int(n.Index)
	id := dd.InvPermZ[low]
	a := n
	for {
		a.Index = uint32(id)
		if a.Parent == nil || int(a.Parent.Index) != save {
			break
		}
		a = a.Parent
	}
}
