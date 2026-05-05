package mc

import (
	"fmt"
	"sync/atomic"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
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
	Syms           map[string]lg.Expr      // cached quantifier -> result mapping
	SymsCtr        int                     // counter for fresh symbols
	Fmlas          []lg.Expr               // accumulated constraints
	SortConstants  *iu.InsMap[string, []*lg.Const] // sort -> constants for invariant
	SortConstants2 *iu.InsMap[string, []*lg.Const] // sort -> constants for transition
	IuCfg          *iu.IvyUtilsConfig      // for IsMacro/ExpandMacro
	FullQI         bool                    // Python: fullqi parameter (default false)
}

// NewQelim creates a new quantifier elimination context.
func NewQelim(sortConstants, sortConstants2 *iu.InsMap[string, []*lg.Const], iuCfg *iu.IvyUtilsConfig, fullQI bool) *Qelim {
	return &Qelim{
		Syms:           make(map[string]lg.Expr),
		SortConstants:  sortConstants,
		SortConstants2: sortConstants2,
		IuCfg:          iuCfg,
		FullQI:         fullQI,
	}
}

// Fresh creates a fresh proposition variable for a quantified expression.
func (q *Qelim) Fresh(exprKey string) *lg.Const {
	name := fmt.Sprintf("__qe[%d]", q.SymsCtr)
	q.SymsCtr++
	c := lg.NewConst(name, lg.Boolean)
	q.Syms[exprKey] = c
	return c
}

// GetConsts returns the constants to instantiate for a given sort.
func (q *Qelim) GetConsts(s lg.Sort, sortConstants *iu.InsMap[string, []*lg.Const]) []*lg.Const {
	if s == nil {
		return nil
	}
	name := s.String()
	if consts, ok := sortConstants.Get2(name); ok {
		return consts
	}
	// For enumerated sorts, generate all values
	if es, ok := s.(*lg.EnumeratedSort); ok {
		consts := make([]*lg.Const, len(es.Extension))
		for i, v := range es.Extension {
			consts[i] = lg.NewConst(v, s)
		}
		return consts
	}
	return nil
}

// isFiniteSort checks if a sort is finite (enumerated or boolean).
func isFiniteSort(s lg.Sort) bool {
	if s == nil {
		return false
	}
	if lg.SortEqual(s, lg.Boolean) {
		return true
	}
	if _, ok := s.(*lg.EnumeratedSort); ok {
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
func (q *Qelim) QE(expr lg.Expr, sortConstants *iu.InsMap[string, []*lg.Const]) lg.Expr {
	switch t := expr.(type) {
	case *lg.ForAll:
		return q.qeQuantifier(t.Variables, t.Body, true, sortConstants)
	case *lg.Exists:
		return q.qeQuantifier(t.Variables, t.Body, false, sortConstants)
	}

	if il.IsMacro(expr, q.IuCfg) {
		return q.QE(il.ExpandMacro(expr), sortConstants)
	}

	// Recurse into children
	children := expr.Children()
	if len(children) == 0 {
		return expr
	}
	newChildren := make([]lg.Expr, len(children))
	for i, c := range children {
		newChildren[i] = q.QE(c, sortConstants)
	}
	return cloneNormal(expr, newChildren)
}

// qeQuantifier handles quantifier elimination for a single quantifier.
func (q *Qelim) qeQuantifier(vars []*lg.Variable, body lg.Expr, isForall bool, sortConstants *iu.InsMap[string, []*lg.Const]) lg.Expr {
	// Check cache
	key := fmt.Sprintf("%v:%v:%v", vars, body, isForall)
	if old, ok := q.Syms[key]; ok {
		return old
	}

	// Get constants for each variable's sort
	constSets := make([][]*lg.Const, len(vars))
	for i, v := range vars {
		constSets[i] = q.GetConsts(v.VSort, sortConstants)
	}

	// Generate all combinations (cartesian product)
	combos := cartesianProduct(constSets)

	// Build substitution maps and instantiate
	var insts []lg.Expr
	for _, combo := range combos {
		subs := make(map[string]lg.Expr, len(vars))
		for i, v := range vars {
			subs[v.Name] = combo[i]
		}
		inst := lu.SubstituteByName(body, subs)
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
			return &lg.And{Terms: insts}
		}
		return &lg.Or{Terms: insts}
	}

	// Infinite sorts: introduce fresh proposition with constraints
	res := q.Fresh(key)
	for _, inst := range insts {
		var constraint lg.Expr
		if isForall {
			constraint = &lg.Implies{T1: res, T2: inst}
		} else {
			constraint = &lg.Implies{T1: inst, T2: res}
		}
		q.Fmlas = append(q.Fmlas, constraint)
	}
	return res
}

// Apply applies quantifier elimination to a transition relation and invariant.
// Python: ivy_mc.py:903-914
func (q *Qelim) Apply(transFmlas, transDefs []lg.Expr, invariant lg.Expr, indhyps []lg.Expr) ([]lg.Expr, []lg.Expr, lg.Expr) {
	// Apply to transition relation
	// Python: constants = self.sort_constants2 if fullqi.get() else self.sort_constants
	constants := q.SortConstants
	if q.FullQI {
		constants = q.SortConstants2
	}
	newDefs := make([]lg.Expr, len(transDefs))
	for i, def := range transDefs {
		newDefs[i] = q.QE(def, constants)
	}
	newFmlas := make([]lg.Expr, len(transFmlas))
	for i, fmla := range transFmlas {
		newFmlas[i] = q.QE(closeFormula(fmla), constants)
	}

	// Apply to invariant
	newInv := q.QE(invariant, q.SortConstants)

	// Apply to inductive hypotheses
	newHyps := make([]lg.Expr, len(indhyps))
	for i, hyp := range indhyps {
		newHyps[i] = q.QE(hyp, q.SortConstants2)
	}

	// Combine: new_fmlas + indhyps + constraints
	allFmlas := make([]lg.Expr, 0, len(newFmlas)+len(newHyps)+len(q.Fmlas))
	allFmlas = append(allFmlas, newFmlas...)
	allFmlas = append(allFmlas, newHyps...)
	allFmlas = append(allFmlas, q.Fmlas...)

	return allFmlas, newDefs, newInv
}

// --- Helper functions ---

// cartesianProduct computes the cartesian product of multiple slices.
func cartesianProduct(sets [][]*lg.Const) [][]*lg.Const {
	if len(sets) == 0 {
		return [][]*lg.Const{{}}
	}
	first := sets[0]
	rest := cartesianProduct(sets[1:])
	var result [][]*lg.Const
	for _, v := range first {
		for _, r := range rest {
			combo := make([]*lg.Const, 1+len(r))
			combo[0] = v
			copy(combo[1:], r)
			result = append(result, combo)
		}
	}
	return result
}

// closeFormula universally closes a formula (binds all free variables).
func closeFormula(fmla lg.Expr) lg.Expr {
	fvs := lu.FreeVariablesList(fmla)
	if len(fvs) == 0 {
		return fmla
	}
	return &lg.ForAll{Variables: fvs, Body: fmla}
}

// ElimIteKey generates a fresh variable name for ITE elimination.
func ElimIteKey(sortName string, iteCtr *int64) string {
	ctr := nextIteCtr(iteCtr)
	return fmt.Sprintf("__ite[%d]:%s", ctr, sortName)
}

