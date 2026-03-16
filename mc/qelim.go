package mc

import (
	"fmt"
	"sync/atomic"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
)

// Global counter for ITE elimination fresh variables.
var iteCtr int64

// NextIteCtr returns the next unique ITE counter value.
func NextIteCtr() int64 {
	return atomic.AddInt64(&iteCtr, 1) - 1
}

// Qelim implements quantifier elimination by finite instantiation.
// For each quantified subformula, it either expands it (if the sort is finite)
// or replaces it with a fresh proposition variable and generates constraints.
//
// Python: ivy_mc.py:881-914
type Qelim struct {
	Syms           map[string]lg.Node      // cached quantifier -> result mapping
	SymsCtr        int                     // counter for fresh symbols
	Fmlas          []lg.Node               // accumulated constraints
	SortConstants  map[string][]*lg.Const  // sort -> constants for invariant
	SortConstants2 map[string][]*lg.Const  // sort -> constants for transition
}

// NewQelim creates a new quantifier elimination context.
func NewQelim(sortConstants, sortConstants2 map[string][]*lg.Const) *Qelim {
	return &Qelim{
		Syms:           make(map[string]lg.Node),
		SortConstants:  sortConstants,
		SortConstants2: sortConstants2,
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
func (q *Qelim) GetConsts(s lg.Sort, sortConstants map[string][]*lg.Const) []*lg.Const {
	if s == nil {
		return nil
	}
	name := s.String()
	if consts, ok := sortConstants[name]; ok {
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
	if s == lg.Boolean {
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
func (q *Qelim) QE(expr lg.Node, sortConstants map[string][]*lg.Const) lg.Node {
	switch t := expr.(type) {
	case *lg.ForAll:
		return q.qeQuantifier(t.Variables, t.Body, true, sortConstants)
	case *lg.Exists:
		return q.qeQuantifier(t.Variables, t.Body, false, sortConstants)
	}

	// Recurse into children
	children := expr.Children()
	if len(children) == 0 {
		return expr
	}
	newChildren := make([]lg.Node, len(children))
	changed := false
	for i, c := range children {
		nc := q.QE(c, sortConstants)
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return expr
	}
	return il.CloneNode(expr, newChildren)
}

// qeQuantifier handles quantifier elimination for a single quantifier.
func (q *Qelim) qeQuantifier(vars []*lg.Var, body lg.Node, isForall bool, sortConstants map[string][]*lg.Const) lg.Node {
	// Check cache
	key := fmt.Sprintf("%v:%v:%v", vars, body, isForall)
	if old, ok := q.Syms[key]; ok {
		return old
	}

	// Get constants for each variable's sort
	constSets := make([][]*lg.Const, len(vars))
	for i, v := range vars {
		constSets[i] = q.GetConsts(v.VSort, sortConstants)
		if len(constSets[i]) == 0 {
			// No constants for this sort — can't fully eliminate
			constSets[i] = []*lg.Const{lg.NewConst("_dummy_"+v.Name, v.VSort)}
		}
	}

	// Generate all combinations (cartesian product)
	combos := cartesianProduct(constSets)

	// Build substitution maps and instantiate
	var insts []lg.Node
	for _, combo := range combos {
		subs := make(map[string]lg.Node, len(vars))
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
		var constraint lg.Node
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
func (q *Qelim) Apply(transFmlas, transDefs []lg.Node, invariant lg.Node, indhyps []lg.Node) ([]lg.Node, []lg.Node, lg.Node) {
	// Apply to transition relation
	constants := q.SortConstants2
	newDefs := make([]lg.Node, len(transDefs))
	for i, def := range transDefs {
		newDefs[i] = q.QE(def, constants)
	}
	newFmlas := make([]lg.Node, len(transFmlas))
	for i, fmla := range transFmlas {
		newFmlas[i] = q.QE(closeFormula(fmla), constants)
	}

	// Apply to invariant
	newInv := q.QE(invariant, q.SortConstants)

	// Apply to inductive hypotheses
	newHyps := make([]lg.Node, len(indhyps))
	for i, hyp := range indhyps {
		newHyps[i] = q.QE(hyp, q.SortConstants2)
	}

	// Combine: new_fmlas + indhyps + constraints
	allFmlas := make([]lg.Node, 0, len(newFmlas)+len(newHyps)+len(q.Fmlas))
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
func closeFormula(fmla lg.Node) lg.Node {
	fvs := lu.FreeVariablesList(fmla)
	if len(fvs) == 0 {
		return fmla
	}
	return &lg.ForAll{Variables: fvs, Body: fmla}
}

// ElimIteKey generates a fresh variable name for ITE elimination.
func ElimIteKey(sortName string) string {
	ctr := NextIteCtr()
	return fmt.Sprintf("__ite[%d]:%s", ctr, sortName)
}

// InstantiateAxiomsStub is a stub for pattern-based eager axiom instantiation.
func InstantiateAxiomsStub() []lg.Node {
	return nil
}
