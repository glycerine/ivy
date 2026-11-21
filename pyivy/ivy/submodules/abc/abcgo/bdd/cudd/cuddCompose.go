package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

var (
	Cudd_SetVarMap = core.Cudd_SetVarMap
	Cudd_bddVarMap = core.Cudd_bddVarMap
)

func Cudd_bddCompose(dd *DdManager, f, g *DdNode, variable int) *DdNode {
	return core.Cudd_bddCompose(dd, f, g, variable)
}

func Cudd_addCompose(dd *DdManager, f, g *DdNode, variable int) *DdNode {
	return core.Cudd_addCompose(dd, f, g, variable)
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
	for i := 0; i+2 < len(variables); i += 3 {
		if variables[i] == nil || variables[i+1] == nil {
			return nil
		}
		a, b := int(Cudd_NodeReadIndex(variables[i])), int(Cudd_NodeReadIndex(variables[i+1]))
		if a >= len(permutation) || b >= len(permutation) {
			return nil
		}
		permutation[a], permutation[b] = b, a
	}
	return core.Cudd_bddPermute(dd, f, permutation)
}

func Cudd_addVectorCompose(dd *DdManager, f *DdNode, vector []*DdNode) *DdNode {
	return core.Cudd_addVectorCompose(dd, f, vector)
}

func Cudd_addGeneralVectorCompose(dd *DdManager, f *DdNode, on, off []*DdNode) *DdNode {
	return core.Cudd_addGeneralVectorCompose(dd, f, on, off)
}

func Cudd_addNonSimCompose(dd *DdManager, f *DdNode, vector []*DdNode) *DdNode {
	return core.Cudd_addNonSimCompose(dd, f, vector)
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
	if dd == nil || key == nil || cube == nil || lastSub < 0 || int(lastSub) > len(vector) {
		return nil
	}
	active := make([]*DdNode, len(dd.Vars))
	for i := range active {
		active[i] = core.Cudd_addIthVar(dd, i)
	}
	for cube != dd.One {
		if !core.Cudd_IsAdd(cube) || Cudd_IsConstant(cube) || cube.E != dd.Zero || cube.Index >= uint32(lastSub) || int(cube.Index) >= len(active) {
			return nil
		}
		active[cube.Index] = vector[cube.Index]
		cube = cube.T
	}
	return core.Cudd_addNonSimCompose(dd, f, active)
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
		permutation[a], permutation[b] = b, a
	}
	return permutation, true
}
