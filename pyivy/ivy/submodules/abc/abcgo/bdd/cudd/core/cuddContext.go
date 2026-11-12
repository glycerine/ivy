package core

import "context"

// BDDAndContext checks cancellation during the authoritative ITE recursion.
// A cancelled operation returns nil; already completed canonical nodes remain
// usable and a subsequent operation may reuse them. Managers are not concurrent.
func BDDAndContext(ctx context.Context, dd *DdManager, f, g *DdNode) *DdNode {
	if ctx == nil || dd == nil || !dd.valid(f, g) {
		return nil
	}
	return dd.iteChecked(f, g, dd.LogicZero, func() bool { return ctx.Err() == nil })
}

// BDDAndAbstractContext fuses conjunction and existential abstraction, avoiding
// an intermediate conjunction. Only positive variable cubes are accepted.
func BDDAndAbstractContext(ctx context.Context, dd *DdManager, f, g, cube *DdNode) *DdNode {
	if ctx == nil || dd == nil || !dd.valid(f, g, cube) {
		return nil
	}
	for c := cube; c != dd.One; c = c.T {
		if Cudd_IsConstant(c) || Cudd_IsComplement(c) || c.E != dd.LogicZero {
			return nil
		}
	}
	return dd.andAbstractChecked(f, g, cube, func() bool { return ctx.Err() == nil })
}
func (dd *DdManager) andAbstractChecked(f, g, cube *DdNode, check func() bool) *DdNode {
	memo := map[bddITEKey]*DdNode{}
	var visit func(*DdNode, *DdNode, *DdNode) *DdNode
	visit = func(f, g, cube *DdNode) *DdNode {
		if check != nil && !check() {
			return nil
		}
		if f == dd.LogicZero || g == dd.LogicZero || f == Cudd_Not(g) {
			return dd.LogicZero
		}
		if f == dd.One && g == dd.One {
			return dd.One
		}
		if cube == dd.One {
			return dd.iteChecked(f, g, dd.LogicZero, check)
		}
		if f.ID > g.ID {
			f, g = g, f
		}
		top := min(dd.level(f), dd.level(g))
		for dd.level(cube) < top {
			cube = cube.T
		}
		if cube == dd.One {
			return dd.iteChecked(f, g, dd.LogicZero, check)
		}
		key := bddITEKey{f, g, cube}
		if r := memo[key]; r != nil {
			return r
		}
		ft, fe := dd.cofactors(f, top)
		gt, ge := dd.cofactors(g, top)
		next := cube
		quantify := dd.level(cube) == top
		if quantify {
			next = cube.T
		}
		t := visit(ft, gt, next)
		if t == nil {
			return nil
		}
		if quantify && t == dd.One {
			memo[key] = t
			return t
		}
		e := visit(fe, ge, next)
		if e == nil {
			return nil
		}
		var r *DdNode
		if quantify {
			r = dd.iteChecked(t, dd.One, e, check)
		} else {
			r = dd.node(uint32(dd.InvPerm[top]), t, e)
		}
		if r != nil {
			memo[key] = r
		}
		return r
	}
	return visit(f, g, cube)
}

// BDDTransferContext performs a simultaneous index mapping between managers.
// Unlike the old C routine's misleading ADD documentation this API accepts
// BDDs only; numerical ADD transfer remains available through Transfer.
func BDDTransferContext(ctx context.Context, source, destination *DdManager, f *DdNode, permutation []int) *DdNode {
	if ctx == nil || source == nil || destination == nil || destination.closed || !source.valid(f) {
		return nil
	}
	check := func() bool { return ctx.Err() == nil }
	for _, index := range source.Support(f) {
		if !check() {
			return nil
		}
		target := index
		if permutation != nil {
			if index >= len(permutation) {
				return nil
			}
			target = permutation[index]
		}
		if target < 0 || uint64(target) >= uint64(bddConstantIndex) {
			return nil
		}
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(edge *DdNode) *DdNode {
		if !check() {
			return nil
		}
		n := Cudd_Regular(edge)
		r := memo[n]
		if r == nil {
			if Cudd_IsConstant(n) {
				r = destination.One
			} else {
				target := int(n.Index)
				if permutation != nil {
					target = permutation[target]
				}
				t := visit(n.T)
				if t == nil {
					return nil
				}
				e := visit(n.E)
				if e == nil {
					return nil
				}
				for len(destination.Vars) <= target {
					if !check() {
						return nil
					}
					Cudd_bddNewVar(destination)
				}
				r = destination.iteChecked(destination.Vars[target], t, e, check)
			}
			if r == nil {
				return nil
			}
			memo[n] = r
		}
		return Cudd_NotCond(r, Cudd_IsComplement(edge))
	}
	return visit(f)
}
