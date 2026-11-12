/**CFile****************************************************************

  FileName    [extraBddAuto.c]

  PackageName [extra]

  Synopsis    [Computation of autosymmetries.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - September 1, 2003.]

  Revision    [$Id: extraBddAuto.c,v 1.0 2003/05/21 18:03:50 alanmi Exp $]

***********************************************************************/

package extrab

import cudd "github.com/glycerine/abcgo/bdd/cudd/core"

type autoPair struct{ f, g *DdNode }
type autoSigned struct {
	f        *DdNode
	negative bool
}
type autoEngine struct {
	dd                         *DdManager
	spaces                     map[autoPair]*DdNode
	signs, equations, matrices map[autoSigned]*DdNode
	canonical                  map[*DdNode]*DdNode
}

func newAuto(dd *DdManager) *autoEngine {
	return &autoEngine{dd, make(map[autoPair]*DdNode), make(map[autoSigned]*DdNode), make(map[autoSigned]*DdNode), make(map[autoSigned]*DdNode), make(map[*DdNode]*DdNode)}
}
func autoBDD(dd *DdManager, roots ...*DdNode) bool {
	if dd == nil {
		return false
	}
	for _, f := range roots {
		if !cudd.Cudd_bddLeq(dd, f, f) {
			return false
		}
	}
	return true
}
func autoChildren(f *DdNode) (lo, hi *DdNode) {
	r := cudd.Cudd_Regular(f)
	lo, hi = r.E, r.T
	if cudd.Cudd_IsComplement(f) {
		lo, hi = cudd.Cudd_Not(lo), cudd.Cudd_Not(hi)
	}
	return
}
func (e *autoEngine) space(f, g *DdNode) *DdNode {
	dd := e.dd
	if cudd.Cudd_IsConstant(f) {
		if f == g {
			return dd.One
		}
		return dd.LogicZero
	}
	if cudd.Cudd_IsConstant(g) {
		return dd.LogicZero
	}
	if f.ID > g.ID {
		f, g = g, f
	}
	key := autoPair{f, g}
	if r := e.spaces[key]; r != nil {
		return r
	}
	rf, rg := cudd.Cudd_Regular(f), cudd.Cudd_Regular(g)
	lf, lg := dd.Perm[rf.Index], dd.Perm[rg.Index]
	index := dd.InvPerm[min(lf, lg)]
	f0, f1 := f, f
	if lf <= lg {
		f0, f1 = autoChildren(f)
	}
	g0, g1 := g, g
	if lg <= lf {
		g0, g1 = autoChildren(g)
	}
	lo := cudd.Cudd_bddAnd(dd, e.space(f0, g0), e.space(f1, g1))
	hi := cudd.Cudd_bddAnd(dd, e.space(f0, g1), e.space(f1, g0))
	r := cudd.Cudd_bddIte(dd, dd.Vars[index], hi, lo)
	e.spaces[key] = r
	return r
}
func (e *autoEngine) sign(f *DdNode, negative bool) *DdNode {
	dd := e.dd
	if cudd.Cudd_IsConstant(f) {
		if negative {
			return dd.LogicZero
		}
		return dd.One
	}
	key := autoSigned{f, negative}
	if r := e.signs[key]; r != nil {
		return r
	}
	lo, hi := autoChildren(f)
	a := cudd.Cudd_bddAnd(dd, e.sign(lo, negative), e.sign(hi, negative))
	b := cudd.Cudd_bddAnd(dd, e.sign(lo, !negative), e.sign(hi, !negative))
	r := cudd.Cudd_bddIte(dd, dd.Vars[cudd.Cudd_Regular(f).Index], b, a)
	e.signs[key] = r
	return r
}
func (e *autoEngine) canon(f *DdNode) *DdNode {
	if cudd.Cudd_IsConstant(f) {
		return f
	}
	if r := e.canonical[f]; r != nil {
		return r
	}
	lo, hi := autoChildren(f)
	var r *DdNode
	if lo == e.dd.LogicZero {
		r = e.canon(hi)
	} else if hi == e.dd.LogicZero {
		r = e.canon(lo)
	} else {
		r = cudd.Cudd_bddAnd(e.dd, e.dd.Vars[cudd.Cudd_Regular(f).Index], e.canon(lo))
	}
	e.canonical[f] = r
	return r
}
func (e *autoEngine) zCompose(index int, with, without *DdNode) *DdNode {
	return cudd.Cudd_zddUnion(e.dd, without, cudd.Cudd_zddChange(e.dd, with, index))
}
func (e *autoEngine) equation(f *DdNode, negative bool) *DdNode {
	dd := e.dd
	if f == dd.LogicZero {
		return dd.One
	}
	if f == dd.One {
		return dd.Zero
	}
	key := autoSigned{f, negative}
	if r := e.equations[key]; r != nil {
		return r
	}
	lo, hi := autoChildren(f)
	index := 2 * int(cudd.Cudd_Regular(f).Index)
	var r *DdNode
	if lo == dd.LogicZero {
		r = e.equation(hi, negative)
		if !negative {
			r = e.zCompose(index, dd.One, r)
		}
	} else if hi == dd.LogicZero {
		r = e.equation(lo, negative)
		if negative {
			r = e.zCompose(index, dd.One, r)
		}
	} else {
		a := e.equation(lo, negative)
		b := e.equation(hi, negative)
		c := e.equation(hi, !negative)
		r = e.zCompose(index, cudd.Cudd_zddIntersect(dd, a, c), cudd.Cudd_zddIntersect(dd, a, b))
	}
	e.equations[key] = r
	return r
}
func (e *autoEngine) matrix(f *DdNode, negative bool) *DdNode {
	dd := e.dd
	if f == dd.Zero {
		return dd.One
	}
	if f == dd.One {
		if negative {
			return dd.LogicZero
		}
		return dd.One
	}
	key := autoSigned{f, negative}
	if r := e.matrices[key]; r != nil {
		return r
	}
	a := cudd.Cudd_bddAnd(dd, e.matrix(f.E, negative), e.matrix(f.T, negative))
	b := cudd.Cudd_bddAnd(dd, e.matrix(f.E, negative), e.matrix(f.T, !negative))
	v := cudd.Cudd_bddIthVar(dd, int(f.Index)/2)
	r := cudd.Cudd_bddIte(dd, v, b, a)
	e.matrices[key] = r
	return r
}
func autoMatrixValid(dd *DdManager, f *DdNode) bool {
	if dd == nil || f == nil || cudd.Cudd_zddUnion(dd, f, dd.Zero) == nil {
		return false
	}
	seen := map[*DdNode]bool{}
	var visit func(*DdNode) bool
	visit = func(f *DdNode) bool {
		if cudd.Cudd_IsConstant(f) || seen[f] {
			return true
		}
		seen[f] = true
		return f.Index&1 == 0 && int(f.Index)/2 < len(dd.Vars) && visit(f.T) && visit(f.E)
	}
	return visit(f)
}
func Extra_bddSpaceFromFunction(dd *DdManager, f, g *DdNode) *DdNode {
	if !autoBDD(dd, f, g) {
		return nil
	}
	return newAuto(dd).space(f, g)
}
func Extra_bddSpaceFromFunctionPos(dd *DdManager, f *DdNode) *DdNode {
	if !autoBDD(dd, f) {
		return nil
	}
	return newAuto(dd).sign(f, false)
}
func Extra_bddSpaceFromFunctionNeg(dd *DdManager, f *DdNode) *DdNode {
	if !autoBDD(dd, f) {
		return nil
	}
	return newAuto(dd).sign(f, true)
}
func Extra_bddSpaceCanonVars(dd *DdManager, f *DdNode) *DdNode {
	if !autoBDD(dd, f) {
		return nil
	}
	return newAuto(dd).canon(f)
}
func Extra_bddSpaceReduce(dd *DdManager, f, vars *DdNode) *DdNode {
	if !autoBDD(dd, f, vars) {
		return nil
	}
	cube := dd.One
	for _, index := range dd.Support(vars) {
		cube = cudd.Cudd_bddAnd(dd, cube, cudd.Cudd_Not(dd.Vars[index]))
	}
	return cudd.Cudd_Cofactor(dd, f, cube)
}
func Extra_bddSpaceEquationsPos(dd *DdManager, f *DdNode) *DdNode {
	if !autoBDD(dd, f) {
		return nil
	}
	return newAuto(dd).equation(f, false)
}
func Extra_bddSpaceEquationsNeg(dd *DdManager, f *DdNode) *DdNode {
	if !autoBDD(dd, f) {
		return nil
	}
	return newAuto(dd).equation(f, true)
}
func Extra_bddSpaceEquations(dd *DdManager, f *DdNode) *DdNode {
	if !autoBDD(dd, f) {
		return nil
	}
	e := newAuto(dd)
	return cudd.Cudd_zddUnion(dd, e.equation(f, false), e.equation(f, true))
}
func Extra_bddSpaceFromMatrixPos(dd *DdManager, f *DdNode) *DdNode {
	if !autoMatrixValid(dd, f) {
		return nil
	}
	return newAuto(dd).matrix(f, false)
}
func Extra_bddSpaceFromMatrixNeg(dd *DdManager, f *DdNode) *DdNode {
	if !autoMatrixValid(dd, f) {
		return nil
	}
	return newAuto(dd).matrix(f, true)
}
func Extra_bddSpaceFromFunctionFast(dd *DdManager, f *DdNode) *DdNode {
	if !autoBDD(dd, f) {
		return nil
	}
	support := dd.Support(f)
	if len(dd.Vars) < 2*len(support) {
		return nil
	}
	permutation, back := make([]int, len(dd.Vars)), make([]int, len(dd.Vars))
	vector := append([]*DdNode(nil), dd.Vars...)
	for i := range permutation {
		permutation[i] = i
		back[i] = i
	}
	cube := dd.One
	for k, index := range support {
		even, odd := dd.InvPerm[2*k], dd.InvPerm[2*k+1]
		permutation[index] = even
		back[odd] = index
		vector[even] = cudd.Cudd_bddXor(dd, dd.Vars[even], dd.Vars[odd])
		cube = cudd.Cudd_bddAnd(dd, cube, dd.Vars[even])
	}
	first := cudd.Cudd_bddPermute(dd, f, permutation)
	second := cudd.Cudd_bddVectorCompose(dd, first, vector)
	different := cudd.Cudd_bddXor(dd, first, second)
	shift := cudd.Cudd_Not(cudd.Cudd_bddExistAbstract(dd, different, cube))
	return cudd.Cudd_bddPermute(dd, shift, back)
}
func Extra_zddLitCountComb(dd *DdManager, f *DdNode) int {
	if dd == nil || f == nil || cudd.Cudd_zddUnion(dd, f, dd.Zero) == nil {
		return -1
	}
	if f == dd.Zero {
		return 0
	}
	count := 0
	for f != dd.One {
		if f.E != dd.Zero {
			return -1
		}
		count++
		f = f.T
	}
	return count
}
func autoSelectSubset(dd *DdManager, f *DdNode) *DdNode {
	cube := dd.One
	for f != dd.One {
		if f == dd.Zero {
			return dd.Zero
		}
		if f.E != dd.Zero {
			f = f.E
		} else {
			cube = cudd.Cudd_zddChange(dd, cube, int(f.Index))
			f = f.T
		}
	}
	return cube
}

