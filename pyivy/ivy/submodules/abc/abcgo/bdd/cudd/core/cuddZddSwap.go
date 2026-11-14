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
	// Save the original x layer before changing any keys. The C implementation
	// removes y-dependent x nodes from its x bucket and processes that FIFO.
	xnodes := make([]*DdNode, 0)
	for _, n := range dd.zddUnique {
		if n.Index == uint32(xi) {
			xnodes = append(xnodes, n)
		}
	}
	// A temporary canonical table for newly created x nodes. Existing x nodes
	// that stay at x are also candidates, matching the source's collision scan.
	work := make(map[bddUniqueKey]*DdNode, len(dd.zddUnique))
	for k, n := range dd.zddUnique {
		if n.Index != uint32(xi) {
			work[k] = n
		}
	}
	for _, n := range xnodes {
		if n.T.Index != uint32(yi) && n.E.Index != uint32(yi) {
			work[bddUniqueKey{n.Index, n.T, n.E}] = n
		}
	}
	internX := func(t, e *DdNode) *DdNode {
		if t == dd.Zero {
			return e
		}
		k := bddUniqueKey{uint32(xi), t, e}
		if n := work[k]; n != nil {
			return n
		}
		n := &DdNode{Index: uint32(xi), T: t, E: e, ID: dd.nextID, manager: dd, family: 1}
		n.inverse = &DdNode{Index: uint32(xi), ID: dd.nextID + 1, manager: dd, regular: n, inverse: n}
		dd.nextID += 2
		work[k] = n
		return n
	}
	for _, f := range xnodes {
		f1, f0 := f.T, f.E
		if f1.Index != uint32(yi) && f0.Index != uint32(yi) {
			continue
		}
		f11, f10 := dd.Zero, f1
		if f1.Index == uint32(yi) {
			f11, f10 = f1.T, f1.E
		}
		f01, f00 := dd.Zero, f0
		if f0.Index == uint32(yi) {
			f01, f00 = f0.T, f0.E
		}
		var nt *DdNode
		if f11 == dd.Zero {
			nt = f01
		} else {
			nt = internX(f11, f01)
		}
		var ne *DdNode
		if f10 == dd.Zero {
			ne = f00
		} else {
			ne = internX(f10, f00)
		}
		f.Index, f.T, f.E = uint32(yi), nt, ne
	}
	// Install rewritten nodes and update the level permutation. As in CUDD,
	// existing node handles are retained; the Go manager has no subtable ref
	// counts, so zero-ref collection is deliberately not performed here.
	for _, n := range xnodes {
		if n.Index == uint32(yi) {
			delete(work, bddUniqueKey{uint32(xi), n.T, n.E})
			work[bddUniqueKey{n.Index, n.T, n.E}] = n
		}
	}
	dd.zddUnique = work
	dd.PermZ[xi], dd.PermZ[yi] = y, x
	dd.InvPermZ[x], dd.InvPermZ[y] = yi, xi
	if len(dd.Univ) > y && dd.Univ[x] != nil {
		dd.Univ[y] = dd.Univ[x].T
	}
	clear(dd.cache)
	clear(dd.computed)
	return true
}
