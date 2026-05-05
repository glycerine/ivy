package goivy

import (
	"fmt"
)

// SortInfer performs sort inference on a term, optionally constraining it
// to the given sort. Returns the concretized term.
// Corresponds to Python's sort_infer.
func SortInfer(term Expr, sort Sort) (Expr, error) {
	res, err := ConcretizeSorts(term, sort)
	if err != nil {
		return nil, err
	}
	if err := CheckConcretelySorted(res, nil); err != nil {
		return nil, err
	}
	return res, nil
}

// SortInferList performs sort inference on a list of terms using a shared
// unification environment so that variables/constants with the same name
// across terms are unified to the same sort.
// Corresponds to Python's sort_infer_list.
func SortInferList(terms []Expr, sorts []Sort, unsortedVarNames map[string]bool) ([]Expr, error) {
	result, err := ConcretizeTerms(terms, sorts)
	if err != nil {
		return nil, err
	}
	for _, t := range result {
		if err := CheckConcretelySorted(t, unsortedVarNames); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Sortify adds sorts to an untyped AST by looking up symbols in the
// current signature context. Recursively processes children.
// Corresponds to Python's sortify.
func Sortify(sig *Sig, node Expr) Expr {
	args := NodeArgs(node)
	newArgs := make([]Expr, len(args))
	for i, arg := range args {
		newArgs[i] = Sortify(sig, arg)
	}

	// If this is an Apply with a TopSort-sorted function, look it up
	if app, ok := node.(*Apply); ok {
		if _, isTop := app.Func.NodeSort().(*TopSort); isTop {
			if c, ok := app.Func.(*Const); ok {
				sym, err := sig.FindSymbol(c.Name, false)
				if err == nil {
					// Reconstruct the apply with the resolved symbol
					result, applyErr := NewApply(sym, newArgs...)
					if applyErr == nil {
						return result
					}
				}
			}
		}
	}

	return CloneNode(node, newArgs)
}

// CheckConcretelySorted checks that all variables and constants in the
// term have concrete sorts (no TopSort or polymorphic elements).
// Returns an error if any unsorted element is found.
// unsortedVarNames, if non-nil, lists variable names to exempt from the check.
// Corresponds to Python's check_concretely_sorted (ivy_logic.py:1194-1200):
//
//	for x in chain(lu.used_variables(term), lu.used_constants(term)):
//	    if lg.contains_topsort(x.sort) or lg.is_polymorphic(x.sort):
//	        ...
//
// Note: Python checks the SORT of x, not x itself. A Const named "X" or "<="
// is name-polymorphic, but its inferred sort is concrete — Python deems it OK.
// Earlier Go versions checked the node itself (lg.IsPolymorphic(v)), which
// over-rejected name-polymorphic constants whose sort was actually concrete.
func CheckConcretelySorted(term Expr, unsortedVarNames map[string]bool) error {
	usedVars := UsedVariables(term)
	for _, v := range usedVars {
		if unsortedVarNames != nil {
			if vv, ok := v.(*LogicVariable); ok && unsortedVarNames[vv.Name] {
				continue
			}
		}
		s := v.NodeSort()
		if ContainsTopSort(s) || IsPolymorphic(s) {
			return &IvyError{
				Msg: fmt.Sprintf("cannot infer sort of %s in %s", v, term),
			}
		}
	}
	usedSyms := UsedSymbolsAst(term)
	for _, sym := range usedSyms.All() {
		name := ExprName(sym)
		if unsortedVarNames != nil && unsortedVarNames[name] {
			continue
		}
		s := sym.NodeSort()
		if ContainsTopSort(s) || IsPolymorphic(s) {
			return &IvyError{
				Msg: fmt.Sprintf("cannot infer sort of %s in %s", sym, term),
			}
		}
	}
	return nil
}

// AllConcretelySorted trivially returns nil.
// Matches Python ivy_logic.py:1165-1166: def all_concretely_sorted(*terms): return True
func AllConcretelySorted(terms ...Expr) error {
	return nil
}