// Extra_bddSpaceExorGates returns referenced equations indexed by their unique
// reduced-function variable. Nil means the equation/basis precondition failed.
func Extra_bddSpaceExorGates(dd *DdManager, reduced, equations *DdNode) []*DdNode {
	if !autoBDD(dd, reduced) || !autoMatrixValid(dd, equations) {
		return nil
	}
	support := map[int]bool{}
	for _, index := range dd.Support(reduced) {
		support[index] = true
	}
	result := make([]*DdNode, len(dd.Vars))
	fail := func() []*DdNode {
		for _, f := range result {
			if f != nil {
				cudd.Cudd_RecursiveDeref(dd, f)
			}
		}
		return nil
	}
	for equations != dd.Zero {
		cube := autoSelectSubset(dd, equations)
		equations = cudd.Cudd_zddDiff(dd, equations, cube)
		index := -1
		for node := cube; node != dd.One; node = node.T {
			v := int(node.Index) / 2
			if support[v] {
				if index != -1 {
					return fail()
				}
				index = v
			}
		}
		if index < 0 {
			return fail()
		}
		if Extra_zddLitCountComb(dd, cube) > 1 {
			if result[index] != nil {
				return fail()
			}
			result[index] = cube
			cudd.Cudd_Ref(cube)
		}
	}
	return result
}
func extraBddSpaceFromFunction(dd *DdManager, f, g *DdNode) *DdNode {
	return Extra_bddSpaceFromFunction(dd, f, g)
}
func extraBddSpaceFromFunctionPos(dd *DdManager, f *DdNode) *DdNode {
	return Extra_bddSpaceFromFunctionPos(dd, f)
}
func extraBddSpaceFromFunctionNeg(dd *DdManager, f *DdNode) *DdNode {
	return Extra_bddSpaceFromFunctionNeg(dd, f)
}
func extraBddSpaceCanonVars(dd *DdManager, f *DdNode) *DdNode { return Extra_bddSpaceCanonVars(dd, f) }
func extraBddSpaceEquationsPos(dd *DdManager, f *DdNode) *DdNode {
	return Extra_bddSpaceEquationsPos(dd, f)
}
func extraBddSpaceEquationsNeg(dd *DdManager, f *DdNode) *DdNode {
	return Extra_bddSpaceEquationsNeg(dd, f)
}
func extraBddSpaceFromMatrixPos(dd *DdManager, f *DdNode) *DdNode {
	return Extra_bddSpaceFromMatrixPos(dd, f)
}
func extraBddSpaceFromMatrixNeg(dd *DdManager, f *DdNode) *DdNode {
	return Extra_bddSpaceFromMatrixNeg(dd, f)
}
