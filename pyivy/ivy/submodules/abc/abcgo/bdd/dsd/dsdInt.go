package dsd

import bdd "github.com/glycerine/abcgo/bdd/cudd/core"

func function(p *Dsd_Node_t) *DdNode { return bdd.Cudd_NotCond(Dsd_Regular(p).G, Dsd_IsComplement(p)) }
func level(dd *DdManager, f *DdNode) int {
	r := bdd.Cudd_Regular(f)
	if bdd.Cudd_IsConstant(r) {
		return int(bdd.CUDD_CONST_INDEX)
	}
	return dd.Perm[r.Index]
}
func split(dd *DdManager, f *DdNode, lev int) (*DdNode, *DdNode) {
	if level(dd, f) != lev {
		return f, f
	}
	r := bdd.Cudd_Regular(f)
	return bdd.Cudd_NotCond(r.E, bdd.Cudd_IsComplement(f)), bdd.Cudd_NotCond(r.T, bdd.Cudd_IsComplement(f))
}
func suppContains(dd *DdManager, a, b *DdNode) bool { // support cubes: does a contain b?
	for b != dd.One {
		if a == dd.One {
			return false
		}
		la, lb := level(dd, a), level(dd, b)
		if la > lb {
			return false
		}
		if la == lb {
			b = b.T
		}
		a = a.T
	}
	return true
}
func suppOverlap(dd *DdManager, a, b *DdNode) bool {
	for a != dd.One && b != dd.One {
		la, lb := level(dd, a), level(dd, b)
		if la == lb {
			return true
		}
		if la < lb {
			a = a.T
		} else {
			b = b.T
		}
	}
	return false
}
func suppSize(dd *DdManager, a *DdNode) int {
	n := 0
	for a != dd.One {
		n++
		a = a.T
	}
	return n
}
func suppHas(dd *DdManager, s, v *DdNode) bool {
	target := level(dd, v)
	for s != dd.One {
		l := level(dd, s)
		if l == target {
			return true
		}
		if l > target {
			return false
		}
		s = s.T
	}
	return false
}
