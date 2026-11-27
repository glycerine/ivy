package dsd

import bdd "github.com/glycerine/abcgo/bdd/cudd/core"

// Support-cube and cube helpers of extraBddMisc.c (Extra_bddSupp*,
// Extra_bddGetOneCube, Extra_bddFindOneCube) and cuddUtil.c's DAG counts,
// walking the BDD pointers as the C does.

// extraBddSuppSize is Extra_bddSuppSize.
func extraBddSuppSize(dd *DdManager, bSupp *DdNode) int {
	Counter := 0
	for bSupp != dd.One {
		bSupp = bSupp.T
		Counter++
	}
	return Counter
}

// extraBddSuppContainVar is Extra_bddSuppContainVar.
func extraBddSuppContainVar(dd *DdManager, bS, bVar *DdNode) bool {
	for ; bS != dd.One; bS = bS.T {
		if bS.Index == bVar.Index {
			return true
		}
	}
	return false
}

// extraBddSuppOverlapping is Extra_bddSuppOverlapping.
func extraBddSuppOverlapping(dd *DdManager, S1, S2 *DdNode) bool {
	for S1.Index != bdd.CUDD_CONST_INDEX && S2.Index != bdd.CUDD_CONST_INDEX {
		// if the top vars are the same, they intersect
		if S1.Index == S2.Index {
			return true
		}
		// if the top vars are different, skip the one, which is higher
		if dd.Perm[S1.Index] < dd.Perm[S2.Index] {
			S1 = S1.T
		} else {
			S2 = S2.T
		}
	}
	return false
}

// suppOverlap reports whether two support cubes share a variable.
func suppOverlap(dd *DdManager, a, b *DdNode) bool { return extraBddSuppOverlapping(dd, a, b) }

// extraBddSuppCheckContainment is Extra_bddSuppCheckContainment: returns
// true if one support is contained in another; in this case, bLarge
// (bSmall) is assigned to point to the larger (smaller) support. If the
// supports are identical, returns false and does not assign the supports.
func extraBddSuppCheckContainment(dd *DdManager, bL, bH *DdNode, bLarge, bSmall **DdNode) bool {
	bSL := bL
	bSH := bH
	fLcontainsH := true
	fHcontainsL := true
	var TopVar uint32

	if bSL == bSH {
		return false
	}

	for bSL != dd.One || bSH != dd.One {
		if bSL == dd.One { // Low component has no vars; High components has some vars
			fLcontainsH = false
			if !fHcontainsL {
				return false
			}
			break
		}

		if bSH == dd.One { // similarly
			fHcontainsL = false
			if !fLcontainsH {
				return false
			}
			break
		}

		// determine the topmost var of the supports by comparing their levels
		if dd.Perm[bSL.Index] < dd.Perm[bSH.Index] {
			TopVar = bSL.Index
		} else {
			TopVar = bSH.Index
		}

		if TopVar == bSL.Index && TopVar == bSH.Index { // they are on the same level
			// it does not tell us anything about their containment
			// skip this var
			bSL = bSL.T
			bSH = bSH.T
		} else if TopVar == bSL.Index { // and TopVar != bSH->index
			// Low components is higher and contains more vars
			// it is not possible that High component contains Low
			fHcontainsL = false
			// skip this var
			bSL = bSL.T
		} else { // if ( TopVar == bSH->index ) // and TopVar != bSL->index
			// similarly
			fLcontainsH = false
			// skip this var
			bSH = bSH.T
		}

		// check the stopping condition
		if !fHcontainsL && !fLcontainsH {
			return false
		}
	}
	// only one of them can be true at the same time
	if fHcontainsL && fLcontainsH {
		panic("dsd: Extra_bddSuppCheckContainment: !fHcontainsL || !fLcontainsH")
	}
	if fHcontainsL {
		*bLarge = bH
		*bSmall = bL
	} else { // fLcontainsH
		*bLarge = bL
		*bSmall = bH
	}
	return true
}

