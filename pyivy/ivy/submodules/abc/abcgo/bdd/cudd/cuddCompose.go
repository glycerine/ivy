package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

func Cudd_bddCompose(dd *DdManager, f, g *DdNode, variable int) *DdNode {
	return core.Cudd_bddCompose(dd, f, g, variable)
}

func Cudd_addCompose(dd *DdManager, f, g *DdNode, variable int) *DdNode {
	if dd == nil || f == nil || g == nil || variable < 0 || variable >= len(dd.Vars) {
		return nil
	}
	return addComposeAt(dd, f, g, uint32(variable), make(map[*DdNode]*DdNode))
}

func Cudd_addPermute(dd *DdManager, f *DdNode, permutation []int) *DdNode {
	return core.Cudd_addPermute(dd, f, permutation)
}

func Cudd_addSwapVariables(dd *DdManager, f *DdNode, x, y []*DdNode) *DdNode {
	permutation, ok := swapPermutation(dd, x, y)
	if !ok {
		return nil
	}
	return core.Cudd_addPermute(dd, f, permutation)
}

func Cudd_bddPermute(dd *DdManager, f *DdNode, permutation []int) *DdNode {
	return core.Cudd_bddPermute(dd, f, permutation)
}

func Cudd_bddSwapVariables(dd *DdManager, f *DdNode, x, y []*DdNode) *DdNode {
	permutation, ok := swapPermutation(dd, x, y)
	if !ok {
		return nil
	}
	return core.Cudd_bddPermute(dd, f, permutation)
}

func Cudd_bddAdjPermuteX(dd *DdManager, f *DdNode, variables []*DdNode) *DdNode {
	if dd == nil || len(variables) > len(dd.Vars) {
		return nil
	}
	permutation := identityPermutation(len(dd.Vars))
	for i := 0; i+1 < len(variables); i += 3 {
		if variables[i] == nil || variables[i+1] == nil {
			return nil
		}
		a, b := int(Cudd_NodeReadIndex(variables[i])), int(Cudd_NodeReadIndex(variables[i+1]))
		if a >= len(permutation) || b >= len(permutation) {
			return nil
		}
		permutation[a], permutation[b] = permutation[b], permutation[a]
	}
	return core.Cudd_bddPermute(dd, f, permutation)
}

func Cudd_addVectorCompose(dd *DdManager, f *DdNode, vector []*DdNode) *DdNode {
	return addVectorCompose(dd, f, vector)
}

func Cudd_addGeneralVectorCompose(dd *DdManager, f *DdNode, on, off []*DdNode) *DdNode {
	if dd == nil || len(on) < len(dd.Vars) || len(off) < len(dd.Vars) {
		return nil
	}
	vector := make([]*DdNode, len(dd.Vars))
	for i := range vector {
		if on[i] == nil || off[i] == nil {
			return nil
		}
		original := core.Cudd_addIthVar(dd, i)
		vector[i] = core.Cudd_addIte(dd, on[i], core.Cudd_addConst(dd, 1), core.Cudd_addIte(dd, off[i], core.Cudd_addConst(dd, 0), original))
	}
	return addVectorCompose(dd, f, vector)
}

func Cudd_addNonSimCompose(dd *DdManager, f *DdNode, vector []*DdNode) *DdNode {
	if dd == nil || len(vector) < len(dd.Vars) {
		return nil
	}
	result := f
	for i, replacement := range vector[:len(dd.Vars)] {
		if replacement == nil {
			return nil
		}
		result = Cudd_addCompose(dd, result, replacement, i)
		if result == nil {
			return nil
		}
	}
	return result
}

func Cudd_bddVectorCompose(dd *DdManager, f *DdNode, vector []*DdNode) *DdNode {
	return core.Cudd_bddVectorCompose(dd, f, vector)
}

func cuddBddComposeRecur(dd *DdManager, f, g, projection *DdNode) *DdNode {
	if projection == nil || Cudd_IsConstant(projection) {
		return nil
	}
	return core.Cudd_bddCompose(dd, f, g, int(Cudd_NodeReadIndex(projection)))
}

func cuddAddComposeRecur(dd *DdManager, f, g, projection *DdNode) *DdNode {
	if projection == nil || Cudd_IsConstant(projection) {
		return nil
	}
	return Cudd_addCompose(dd, f, g, int(Cudd_NodeReadIndex(projection)))
}

func cuddAddNonSimComposeRecur(dd *DdManager, f *DdNode, vector []*DdNode, key, cube *DdNode, lastSub int32) *DdNode {
	if dd == nil || f == nil || key == nil {
		return nil
	}
	index := int(Cudd_NodeReadIndex(key))
	if index < 0 || index >= len(vector) || vector[index] == nil {
		return nil
	}
	return Cudd_addCompose(dd, f, vector[index], index)
}

func addComposeAt(dd *DdManager, f, replacement *DdNode, index uint32, memo map[*DdNode]*DdNode) *DdNode {
	if Cudd_IsConstant(f) {
		return f
	}
	if result := memo[f]; result != nil {
		return result
	}
	r := Cudd_Regular(f)
	t := addComposeAt(dd, Cudd_T(r), replacement, index, memo)
	e := addComposeAt(dd, Cudd_E(r), replacement, index, memo)
	var result *DdNode
	if r.Index == index {
		result = core.Cudd_addIte(dd, replacement, t, e)
	} else {
		result = core.Cudd_addIte(dd, core.Cudd_addIthVar(dd, int(r.Index)), t, e)
	}
	memo[f] = result
	return result
}

func addVectorCompose(dd *DdManager, f *DdNode, vector []*DdNode) *DdNode {
	if dd == nil || f == nil || len(vector) < len(dd.Vars) {
		return nil
	}
	memo := make(map[*DdNode]*DdNode)
	var visit func(*DdNode) *DdNode
	visit = func(node *DdNode) *DdNode {
		if Cudd_IsConstant(node) {
			return node
		}
		if result := memo[node]; result != nil {
			return result
		}
		r := Cudd_Regular(node)
		t, e := visit(Cudd_T(r)), visit(Cudd_E(r))
		if int(r.Index) >= len(vector) || vector[r.Index] == nil {
			return nil
		}
		result := core.Cudd_addIte(dd, vector[r.Index], t, e)
		memo[node] = result
		return result
	}
	return visit(f)
}

func identityPermutation(n int) []int {
	result := make([]int, n)
	for i := range result {
		result[i] = i
	}
	return result
}

func swapPermutation(dd *DdManager, x, y []*DdNode) ([]int, bool) {
	if dd == nil || len(x) != len(y) || len(x) > len(dd.Vars) {
		return nil, false
	}
	permutation := identityPermutation(len(dd.Vars))
	for i := range x {
		if x[i] == nil || y[i] == nil {
			return nil, false
		}
		a, b := int(Cudd_NodeReadIndex(x[i])), int(Cudd_NodeReadIndex(y[i]))
		if a >= len(permutation) || b >= len(permutation) {
			return nil, false
		}
		permutation[a], permutation[b] = permutation[b], permutation[a]
	}
	return permutation, true
}
