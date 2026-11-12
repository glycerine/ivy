package dsd

import bdd "github.com/glycerine/abcgo/bdd/cudd/core"

// oneCube chooses the low nonzero branch, as Extra_bddGetOneCube does.
func oneCube(dd *DdManager, f *DdNode) *DdNode {
	if f == dd.LogicZero {
		panic("dsd: no cube in the false function")
	}
	if f == dd.One {
		return dd.One
	}
	r := bdd.Cudd_Regular(f)
	lo, hi := split(dd, f, level(dd, f))
	v := dd.Vars[r.Index]
	if lo != dd.LogicZero {
		return bdd.Cudd_bddAnd(dd, bdd.Cudd_Not(v), oneCube(dd, lo))
	}
	return bdd.Cudd_bddAnd(dd, v, oneCube(dd, hi))
}

// Extra_bddNodePointedByCube restricts f along a satisfying cube. Unlike the C
// pointer walk, it also handles a constant f and cubes omitting variables of f.
func Extra_bddNodePointedByCube(dd *DdManager, f, cube *DdNode) *DdNode {
	result := bdd.Cudd_Cofactor(dd, f, cube)
	if result == nil {
		panic("dsd: invalid restriction cube")
	}
	return result
}

// Dsd_TreeGetPrimeFunction returns the local function on the first nDecs BDD
// levels. Grouping each component's support into adjacent levels lets the
// recursive remapping skip a whole component with one representative cube.
func Dsd_TreeGetPrimeFunction(dd *DdManager, p *Dsd_Node_t) *DdNode {
	p = Dsd_Regular(p)
	n := len(p.pDecs)
	if p.Type != DSD_NODE_PRIME {
		panic("dsd: expected prime node")
	}
	permutation := make([]int, len(dd.Vars))
	for i := range permutation {
		permutation[i] = -1
	}
	variableToFormal := make([]int, len(dd.Vars))
	for i := range variableToFormal {
		variableToFormal[i] = -1
	}
	cubes0, cubes1 := make([]*DdNode, n), make([]*DdNode, n)
	next := 0
	for i, child := range p.pDecs {
		r := Dsd_Regular(child)
		for s := r.S; s != dd.One; s = s.T {
			old := int(s.Index)
			if permutation[old] >= 0 {
				panic("dsd: overlapping prime component supports")
			}
			index := dd.InvPerm[next]
			permutation[old] = index
			variableToFormal[index] = i
			next++
		}
		cubes0[i] = oneCube(dd, bdd.Cudd_Not(function(child)))
		cubes1[i] = oneCube(dd, function(child))
	}
	for _, old := range dd.InvPerm {
		if permutation[old] < 0 {
			permutation[old] = dd.InvPerm[next]
			next++
		}
	}
	f := bdd.Cudd_bddPermute(dd, p.G, permutation)
	for i := range cubes0 {
		cubes0[i] = bdd.Cudd_bddPermute(dd, cubes0[i], permutation)
		cubes1[i] = bdd.Cudd_bddPermute(dd, cubes1[i], permutation)
	}
	memo := map[*DdNode]*DdNode{}
	var remap func(*DdNode) *DdNode
	remap = func(f *DdNode) *DdNode {
		if bdd.Cudd_IsConstant(f) {
			return f
		}
		if result := memo[f]; result != nil {
			return result
		}
		i := variableToFormal[bdd.Cudd_Regular(f).Index]
		if i < 0 {
			panic("dsd: prime function exceeds component support")
		}
		lo := remap(Extra_bddNodePointedByCube(dd, f, cubes0[i]))
		hi := remap(Extra_bddNodePointedByCube(dd, f, cubes1[i]))
		result := bdd.Cudd_bddIte(dd, dd.Vars[dd.InvPerm[i]], hi, lo)
		memo[f] = result
		return result
	}
	return remap(f)
}

// The legacy routine optionally maps formal inputs to variable indices rather
// than levels. Without remapping it uses each component's first support index.
func Dsd_TreeGetPrimeFunctionOld(dd *DdManager, p *Dsd_Node_t, fRemap int) *DdNode {
	p = Dsd_Regular(p)
	local := Dsd_TreeGetPrimeFunction(dd, p)
	permutation := make([]int, len(dd.Vars))
	for i := range permutation {
		permutation[i] = i
	}
	for i, c := range p.pDecs {
		index := int(Dsd_Regular(c).S.Index)
		if fRemap != 0 {
			index = i
		}
		permutation[dd.InvPerm[i]] = index
	}
	return bdd.Cudd_bddPermute(dd, local, permutation)
}