// extraBddGetOneCube is Extra_bddGetOneCube: the cube with the smallest
// bits-to-integer value.
func extraBddGetOneCube(dd *DdManager, bFunc *DdNode) *DdNode {
	var bFunc0, bFunc1, bRes0, bRes1, bRes *DdNode

	bFuncR := bdd.Cudd_Regular(bFunc)
	if bdd.Cudd_IsConstant(bFuncR) {
		return bFunc
	}

	// cofactor
	if bdd.Cudd_IsComplement(bFunc) {
		bFunc0 = bdd.Cudd_Not(bFuncR.E)
		bFunc1 = bdd.Cudd_Not(bFuncR.T)
	} else {
		bFunc0 = bFuncR.E
		bFunc1 = bFuncR.T
	}

	// try to find the cube with the negative literal
	bRes0 = extraBddGetOneCube(dd, bFunc0)
	bdd.Cudd_Ref(bRes0)

	if bRes0 != dd.LogicZero {
		bRes = bdd.Cudd_bddAnd(dd, bRes0, bdd.Cudd_Not(dd.Vars[bFuncR.Index]))
		bdd.Cudd_Ref(bRes)
		bdd.Cudd_RecursiveDeref(dd, bRes0)
	} else {
		bdd.Cudd_RecursiveDeref(dd, bRes0)
		// try to find the cube with the positive literal
		bRes1 = extraBddGetOneCube(dd, bFunc1)
		bdd.Cudd_Ref(bRes1)
		if bRes1 == dd.LogicZero {
			panic("dsd: Extra_bddGetOneCube: bRes1 != b0")
		}
		bRes = bdd.Cudd_bddAnd(dd, bRes1, dd.Vars[bFuncR.Index])
		bdd.Cudd_Ref(bRes)
		bdd.Cudd_RecursiveDeref(dd, bRes1)
	}

	bdd.Cudd_Deref(bRes)
	return bRes
}

// extraBddFindOneCube is Extra_bddFindOneCube: the cube of
// Cudd_bddPickOneCube (random choices where both branches are nonzero).
func extraBddFindOneCube(dd *DdManager, bF *DdNode) *DdNode {
	size := len(dd.Vars)
	s_Temp := make([]byte, size)
	for i := range s_Temp {
		s_Temp[i] = 2
	}
	node := bF
	for node != dd.One {
		if node == dd.LogicZero {
			return nil
		}
		N := bdd.Cudd_Regular(node)
		T, E := N.T, N.E
		if bdd.Cudd_IsComplement(node) {
			T, E = bdd.Cudd_Not(T), bdd.Cudd_Not(E)
		}
		if T == dd.LogicZero {
			s_Temp[N.Index] = 0
			node = E
		} else if E == dd.LogicZero {
			s_Temp[N.Index] = 1
			node = T
		} else {
			dir := byte((bdd.Cudd_Random() & 0x2000) >> 13)
			s_Temp[N.Index] = dir
			if dir != 0 {
				node = T
			} else {
				node = E
			}
		}
	}

	// start the cube
	bCube := dd.One
	bdd.Cudd_Ref(bCube)
	for v := 0; v < size; v++ {
		if s_Temp[v] == 0 {
			bTemp := bCube
			bCube = bdd.Cudd_bddAnd(dd, bTemp, bdd.Cudd_Not(dd.Vars[v]))
			bdd.Cudd_Ref(bCube)
			bdd.Cudd_RecursiveDeref(dd, bTemp)
		} else if s_Temp[v] == 1 {
			bTemp := bCube
			bCube = bdd.Cudd_bddAnd(dd, bTemp, dd.Vars[v])
			bdd.Cudd_Ref(bCube)
			bdd.Cudd_RecursiveDeref(dd, bTemp)
		}
	}
	bdd.Cudd_Deref(bCube)
	return bCube
}

// cuddSharingSize is Cudd_SharingSize: the number of distinct nodes
// (the constant included) in an array of DDs.
func cuddSharingSize(nodeArray []*DdNode) int {
	seen := make(map[*DdNode]struct{})
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		n = bdd.Cudd_Regular(n)
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		if !bdd.Cudd_IsConstant(n) {
			visit(n.T)
			visit(n.E)
		}
	}
	for _, f := range nodeArray {
		visit(f)
	}
	return len(seen)
}

// cuddSupportSize is Cudd_SupportSize.
func cuddSupportSize(dd *DdManager, f *DdNode) int {
	seenVar := make([]bool, len(dd.Vars))
	seen := make(map[*DdNode]struct{})
	count := 0
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		n = bdd.Cudd_Regular(n)
		if bdd.Cudd_IsConstant(n) {
			return
		}
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		if !seenVar[n.Index] {
			seenVar[n.Index] = true
			count++
		}
		visit(n.T)
		visit(n.E)
	}
	visit(f)
	return count
}
