package mc

import (
	"fmt"
	"sync/atomic"
)

// Global counter for ITE elimination fresh variables.
var iteCtr int64

// NextIteCtr returns the next unique ITE counter value.
func NextIteCtr() int64 {
	return atomic.AddInt64(&iteCtr, 1) - 1
}

// Qelim implements quantifier elimination by finite instantiation.
// For each quantified subformula, it generates constraint instances.
type Qelim struct {
	Syms           map[string]string // map from quantified formula repr to proposition variable name
	SymsCtr        int               // counter for fresh symbols
	Fmlas          []string          // accumulated constraints
	SortConstants  map[string][]string // sort -> constants for invariant
	SortConstants2 map[string][]string // sort -> constants for transition
}

// NewQelim creates a new quantifier elimination context.
func NewQelim(sortConstants, sortConstants2 map[string][]string) *Qelim {
	return &Qelim{
		Syms:           make(map[string]string),
		SortConstants:  sortConstants,
		SortConstants2: sortConstants2,
	}
}

// Fresh creates a fresh proposition variable name for a quantified expression.
func (q *Qelim) Fresh(exprKey string) string {
	name := fmt.Sprintf("__qe[%d]", q.SymsCtr)
	q.Syms[exprKey] = name
	q.SymsCtr++
	return name
}

// GetConsts returns the constants to instantiate for a given sort.
// If the sort is finite, it enumerates all values.
func (q *Qelim) GetConsts(sortName string, sortConstants map[string][]string) []string {
	// In the Python version, this checks is_finite_sort and calls sort_values.
	// Here we provide what's in the sort constants map.
	if consts, ok := sortConstants[sortName]; ok {
		return consts
	}
	return nil
}

// ElimIteKey generates a fresh variable name for ITE elimination.
func ElimIteKey(sortName string) string {
	ctr := NextIteCtr()
	return fmt.Sprintf("__ite[%d]:%s", ctr, sortName)
}

// InstantiateAxiomsStub is a stub for pattern-based eager axiom instantiation.
// The full implementation requires the module and solver infrastructure.
// It returns an empty list of instantiated axioms.
func InstantiateAxiomsStub(
	stvars []string,
	transDefs []string,
	transFmlas []string,
	invariant string,
	sortConstants map[string][]string,
	funs []string,
) []string {
	// Stub: return empty list.
	// Full implementation would:
	// 1. Expand axiom schemata
	// 2. Extract triggers from axioms
	// 3. Match triggers against defs/fmlas/invariant
	// 4. Instantiate matched axioms
	return nil
}
