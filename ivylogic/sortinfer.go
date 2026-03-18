package ivylogic

import (
	"fmt"

	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	"github.com/glycerine/goivy/typeinfer"
)

// SortInfer performs sort inference on a term, optionally constraining it
// to the given sort. Returns the concretized term.
// Corresponds to Python's sort_infer.
func SortInfer(term lg.Expr, sort lg.Sort) (lg.Expr, error) {
	res, err := typeinfer.ConcretizeSorts(term, sort)
	if err != nil {
		return nil, err
	}
	if err := CheckConcretelySorted(res, nil); err != nil {
		return nil, err
	}
	return res, nil
}

// SortInferList performs sort inference on a list of terms.
// Corresponds to Python's sort_infer_list.
func SortInferList(terms []lg.Expr) ([]lg.Expr, error) {
	result := make([]lg.Expr, len(terms))
	for i, t := range terms {
		res, err := typeinfer.ConcretizeSorts(t, nil)
		if err != nil {
			return nil, err
		}
		if err := CheckConcretelySorted(res, nil); err != nil {
			return nil, err
		}
		result[i] = res
	}
	return result, nil
}

// Sortify adds sorts to an untyped AST by looking up symbols in the
// current signature context. Recursively processes children.
// Corresponds to Python's sortify.
func Sortify(sig *Sig, node lg.Expr) lg.Expr {
	args := NodeArgs(node)
	newArgs := make([]lg.Expr, len(args))
	for i, arg := range args {
		newArgs[i] = Sortify(sig, arg)
	}

	// If this is an Apply with a TopSort-sorted function, look it up
	if app, ok := node.(*lg.Apply); ok {
		if _, isTop := app.Func.NodeSort().(*lg.TopSort); isTop {
			if c, ok := app.Func.(*lg.Symbol); ok {
				sym, err := sig.FindSymbol(c.Name, false)
				if err == nil {
					// Reconstruct the apply with the resolved symbol
					result, applyErr := lg.NewApply(sym, newArgs...)
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
// Corresponds to Python's check_concretely_sorted.
func CheckConcretelySorted(term lg.Expr, unsortedVarNames map[string]bool) error {
	usedVars := lu.UsedVariables(term)
	for _, v := range usedVars {
		if unsortedVarNames != nil {
			if vv, ok := v.(*lg.Variable); ok && unsortedVarNames[vv.Name] {
				continue
			}
		}
		if lg.ContainsTopSort(v) || lg.IsPolymorphic(v) {
			return &lg.IvyError{
				Msg: fmt.Sprintf("cannot infer sort of %s in %s", v, term),
			}
		}
	}
	usedConsts := lu.UsedConstants(term)
	for _, c := range usedConsts {
		if unsortedVarNames != nil {
			if cc, ok := c.(*lg.Symbol); ok && unsortedVarNames[cc.Name] {
				continue
			}
		}
		if lg.ContainsTopSort(c) || lg.IsPolymorphic(c) {
			return &lg.IvyError{
				Msg: fmt.Sprintf("cannot infer sort of %s in %s", c, term),
			}
		}
	}
	return nil
}

// AllConcretelySorted trivially returns nil.
// Matches Python ivy_logic.py:1165-1166: def all_concretely_sorted(*terms): return True
func AllConcretelySorted(terms ...lg.Expr) error {
	return nil
}
