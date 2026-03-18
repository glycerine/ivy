// Resort functions for remapping sorts throughout module contents.
// Corresponds to Python's resort_* functions in ivy_module.py.
package module

import (
	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
)

// sortSubsMap extracts the SortKey→NewSort map from a SortRefinement map,
// suitable for passing to logicutil.ResortSort / logicutil.ResortAst.
func sortSubsMap(rn map[lg.NodeKey]*SortRefinement) map[lg.NodeKey]lg.Sort {
	m := make(map[lg.NodeKey]lg.Sort, len(rn))
	for k, sr := range rn {
		m[k] = sr.New
	}
	return m
}

// ResortModule remaps all sorts in the module using the given substitution.
func (m *Module) ResortModule(subs map[lg.NodeKey]*SortRefinement) {
	ss := sortSubsMap(subs)
	// Resort definitions
	for _, lf := range m.Definitions {
		lf.Formula = lu.ResortAst(lf.Formula.(lg.Expr), ss).(ast.Node)
	}
	// Resort axioms
	for _, lf := range m.LabeledAxioms {
		lf.Formula = lu.ResortAst(lf.Formula.(lg.Expr), ss).(ast.Node)
	}
	// Resort props
	for _, lf := range m.LabeledProps {
		lf.Formula = lu.ResortAst(lf.Formula.(lg.Expr), ss).(ast.Node)
	}
	// Resort inits
	for _, lf := range m.LabeledInits {
		lf.Formula = lu.ResortAst(lf.Formula.(lg.Expr), ss).(ast.Node)
	}
	// Resort conjectures
	for _, lf := range m.LabeledConjs {
		lf.Formula = lu.ResortAst(lf.Formula.(lg.Expr), ss).(ast.Node)
	}
	// Resort assumed invariants
	for _, lf := range m.AssumedInvs {
		lf.Formula = lu.ResortAst(lf.Formula.(lg.Expr), ss).(ast.Node)
	}

	// Resort signature
	ResortSig(m.Sig, subs)
}

// ResortLabeledAsts remaps sorts in a slice of labeled formulas.
func ResortLabeledAsts(asts []*ast.LabeledFormula, subs map[lg.NodeKey]*SortRefinement) {
	ss := sortSubsMap(subs)
	for _, lf := range asts {
		lf.Formula = lu.ResortAst(lf.Formula.(lg.Expr), ss).(ast.Node)
	}
}

// ResortSig remaps all sorts in a signature.
func ResortSig(sig *il.Sig, subs map[lg.NodeKey]*SortRefinement) {
	if sig == nil {
		return
	}
	ss := sortSubsMap(subs)
	// Resort sort entries
	newSorts := make(map[string]lg.Sort, len(sig.Sorts))
	for _, sort := range sig.Sorts {
		newSort := lu.ResortSort(sort, ss)
		newName := il.SortName(newSort)
		newSorts[newName] = newSort
	}
	sig.Sorts = newSorts

	// Resort symbol entries
	for name, entry := range sig.Symbols {
		entry.Sort = resortSymbolSort(entry.Sort, ss)
		if entry.Union != nil {
			for i, s := range entry.Union.Sorts {
				entry.Union.Sorts[i] = resortSymbolSort(s, ss)
			}
		}
		_ = name
	}
}

// resortSymbolSort remaps a symbol's sort.
func resortSymbolSort(s lg.Sort, subs map[lg.NodeKey]lg.Sort) lg.Sort {
	if fs, ok := s.(*lg.FunctionSort); ok {
		dom := fs.Domain()
		newDom := make([]lg.Sort, len(dom))
		changed := false
		for i, d := range dom {
			nd := lu.ResortSort(d, subs)
			newDom[i] = nd
			if nd != d {
				changed = true
			}
		}
		rng := lu.ResortSort(fs.Range(), subs)
		if rng != fs.Range() {
			changed = true
		}
		if !changed {
			return s
		}
		all := make([]lg.Sort, len(newDom)+1)
		copy(all, newDom)
		all[len(newDom)] = rng
		nfs, err := lg.NewFunctionSort(all...)
		if err != nil {
			return s
		}
		return nfs
	}
	return lu.ResortSort(s, subs)
}

// ResortMapSymbolSort remaps sorts in a map of sort entries.
func ResortMapSymbolSort(m map[string]lg.Sort, subs map[lg.NodeKey]*SortRefinement) {
	ss := sortSubsMap(subs)
	for name, sort := range m {
		m[name] = resortSymbolSort(sort, ss)
	}
}

// ResortSymbols remaps sorts in a slice of constants.
func ResortSymbols(syms []*lg.Symbol, subs map[lg.NodeKey]*SortRefinement) []*lg.Symbol {
	ss := sortSubsMap(subs)
	result := make([]*lg.Symbol, len(syms))
	for i, sym := range syms {
		newSort := resortSymbolSort(sym.CSort, ss)
		if newSort != sym.CSort {
			result[i] = lg.NewSymbol(sym.Name, newSort)
		} else {
			result[i] = sym
		}
	}
	return result
}

// ResortNameAstPairs remaps sorts in a slice of (name, ast) pairs.
// Corresponds to Python's resort_name_ast_pairs.
func ResortNameAstPairs(pairs []NameAstPair, subs map[lg.NodeKey]*SortRefinement) []NameAstPair {
	ss := sortSubsMap(subs)
	result := make([]NameAstPair, len(pairs))
	for i, p := range pairs {
		result[i] = NameAstPair{Name: p.Name, Ast: lu.ResortAst(p.Ast, ss)}
	}
	return result
}

// NameAstPair holds a (name, AST node) pair used in resort_name_ast_pairs.
type NameAstPair struct {
	Name string
	Ast  lg.Expr
}

// ResortAliasesMap remaps sort aliases according to the sort refinement.
// For each (s1 -> s2) in subs, adds s1.name -> s2.name to the alias map.
// Corresponds to Python's resort_aliases_map.
func ResortAliasesMap(amap map[string]string, subs map[lg.NodeKey]*SortRefinement) map[string]string {
	result := make(map[string]string, len(amap)+len(subs))
	for k, v := range amap {
		result[k] = v
	}
	for _, sr := range subs {
		result[il.SortName(sr.Old)] = il.SortName(sr.New)
	}
	return result
}

// InstantiateNonEPR instantiates non-EPR formulas with the given ground terms.
// Corresponds to Python's instantiate_non_epr.
func InstantiateNonEPR(nonEPR map[string]lg.Expr, groundTerms []lg.Expr) []lg.Expr {
	var result []lg.Expr
	for _, fmla := range nonEPR {
		// For each non-EPR formula, substitute ground terms for the
		// non-variable parameters.
		// This is a simplified version; the full implementation would
		// iterate over all valid instantiations.
		result = append(result, fmla)
	}
	return result
}
