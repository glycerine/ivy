package goivy

import (
	"fmt"
	"sync/atomic"
)

// nextIteCtr increments the ITE counter on the given pointer and returns the old value.
func nextIteCtr(ctr *int64) int64 {
	return atomic.AddInt64(ctr, 1) - 1
}

// Qelim implements quantifier elimination by finite instantiation.
// For each quantified subformula, it either expands it (if the sort is finite)
// or replaces it with a fresh proposition variable and generates constraints.
//
// Python: ivy_mc.py:881-914
type Qelim struct {
	Syms           map[string]Expr           // cached quantifier -> result mapping
	SymsCtr        int                       // counter for fresh symbols
	Fmlas          []Expr                    // accumulated constraints
	SortConstants  *InsMap[string, []*Const] // sort -> constants for invariant
	SortConstants2 *InsMap[string, []*Const] // sort -> constants for transition
	IuCfg          *IvyUtilsConfig           // for IsMacro/ExpandMacro
	FullQI         bool                      // Python: fullqi parameter (default false)
	Interp         map[string]interface{}    // Python: il.sig.interp — for get_sort_theory
}

// NewQelim creates a new quantifier elimination context.
func NewQelim(sortConstants, sortConstants2 *InsMap[string, []*Const], iuCfg *IvyUtilsConfig, fullQI bool, interp map[string]interface{}) *Qelim {
	return &Qelim{
		Syms:           make(map[string]Expr),
		SortConstants:  sortConstants,
		SortConstants2: sortConstants2,
		IuCfg:          iuCfg,
		FullQI:         fullQI,
		Interp:         interp,
	}
}

// Fresh creates a fresh proposition variable for a quantified expression.
func (q *Qelim) Fresh(exprKey string) *Const {
	name := fmt.Sprintf("__qe[%d]", q.SymsCtr)
	q.SymsCtr++
	c := NewConst(name, Boolean)
	q.Syms[exprKey] = c
	return c
}

// isFiniteSort checks if a sort is finite by direct type inspection
// (enumerated, boolean). Does NOT consult the interp map — callers that
// need to recognise interpreted finite sorts (range, bitvector) must use
// IsFiniteSortWithInterp from mc_helpers.go.
//
// This narrow helper is kept for non-Qelim callers in mc_propabs.go,
// mc_toaiger.go, and mc_transforms.go whose interp-awareness is a separate
// porting task. Qelim itself uses IsFiniteSortWithInterp.
func isFiniteSort(s Sort) bool {
	if s == nil {
		return false
	}
	if SortEqual(s, Boolean) {
		return true
	}
	if _, ok := s.(*LogicEnumeratedSort); ok {
		return true
	}
	return false
}

// GetConsts returns the values to instantiate for a given sort.
// Python: get_consts(sort, sort_constants) at ivy_mc.py:908-911:
//
//	if is_finite_sort(sort): return sort_values(sort)
//	return sort_constants[sort]
//
// For finite sorts (enumerated, range, bitvector, boolean — checked through
// the interp map via GetSortTheory) it returns the canonical ground values
// from sortValuesAsExprs; sort_constants is ignored in that case (Python
// parity: sort_constants may carry mined skolems / action params, but the
// finite-sort enumeration uses the canonical [0..N] / extension instead).
//
// For non-finite sorts, it falls back to sortConstants[name], widening the
// []*Const slice into []Expr so callers can substitute Boolean Or()/And()
// uniformly with Const symbols.
func (q *Qelim) GetConsts(s Sort, sortConstants *InsMap[string, []*Const]) []Expr {
	if s == nil {
		return nil
	}
	if IsFiniteSortWithInterp(s, q.Interp) {
		return sortValuesAsExprs(s, q.Interp)
	}
	name := s.String()
	if consts, ok := sortConstants.Get2(name); ok {
		out := make([]Expr, len(consts))
		for i, c := range consts {
			out[i] = c
		}
		return out
	}
	return nil
}

// QE performs quantifier elimination on an expression.
// For universally quantified formulas over finite sorts, it expands to conjunction.
// For existentially quantified formulas over finite sorts, it expands to disjunction.
// For infinite sorts, it creates a fresh proposition with constraints.
//
// Python: ivy_mc.py:881-902
func (q *Qelim) QE(expr Expr, sortConstants *InsMap[string, []*Const]) Expr {
	switch t := expr.(type) {
	case *ForAll:
		return q.qeQuantifier(t.Variables, t.Body, true, sortConstants)
	case *LogicExists:
		return q.qeQuantifier(t.Variables, t.Body, false, sortConstants)
	}

	if IsMacro(expr, q.IuCfg) {
		return q.QE(ExpandMacro(expr), sortConstants)
	}

	// Recurse into children
	children := expr.Children()
	if len(children) == 0 {
		return expr
	}
	newChildren := make([]Expr, len(children))
	for i, c := range children {
		newChildren[i] = q.QE(c, sortConstants)
	}
	return cloneNormal(expr, newChildren)
}

