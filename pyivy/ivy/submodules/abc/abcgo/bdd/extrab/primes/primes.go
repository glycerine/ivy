// Package primes implements Extra_zddPrimes from ABC's extraBddMisc.c
// (Alan Mishchenko). It shares CUDD core's sole BDD/ZDD representation.
package primes

import cudd "github.com/glycerine/abcgo/bdd/cudd/core"

// Extra_zddPrimes returns all prime implicants as a paired-literal ZDD:
// index 2*i means positive variable i; 2*i+1 means negative variable i.
// False maps to the empty family, true to the family containing the empty cube.
// Like CUDD results, the return is unreferenced; callers retain it with Cudd_Ref
// when performing a subsequent heap reorder. Literal variables grow as needed.
func Extra_zddPrimes(dd *cudd.DdManager, f *cudd.DdNode) *cudd.DdNode {
	if dd == nil || !cudd.Cudd_bddLeq(dd, f, f) {
		return nil
	}
	memo := make(map[*cudd.DdNode]*cudd.DdNode)
	var visit func(*cudd.DdNode) *cudd.DdNode
	visit = func(f *cudd.DdNode) *cudd.DdNode {
		if f == dd.LogicZero {
			return dd.Zero
		}
		if f == dd.One {
			return dd.One
		}
		if r := memo[f]; r != nil {
			return r
		}
		regular := cudd.Cudd_Regular(f)
		lo, hi := regular.E, regular.T
		if cudd.Cudd_IsComplement(f) {
			lo, hi = cudd.Cudd_Not(lo), cudd.Cudd_Not(hi)
		}
		intersection := cudd.Cudd_bddAnd(dd, lo, hi)
		p0, p1 := visit(lo), visit(hi)
		var absent, positive, negative *cudd.DdNode
		switch intersection {
		case lo:
			absent = p0
			negative = dd.Zero
			positive = cudd.Cudd_zddDiff(dd, p1, p0)
		case hi:
			absent = p1
			positive = dd.Zero
			negative = cudd.Cudd_zddDiff(dd, p0, p1)
		default:
			absent = visit(intersection)
			negative = cudd.Cudd_zddDiff(dd, p0, absent)
			positive = cudd.Cudd_zddDiff(dd, p1, absent)
		}
		// ZDD change is the canonical equivalent of extraComposeCover. It also
		// supports valid independent BDD/ZDD orders without manual node insertion.
		positive = cudd.Cudd_zddChange(dd, positive, 2*int(regular.Index))
		negative = cudd.Cudd_zddChange(dd, negative, 2*int(regular.Index)+1)
		result := cudd.Cudd_zddUnion(dd, absent, cudd.Cudd_zddUnion(dd, positive, negative))
		memo[f] = result
		return result
	}
	return visit(f)
}
