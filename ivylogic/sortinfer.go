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
func SortInfer(term lg.Node, sort lg.Sort) (lg.Node, error) {
	res, err := typeinfer.ConcretizeSorts(term, sort)
	if err != nil {
		return nil, err
	}
	if err := CheckConcretelySorted(res); err != nil {
		return nil, err
	}
	return res, nil
}

// SortInferList performs sort inference on a list of terms.
// Corresponds to Python's sort_infer_list.
func SortInferList(terms []lg.Node) ([]lg.Node, error) {
	result := make([]lg.Node, len(terms))
	for i, t := range terms {
		res, err := typeinfer.ConcretizeSorts(t, nil)
		if err != nil {
			return nil, err
		}
		if err := CheckConcretelySorted(res); err != nil {
			return nil, err
		}
		result[i] = res
	}
	return result, nil
}

// Sortify adds sorts to an untyped AST by looking up symbols in the
// current signature context. Recursively processes children.
// Corresponds to Python's sortify.
func Sortify(sig *Sig, node lg.Node) lg.Node {
	args := NodeArgs(node)
	newArgs := make([]lg.Node, len(args))
	for i, arg := range args {
		newArgs[i] = Sortify(sig, arg)
	}

	// If this is an Apply with a TopSort-sorted function, look it up
	if app, ok := node.(*lg.Apply); ok {
		if _, isTop := app.Func.NodeSort().(*lg.TopSort); isTop {
			if c, ok := app.Func.(*lg.Const); ok {
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
// Corresponds to Python's check_concretely_sorted.
func CheckConcretelySorted(term lg.Node) error {
	usedVars := lu.UsedVariables(term)
	for v := range usedVars {
		if lg.ContainsTopSort(v) || lg.IsPolymorphic(v) {
			return &lg.IvyError{
				Msg: fmt.Sprintf("cannot infer sort of %s in %s", v, term),
			}
		}
	}
	usedConsts := lu.UsedConstants(term)
	for c := range usedConsts {
		if lg.ContainsTopSort(c) || lg.IsPolymorphic(c) {
			return &lg.IvyError{
				Msg: fmt.Sprintf("cannot infer sort of %s in %s", c, term),
			}
		}
	}
	return nil
}

// AllConcretelySorted checks that all given terms are concretely sorted.
// Returns nil if all are concretely sorted, or the first error found.
func AllConcretelySorted(terms ...lg.Node) error {
	for _, t := range terms {
		if err := CheckConcretelySorted(t); err != nil {
			return err
		}
	}
	return nil
}
