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
}

// NewQelim creates a new quantifier elimination context.
func NewQelim(sortConstants, sortConstants2 *InsMap[string, []*Const], iuCfg *IvyUtilsConfig, fullQI bool) *Qelim {
	return &Qelim{
		Syms:           make(map[string]Expr),
		SortConstants:  sortConstants,
		SortConstants2: sortConstants2,
		IuCfg:          iuCfg,
		FullQI:         fullQI,
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

// GetConsts returns the constants to instantiate for a given sort.
func (q *Qelim) GetConsts(s Sort, sortConstants *InsMap[string, []*Const]) []*Const {
	if s == nil {
		return nil
	}
	name := s.String()
	if consts, ok := sortConstants.Get2(name); ok {
		return consts
	}
	// For enumerated sorts, generate all values
	if es, ok := s.(*LogicEnumeratedSort); ok {
		consts := make([]*Const, len(es.Extension))
		for i, v := range es.Extension {
			consts[i] = NewConst(v, s)
		}
		return consts
	}
	return nil
}

// isFiniteSort checks if a sort is finite (enumerated or boolean).
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

	// Get constants for each variable's sort
	constSets := make([][]*Const, len(vars))
	for i, v := range vars {
		constSets[i] = q.GetConsts(v.VSort, sortConstants)
	}

	// Generate all combinations (cartesian product)
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

	// If all sorts are finite, expand directly
	allFinite := true
	for _, v := range vars {
		if !isFiniteSort(v.VSort) {
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
func mcCartesianProduct(sets [][]*Const) [][]*Const {
	if len(sets) == 0 {
		return [][]*Const{{}}
	}
	first := sets[0]
	rest := mcCartesianProduct(sets[1:])
	var result [][]*Const
	for _, v := range first {
		for _, r := range rest {
			combo := make([]*Const, 1+len(r))
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
