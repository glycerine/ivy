package mc

import (
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	tr "github.com/glycerine/goivy/transrel"
)

// MineConstants collects constants to use for quantifier instantiation.
// It looks at the invariant and module parameters, collecting non-function-sort
// symbols that are either Skolem or module parameters.
//
// Python: ivy_mc.py:778-784
func MineConstants(mod *module.Module, trans *co.Clauses, invariant lg.Node) map[string][]*lg.Const {
	res := make(map[string][]*lg.Const)

	// Collect symbols from invariant and module params
	var fmlas []lg.Node
	fmlas = append(fmlas, invariant)
	for _, p := range mod.Params {
		fmlas = append(fmlas, p)
	}

	paramSet := make(map[string]bool)
	for _, p := range mod.Params {
		paramSet[p.Name] = true
	}

	seen := make(map[string]bool)
	for _, fmla := range fmlas {
		syms := co.UsedSymbolsAST(fmla)
		for _, symNode := range syms {
			sym := symNode.(*lg.Const)
			if seen[sym.Name] {
				continue
			}
			if il.IsFunctionSort(sym.CSort) {
				continue
			}
			if tr.IsSkolem(sym.Name) || paramSet[sym.Name] {
				seen[sym.Name] = true
				sortKey := sortKeyStr(sym.CSort)
				res[sortKey] = append(res[sortKey], sym)
			}
		}
	}
	return res
}

// MineConstants2 collects a broader set of constants for quantifier instantiation.
// It looks at all symbols in the invariant and transition relation.
//
// Python: ivy_mc.py:786-794
func MineConstants2(mod *module.Module, trans *co.Clauses, invariant lg.Node) map[string][]*lg.Const {
	res := make(map[string][]*lg.Const)
	seen := make(map[string]bool)

	// Collect symbols from invariant
	syms := co.UsedSymbolsAST(invariant)
	for _, symNode := range syms {
		sym := symNode.(*lg.Const)
		if seen[sym.Name] {
			continue
		}
		if !il.IsFunctionSort(sym.CSort) {
			seen[sym.Name] = true
			sortKey := sortKeyStr(sym.CSort)
			res[sortKey] = append(res[sortKey], sym)
		}
	}

	// Collect symbols from transition clauses
	if trans != nil {
		for _, sym := range co.SymbolsClauses(trans) {
			if seen[sym.Name] {
				continue
			}
			if !il.IsFunctionSort(sym.CSort) {
				seen[sym.Name] = true
				sortKey := sortKeyStr(sym.CSort)
				res[sortKey] = append(res[sortKey], sym)
			}
		}
	}
	return res
}

// sortKeyStr returns a string key for a sort, used as map key for sort→constants.
func sortKeyStr(s lg.Sort) string {
	if s == nil {
		return "<nil>"
	}
	return s.String()
}

// PrevExpr checks if an expression is the "next-state" version of
// an expression involving only state variables. If so, returns the
// "current-state" version. This is used to link abstract variables
// across time steps in the AIGER encoding.
//
// An expression qualifies if:
// - It does NOT contain any current-state variables or Skolems that aren't in sort_constants
// - It DOES contain at least one new_ symbol
// If so, it returns the expression with new_ replaced by current.
//
// Python: ivy_mc.py:802-810
func PrevExpr(stVarSet map[string]bool, expr lg.Node, sortConstants map[string][]*lg.Const) lg.Node {
	symsMap := co.UsedSymbolsAST(expr)

	// Check: expression must not contain current-state vars or non-constant Skolems
	for _, symNode := range symsMap {
		sym := symNode.(*lg.Const)
		if stVarSet[sym.Name] {
			return nil
		}
		if tr.IsSkolem(sym.Name) {
			sortKey := sortKeyStr(sym.CSort)
			inConstants := false
			for _, sc := range sortConstants[sortKey] {
				if sc.Name == sym.Name {
					inConstants = true
					break
				}
			}
			if !inConstants {
				return nil
			}
		}
	}

	// Find new_ symbols
	var newSyms []*lg.Const
	for _, symNode := range symsMap {
		sym := symNode.(*lg.Const)
		if tr.IsNew(sym.Name) {
			newSyms = append(newSyms, sym)
		}
	}

	if len(newSyms) == 0 {
		return nil
	}

	// Build renaming: new_X → X
	renaming := make(map[string]*lg.Const, len(newSyms))
	for _, sym := range newSyms {
		oldName := tr.NewOf(sym.Name)
		renaming[sym.Name] = lg.NewConst(oldName, sym.CSort)
	}

	return co.RenameAST(expr, renaming)
}
