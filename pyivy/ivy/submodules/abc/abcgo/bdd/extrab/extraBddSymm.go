/**CFile****************************************************************

  FileName    [extraBddSymm.c]

  PackageName [extra]

  Synopsis    [Efficient methods to compute the information about
  symmetric variables using the algorithm presented in the paper:
  A. Mishchenko. Fast Computation of Symmetries in Boolean Functions.
  Transactions on CAD, Nov. 2003.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - September 1, 2003.]

  Revision    [$Id: extraBddSymm.c,v 1.0 2003/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

package extrab

import (
	"fmt"
	"math"
	"strings"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

type symmPairKey struct{ f, vars *DdNode }
type symmVarsKey struct{ f, g, vars *DdNode }
type symmTupleKey struct {
	vars *DdNode
	k    int
}
type symmEngine struct {
	dd         *DdManager
	pairs      map[symmPairKey]*DdNode
	variables  map[symmVarsKey]*DdNode
	singletons map[*DdNode]*DdNode
	tuples     map[symmTupleKey]*DdNode
}

func newSymm(dd *DdManager) *symmEngine {
	return &symmEngine{dd, make(map[symmPairKey]*DdNode), make(map[symmVarsKey]*DdNode), make(map[*DdNode]*DdNode), make(map[symmTupleKey]*DdNode)}
}
func symmDomain(dd *DdManager, vars *DdNode, functions ...*DdNode) bool {
	indices, ok := extraPositiveCube(dd, vars)
	if !ok || !extraValidBDD(dd, functions...) {
		return false
	}
	present := map[int]bool{}
	for _, i := range indices {
		present[i] = true
	}
	for _, f := range functions {
		for _, i := range dd.Support(f) {
			if !present[i] {
				return false
			}
		}
	}
	return true
}
func (e *symmEngine) singleton(vars *DdNode) *DdNode {
	if vars == e.dd.One {
		return e.dd.One
	}
	if f := e.singletons[vars]; f != nil {
		return f
	}
	result := cudd.Cudd_zddUnion(e.dd, e.singleton(vars.T), cudd.Cudd_zddChange(e.dd, e.dd.One, 2*int(vars.Index)))
	e.singletons[vars] = result
	return result
}
func (e *symmEngine) tuple(vars *DdNode, k int) *DdNode {
	if k < 0 {
		return e.dd.Zero
	}
	if vars == e.dd.One {
		if k == 0 {
			return e.dd.One
		}
		return e.dd.Zero
	}
	key := symmTupleKey{vars, k}
	if f := e.tuples[key]; f != nil {
		return f
	}
	without := e.tuple(vars.T, k)
	with := e.tuple(vars.T, k-1)
	result := cudd.Cudd_zddUnion(e.dd, without, cudd.Cudd_zddChange(e.dd, with, 2*int(vars.Index)))
	e.tuples[key] = result
	return result
}
func (e *symmEngine) getVars(f, g, vars *DdNode) *DdNode {
	dd := e.dd
	if cudd.Cudd_IsConstant(f) && cudd.Cudd_IsConstant(g) {
		if f == g {
			return e.singleton(vars)
		}
		return dd.Zero
	}
	key := symmVarsKey{f, g, vars}
	if r := e.variables[key]; r != nil {
		return r
	}
	lf, lg := math.MaxInt, math.MaxInt
	if !cudd.Cudd_IsConstant(f) {
		lf = dd.Perm[cudd.Cudd_Regular(f).Index]
	}
	if !cudd.Cudd_IsConstant(g) {
		lg = dd.Perm[cudd.Cudd_Regular(g).Index]
	}
	level := min(lf, lg)
	aligned := vars
	for aligned != dd.One && dd.Perm[aligned.Index] < level {
		aligned = aligned.T
	}
	f0, f1 := f, f
	if lf == level {
		f0, f1 = extraBDDCofactors(f)
	}
	g0, g1 := g, g
	if lg == level {
		g0, g1 = extraBDDCofactors(g)
	}
	result := e.getVars(f0, g0, aligned.T)
	if result != dd.Zero {
		result = cudd.Cudd_zddIntersect(dd, result, e.getVars(f1, g1, aligned.T))
	}
	if f0 == g1 {
		result = cudd.Cudd_zddUnion(dd, result, cudd.Cudd_zddChange(dd, dd.One, 2*int(aligned.Index)))
	}
	if f == g && vars != aligned {
		extra := cudd.Cudd_bddExistAbstract(dd, vars, aligned)
		result = cudd.Cudd_zddUnion(dd, result, e.singleton(extra))
	}
	e.variables[key] = result
	return result
}
func (e *symmEngine) pair(f, vars *DdNode) *DdNode {
	dd := e.dd
	if cudd.Cudd_IsConstant(f) {
		return e.tuple(vars, 2)
	}
	key := symmPairKey{f, vars}
	if r := e.pairs[key]; r != nil {
		return r
	}
	regular := cudd.Cudd_Regular(f)
	level := dd.Perm[regular.Index]
	aligned := vars
	for aligned != dd.One && dd.Perm[aligned.Index] < level {
		aligned = aligned.T
	}
	lo, hi := extraBDDCofactors(f)
	result := e.pair(lo, aligned.T)
	if result != dd.Zero {
		result = cudd.Cudd_zddIntersect(dd, result, e.pair(hi, aligned.T))
	}
	partners := e.getVars(hi, lo, aligned.T)
	result = cudd.Cudd_zddUnion(dd, result, cudd.Cudd_zddChange(dd, partners, 2*int(regular.Index)))
	if vars != aligned {
		unused := Extra_bddReduceVarSet(dd, vars, f)
		result = cudd.Cudd_zddUnion(dd, result, e.tuple(unused, 2))
	}
	e.pairs[key] = result
	return result
}
func Extra_zddSymmPairsCompute(dd *DdManager, f, vars *DdNode) *DdNode {
	if !symmDomain(dd, vars, f) {
		return nil
	}
	return newSymm(dd).pair(f, vars)
}
func Extra_zddGetSymmetricVars(dd *DdManager, f, g, vars *DdNode) *DdNode {
	if !symmDomain(dd, vars, f, g) {
		return nil
	}
	return newSymm(dd).getVars(f, g, vars)
}

// Extra_zddGetSingletons preserves the source's empty-set base case in addition
// to its positive literal singleton members.
func Extra_zddGetSingletons(dd *DdManager, vars *DdNode) *DdNode {
	if _, valid := extraPositiveCube(dd, vars); !valid {
		return nil
	}
	return newSymm(dd).singleton(vars)
}
func Extra_bddReduceVarSet(dd *DdManager, vars, f *DdNode) *DdNode {
	indices, valid := extraPositiveCube(dd, vars)
	if !valid || !extraValidBDD(dd, f) {
		return nil
	}
	support := extraSupportSet(dd, f)
	result := dd.One
	for _, i := range indices {
		if !support[i] {
			result = cudd.Cudd_bddAnd(dd, result, dd.Vars[i])
		}
	}
	return result
}
func Extra_zddTuplesFromBdd(dd *DdManager, k int, vars *DdNode) *DdNode {
	indices, valid := extraPositiveCube(dd, vars)
	if !valid || k < 0 || k > len(indices) {
		return nil
	}
	return newSymm(dd).tuple(vars, k)
}
func Extra_zddSelectOneSubset(dd *DdManager, f *DdNode) *DdNode {
	if dd == nil || f == nil || cudd.Cudd_zddUnion(dd, f, dd.Zero) == nil {
		return nil
	}
	return extraSelectSubset(dd, f)
}
func Extra_SymmPairsAllocate(n int) *Extra_SymmInfo_t {
	if n < 0 {
		return nil
	}
	p := &Extra_SymmInfo_t{NVars: n, Vars: make([]int, n), Symms: make([][]bool, n)}
	for i := range p.Symms {
		p.Symms[i] = make([]bool, n)
	}
	return p
}
func Extra_SymmPairsDissolve(p *Extra_SymmInfo_t) {
	if p != nil {
		*p = Extra_SymmInfo_t{}
	}
}
func Extra_SymmPairsPrint(p *Extra_SymmInfo_t) {
	if p == nil {
		return
	}
	fmt.Println()
	for i := 0; i < p.NVars; i++ {
		fmt.Print(strings.Repeat(" ", i+1))
		for j := i + 1; j < p.NVars; j++ {
			if p.Symms[i][j] {
				fmt.Print("1")
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println()
	}
}
func Extra_SymmPairsCreateFromZdd(dd *DdManager, pairs, vars *DdNode) *Extra_SymmInfo_t {
	support, valid := extraPositiveCube(dd, vars)
	if !valid || dd == nil || pairs == nil || cudd.Cudd_zddUnion(dd, pairs, dd.Zero) == nil {
		return nil
	}
	p := Extra_SymmPairsAllocate(len(support))
	p.NVarsMax = len(dd.Vars)
	copy(p.Vars, support)
	positions := map[int]int{}
	for i, v := range support {
		positions[v] = i
	}
	var visit func(*DdNode, []int) bool
	visit = func(f *DdNode, selected []int) bool {
		if f == dd.Zero {
			return true
		}
		if f == dd.One {
			if len(selected) != 2 {
				return false
			}
			a, okA := positions[selected[0]]
			b, okB := positions[selected[1]]
			if !okA || !okB || a == b {
				return false
			}
			if a > b {
				a, b = b, a
			}
			p.Symms[a][b] = true
			p.NSymms++
			return true
		}
		if f.Index&1 != 0 || len(selected) > 2 {
			return false
		}
		return visit(f.E, selected) && visit(f.T, append(append([]int(nil), selected...), int(f.Index)/2))
	}
	if !visit(pairs, nil) {
		return nil
	}
	return p
}
func Extra_SymmPairsCompute(dd *DdManager, f *DdNode) *Extra_SymmInfo_t {
	if !extraValidBDD(dd, f) {
		return nil
	}
	vars := cudd.Cudd_Support(dd, f)
	return Extra_SymmPairsCreateFromZdd(dd, Extra_zddSymmPairsCompute(dd, f, vars), vars)
}
func Extra_bddCheckVarsSymmetricNaive(dd *DdManager, f *DdNode, a, b int) bool {
	if !extraValidBDD(dd, f) || a < 0 || b < 0 || a == b || a >= len(dd.Vars) || b >= len(dd.Vars) {
		return false
	}
	zeroOne := cudd.Cudd_bddAnd(dd, cudd.Cudd_Not(dd.Vars[a]), dd.Vars[b])
	oneZero := cudd.Cudd_bddAnd(dd, dd.Vars[a], cudd.Cudd_Not(dd.Vars[b]))
	return cudd.Cudd_Cofactor(dd, f, zeroOne) == cudd.Cudd_Cofactor(dd, f, oneZero)
}
func Extra_SymmPairsComputeNaive(dd *DdManager, f *DdNode) *Extra_SymmInfo_t {
	if !extraValidBDD(dd, f) {
		return nil
	}
	support := dd.Support(f)
	p := Extra_SymmPairsAllocate(len(support))
	p.NVarsMax = len(dd.Vars)
	copy(p.Vars, support)
	for i := range support {
		for j := i + 1; j < len(support); j++ {
			if Extra_bddCheckVarsSymmetricNaive(dd, f, support[i], support[j]) {
				p.Symms[i][j], p.Symms[j][i] = true, true
				p.NSymms++
			}
		}
	}
	return p
}

// Typed status replaces the source's overloaded BDD-one/BDD-zero/ADD-zero tags.
type symmetryStatus uint8

const (
	symmetryImpossible symmetryStatus = iota
	symmetryPossible
	symmetrySecondSeen
)

type symmetryPathKey struct {
	f         *DdNode
	firstSeen bool
}

func symmPathCheck(dd *DdManager, f *DdNode, first, second int, seen bool) symmetryStatus {
	memo := map[symmetryPathKey]symmetryStatus{}
	var visit func(*DdNode, bool) symmetryStatus
	visit = func(f *DdNode, seen bool) symmetryStatus {
		if f == dd.LogicZero {
			return symmetryPossible
		}
		key := symmetryPathKey{f, seen}
		if r, ok := memo[key]; ok {
			return r
		}
		level := math.MaxInt
		if !cudd.Cudd_IsConstant(f) {
			level = dd.Perm[cudd.Cudd_Regular(f).Index]
		}
		firstLevel := first
		if seen {
			firstLevel = -1
		}
		var r symmetryStatus
		if level < second {
			lo, hi := extraBDDCofactors(f)
			switch {
			case level < firstLevel:
				r = visit(lo, seen)
				if r != symmetryImpossible {
					r = visit(hi, seen)
				}
			case level == firstLevel:
				a, b := visit(lo, true), visit(hi, true)
				if a == symmetryImpossible || b == symmetryImpossible {
					r = symmetryImpossible
				} else if a == symmetrySecondSeen || b == symmetrySecondSeen {
					r = symmetryPossible
				} else {
					r = symmetryImpossible
				}
			default:
				a, b := visit(lo, seen), visit(hi, seen)
				if a == symmetryImpossible || b == symmetryImpossible {
					r = symmetryImpossible
				} else if a == symmetrySecondSeen || b == symmetrySecondSeen {
					r = symmetrySecondSeen
				} else {
					r = symmetryPossible
				}
			}
		} else if level == second {
			if seen {
				r = symmetrySecondSeen
			} else {
				r = symmetryImpossible
			}
		} else {
			r = symmetryPossible
		}
		memo[key] = r
		return r
	}
	return visit(f, seen)
}
func extraBddCheckVarsSymmetric(dd *DdManager, f, vars *DdNode) *DdNode {
	if !extraValidBDD(dd, f, vars) {
		return nil
	}
	seen := cudd.Cudd_IsComplement(vars)
	regular := cudd.Cudd_Regular(vars)
	indices, valid := extraPositiveCube(dd, regular)
	if !valid || len(indices) < 1 || len(indices) > 2 || seen && len(indices) != 1 {
		return nil
	}
	first, second := -1, dd.Perm[indices[len(indices)-1]]
	if len(indices) == 2 {
		first = dd.Perm[indices[0]]
	}
	switch symmPathCheck(dd, f, first, second, seen) {
	case symmetryPossible:
		return dd.One
	case symmetrySecondSeen:
		return dd.Zero
	default:
		return dd.LogicZero
	}
}
func Extra_bddCheckVarsSymmetric(dd *DdManager, f *DdNode, a, b int) bool {
	if !extraValidBDD(dd, f) || a < 0 || b < 0 || a == b || a >= len(dd.Vars) || b >= len(dd.Vars) {
		return false
	}
	vars := cudd.Cudd_bddAnd(dd, dd.Vars[a], dd.Vars[b])
	return extraBddCheckVarsSymmetric(dd, f, vars) == dd.One
}
func extraZddSymmPairsCompute(dd *DdManager, f, vars *DdNode) *DdNode {
	return Extra_zddSymmPairsCompute(dd, f, vars)
}
func extraZddGetSymmetricVars(dd *DdManager, f, g, vars *DdNode) *DdNode {
	return Extra_zddGetSymmetricVars(dd, f, g, vars)
}
func extraZddGetSingletons(dd *DdManager, vars *DdNode) *DdNode {
	return Extra_zddGetSingletons(dd, vars)
}
func extraBddReduceVarSet(dd *DdManager, vars, f *DdNode) *DdNode {
	return Extra_bddReduceVarSet(dd, vars, f)
}
func extraZddTuplesFromBdd(dd *DdManager, kVars, vars *DdNode) *DdNode {
	indices, valid := extraPositiveCube(dd, kVars)
	if !valid {
		return nil
	}
	return Extra_zddTuplesFromBdd(dd, len(indices), vars)
}
func extraZddSelectOneSubset(dd *DdManager, f *DdNode) *DdNode {
	return Extra_zddSelectOneSubset(dd, f)
}
