package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

func Cudd_Cofactor(dd *DdManager, f, cube *DdNode) *DdNode {
	return core.Cudd_Cofactor(dd, f, cube)
}

func cuddGetBranches(g *DdNode, g1, g0 **DdNode) {
	if g1 == nil || g0 == nil || g == nil {
		return
	}
	regular := Cudd_Regular(g)
	*g1, *g0 = Cudd_T(regular), Cudd_E(regular)
	if Cudd_IsComplement(g) {
		*g1, *g0 = Cudd_Not(*g1), Cudd_Not(*g0)
	}
}

func cuddCheckCube(dd *DdManager, g *DdNode) int32 {
	if dd == nil || g == nil {
		return 0
	}
	if g == Cudd_ReadOne(dd) {
		return 1
	}
	if Cudd_IsConstant(g) {
		return 0
	}
	var t, e *DdNode
	cuddGetBranches(g, &t, &e)
	zero := Cudd_ReadLogicZero(dd)
	if e == zero {
		return cuddCheckCube(dd, t)
	}
	if t == zero {
		return cuddCheckCube(dd, e)
	}
	return 0
}

func cuddCofactorRecur(dd *DdManager, f, g *DdNode) *DdNode {
	return core.Cudd_Cofactor(dd, f, g)
}
