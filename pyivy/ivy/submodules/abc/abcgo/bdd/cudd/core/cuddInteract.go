package core

import "slices"

// Rows of the packed triangle end at the larger variable index. Existing
// pairs retain their bit positions when a new variable is appended.
func interactionRow(y int) uint64 { return uint64(y) * uint64(y-1) / 2 }

func (dd *DdManager) setInteractionRange(first, last uint64) {
	lo, hi := int(first/64), int(last/64)
	if hi >= len(dd.interact) {
		dd.interact = append(dd.interact, make([]uint64, hi+1-len(dd.interact))...)
	}
	left := ^uint64(0) << (first % 64)
	right := ^uint64(0) >> (63 - last%64)
	if lo == hi {
		dd.interact[lo] |= left & right
		return
	}
	dd.interact[lo] |= left
	for i := lo + 1; i < hi; i++ {
		dd.interact[i] = ^uint64(0)
	}
	dd.interact[hi] |= right
}

// Cudd_SetInteraction and Cudd_TestInteraction maintain the symmetric
// variable interaction relation used by adjacent-swap reordering.
func Cudd_SetInteraction(dd *DdManager, x, y int) {
	if x > y {
		x, y = y, x
	}
	if dd != nil && x >= 0 && x < y && y < len(dd.Vars) {
		bit := interactionRow(y) + uint64(x)
		dd.setInteractionRange(bit, bit)
	}
}

func Cudd_TestInteraction(dd *DdManager, x, y int) bool {
	if x > y {
		x, y = y, x
	}
	if dd == nil || x < 0 || x == y || y >= len(dd.Vars) {
		return false
	}
	bit := interactionRow(y) + uint64(x)
	return bit/64 < uint64(len(dd.interact)) && dd.interact[bit/64]&(uint64(1)<<(bit%64)) != 0
}

func (dd *DdManager) discardInteractionsAfterSize(size int) {
	bits := interactionRow(size)
	words := bits / 64
	if words >= uint64(len(dd.interact)) {
		return
	}
	if tail := bits % 64; tail != 0 {
		dd.interact[words] &= (uint64(1) << tail) - 1
		words++
	}
	dd.interact = dd.interact[:words]
}

// Cudd_ComputeInteractions sets every pair of variable indices in the support
// of a resident BDD or ADD root. Process levels from top to bottom so a node
// reached from above never needs a separate search as a root.
func Cudd_ComputeInteractions(dd *DdManager) {
	if dd == nil || dd.closed {
		return
	}
	clear(dd.interact)
	var work []*DdNode
	var support []int
	type interval struct{ first, last int }
	var runs []interval
	present := make([]bool, len(dd.Vars))
	const (
		previousRoot = 1
		currentRoot  = 2
	)
	enqueue := func(n *DdNode) {
		n = Cudd_Regular(n)
		if n.Index != bddConstantIndex && n.interactionMark != currentRoot {
			n.interactionMark = currentRoot
			work = append(work, n)
		}
	}
	visit := func(entry uniqueEntry) {
		root := entry.node
		if root.interactionMark != 0 || Cudd_IsConstant(entry.t) && Cudd_IsConstant(entry.e) {
			return
		}
		work = work[:0]
		support = support[:0]
		enqueue(root)
		for i := 0; i < len(work); i++ {
			n := work[i]
			if !present[n.Index] {
				present[n.Index] = true
				support = append(support, int(n.Index))
			}
			enqueue(n.T)
			enqueue(n.E)
		}
		for _, n := range work {
			n.interactionMark = previousRoot
		}
		slices.Sort(support)
		runs = runs[:0]
		for _, y := range support {
			present[y] = false
			row := interactionRow(y)
			for _, run := range runs {
				dd.setInteractionRange(row+uint64(run.first), row+uint64(run.last))
			}
			last := len(runs) - 1
			if last >= 0 && runs[last].last+1 == y {
				runs[last].last = y
			} else {
				runs = append(runs, interval{y, y})
			}
		}
	}
	for _, index := range dd.InvPerm {
		dd.unique.forLevel(index, visit)
		dd.addUnique.forLevel(index, visit)
	}
	for _, table := range []*uniqueTable{&dd.unique, &dd.addUnique} {
		for _, n := range table.entries {
			n.interactionMark = 0
		}
	}
}
