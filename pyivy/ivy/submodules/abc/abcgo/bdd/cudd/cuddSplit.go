package cudd

import (
	"math"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// Cudd_SplitSet returns a subset of S containing m minterms over xVars.
// As in CUDD, S must depend only on the variables listed in xVars.
func Cudd_SplitSet(manager *DdManager, S *DdNode, xVars []*DdNode, n int32, m float64) *DdNode {
	if manager == nil || n < 0 || int(n) > len(xVars) || math.IsNaN(m) || m < 0 {
		return nil
	}
	zero, one := core.Cudd_ReadLogicZero(manager), core.Cudd_ReadOne(manager)
	if m == 0 {
		return zero
	}
	if S == nil || S == zero {
		return nil
	}
	max := math.Ldexp(1, int(n))
	if m > max {
		return nil
	}
	seen := make([]bool, len(manager.Vars))
	for _, v := range xVars[:n] {
		if v == nil {
			return nil
		}
		idx := int(core.Cudd_NodeReadIndex(v))
		if idx < 0 || idx >= len(manager.Perm) {
			return nil
		}
		seen[manager.Perm[idx]] = true
	}
	if S == one {
		if m == max {
			return S
		}
		return selectMintermsFromUniverse(manager, seen, m)
	}
	counts := make(map[*DdNode]float64)
	num := bddAnnotateMintermCount(manager, S, max, counts)
	if m == num {
		return S
	}
	return cuddSplitSetRecur(manager, counts, seen, S, m, max, 0)
}

// cuddSplitSetRecur follows the CUDD split decision tree. `seen` marks
// variables already fixed by ancestors; `counts` memoizes BDD minterm counts.
func cuddSplitSetRecur(manager *DdManager, counts map[*DdNode]float64, seen []bool, p *DdNode, n, max float64, index int32) *DdNode {
	if core.Cudd_IsConstant(p) {
		return selectMintermsFromUniverse(manager, seen, n)
	}
	r := core.Cudd_Regular(p)
	variable := int(core.Cudd_NodeReadIndex(r))
	level := manager.Perm[variable]
	seen[level] = false
	then, otherwise := core.Cudd_T(r), core.Cudd_E(r)
	if core.Cudd_IsComplement(p) {
		then, otherwise = core.Cudd_Not(then), core.Cudd_Not(otherwise)
	}
	remaining := 0
	for _, used := range seen {
		if used {
			remaining++
		}
	}
	numT := splitNodeCount(manager, then, remaining, counts)
	numE := splitNodeCount(manager, otherwise, remaining, counts)
	v := core.Cudd_bddIthVar(manager, variable)
	if numT == n {
		return core.Cudd_bddAnd(manager, v, then)
	}
	if numE == n {
		return core.Cudd_bddAnd(manager, core.Cudd_Not(v), otherwise)
	}
	if numT < n {
		q := cuddSplitSetRecur(manager, counts, seen, otherwise, n-numT, max, index+1)
		return core.Cudd_bddIte(manager, v, then, q)
	}
	if numE < n {
		q := cuddSplitSetRecur(manager, counts, seen, then, n-numE, max, index+1)
		return core.Cudd_bddIte(manager, v, q, otherwise)
	}
	if core.Cudd_IsConstant(then) {
		q := selectMintermsFromUniverse(manager, seen, n)
		return core.Cudd_bddAnd(manager, v, q)
	}
	if core.Cudd_IsConstant(otherwise) {
		q := selectMintermsFromUniverse(manager, seen, n)
		return core.Cudd_bddAnd(manager, core.Cudd_Not(v), q)
	}
	if numT < numE {
		q := cuddSplitSetRecur(manager, counts, seen, then, n, max, index+1)
		return core.Cudd_bddAnd(manager, v, q)
	}
	q := cuddSplitSetRecur(manager, counts, seen, otherwise, n, max, index+1)
	return core.Cudd_bddAnd(manager, core.Cudd_Not(v), q)
}

func selectMintermsFromUniverse(manager *DdManager, varSeen []bool, n float64) *DdNode {
	vars := make([]*DdNode, 0, len(varSeen))
	for level := len(varSeen) - 1; level >= 0; level-- {
		if varSeen[level] {
			vars = append(vars, core.Cudd_bddIthVar(manager, manager.InvPerm[level]))
		}
	}
	return mintermsFromUniverse(manager, vars, int32(len(vars)), n, 0)
}

func mintermsFromUniverse(manager *DdManager, vars []*DdNode, numVars int32, n float64, index int32) *DdNode {
	one, zero := core.Cudd_ReadOne(manager), core.Cudd_ReadLogicZero(manager)
	max := math.Ldexp(1, int(numVars))
	if n == max {
		return one
	}
	if n == 0 {
		return zero
	}
	if numVars <= 0 || int(index) >= len(vars) {
		return nil
	}
	max2 := max / 2
	if n == max2 {
		return vars[index]
	}
	if n > max2 {
		q := mintermsFromUniverse(manager, vars, numVars-1, n-max2, index+1)
		return core.Cudd_bddIte(manager, vars[index], one, q)
	}
	q := mintermsFromUniverse(manager, vars, numVars-1, n, index+1)
	return core.Cudd_bddAnd(manager, vars[index], q)
}

func bddAnnotateMintermCount(manager *DdManager, node *DdNode, max float64, table map[*DdNode]float64) float64 {
	if node == nil {
		return 0
	}
	if n, ok := table[node]; ok {
		return n
	}
	if core.Cudd_IsConstant(node) {
		if node == core.Cudd_ReadOne(manager) {
			return max
		}
		return 0
	}
	// The total count is scaled to the caller's universe. This annotation is
	// retained separately from splitNodeCount to match the source algorithm.
	p := core.Cudd_CountMinterm(manager, node, len(manager.Vars))
	full := math.Ldexp(1, len(manager.Vars))
	count := p * max / full
	table[node] = count
	return count
}

func splitNodeCount(manager *DdManager, node *DdNode, remaining int, counts map[*DdNode]float64) float64 {
	if node == nil || node == core.Cudd_ReadLogicZero(manager) {
		return 0
	}
	if node == core.Cudd_ReadOne(manager) {
		return math.Ldexp(1, remaining)
	}
	return math.Ldexp(core.Cudd_CountMinterm(manager, node, len(manager.Vars)), remaining-len(manager.Vars))
}
