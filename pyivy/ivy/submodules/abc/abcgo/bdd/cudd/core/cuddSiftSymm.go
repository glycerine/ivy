package core

import "sort"

// Symmetric sifting (cuddSymmetry.c) on a siftHeap. next[l] links the
// levels of a symmetry group: each level points to the next one below, and
// the bottom points back to the top, as subtables[l].next in CUDD.

// symmCheck is cuddSymmCheck(x, y).
func (h *siftHeap) symmCheck(x, y int) bool {
	if h.levelKeys(x) == 1 {
		return false
	}
	yindex := int32(h.invperm[y])
	if h.levelKeys(y) == 1 && h.varRefIsolated(int(yindex)) {
		return false
	}
	xsymmy, xsymmyp := true, true
	arccount := 0
	for _, f := range h.levels[x].nodes {
		nf := h.nodes[f]
		f1 := nf.t
		f0 := siftMk(siftReg(nf.e), false)
		comple := siftCompl(nf.e)
		var f11, f10, f01, f00 siftEdge
		if h.indexOf(f1) == yindex {
			arccount++
			f11, f10 = h.nodes[siftReg(f1)].t, h.nodes[siftReg(f1)].e
		} else {
			if h.indexOf(f0) != yindex {
				// an isolated projection function may bypass layer y
				if f1 != 0 || f0 != 0 || nf.ref != 1 {
					return false
				}
			}
			f11, f10 = f1, f1
		}
		if h.indexOf(f0) == yindex {
			arccount++
			f01, f00 = h.nodes[siftReg(f0)].t, h.nodes[siftReg(f0)].e
		} else {
			f01, f00 = f0, f0
		}
		if comple {
			f01, f00 = siftNot(f01), siftNot(f00)
		}
		if f1 != 0 || f0 != 0 || nf.ref != 1 {
			xsymmy = xsymmy && f01 == f10
			xsymmyp = xsymmyp && f11 == f00
			if !xsymmy && !xsymmyp {
				return false
			}
		}
	}
	total := -1 // minus the projection function
	for _, f := range h.levels[y].nodes {
		total += int(h.nodes[f].ref)
	}
	return arccount == total
}

// symmGroupMove is ddSymmGroupMove; it pushes one move.
func (h *siftHeap) symmGroupMove(x, y int, moves *[]siftMove) int {
	size := -1
	xbot, xtop := x, h.next[x]
	xsize := xbot - xtop + 1
	ybot := y
	for ybot < h.next[ybot] {
		ybot = h.next[ybot]
	}
	ytop := y
	ysize := ybot - ytop + 1
	swapx, swapy := -1, -1
	for i := 1; i <= ysize; i++ {
		for j := 1; j <= xsize; j++ {
			size = h.swap(x)
			swapx, swapy = x, y
			y = x
			x = y - 1
		}
		y = ytop + i
		x = y - 1
	}
	h.fixGroups(xtop, xsize, ysize)
	*moves = append([]siftMove{{swapx, swapy, size}}, *moves...)
	return size
}

// fixGroups relinks the two groups after a group move ("fix symmetries").
func (h *siftHeap) fixGroups(xtop, xsize, ysize int) {
	y := xtop
	for i := 0; i < ysize-1; i++ {
		h.next[y] = y + 1
		y++
	}
	h.next[y] = xtop
	x := y + 1
	newxtop := x
	for i := 0; i < xsize-1; i++ {
		h.next[x] = x + 1
		x++
	}
	h.next[x] = newxtop
}

// symmGroupMoveBackward is ddSymmGroupMoveBackward.
func (h *siftHeap) symmGroupMoveBackward(x, y int) int {
	size := -1
	xbot, xtop := x, h.next[x]
	xsize := xbot - xtop + 1
	ybot := y
	for ybot < h.next[ybot] {
		ybot = h.next[ybot]
	}
	ytop := y
	ysize := ybot - ytop + 1
	for i := 1; i <= ysize; i++ {
		for j := 1; j <= xsize; j++ {
			size = h.swap(x)
			y = x
			x = y - 1
		}
		y = ytop + i
		x = y - 1
	}
	h.fixGroups(xtop, xsize, ysize)
	return size
}

func (h *siftHeap) zInteract(z, index int) bool {
	zindex := h.invperm[z]
	return zindex == index || h.testInteract(zindex, index)
}

