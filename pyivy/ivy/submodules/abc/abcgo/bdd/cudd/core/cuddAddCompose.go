package core

// addProjection tests a positive or negative projection without interning it.
func addProjection(n *DdNode, index int, t, e *DdNode) bool {
	return n.Index == uint32(index) && n.T == t && n.E == e
}

func Cudd_addCompose(dd *DdManager, f, replacement *DdNode, variable int) *DdNode {
	if !dd.validAdd(f, replacement) || variable < 0 || variable >= len(dd.Vars) {
		return nil
	}
	level := dd.Perm[variable]
	memo := make(map[*DdNode]*DdNode)
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if dd.level(n) > level {
			return n
		}
		if int(n.Index) == variable {
			return Cudd_addIte(dd, replacement, n.T, n.E)
		}
		if found := memo[n]; found != nil {
			return found
		}
		t, e := visit(n.T), visit(n.E)
		var result *DdNode
		if dd.Perm[n.Index] < min(dd.level(t), dd.level(e)) {
			result = dd.addNode(n.Index, t, e)
		} else {
			result = Cudd_addIte(dd, Cudd_addIthVar(dd, int(n.Index)), t, e)
		}
		memo[n] = result
		return result
	}
	return visit(f)
}

func Cudd_addNonSimCompose(dd *DdManager, f *DdNode, vector []*DdNode) *DdNode {
	if !dd.validAdd(f) || len(vector) < len(dd.Vars) || !dd.validAdd(vector[:len(dd.Vars)]...) {
		return nil
	}
	for level := len(dd.Vars) - 1; level >= 0; level-- {
		index := dd.InvPerm[level]
		if !addProjection(vector[index], index, dd.One, dd.Zero) {
			f = Cudd_addCompose(dd, f, vector[index], index)
		}
	}
	return f
}

func Cudd_addVectorCompose(dd *DdManager, f *DdNode, vector []*DdNode) *DdNode {
	return dd.composeADDVector(f, vector, nil)
}

func Cudd_addGeneralVectorCompose(dd *DdManager, f *DdNode, on, off []*DdNode) *DdNode {
	if dd == nil || len(off) < len(dd.Vars) {
		return nil
	}
	return dd.composeADDVector(f, on, off)
}

func (dd *DdManager) composeADDVector(f *DdNode, on, off []*DdNode) *DdNode {
	if !dd.validAdd(f) || len(on) < len(dd.Vars) || !dd.validAdd(on[:len(dd.Vars)]...) || off != nil && !dd.validAdd(off[:len(dd.Vars)]...) {
		return nil
	}
	deepest := -1
	for level := len(dd.Vars) - 1; level >= 0; level-- {
		i := dd.InvPerm[level]
		if !addProjection(on[i], i, dd.One, dd.Zero) || off != nil && !addProjection(off[i], i, dd.Zero, dd.One) {
			deepest = level
			break
		}
	}
	memo := make(map[*DdNode]*DdNode)
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if dd.level(n) > deepest {
			return n
		}
		if found := memo[n]; found != nil {
			return found
		}
		t, e := visit(n.T), visit(n.E)
		var result *DdNode
		if off == nil {
			result = Cudd_addIte(dd, on[n.Index], t, e)
		} else {
			yes := dd.addArithmetic(cacheADDTimes, on[n.Index], t)
			no := dd.addArithmetic(cacheADDTimes, off[n.Index], e)
			result = dd.addArithmetic(cacheADDPlus, yes, no)
		}
		if n.Ref != 1 {
			memo[n] = result
		}
		return result
	}
	return visit(f)
}
