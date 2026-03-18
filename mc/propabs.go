package mc

import (
	"fmt"
	"sync/atomic"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

// Global counter for propositional abstraction.
var propAbsCtr int64

// NextPropAbsCtr returns the next unique propositional abstraction counter value.
func NextPropAbsCtr() int64 {
	return atomic.AddInt64(&propAbsCtr, 1) - 1
}

// PropAbs holds the state for propositional abstraction of non-finite atoms.
// Non-propositional atoms (quantifiers, non-finite-sort applications) are
// replaced with fresh boolean variables.
//
// Python: ivy_mc.py:1287-1318
type PropAbs struct {
	// Map from expression key to abstract proposition
	Map map[string]*lg.Symbol
	// Counter for fresh symbols
	Ctr int
	// New state variables introduced by abstraction
	NewStVars []*lg.Symbol
	// Finite symbols (not abstracted)
	FiniteSyms    []*lg.Symbol
	FiniteSymsSet map[string]bool
	// State variable set (for prev_expr detection)
	StVarSet map[string]bool
	// Sort constants (for prev_expr detection)
	SortConstants map[string][]*lg.Symbol
	// Accumulated formulas from abstraction
	Fmlas []lg.Expr
}

// NewPropAbs creates a new propositional abstraction context.
func NewPropAbs(stVarSet map[string]bool, sortConstants map[string][]*lg.Symbol) *PropAbs {
	return &PropAbs{
		Map:           make(map[string]*lg.Symbol),
		FiniteSymsSet: make(map[string]bool),
		StVarSet:      stVarSet,
		SortConstants: sortConstants,
	}
}

// newProp returns the abstract proposition for an expression.
// If the expression is a "prev_expr" (refers to next-state of a state var),
// it links the new variable to the old one.
// Python: ivy_mc.py:1287-1303
func (pa *PropAbs) newProp(expr lg.Expr) *lg.Symbol {
	key := fmt.Sprint(expr)
	if res, ok := pa.Map[key]; ok {
		return res
	}

	// Check if this is a prev_expr (next-state of a known state variable)
	if prevExpr := pa.prevExpr(expr); prevExpr != nil {
		prevAbs := pa.newProp(prevExpr)
		pa.NewStVars = append(pa.NewStVars, prevAbs)
		// Create next-state version
		nextName := fmt.Sprintf("__abs[%d]", pa.Ctr)
		pa.Ctr++
		res := lg.NewSymbol(nextName, lg.Boolean)
		pa.Map[key] = res
		return res
	}

	name := fmt.Sprintf("__abs[%d]", pa.Ctr)
	pa.Ctr++
	res := lg.NewSymbol(name, lg.Boolean)
	pa.Map[key] = res
	return res
}

// prevExpr checks if an expression is the "next-state" version of
// an expression involving only state variables. If so, returns the
// "current-state" version. This is used to link abstract variables
// across time steps.
//
// Python: ivy_mc.py prev_expr()
func (pa *PropAbs) prevExpr(expr lg.Expr) lg.Expr {
	return PrevExpr(pa.StVarSet, expr, pa.SortConstants)
}

// MkPropAbs performs propositional abstraction on an expression.
// Replaces non-propositional atoms with fresh boolean variables.
//
// An atom is abstracted if:
// - It is a quantifier (forall/exists)
// - It has arguments with non-finite sorts
// - It is an uninterpreted function application
//
// Constants that are finite-sort and not constructors/numerals are
// tracked as finite symbols.
//
// Python: ivy_mc.py:1308-1318
func (pa *PropAbs) MkPropAbs(expr lg.Expr) lg.Expr {
	// Check if this needs abstraction
	needsAbstraction := false

	switch t := expr.(type) {
	case *lg.ForAll, *lg.Exists:
		needsAbstraction = true
	case *lg.Apply:
		// Check if any argument has non-finite sort
		for _, arg := range t.Terms {
			if !isFiniteSort(arg.NodeSort()) {
				needsAbstraction = true
				break
			}
		}
		// Uninterpreted function application
		if !needsAbstraction {
			if c, ok := t.Func.(*lg.Symbol); ok {
				if !isInterpretedSymbol(c) {
					needsAbstraction = true
				}
			}
		}
	default:
		// Not a compound expression needing abstraction
	}

	if needsAbstraction {
		return pa.newProp(expr)
	}

	// Track finite symbols (non-numeral, non-constructor constants)
	if c, ok := expr.(*lg.Symbol); ok {
		if !il.IsNumeral(c) && !pa.FiniteSymsSet[c.Name] {
			pa.FiniteSymsSet[c.Name] = true
			pa.FiniteSyms = append(pa.FiniteSyms, c)
		}
	}

	// Recurse into children
	children := expr.Children()
	if len(children) == 0 {
		return expr
	}
	newChildren := make([]lg.Expr, len(children))
	changed := false
	for i, child := range children {
		nc := pa.MkPropAbs(child)
		newChildren[i] = nc
		if nc != child {
			changed = true
		}
	}
	if !changed {
		return expr
	}
	return il.CloneNode(expr, newChildren)
}

// Apply applies propositional abstraction to transition formulas and definitions.
// Returns (new_fmlas, new_defs).
func (pa *PropAbs) Apply(transFmlas, transDefs []lg.Expr) ([]lg.Expr, []lg.Expr) {
	newDefs := make([]lg.Expr, len(transDefs))
	for i, def := range transDefs {
		newDefs[i] = pa.MkPropAbs(def)
	}
	newFmlas := make([]lg.Expr, len(transFmlas))
	for i, fmla := range transFmlas {
		newFmlas[i] = pa.MkPropAbs(closeFormula(fmla))
	}
	return newFmlas, newDefs
}

// isInterpretedSymbol checks if a symbol has a built-in interpretation.
func isInterpretedSymbol(c *lg.Symbol) bool {
	// Common interpreted symbols
	switch c.Name {
	case "+", "-", "*", "/", "<", "<=", ">", ">=", "=":
		return true
	}
	return false
}