// qeQuantifier handles quantifier elimination for a single quantifier.
func (q *Qelim) qeQuantifier(vars []*LogicVariable, body Expr, isForall bool, sortConstants *InsMap[string, []*Const]) Expr {
	// Check cache
	key := fmt.Sprintf("%v:%v:%v", vars, body, isForall)
	if old, ok := q.Syms[key]; ok {
		return old
	}

	// Get values for each variable's sort. Python: consts = [self.get_consts(x.sort,sort_constants) for x in expr.variables].
	constSets := make([][]Expr, len(vars))
	for i, v := range vars {
		constSets[i] = q.GetConsts(v.VSort, sortConstants)
	}

	// Generate all combinations (cartesian product). Python: itertools.product(*consts).
	combos := mcCartesianProduct(constSets)

	// Build substitution maps and instantiate
	var insts []Expr
	for _, combo := range combos {
		subs := make(map[string]Expr, len(vars))
		for i, v := range vars {
			subs[v.Name] = combo[i]
		}
		inst := SubstituteByName(body, subs)
		inst = q.QE(inst, sortConstants)
		insts = append(insts, inst)
	}

	// If all sorts are finite, expand directly. Python: all(is_finite_sort(x.sort) for x in expr.variables).
	allFinite := true
	for _, v := range vars {
		if !IsFiniteSortWithInterp(v.VSort, q.Interp) {
			allFinite = false
			break
		}
	}

	if allFinite {
		if isForall {
			return &LogicAnd{Terms: insts}
		}
		return &LogicOr{Terms: insts}
	}

	// Infinite sorts: introduce fresh proposition with constraints
	res := q.Fresh(key)
	for _, inst := range insts {
		var constraint Expr
		if isForall {
			constraint = &LogicImplies{T1: res, T2: inst}
		} else {
			constraint = &LogicImplies{T1: inst, T2: res}
		}
		q.Fmlas = append(q.Fmlas, constraint)
	}
	return res
}

// Apply applies quantifier elimination to a transition relation and invariant.
// Python: ivy_mc.py:903-914
func (q *Qelim) Apply(transFmlas, transDefs []Expr, invariant Expr, indhyps []Expr) ([]Expr, []Expr, Expr) {
	// Apply to transition relation
	// Python: constants = self.sort_constants2 if fullqi.get() else self.sort_constants
	constants := q.SortConstants
	if q.FullQI {
		constants = q.SortConstants2
	}
	newDefs := make([]Expr, len(transDefs))
	for i, def := range transDefs {
		newDefs[i] = q.QE(def, constants)
	}
	newFmlas := make([]Expr, len(transFmlas))
	for i, fmla := range transFmlas {
		newFmlas[i] = q.QE(closeFormula(fmla), constants)
	}

	// Apply to invariant
	newInv := q.QE(invariant, q.SortConstants)

	// Apply to inductive hypotheses
	newHyps := make([]Expr, len(indhyps))
	for i, hyp := range indhyps {
		newHyps[i] = q.QE(hyp, q.SortConstants2)
	}

	// Combine: new_fmlas + indhyps + constraints
	allFmlas := make([]Expr, 0, len(newFmlas)+len(newHyps)+len(q.Fmlas))
	allFmlas = append(allFmlas, newFmlas...)
	allFmlas = append(allFmlas, newHyps...)
	allFmlas = append(allFmlas, q.Fmlas...)

	return allFmlas, newDefs, newInv
}

// --- Helper functions ---

// cartesianProduct computes the cartesian product of multiple slices.
// Python: itertools.product(*consts) at ivy_mc.py:919.
func mcCartesianProduct(sets [][]Expr) [][]Expr {
	if len(sets) == 0 {
		return [][]Expr{{}}
	}
	first := sets[0]
	rest := mcCartesianProduct(sets[1:])
	var result [][]Expr
	for _, v := range first {
		for _, r := range rest {
			combo := make([]Expr, 1+len(r))
			combo[0] = v
			copy(combo[1:], r)
			result = append(result, combo)
		}
	}
	return result
}

// closeFormula universally closes a formula (binds all free variables).
func closeFormula(fmla Expr) Expr {
	fvs := FreeVariablesList(fmla)
	if len(fvs) == 0 {
		return fmla
	}
	return &ForAll{Variables: fvs, Body: fmla}
}

// ElimIteKey generates a fresh variable name for ITE elimination.
func ElimIteKey(sortName string, iteCtr *int64) string {
	ctr := nextIteCtr(iteCtr)
	return fmt.Sprintf("__ite[%d]:%s", ctr, sortName)
}
