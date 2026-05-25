package goivy

import (
	"fmt"
	//"sync/atomic"
)

// unused globals, comment out; add on a Config if needed.
// Global counter for propositional abstraction.
//var propAbsCtr int64
//
// NextPropAbsCtr returns the next unique propositional abstraction counter value.
//func NextPropAbsCtr() int64 {
//	return atomic.AddInt64(&propAbsCtr, 1) - 1
//}

// PropAbs holds the state for propositional abstraction of non-finite atoms.
// Non-propositional atoms (quantifiers, non-finite-sort applications) are
// replaced with fresh variables of the abstracted expression's sort.
//
// Python: ivy_mc.py:1287-1318
type PropAbs struct {
	// Map from expression key to abstract proposition
	Map *InsMap[string, *Const]
	// OrigExprs maps the same string keys to the original expressions,
	// so we can check immutability of the original (not the abstract var).
	// Python keeps this naturally since prop_abs maps expr->var directly.
	OrigExprs *InsMap[string, Expr]
	// Counter for fresh symbols
	Ctr int
	// New state variables introduced by abstraction
	NewStVars []*Const
	// Finite symbols (not abstracted)
	FiniteSyms    []*Const
	FiniteSymsSet map[string]bool
	// State variable set (for prev_expr detection)
	StVarSet map[string]bool
	// Sort constants (for prev_expr detection)
	SortConstants *InsMap[string, []*Const]
	// Sort interpretation table used to recognize finite interpreted sorts.
	Interp map[string]interface{}
	// Constructors are excluded from finite symbol state tracking.
	Constructors map[string]bool
	// Accumulated formulas from abstraction
	Fmlas []Expr
}

// NewPropAbs creates a new propositional abstraction context.
func NewPropAbs(stVarSet map[string]bool, sortConstants *InsMap[string, []*Const], interp ...map[string]interface{}) *PropAbs {
	var sortInterp map[string]interface{}
	if len(interp) > 0 {
		sortInterp = interp[0]
	}
	return &PropAbs{
		Map:           NewInsMap[string, *Const](),
		OrigExprs:     NewInsMap[string, Expr](),
		FiniteSymsSet: make(map[string]bool),
		StVarSet:      stVarSet,
		SortConstants: sortConstants,
		Interp:        sortInterp,
	}
}

func (pa *PropAbs) isFiniteSort(s Sort) bool {
	if pa.Interp != nil {
		return IsFiniteSortWithInterp(s, pa.Interp)
	}
	return isFiniteSort(s)
}

func (pa *PropAbs) isConstructor(c *Const) bool {
	return pa != nil && pa.Constructors != nil && pa.Constructors[c.Name]
}

// newProp returns the abstract variable for an expression.
// If the expression is a "prev_expr" (refers to next-state of a state var),
// it links the new variable to the old one.
// Python: ivy_mc.py:1287-1303
func (pa *PropAbs) newProp(expr Expr) *Const {
	key := string(expr.Sexp())
	if res, ok := pa.Map.Get2(key); ok {
		return res
	}

	// Check if this is a prev_expr (next-state of a known state variable)
	if prevExpr := pa.prevExpr(expr); prevExpr != nil {
		prevAbs := pa.newProp(prevExpr)
		pa.NewStVars = append(pa.NewStVars, prevAbs)
		// Python: res = tr.new(pva) — reuses pva's name with new_ prefix,
		// does NOT consume a fresh counter value.
		res := NewConst(ActionNewName(prevAbs.Name), prevAbs.CSort)
		pa.Map.Set(key, res)
		pa.OrigExprs.Set(key, expr)
		return res
	}

	name := fmt.Sprintf("__abs[%d]", pa.Ctr)
	pa.Ctr++
	res := NewConst(name, expr.NodeSort())
	pa.Map.Set(key, res)
	pa.OrigExprs.Set(key, expr)
	return res
}

// prevExpr checks if an expression is the "next-state" version of
// an expression involving only state variables. If so, returns the
// "current-state" version. This is used to link abstract variables
// across time steps.
//
// Python: ivy_mc.py prev_expr()
func (pa *PropAbs) prevExpr(expr Expr) Expr {
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
func (pa *PropAbs) MkPropAbs(expr Expr) Expr {
	// Check if this needs abstraction.
	// Python: if (is_quantifier(expr) or
	//            len(expr.args) > 0 and (
	//              any(not is_finite_sort(a.sort) for a in expr.args)
	//              or is_app(expr) and not is_interpreted_symbol(expr.func))):
	needsAbstraction := false

	switch expr.(type) {
	case *ForAll, *LogicExists:
		needsAbstraction = true
	default:
		children := expr.Children()
		if len(children) > 0 {
			for _, child := range children {
				if !pa.isFiniteSort(child.NodeSort()) {
					needsAbstraction = true
					break
				}
			}
			if !needsAbstraction {
				if app, ok := expr.(*Apply); ok {
					if c, ok := app.Func.(*Const); ok {
						if !isInterpretedSymbol(c) {
							needsAbstraction = true
						}
					}
				}
			}
		}
	}

	if needsAbstraction {
		return pa.newProp(expr)
	}

	// Track finite symbols (non-numeral, non-constructor constants)
	if c, ok := expr.(*Const); ok {
		if !IsNumeral(c) && !pa.isConstructor(c) && !IsSkolem(c.Name) && !pa.FiniteSymsSet[c.Name] {
			pa.FiniteSymsSet[c.Name] = true
			pa.FiniteSyms = append(pa.FiniteSyms, c)
		}
	}

	// Recurse into children
	children := expr.Children()
	if len(children) == 0 {
		return expr
	}
	newChildren := make([]Expr, len(children))
	for i, child := range children {
		newChildren[i] = pa.MkPropAbs(child)
	}
	return CloneNode(expr, newChildren)
}

// Apply applies propositional abstraction to transition formulas and definitions.
// Returns (new_fmlas, new_defs).
func (pa *PropAbs) Apply(transFmlas, transDefs []Expr) ([]Expr, []Expr) {
	newDefs := make([]Expr, len(transDefs))
	for i, def := range transDefs {
		newDefs[i] = pa.MkPropAbs(def)
	}
	newFmlas := make([]Expr, len(transFmlas))
	for i, fmla := range transFmlas {
		newFmlas[i] = pa.MkPropAbs(closeFormula(fmla))
	}
	return newFmlas, newDefs
}

// isInterpretedSymbol checks if a symbol has a built-in interpretation.
func isInterpretedSymbol(c *Const) bool {
	// Common interpreted symbols
	switch c.Name {
	case "+", "-", "*", "/", "<", "<=", ">", ">=", "=":
		return true
	}
	return false
}
