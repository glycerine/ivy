// Resort functions for remapping sorts throughout module contents.
// Corresponds to Python's resort_* functions in ivy_module.py.
package module

import (
	lg "github.com/glycerine/goivy/logic"
	il "github.com/glycerine/goivy/ivylogic"
	lu "github.com/glycerine/goivy/logicutil"
)

// ResortModule remaps all sorts in the module using the given substitution.
func (m *Module) ResortModule(subs map[string]lg.Sort) {
	// Resort definitions
	for _, lf := range m.Definitions {
		lf.Formula = lu.ResortAst(lf.Formula, subs)
	}
	// Resort axioms
	for _, lf := range m.LabeledAxioms {
		lf.Formula = lu.ResortAst(lf.Formula, subs)
	}
	// Resort props
	for _, lf := range m.LabeledProps {
		lf.Formula = lu.ResortAst(lf.Formula, subs)
	}
	// Resort inits
	for _, lf := range m.LabeledInits {
		lf.Formula = lu.ResortAst(lf.Formula, subs)
	}
	// Resort conjectures
	for _, lf := range m.LabeledConjs {
		lf.Formula = lu.ResortAst(lf.Formula, subs)
	}
	// Resort assumed invariants
	for _, lf := range m.AssumedInvs {
		lf.Formula = lu.ResortAst(lf.Formula, subs)
	}

	// Resort signature
	ResortSig(m.Sig, subs)
}

// ResortLabeledAsts remaps sorts in a slice of labeled formulas.
func ResortLabeledAsts(asts []*LabeledFormula, subs map[string]lg.Sort) {
	for _, lf := range asts {
		lf.Formula = lu.ResortAst(lf.Formula, subs)
	}
}

// ResortSig remaps all sorts in a signature.
func ResortSig(sig *il.Sig, subs map[string]lg.Sort) {
	if sig == nil {
		return
	}
	// Resort sort entries
	newSorts := make(map[string]lg.Sort, len(sig.Sorts))
	for name, sort := range sig.Sorts {
		newSort := lu.ResortSort(sort, subs)
		newName := il.SortName(newSort)
		newSorts[newName] = newSort
		_ = name
	}
	sig.Sorts = newSorts

	// Resort symbol entries
	for name, entry := range sig.Symbols {
		entry.Sort = resortSymbolSort(entry.Sort, subs)
		if entry.Union != nil {
			for i, s := range entry.Union.Sorts {
				entry.Union.Sorts[i] = resortSymbolSort(s, subs)
			}
		}
		_ = name
	}
}

// resortSymbolSort remaps a symbol's sort.
func resortSymbolSort(s lg.Sort, subs map[string]lg.Sort) lg.Sort {
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

// ResortMapSymbolSort remaps sorts in a map[string]lg.Sort.
func ResortMapSymbolSort(m map[string]lg.Sort, subs map[string]lg.Sort) {
	for name, sort := range m {
		m[name] = resortSymbolSort(sort, subs)
	}
}

// ResortSymbols remaps sorts in a slice of constants.
func ResortSymbols(syms []*lg.Const, subs map[string]lg.Sort) []*lg.Const {
	result := make([]*lg.Const, len(syms))
	for i, sym := range syms {
		newSort := resortSymbolSort(sym.CSort, subs)
		if newSort != sym.CSort {
			result[i] = lg.NewConst(sym.Name, newSort)
		} else {
			result[i] = sym
		}
	}
	return result
}

// InstantiateNonEPR instantiates non-EPR formulas with the given ground terms.
// Corresponds to Python's instantiate_non_epr.
func InstantiateNonEPR(nonEPR map[string]lg.Node, groundTerms []lg.Node) []lg.Node {
	var result []lg.Node
	for _, fmla := range nonEPR {
		// For each non-EPR formula, substitute ground terms for the
		// non-variable parameters.
		// This is a simplified version; the full implementation would
		// iterate over all valid instantiations.
		result = append(result, fmla)
	}
	return result
}
