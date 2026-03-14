package mc

import (
	"fmt"
	"sync/atomic"
)

// Global counter for propositional abstraction.
var propAbsCtr int64

// NextPropAbsCtr returns the next unique propositional abstraction counter value.
func NextPropAbsCtr() int64 {
	return atomic.AddInt64(&propAbsCtr, 1) - 1
}

// PropAbs holds the state for propositional abstraction of non-finite atoms.
type PropAbs struct {
	// Map from expression string to abstract proposition name
	Map map[string]string
	// Counter for fresh symbols
	ctr int
	// New state variables introduced by abstraction
	NewStVars []string
	// Finite symbols (not abstracted)
	FiniteSyms    []string
	FiniteSymsSet map[string]bool
}

// NewPropAbs creates a new propositional abstraction context.
func NewPropAbs() *PropAbs {
	return &PropAbs{
		Map:           make(map[string]string),
		FiniteSymsSet: make(map[string]bool),
	}
}

// NewProp returns the abstract proposition for an expression.
// If the expression has not been seen before, a fresh symbol is created.
func (pa *PropAbs) NewProp(expr string) string {
	if res, ok := pa.Map[expr]; ok {
		return res
	}
	name := fmt.Sprintf("__abs[%d]", pa.ctr)
	pa.Map[expr] = name
	pa.ctr++
	return name
}

// MineConstantsStub is a stub for mining constants from formulas.
// The full implementation requires the module infrastructure.
// Returns an empty map.
func MineConstantsStub() map[string][]string {
	return make(map[string][]string)
}

// ToTableLookupStub is a stub for converting function applications to table lookups.
// The full implementation requires the complete logic infrastructure.
func ToTableLookupStub() {
	// Stub: no-op
}
