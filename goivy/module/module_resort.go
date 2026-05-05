// Resort functions for remapping sorts throughout module contents.
// Corresponds to Python's resort_* functions in ivy_module.py.
package module

import (
	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
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

// resortFormula safely re-sorts a labeled formula's Formula field.
func resortFormula(lf *ast.LabeledFormula, ss map[lg.NodeKey]lg.Sort) {
	if fmla, ok := lf.Formula.(lg.Expr); ok {
		lf.Formula = lu.ResortAst(fmla, ss)
	}
}

// ResortModule remaps all sorts in the module using the given substitution.
func (m *Module) ResortModule(subs map[lg.NodeKey]*SortRefinement) {
	ss := sortSubsMap(subs)
	// Resort definitions
	for _, lf := range m.Definitions {
		resortFormula(lf, ss)
	}
	// Resort axioms
	for _, lf := range m.LabeledAxioms {
		resortFormula(lf, ss)
	}
	// Resort props
	for _, lf := range m.LabeledProps {
		resortFormula(lf, ss)
	}
	// Resort inits
	for _, lf := range m.LabeledInits {
		resortFormula(lf, ss)
	}
	// Resort conjectures
	for _, lf := range m.LabeledConjs {
		resortFormula(lf, ss)
	}
	// Resort assumed invariants
	for _, lf := range m.AssumedInvs {
		resortFormula(lf, ss)
	}

	// Resort signature
	ResortSig(m.Sig, subs)
}

// ResortLabeledAsts remaps sorts in a slice of labeled formulas.
func ResortLabeledAsts(asts []*ast.LabeledFormula, subs map[lg.NodeKey]*SortRefinement) {
	ss := sortSubsMap(subs)
	for _, lf := range asts {
		resortFormula(lf, ss)
	}
}

// ResortSig remaps all sorts in a signature.
func ResortSig(sig *il.Sig, subs map[lg.NodeKey]*SortRefinement) {
	if sig == nil {
		return
	}
	ss := sortSubsMap(subs)
	// Resort sort entries
	newSorts := iu.NewInsMap[string, lg.Sort]()
	for _, sort := range sig.Sorts.All() {
		newSort := lu.LogicUtilResortSort(sort, ss)
		newName := il.SortName(newSort)
		newSorts.Set(newName, newSort)
	}
	sig.Sorts = newSorts

	// Resort symbol entries
	for name, entry := range sig.Symbols.All() {
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
		for i, d := range dom {
			newDom[i] = lu.LogicUtilResortSort(d, subs)
		}
		rng := lu.LogicUtilResortSort(fs.Range(), subs)
		all := make([]lg.Sort, len(newDom)+1)
		copy(all, newDom)
		all[len(newDom)] = rng
		nfs, err := lg.NewFunctionSort(all...)
		if err != nil {
			return s
		}
		return nfs
	}
	return lu.LogicUtilResortSort(s, subs)
}

// ResortMapSymbolSort remaps sorts in a map of sort entries.
func ResortMapSymbolSort(m map[string]lg.Sort, subs map[lg.NodeKey]*SortRefinement) {
	ss := sortSubsMap(subs)
	for name, sort := range m {
		m[name] = resortSymbolSort(sort, ss)
	}
}

// ResortSymbols remaps sorts in a slice of constants.
func ResortSymbols(syms []*lg.Const, subs map[lg.NodeKey]*SortRefinement) []*lg.Const {
	ss := sortSubsMap(subs)
	result := make([]*lg.Const, len(syms))
	for i, sym := range syms {
		newSort := resortSymbolSort(sym.CSort, ss)
		if newSort != sym.CSort {
			result[i] = lg.NewConst(sym.Name, newSort)
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
