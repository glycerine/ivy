package core

// Window reordering follows cuddWindow.c. Candidate orders are evaluated on
// scratch canonical heaps; committing a selected order preserves live handles.

func Cudd_WindowReorder(dd *DdManager, low, high int, method Cudd_ReorderingType) int {
	if dd == nil || low < 0 || high >= len(dd.Vars) || low > high {
		return 0
	}
	switch method {
	case CUDD_REORDER_WINDOW2:
		return window2(dd, low, high, false)
	case CUDD_REORDER_WINDOW3:
		return window3(dd, low, high, false)
	case CUDD_REORDER_WINDOW4:
		return window4(dd, low, high, false)
	case CUDD_REORDER_WINDOW2_CONV:
		return window2(dd, low, high, true)
	case CUDD_REORDER_WINDOW3_CONV:
		return window3(dd, low, high, true)
	case CUDD_REORDER_WINDOW4_CONV:
		return window4(dd, low, high, true)
	default:
		return 0
	}
}

func windowCost(dd *DdManager, order []int) int {
	// reorderBuild counts nodes reachable from referenced roots and constants.
	// Unreferenced projections are pinned separately by liveReorderRoots but are
	// deliberately excluded from cost, matching C's keys - isolated measure.
	return reorderBuild(dd, dd.liveReorderRoots(), order).cost
}

func windowPreprocessSize(dd *DdManager) int {
	roots := dd.liveReorderRoots()
	order := append([]int(nil), dd.InvPerm...)
	heap := reorderBuild(dd, roots, order)
	size := heap.cost
	// The heap cost intentionally omits isolated projection functions. Cudd_ReduceHeap's
	// minsize check runs before that exclusion, so count projection handles absent
	// from the active root graph here.
	for _, variable := range roots.pins {
		if fresh := heap.memo[Cudd_Regular(variable)]; fresh == nil || !heap.nodes[fresh] {
			size++
		}
	}
	return size
}

func swapOrder(order []int, a, b int) []int {
	next := append([]int(nil), order...)
	next[a], next[b] = next[b], next[a]
	return next
}

func window2(dd *DdManager, low, high int, converge bool) int {
	if high-low < 1 {
		return 0
	}
	for {
		changed := false
		for x := low; x < high; x++ {
			order := append([]int(nil), dd.InvPerm...)
			candidate := swapOrder(order, x, x+1)
			if windowCost(dd, candidate) < windowCost(dd, order) {
				if Cudd_ShuffleHeap(dd, candidate) == 0 {
					return 0
				}
				changed = true
			}
		}
		if !converge || !changed {
			return 1
		}
	}
}

var window3Permutations = [][]int{
	{0, 1, 2}, // ABC
	{1, 0, 2}, // BAC
	{1, 2, 0}, // BCA
	{2, 1, 0}, // CBA
	{2, 0, 1}, // CAB
	{0, 2, 1}, // ACB
}

func bestWindow(dd *DdManager, start int, patterns [][]int) (int, []int) {
	base := append([]int(nil), dd.InvPerm...)
	bestOrder := base
	bestCost := windowCost(dd, base)
	bestCode := 1
	for code, pattern := range patterns {
		candidate := append([]int(nil), base...)
		for i, from := range pattern {
			candidate[start+i] = base[start+from]
		}
		cost := windowCost(dd, candidate)
		if cost < bestCost || cost == bestCost && code+1 < bestCode {
			bestCost, bestOrder, bestCode = cost, candidate, code+1
		}
	}
	return bestCode, bestOrder
}

func window3(dd *DdManager, low, high int, converge bool) int {
	if high-low < 2 {
		return window2(dd, low, high, converge)
	}
	for {
		changed := false
		for x := low; x+2 <= high; x++ {
			code, order := bestWindow(dd, x, window3Permutations)
			if code != 1 {
				if Cudd_ShuffleHeap(dd, order) == 0 {
					return 0
				}
				changed = true
			}
		}
		if !converge || !changed {
			return 1
		}
	}
}

// These are the source's Hamiltonian-circuit swap sequence and code order.
// Building the patterns from it prevents an independently maintained list
// from drifting away from cuddWindow.c.
func window4Patterns() []struct {
	code int
	p    []int
	tie  bool
} {
	codes := []int{1, 7, 13, 8, 14, 9, 15, 20, 23, 19, 21, 24, 22, 18, 12, 17, 11, 16, 10, 5, 3, 6, 4, 2}
	// cuddWindow.c only considers lower-code ties at the cases whose condition
	// explicitly includes sizeNew == size && CODE < best.
	tie := []bool{false, false, false, true, false, true, false, false, false, true, true, false, true, true, true, true, true, true, true, true, true, true, true, true}
	swaps := [][2]int{{0, 1}, {2, 3}, {0, 1}, {1, 2}, {2, 3}, {0, 1}, {2, 3}, {1, 2}}
	p := []int{0, 1, 2, 3}
	out := make([]struct {
		code int
		p    []int
		tie  bool
	}, len(codes))
	for i, code := range codes {
		if i != 0 {
			s := swaps[(i-1)%len(swaps)]
			p[s[0]], p[s[1]] = p[s[1]], p[s[0]]
		}
		out[i] = struct {
			code int
			p    []int
			tie  bool
		}{code, append([]int(nil), p...), tie[i]}
	}
	return out
}

func window4Select(bestCode, bestCost, code, cost int, tieAllowed bool) (int, int) {
	if cost < bestCost || cost == bestCost && tieAllowed && code < bestCode {
		return code, cost
	}
	return bestCode, bestCost
}

func window4(dd *DdManager, low, high int, converge bool) int {
	if high-low < 3 {
		return window3(dd, low, high, converge)
	}
	for {
		changed := false
		for x := low; x+3 <= high; x++ {
			base := append([]int(nil), dd.InvPerm...)
			bestOrder, bestCode := base, 1
			bestCost := windowCost(dd, base)
			for _, item := range window4Patterns() {
				candidate := append([]int(nil), base...)
				for i, from := range item.p {
					candidate[x+i] = base[x+from]
				}
				cost := windowCost(dd, candidate)
				newCode, newCost := window4Select(bestCode, bestCost, item.code, cost, item.tie)
				if newCode != bestCode {
					bestCost, bestOrder, bestCode = newCost, candidate, newCode
				}
			}
			if bestCode != 1 {
				if Cudd_ShuffleHeap(dd, bestOrder) == 0 {
					return 0
				}
				changed = true
			}
		}
		if !converge || !changed {
			return 1
		}
	}
}
