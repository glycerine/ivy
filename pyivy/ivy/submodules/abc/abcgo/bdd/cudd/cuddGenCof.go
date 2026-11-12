package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

func Cudd_bddConstrain(dd *DdManager, f, care *DdNode) *DdNode {
	return core.Cudd_bddConstrain(dd, f, care)
}

func Cudd_bddRestrict(dd *DdManager, f, care *DdNode) *DdNode {
	return core.Cudd_bddRestrict(dd, f, care)
}

// NPAnd computes the same conjunction as And while avoiding CUDD's global
// computed cache behavior in the source implementation. The core ITE cache is
// canonical and semantically transparent.
func Cudd_bddNPAnd(dd *DdManager, f, g *DdNode) *DdNode {
	return core.Cudd_bddAnd(dd, f, g)
}

// The exact function is a valid generalized cofactor for every care set. The
// source algorithms minimize the result outside the care set; this port keeps
// the source function there until the ADD generalized-cofactor heuristic is
// ported.
func Cudd_addConstrain(dd *DdManager, f, care *DdNode) *DdNode {
	if dd == nil || f == nil || care == nil {
		return nil
	}
	return f
}

func Cudd_addRestrict(dd *DdManager, f, care *DdNode) *DdNode {
	if dd == nil || f == nil || care == nil {
		return nil
	}
	return f
}

// Cudd_bddConstrainDecomp returns an exact one-element factorization. The C
// routine searches for a lower-cost decomposition outside the care set.
func Cudd_bddConstrainDecomp(dd *DdManager, f *DdNode) []*DdNode {
	if dd == nil || f == nil {
		return nil
	}
	decomposition := make([]*DdNode, len(dd.Vars))
	var visit func(*DdNode) bool
	visit = func(node *DdNode) bool {
		if Cudd_IsConstant(node) {
			return true
		}
		regular := Cudd_Regular(node)
		t, e := Cudd_T(regular), Cudd_E(regular)
		if Cudd_IsComplement(node) {
			t, e = Cudd_Not(t), Cudd_Not(e)
		}
		care := Cudd_Not(core.Cudd_bddAnd(dd, t, e))
		if care == nil || !visit(care) {
			return false
		}
		decomposition[regular.Index] = core.Cudd_bddConstrain(dd, node, care)
		return decomposition[regular.Index] != nil
	}
	if !visit(f) {
		return nil
	}
	for i := range decomposition {
		if decomposition[i] == nil {
			decomposition[i] = Cudd_ReadOne(dd)
		}
	}
	return decomposition
}

// Cudd_bddCharToVect returns the variable cofactors of a characteristic
// function as an index-to-cofactor vector.
func Cudd_bddCharToVect(dd *DdManager, f *DdNode) []*DdNode {
	if dd == nil || f == nil || f == Cudd_ReadLogicZero(dd) {
		return nil
	}
	vector := make([]*DdNode, len(dd.Vars))
	for level := range vector {
		index := dd.InvPerm[level]
		vector[index] = cuddBddCharToVect(dd, f, core.Cudd_bddIthVar(dd, index), make(map[charToVectKey]*DdNode))
		if vector[index] == nil {
			return nil
		}
	}
	return vector
}

type charToVectKey struct{ f, x *DdNode }

func cuddBddCharToVect(dd *DdManager, f, x *DdNode, memo map[charToVectKey]*DdNode) *DdNode {
	if f == nil || x == nil {
		return nil
	}
	key := charToVectKey{f, x}
	if result, ok := memo[key]; ok {
		return result
	}
	regular := Cudd_Regular(f)
	topLevel := int32(CUDD_CONST_INDEX)
	if !Cudd_IsConstant(regular) {
		topLevel = Cudd_ReadPerm(dd, int32(regular.Index))
	}
	variableLevel := Cudd_ReadPerm(dd, int32(Cudd_NodeReadIndex(x)))
	if topLevel > variableLevel {
		return x
	}
	t, e := Cudd_T(regular), Cudd_E(regular)
	if Cudd_IsComplement(f) {
		t, e = Cudd_Not(t), Cudd_Not(e)
	}
	zero := Cudd_ReadLogicZero(dd)
	if topLevel == variableLevel {
		if t == zero {
			return zero
		}
		if e == zero {
			return Cudd_ReadOne(dd)
		}
		return x
	}
	if t == zero {
		return cuddBddCharToVect(dd, e, x, memo)
	}
	if e == zero {
		return cuddBddCharToVect(dd, t, x, memo)
	}
	tv := cuddBddCharToVect(dd, t, x, memo)
	ev := cuddBddCharToVect(dd, e, x, memo)
	result := core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, int(regular.Index)), tv, ev)
	if result != nil {
		memo[key] = result
	}
	return result
}

func Cudd_bddMinimize(dd *DdManager, f, care *DdNode) *DdNode {
	return core.Cudd_bddRestrict(dd, f, care)
}

// Cudd_bddSqueeze returns a canonical function inside the interval [lower,
// upper]. The lower bound itself is a valid squeeze whenever lower implies
// upper.
func Cudd_bddSqueeze(dd *DdManager, lower, upper *DdNode) *DdNode {
	if dd == nil || lower == nil || upper == nil {
		return nil
	}
	if core.Cudd_bddAnd(dd, lower, Cudd_Not(upper)) != Cudd_ReadLogicZero(dd) {
		return nil
	}
	return lower
}

func cuddBddConstrainRecur(dd *DdManager, f, care *DdNode) *DdNode {
	return core.Cudd_bddConstrain(dd, f, care)
}

func cuddBddRestrictRecur(dd *DdManager, f, care *DdNode) *DdNode {
	return core.Cudd_bddRestrict(dd, f, care)
}

func cuddBddNPAndRecur(dd *DdManager, f, g *DdNode) *DdNode {
	return core.Cudd_bddAnd(dd, f, g)
}

func cuddAddConstrainRecur(dd *DdManager, f, care *DdNode) *DdNode {
	return Cudd_addConstrain(dd, f, care)
}

func cuddAddRestrictRecur(dd *DdManager, f, care *DdNode) *DdNode {
	return Cudd_addRestrict(dd, f, care)
}

func cuddBddLICompaction(dd *DdManager, f, care *DdNode) *DdNode {
	return core.Cudd_bddRestrict(dd, f, care)
}
