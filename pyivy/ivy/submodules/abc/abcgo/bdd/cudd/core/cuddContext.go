package core

import "context"

// BDDAndContext checks cancellation during the conjunction recursion.
// A cancelled operation returns nil; already completed canonical nodes remain
// usable and a subsequent operation may reuse them. Managers are not concurrent.
func BDDAndContext(ctx context.Context, dd *DdManager, f, g *DdNode) *DdNode {
	if ctx == nil || dd == nil || !dd.valid(f, g) {
		return nil
	}
	return dd.andChecked(f, g, func() bool { return ctx.Err() == nil })
}

// BDDAndAbstractContext fuses conjunction and existential abstraction, avoiding
// an intermediate conjunction. Only positive variable cubes are accepted.
func BDDAndAbstractContext(ctx context.Context, dd *DdManager, f, g, cube *DdNode) *DdNode {
	if ctx == nil || dd == nil || !dd.valid(f, g, cube) {
		return nil
	}
	if !dd.positiveCube(cube) {
		return nil
	}
	return dd.andAbstractChecked(f, g, cube, func() bool { return ctx.Err() == nil })
}
func (dd *DdManager) andAbstractChecked(f, g, cube *DdNode, check func() bool) *DdNode {
	if check != nil && !check() {
		return nil
	}
	if f == dd.LogicZero || g == dd.LogicZero || f == Cudd_Not(g) {
		return dd.LogicZero
	}
	if cube == dd.One {
		return dd.andChecked(f, g, check)
	}
	if f == dd.One || f == g {
		return dd.existChecked(g, cube, check)
	}
	if g == dd.One {
		return dd.existChecked(f, cube, check)
	}

	if f.ID > g.ID {
		f, g = g, f
	}
	fr, gr := Cudd_Regular(f), Cudd_Regular(g)
	lf, lg := dd.Perm[fr.Index], dd.Perm[gr.Index]
	top := min(lf, lg)
	for dd.Perm[cube.Index] < top {
		cube = cube.T
		if cube == dd.One {
			return dd.andChecked(f, g, check)
		}
	}
	key := computedKey{f: f, g: g, h: cube, kind: cacheBDDAndAbstract}
	hash := key.hash()
	if !dd.singleBDDRef(fr) || !dd.singleBDDRef(gr) {
		if found := dd.cacheLookup(key, hash, true); found != nil {
			return found
		}
	}

	ft, fe, gt, ge := f, f, g, g
	if lf == top {
		ft, fe = fr.T, fr.E
		if fr != f {
			ft, fe = Cudd_Not(ft), Cudd_Not(fe)
		}
	}
	if lg == top {
		gt, ge = gr.T, gr.E
		if gr != g {
			gt, ge = Cudd_Not(gt), Cudd_Not(ge)
		}
	}

	quantify := dd.Perm[cube.Index] == top
	next := cube
	if quantify {
		next = cube.T
	}
	t := dd.andAbstractChecked(ft, gt, next, check)
	if t == nil {
		return nil
	}
	result := t
	if !quantify {
		e := dd.andAbstractChecked(fe, ge, next, check)
		if e == nil {
			return nil
		}
		result = dd.node(uint32(dd.InvPerm[top]), t, e)
	} else if t != dd.One && t != fe && t != ge {
		var e *DdNode
		switch t {
		case Cudd_Not(fe):
			e = dd.existChecked(ge, next, check)
		case Cudd_Not(ge):
			e = dd.existChecked(fe, next, check)
		default:
			e = dd.andAbstractChecked(fe, ge, next, check)
		}
		if e == nil {
			return nil
		}
		if t != e {
			result = Cudd_Not(dd.andChecked(Cudd_Not(t), Cudd_Not(e), check))
		}
	}
	if result != nil && (!dd.singleBDDRef(fr) || !dd.singleBDDRef(gr)) {
		dd.cacheInsert(key, result, hash)
	}
	return result
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
