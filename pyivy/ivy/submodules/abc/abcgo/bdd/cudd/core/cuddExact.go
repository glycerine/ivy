package core

// Cudd_ExactReorder searches every ordering in the requested level interval.
// To keep accidental factorial work bounded, ranges wider than maxExactRange
// return false without changing the manager.
func Cudd_ExactReorder(dd *DdManager, lower, upper int) bool {
	const maxExactRange = 8
	if dd == nil || lower < 0 || upper < lower || upper >= len(dd.InvPerm) || upper-lower+1 > maxExactRange {
		return false
	}
	base := append([]int(nil), dd.InvPerm...)
	items := append([]int(nil), base[lower:upper+1]...)
	best := append([]int(nil), base...)
	bestCost := int(^uint(0) >> 1)
	var search func(int)
	search = func(pos int) {
		if pos == len(items) {
			order := append([]int(nil), base...)
			copy(order[lower:upper+1], items)
			if Cudd_ShuffleHeap(dd, order) == 0 {
				return
			}
			cost := len(dd.unique)
			if cost < bestCost {
				bestCost = cost
				copy(best, order)
			}
			return
		}
		for i := pos; i < len(items); i++ {
			items[pos], items[i] = items[i], items[pos]
			search(pos + 1)
			items[pos], items[i] = items[i], items[pos]
		}
	}
	search(0)
	if bestCost == int(^uint(0)>>1) {
		Cudd_ShuffleHeap(dd, base)
		return false
	}
	return Cudd_ShuffleHeap(dd, best) != 0
}
