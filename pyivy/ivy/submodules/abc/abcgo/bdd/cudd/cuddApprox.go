package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

// The approximation entry points return canonical BDDs that satisfy the
// subset/superset contracts of CUDD. The native C implementation uses a
// density-driven node replacement heuristic. This port instead builds a
// high-density satisfying cube when the requested node threshold calls for
// reduction. That keeps results valid and deterministic while the detailed
// density heuristic is ported.

// Cudd_UnderApprox returns a subset of f. If f already meets threshold it is
// returned unchanged; otherwise a satisfying cube with the greatest number
// of represented assignments is selected.
func Cudd_UnderApprox(dd *DdManager, f *DdNode, numVars, threshold, safe int32, quality float64) *DdNode {
	return underApprox(dd, f, threshold)
}

// Cudd_OverApprox returns the complement of an under-approximation of !f.
func Cudd_OverApprox(dd *DdManager, f *DdNode, numVars, threshold, safe int32, quality float64) *DdNode {
	if dd == nil || f == nil {
		return nil
	}
	return Cudd_Not(underApprox(dd, Cudd_Not(f), threshold))
}

func Cudd_RemapUnderApprox(dd *DdManager, f *DdNode, numVars, threshold int32, quality float64) *DdNode {
	return underApprox(dd, f, threshold)
}

func Cudd_RemapOverApprox(dd *DdManager, f *DdNode, numVars, threshold int32, quality float64) *DdNode {
	if dd == nil || f == nil {
		return nil
	}
	return Cudd_Not(underApprox(dd, Cudd_Not(f), threshold))
}

// Cudd_BiasedUnderApprox keeps f exact where b is true and selects a dense
// satisfying cube from the remaining part of f.
func Cudd_BiasedUnderApprox(dd *DdManager, f, b *DdNode, numVars, threshold int32, quality1, quality0 float64) *DdNode {
	if dd == nil || f == nil || b == nil {
		return nil
	}
	preserved := core.Cudd_bddAnd(dd, f, b)
	biased := core.Cudd_bddAnd(dd, f, Cudd_Not(b))
	return core.Cudd_bddOr(dd, preserved, underApprox(dd, biased, threshold))
}

func Cudd_BiasedOverApprox(dd *DdManager, f, b *DdNode, numVars, threshold int32, quality1, quality0 float64) *DdNode {
	if dd == nil || f == nil || b == nil {
		return nil
	}
	return Cudd_Not(Cudd_BiasedUnderApprox(dd, Cudd_Not(f), b, numVars, threshold, quality1, quality0))
}

func underApprox(dd *DdManager, f *DdNode, threshold int32) *DdNode {
	if dd == nil || f == nil {
		return nil
	}
	if Cudd_IsConstant(f) || threshold <= 0 || int32(Cudd_DagSize(f)) <= threshold {
		return f
	}
	return largestCube(dd, f, map[*DdNode]*DdNode{})
}

// largestCube follows the branch with the larger satisfying subgraph. Its
// result is a conjunction of literals, hence a subset of the input function.
func largestCube(dd *DdManager, f *DdNode, memo map[*DdNode]*DdNode) *DdNode {
	if f == nil || f == Cudd_ReadLogicZero(dd) {
		return Cudd_ReadLogicZero(dd)
	}
	if f == Cudd_ReadOne(dd) {
		return f
	}
	if result, ok := memo[f]; ok {
		return result
	}
	regular := Cudd_Regular(f)
	thenBranch, elseBranch := Cudd_T(regular), Cudd_E(regular)
	if Cudd_IsComplement(f) {
		thenBranch, elseBranch = Cudd_Not(thenBranch), Cudd_Not(elseBranch)
	}
	thenCount := core.Cudd_CountMinterm(dd, thenBranch, len(dd.Vars))
	elseCount := core.Cudd_CountMinterm(dd, elseBranch, len(dd.Vars))
	index := int(Cudd_NodeReadIndex(f))
	variable := core.Cudd_bddIthVar(dd, index)
	var result *DdNode
	if thenCount >= elseCount {
		child := largestCube(dd, thenBranch, memo)
		result = core.Cudd_bddIte(dd, variable, child, Cudd_ReadLogicZero(dd))
	} else {
		child := largestCube(dd, elseBranch, memo)
		result = core.Cudd_bddIte(dd, variable, Cudd_ReadLogicZero(dd), child)
	}
	memo[f] = result
	return result
}