// symmSiftUp is ddSymmSiftingUp.
func (h *siftHeap) symmSiftUp(y, xLow int, dd *DdManager) []siftMove {
	var moves []siftMove
	yindex := h.invperm[y]
	limitSize := h.cost()
	L := limitSize
	gybot := y
	for gybot < h.next[gybot] {
		gybot = h.next[gybot]
	}
	for z := xLow + 1; z <= gybot; z++ {
		if h.zInteract(z, yindex) {
			L -= h.levelKeys(z) - boolToInt(h.varRefIsolated(h.invperm[z]))
		}
	}
	x := y - 1
	for x >= xLow && L <= limitSize {
		gxtop := h.next[x]
		if h.symmCheck(x, y) {
			h.next[x] = y
			i := h.next[y]
			for h.next[i] != y {
				i = h.next[i]
			}
			h.next[i] = gxtop
		} else if h.next[x] == x && h.next[y] == y {
			xindex := h.invperm[x]
			size := h.swap(x)
			if h.testInteract(xindex, yindex) {
				L += h.levelKeys(y) - boolToInt(h.varRefIsolated(xindex))
			}
			moves = append([]siftMove{{x, y, size}}, moves...)
			if float64(size) > float64(limitSize)*dd.maxGrowth {
				return moves
			}
			if size < limitSize {
				limitSize = size
			}
		} else {
			size := h.symmGroupMove(x, y, &moves)
			z := moves[0].y
			for {
				zindex := h.invperm[z]
				if h.testInteract(zindex, yindex) {
					L += h.levelKeys(z) - boolToInt(h.varRefIsolated(zindex))
				}
				z = h.next[z]
				if z == moves[0].y {
					break
				}
			}
			if float64(size) > float64(limitSize)*dd.maxGrowth {
				return moves
			}
			if size < limitSize {
				limitSize = size
			}
		}
		y = gxtop
		x = y - 1
	}
	return moves
}

// symmSiftDown is ddSymmSiftingDown.
func (h *siftHeap) symmSiftDown(x, xHigh int, dd *DdManager) []siftMove {
	var moves []siftMove
	xindex := h.invperm[x]
	gxtop := h.next[x]
	size := h.cost()
	limitSize := size
	R := 0
	for z := xHigh; z > gxtop; z-- {
		if h.zInteract(z, xindex) {
			R += h.levelKeys(z) - boolToInt(h.varRefIsolated(h.invperm[z]))
		}
	}
	y := x + 1
	for y <= xHigh && size-R < limitSize {
		gybot := h.next[y]
		for h.next[gybot] != y {
			gybot = h.next[gybot]
		}
		if h.symmCheck(x, y) {
			gxtop = h.next[x]
			h.next[x] = y
			h.next[gybot] = gxtop
		} else if h.next[x] == x && h.next[y] == y {
			yindex := h.invperm[y]
			if h.testInteract(xindex, yindex) {
				R -= h.levelKeys(y) - boolToInt(h.varRefIsolated(yindex))
			}
			size = h.swap(x)
			moves = append([]siftMove{{x, y, size}}, moves...)
			if float64(size) > float64(limitSize)*dd.maxGrowth {
				return moves
			}
			if size < limitSize {
				limitSize = size
			}
		} else {
			gxtop = h.next[x]
			z := gxtop + 1
			for {
				if h.zInteract(z, xindex) {
					R -= h.levelKeys(z) - boolToInt(h.varRefIsolated(h.invperm[z]))
				}
				z++
				if z > gybot {
					break
				}
			}
			size = h.symmGroupMove(x, y, &moves)
			if float64(size) > float64(limitSize)*dd.maxGrowth {
				return moves
			}
			if size < limitSize {
				limitSize = size
			}
			gxtop = h.next[gybot]
			for z := gxtop + 1; z <= gybot; z++ {
				if h.zInteract(z, xindex) {
					R += h.levelKeys(z) - boolToInt(h.varRefIsolated(h.invperm[z]))
				}
			}
		}
		x = gybot
		y = x + 1
	}
	return moves
}

// symmSiftBackward is ddSymmSiftingBackward.
func (h *siftHeap) symmSiftBackward(moves []siftMove, size int) {
	for _, m := range moves {
		if m.size < size {
			size = m.size
		}
	}
	for _, m := range moves {
		if m.size == size {
			return
		}
		if h.next[m.x] == m.x && h.next[m.y] == m.y {
			h.swap(m.x)
		} else {
			h.symmGroupMoveBackward(m.x, m.y)
		}
	}
}

func (h *siftHeap) groupBottom(i int) int {
	for i < h.next[i] {
		i = h.next[i]
	}
	return i
}

