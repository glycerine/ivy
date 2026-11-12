package core

// Cudd_bddConstrain computes the generalized cofactor f @ c.  It agrees with
// f wherever c is true, and uses the reduced ordered representation to choose
// a canonical result outside that care set.
func Cudd_bddConstrain(dd *DdManager, f, c *DdNode) *DdNode {
	if !dd.valid(f, c) {
		return nil
	}
	return bddConstrain(dd, f, c, make(map[[2]*DdNode]*DdNode))
}

func bddConstrain(dd *DdManager, f, c *DdNode, memo map[[2]*DdNode]*DdNode) *DdNode {
	zero, one := dd.LogicZero, dd.One
	switch {
	case c == one:
		return f
	case c == zero || f == zero:
		return zero
	case Cudd_IsConstant(f):
		return f
	case f == c:
		return one
	case f == Cudd_Not(c):
		return zero
	}
	key := [2]*DdNode{f, c}
	if result, ok := memo[key]; ok {
		return result
	}
	top := min(dd.level(f), dd.level(c))
	index := dd.InvPerm[top]
	ft, fe := dd.cofactors(f, top)
	ct, ce := dd.cofactors(c, top)
	var result *DdNode
	switch {
	case ct == zero:
		result = bddConstrain(dd, fe, ce, memo)
	case ce == zero:
		result = bddConstrain(dd, ft, ct, memo)
	default:
		t := bddConstrain(dd, ft, ct, memo)
		e := bddConstrain(dd, fe, ce, memo)
		if t != nil && e != nil {
			result = dd.node(uint32(index), t, e)
		}
	}
	memo[key] = result
	return result
}

// Cudd_bddRestrict computes Coudert and Madre's generalized cofactor. Its
// result agrees with f on c's on-set and has support contained in f's support.
func Cudd_bddRestrict(dd *DdManager, f, c *DdNode) *DdNode {
	if !dd.valid(f, c) {
		return nil
	}
	result := bddRestrict(dd, f, c, make(map[[2]*DdNode]*DdNode))
	if result == nil {
		return nil
	}
	// The original API avoids returning a larger graph than its input.
	if dagSize(f) <= dagSize(result) {
		return f
	}
	return result
}

func bddRestrict(dd *DdManager, f, c *DdNode, memo map[[2]*DdNode]*DdNode) *DdNode {
	zero, one := dd.LogicZero, dd.One
	switch {
	case c == one:
		return f
	case c == zero || f == zero:
		return zero
	case Cudd_IsConstant(f):
		return f
	case f == c:
		return one
	case f == Cudd_Not(c):
		return zero
	}
	key := [2]*DdNode{f, c}
	if result, ok := memo[key]; ok {
		return result
	}
	topF, topC := dd.level(f), dd.level(c)
	if topC < topF {
		// Eliminate a care-set variable absent from f by existentially
		// projecting it from c: forall? no, OR of its two cofactors.
		ct, ce := dd.cofactors(c, topC)
		projected := Cudd_bddOr(dd, ct, ce)
		result := bddRestrict(dd, f, projected, memo)
		memo[key] = result
		return result
	}
	ft, fe := dd.cofactors(f, topF)
	ct, ce := dd.cofactors(c, topF)
	var result *DdNode
	switch {
	case ct == zero:
		result = bddRestrict(dd, fe, ce, memo)
	case ce == zero:
		result = bddRestrict(dd, ft, ct, memo)
	default:
		t := bddRestrict(dd, ft, ct, memo)
		e := bddRestrict(dd, fe, ce, memo)
		if t != nil && e != nil {
			result = dd.node(uint32(dd.InvPerm[topF]), t, e)
		}
	}
	memo[key] = result
	return result
}

func dagSize(root *DdNode) int {
	seen := make(map[*DdNode]struct{})
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		n = Cudd_Regular(n)
		if n == nil || Cudd_IsConstant(n) {
			return
		}
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		visit(n.T)
		visit(n.E)
	}
	visit(root)
	return len(seen)
}
