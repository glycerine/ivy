package core

// swapZddAdjacentInPlace swaps adjacent ZDD levels x and y, where x < y.
// It is the Go equivalent of cuddZddSwapInPlace: nodes at x that test y are
// factored through y, and their existing handles become y nodes. Other x
// nodes and all unaffected handles remain in place.
func swapZddAdjacentInPlace(dd *DdManager, x, y int) bool {
	if dd == nil || x < 0 || y != x+1 || y >= dd.ZddSize || len(dd.InvPermZ) != dd.ZddSize || len(dd.PermZ) != dd.ZddSize {
		return false
	}
	xi, yi := dd.InvPermZ[x], dd.InvPermZ[y]
	if xi < 0 || yi < 0 || xi >= dd.ZddSize || yi >= dd.ZddSize || dd.PermZ[xi] != x || dd.PermZ[yi] != y {
		return false
	}
	// Only x nodes with a y child change function decomposition. Remove all
	// their old keys before interning new x nodes against the retained layer.
	var moved []*DdNode
	dd.zddUnique.forLevel(xi, func(entry uniqueEntry) {
		if entry.t.Index == uint32(yi) || entry.e.Index == uint32(yi) {
			moved = append(moved, entry.node)
		}
	})
	for _, n := range moved {
		dd.zddUnique.remove(bddUniqueKey{uint32(xi), n.T, n.E})
	}

	for _, f := range moved {
		f1, f0 := f.T, f.E
		f11, f10 := dd.Zero, f1
		if f1.Index == uint32(yi) {
			f11, f10 = f1.T, f1.E
		}
		f01, f00 := dd.Zero, f0
		if f0.Index == uint32(yi) {
			f01, f00 = f0.T, f0.E
		}
		nt, ne := dd.internReorderedZDD(uint32(xi), f11, f01), dd.internReorderedZDD(uint32(xi), f10, f00)
		if !f.childrenReleased {
			Cudd_Ref(nt)
			Cudd_Ref(ne)
			dd.recursiveDeref(f1)
			dd.recursiveDeref(f0)
		}
		f.Index, f.inverse.Index = uint32(yi), uint32(yi)
		f.T, f.E = nt, ne
		dd.zddUnique.set(bddUniqueKey{f.Index, nt, ne}, f)
	}
	if y < len(dd.zddTableSlots) {
		dd.zddTableSlots[x], dd.zddTableSlots[y] = dd.zddTableSlots[y], dd.zddTableSlots[x]
	}
	dd.PermZ[xi], dd.PermZ[yi] = y, x
	dd.InvPermZ[x], dd.InvPermZ[y] = yi, xi
	if len(dd.Univ) > y && dd.Univ[x] != nil {
		dd.Univ[y] = dd.Univ[x].T
	}
	Cudd_CacheFlush(dd)

	return true
}

// A reordered key acquires child references when a live parent first uses it.
func (dd *DdManager) internReorderedZDD(index uint32, t, e *DdNode) *DdNode {
	if t == dd.Zero {
		return e
	}
	key := bddUniqueKey{index, t, e}
	entry := dd.zddUnique.find(key)
	if entry != nil && entry.node != nil {
		return entry.node
	}
	n := dd.newNodePair(index)
	n.T, n.E, n.family = t, e, 1
	// A live rewritten parent acquires the child's arcs through Cudd_Ref.
	// A released parent carries the key without owning its arcs.
	n.childrenReleased = true
	dd.deadZ++
	dd.zddUnique.insert(key, n, entry)
	return n
}