// symmSiftAux is ddSymmSiftingAux.
func (h *siftHeap) symmSiftAux(x, xLow, xHigh int, dd *DdManager) {
	initialSize := h.cost()
	var moveDown, moveUp []siftMove
	if x-xLow > xHigh-x {
		for i := x; i > xLow; i-- {
			if !h.symmCheck(i-1, i) {
				break
			}
			topbot := h.next[i-1]
			h.next[i-1] = i
			h.next[x] = topbot
			i = topbot + 1
		}
	} else {
		for i := x; i < xHigh; i++ {
			if !h.symmCheck(i, i+1) {
				break
			}
			topbot := i + 1
			for topbot < h.next[topbot] {
				topbot = h.next[topbot]
			}
			h.next[topbot] = h.next[i]
			h.next[i] = i + 1
			i = topbot - 1
		}
	}
	x = h.groupBottom(x)
	switch {
	case x == xLow:
		if x == xHigh {
			return
		}
		initGroupSize := 1
		moveDown = h.symmSiftDown(x, xHigh, dd)
		if moveDown == nil {
			return
		}
		x = moveDown[0].y
		i := h.groupBottom(x)
		if initGroupSize == i-x+1 {
			h.symmSiftBackward(moveDown, initialSize)
		} else {
			initialSize = h.cost()
			moveUp = h.symmSiftUp(x, xLow, dd)
			h.symmSiftBackward(moveUp, initialSize)
		}
	case x+1 > xHigh:
		i := x
		x = h.next[x]
		if x == xLow {
			return
		}
		initGroupSize := i - x + 1
		moveUp = h.symmSiftUp(x, xLow, dd)
		if moveUp == nil {
			return
		}
		x = moveUp[0].x
		i = h.next[x]
		if initGroupSize == x-i+1 {
			h.symmSiftBackward(moveUp, initialSize)
		} else {
			initialSize = h.cost()
			moveDown = h.symmSiftDown(x, xHigh, dd)
			h.symmSiftBackward(moveDown, initialSize)
		}
	case x-xLow > xHigh-x:
		moveDown = h.symmSiftDown(x, xHigh, dd)
		var i int
		if moveDown != nil {
			x = moveDown[0].y
			i = h.groupBottom(x)
		} else {
			i = h.groupBottom(x)
			x = h.next[i]
		}
		initGroupSize := i - x + 1
		moveUp = h.symmSiftUp(x, xLow, dd)
		if moveUp != nil {
			x = moveUp[0].x
			i = h.next[x]
		} else {
			i = x
			x = h.groupBottom(x)
		}
		if initGroupSize == x-i+1 {
			h.symmSiftBackward(moveUp, initialSize)
		} else {
			initialSize = h.cost()
			moveDown = h.symmSiftDown(x, xHigh, dd)
			h.symmSiftBackward(moveDown, initialSize)
		}
	default:
		x = h.next[x]
		moveUp = h.symmSiftUp(x, xLow, dd)
		var i int
		if moveUp != nil {
			x = moveUp[0].x
			i = h.next[x]
		} else {
			x = h.groupBottom(x)
			i = h.next[x]
		}
		initGroupSize := x - i + 1
		moveDown = h.symmSiftDown(x, xHigh, dd)
		if moveDown != nil {
			x = moveDown[0].y
			i = h.groupBottom(x)
		} else {
			i = x
			x = h.next[x]
		}
		if initGroupSize == i-x+1 {
			h.symmSiftBackward(moveDown, initialSize)
		} else {
			initialSize = h.cost()
			moveUp = h.symmSiftUp(x, xLow, dd)
			h.symmSiftBackward(moveUp, initialSize)
		}
	}
}

// symmSifting is cuddSymmSifting(lower, upper).
func (h *siftHeap) symmSifting(lower, upper int, dd *DdManager) {
	entry := make([]int, h.n)
	vars := make([]int, h.n)
	for i := 0; i < h.n; i++ {
		entry[i] = h.levelKeys(h.perm[i])
		vars[i] = i
	}
	sort.SliceStable(vars, func(a, b int) bool { return entry[vars[a]] > entry[vars[b]] })
	h.next = make([]int, h.n)
	for i := range h.next {
		h.next[i] = i
	}
	for i := 0; i < min(dd.siftMaxVar, h.n); i++ {
		if h.swaps >= dd.siftMaxSwap {
			break
		}
		x := h.perm[vars[i]]
		if x < lower || x > upper {
			continue
		}
		if h.next[x] == x {
			h.symmSiftAux(x, lower, upper, dd)
		}
	}
}

// symmGroups lists the symmetry groups top to bottom as variable indices.
func (h *siftHeap) symmGroups() [][]int {
	var groups [][]int
	for l := 0; l < h.n; {
		bottom := h.groupBottom(l)
		var g []int
		for z := l; z <= bottom; z++ {
			g = append(g, h.invperm[z])
		}
		groups = append(groups, g)
		l = bottom + 1
	}
	return groups
}
