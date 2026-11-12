/**CFile****************************************************************

  FileName    [extraBddUnate.c]

  PackageName [extra]

  Synopsis    [Efficient methods to compute the information about
  unate variables using an algorithm that is conceptually similar to
  the algorithm for two-variable symmetry computation presented in:
  A. Mishchenko. Fast Computation of Symmetries in Boolean Functions.
  Transactions on CAD, Nov. 2003.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - September 1, 2003.]

  Revision    [$Id: extraBddUnate.c,v 1.0 2003/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

package extrab

import (
	"fmt"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func unateBDD(dd *DdManager, f *DdNode) bool { return dd != nil && cudd.Cudd_bddLeq(dd, f, f) }
func Extra_UnateInfoAllocate(nVars int) *Extra_UnateInfo_t {
	if nVars < 0 {
		return nil
	}
	return &Extra_UnateInfo_t{NVars: nVars, Vars: make([]Extra_UnateVar_t, nVars)}
}
func Extra_UnateInfoDissolve(p *Extra_UnateInfo_t) {
	if p != nil {
		p.Vars = nil
		p.NVars = 0
		p.NUnate = 0
	}
}
func Extra_UnateInfoPrint(p *Extra_UnateInfo_t) {
	if p == nil {
		return
	}
	buffer := make([]byte, p.NVarsMax)
	for i := range buffer {
		buffer[i] = ' '
	}
	for _, v := range p.Vars {
		if v.IVar < 0 || v.IVar >= len(buffer) {
			panic("extrab: invalid unateness variable")
		}
		value := byte('.')
		if v.Neg {
			value = 'n'
		} else if v.Pos {
			value = 'p'
		}
		buffer[v.IVar] = value
	}
	fmt.Println(string(buffer))
}
func Extra_bddCheckUnateNaive(dd *DdManager, f *DdNode, index int) int {
	if !unateBDD(dd, f) || index < 0 || index >= len(dd.Vars) {
		return 0
	}
	v := dd.Vars[index]
	lo := cudd.Cudd_Cofactor(dd, f, cudd.Cudd_Not(v))
	hi := cudd.Cudd_Cofactor(dd, f, v)
	if cudd.Cudd_bddLeq(dd, lo, hi) {
		return 1
	}
	if cudd.Cudd_bddLeq(dd, hi, lo) {
		return -1
	}
	return 0
}
func Extra_UnateComputeSlow(dd *DdManager, f *DdNode) *Extra_UnateInfo_t {
	if !unateBDD(dd, f) {
		return nil
	}
	support := dd.Support(f)
	info := Extra_UnateInfoAllocate(len(support))
	info.NVarsMax = len(dd.Vars)
	for i, index := range support {
		polarity := Extra_bddCheckUnateNaive(dd, f, index)
		info.Vars[i] = Extra_UnateVar_t{index, polarity == 1, polarity == -1}
		if polarity != 0 {
			info.NUnate++
		}
	}
	return info
}
func unateSingleton(dd *DdManager, index int) *DdNode { return cudd.Cudd_zddChange(dd, dd.One, index) }

// Extra_zddGetSingletonsBoth retains the source's empty-set base case: the
// resulting family contains the empty set as well as both literal singletons.
func Extra_zddGetSingletonsBoth(dd *DdManager, vars *DdNode) *DdNode {
	indices, ok := extraPositiveCube(dd, vars)
	if !ok {
		return nil
	}
	result := dd.One
	for _, index := range indices {
		result = cudd.Cudd_zddUnion(dd, result, unateSingleton(dd, 2*index+1))
		result = cudd.Cudd_zddUnion(dd, result, unateSingleton(dd, 2*index))
	}
	return result
}

type unateKey struct{ function, vars *DdNode }

func Extra_zddUnateInfoCompute(dd *DdManager, f, vars *DdNode) *DdNode {
	indices, ok := extraPositiveCube(dd, vars)
	if !ok || !unateBDD(dd, f) {
		return nil
	}
	present := map[int]bool{}
	for _, i := range indices {
		present[i] = true
	}
	for _, i := range dd.Support(f) {
		if !present[i] {
			return nil
		}
	}
	memo := make(map[unateKey]*DdNode)
	var visit func(*DdNode, *DdNode) *DdNode
	visit = func(f, vars *DdNode) *DdNode {
		if cudd.Cudd_IsConstant(f) {
			if vars == dd.One {
				return dd.Zero
			}
			return Extra_zddGetSingletonsBoth(dd, vars)
		}
		key := unateKey{f, vars}
		if r := memo[key]; r != nil {
			return r
		}
		regular := cudd.Cudd_Regular(f)
		level := dd.Perm[regular.Index]
		aligned := vars
		for aligned != dd.One && dd.Perm[aligned.Index] < level {
			aligned = aligned.T
		}
		if aligned == dd.One || aligned.Index != regular.Index {
			return nil
		}
		lo, hi := regular.E, regular.T
		if cudd.Cudd_IsComplement(f) {
			lo, hi = cudd.Cudd_Not(lo), cudd.Cudd_Not(hi)
		}
		result := visit(lo, aligned.T)
		if result != dd.Zero {
			result = cudd.Cudd_zddIntersect(dd, result, visit(hi, aligned.T))
		}
		polarity := -1
		if cudd.Cudd_bddLeq(dd, lo, hi) {
			polarity = 0
		} else if cudd.Cudd_bddLeq(dd, hi, lo) {
			polarity = 1
		}
		if polarity >= 0 {
			result = cudd.Cudd_zddUnion(dd, result, unateSingleton(dd, 2*int(regular.Index)+polarity))
		}
		for extra := vars; extra != aligned; extra = extra.T {
			result = cudd.Cudd_zddUnion(dd, result, unateSingleton(dd, 2*int(extra.Index)+1))
			result = cudd.Cudd_zddUnion(dd, result, unateSingleton(dd, 2*int(extra.Index)))
		}
		memo[key] = result
		return result
	}
	return visit(f, vars)
}
func Extra_UnateInfoCreateFromZdd(dd *DdManager, f, vars *DdNode) *Extra_UnateInfo_t {
	support, ok := extraPositiveCube(dd, vars)
	if !ok || f == nil || cudd.Cudd_zddUnion(dd, f, dd.Zero) == nil {
		return nil
	}
	info := Extra_UnateInfoAllocate(len(support))
	info.NVarsMax = len(dd.Vars)
	positions := map[int]int{}
	for i, index := range support {
		positions[index] = i
		info.Vars[i].IVar = index
	}
	// Traverse the singleton family; malformed cubes and empty-set membership
	// violate this API's source precondition and fail instead of indexing garbage.
	for f != dd.Zero {
		if cudd.Cudd_IsConstant(f) || cudd.Cudd_IsComplement(f) || f.T != dd.One {
			return nil
		}
		index := int(f.Index) / 2
		position, present := positions[index]
		if !present {
			return nil
		}
		if f.Index&1 != 0 {
			info.Vars[position].Neg = true
		} else {
			info.Vars[position].Pos = true
		}
		info.NUnate++
		f = f.E
	}
	return info
}
func Extra_UnateComputeFast(dd *DdManager, f *DdNode) *Extra_UnateInfo_t {
	if !unateBDD(dd, f) {
		return nil
	}
	vars := cudd.Cudd_Support(dd, f)
	result := Extra_zddUnateInfoCompute(dd, f, vars)
	return Extra_UnateInfoCreateFromZdd(dd, result, vars)
}

// Internal entrypoints retain the original API contracts without duplicating recursion.
func extraZddUnateInfoCompute(dd *DdManager, f, vars *DdNode) *DdNode {
	return Extra_zddUnateInfoCompute(dd, f, vars)
}
func extraZddGetSingletonsBoth(dd *DdManager, vars *DdNode) *DdNode {
	return Extra_zddGetSingletonsBoth(dd, vars)
}
