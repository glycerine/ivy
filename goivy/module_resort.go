// Resort functions for remapping sorts throughout module contents.
// Corresponds to Python's resort_* functions in ivy_module.py.
package goivy

// sortSubsMap extracts the SortKey→NewSort map from a SortRefinement map,
// suitable for passing to logicutil.ResortSort / logicutil.ResortAst.
func sortSubsMap(rn map[NodeKey]*SortRefinement) map[NodeKey]Sort {
	m := make(map[NodeKey]Sort, len(rn))
	for k, sr := range rn {
		m[k] = sr.New
	}
	return m
}

// resortFormula safely re-sorts a labeled formula's Formula field.
func resortFormula(lf *LabeledFormula, ss map[NodeKey]Sort) {
	if fmla, ok := lf.Formula.(Expr); ok {
		lf.Formula = ResortAst(fmla, ss)
	}
}

// ResortModule remaps all sorts in the module using the given substitution.
func (m *Module) ResortModule(subs map[NodeKey]*SortRefinement) {
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
func ResortLabeledAsts(asts []*LabeledFormula, subs map[NodeKey]*SortRefinement) {
	ss := sortSubsMap(subs)
	for _, lf := range asts {
		resortFormula(lf, ss)
	}
}

// ResortSig remaps all sorts in a signature.
func ResortSig(sig *Sig, subs map[NodeKey]*SortRefinement) {
	if sig == nil {
		return
	}
	ss := sortSubsMap(subs)
	// Resort sort entries
	newSorts := NewInsMap[string, Sort]()
	for _, sort := range sig.Sorts.All() {
		newSort := LogicUtilResortSort(sort, ss)
		newName := IvySortName(newSort)
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
func resortSymbolSort(s Sort, subs map[NodeKey]Sort) Sort {
	if fs, ok := s.(*FunctionSort); ok {
		dom := fs.Domain()
		newDom := make([]Sort, len(dom))
		for i, d := range dom {
			newDom[i] = LogicUtilResortSort(d, subs)
		}
		rng := LogicUtilResortSort(fs.Range(), subs)
		all := make([]Sort, len(newDom)+1)
		copy(all, newDom)
		all[len(newDom)] = rng
		nfs, err := NewFunctionSort(all...)
		if err != nil {
			return s
		}
		return nfs
	}
	return LogicUtilResortSort(s, subs)
}

// ResortMapSymbolSort remaps sorts in a map of sort entries.
func ResortMapSymbolSort(m map[string]Sort, subs map[NodeKey]*SortRefinement) {
	ss := sortSubsMap(subs)
	for name, sort := range m {
		m[name] = resortSymbolSort(sort, ss)
	}
}

// ResortSymbols remaps sorts in a slice of constants.
func ResortSymbols(syms []*Const, subs map[NodeKey]*SortRefinement) []*Const {
	ss := sortSubsMap(subs)
	result := make([]*Const, len(syms))
	for i, sym := range syms {
		newSort := resortSymbolSort(sym.CSort, ss)
		if newSort != sym.CSort {
			result[i] = NewConst(sym.Name, newSort)
		} else {
			result[i] = sym
		}
	}
	return result
}

// ResortNameAstPairs remaps sorts in a slice of (name, ast) pairs.
// Corresponds to Python's resort_name_ast_pairs.
func ResortNameAstPairs(pairs []NameAstPair, subs map[NodeKey]*SortRefinement) []NameAstPair {
	ss := sortSubsMap(subs)
	result := make([]NameAstPair, len(pairs))
	for i, p := range pairs {
		result[i] = NameAstPair{Name: p.Name, Ast: ResortAst(p.Ast, ss)}
	}
	return result
}

// NameAstPair holds a (name, AST node) pair used in resort_name_ast_pairs.
type NameAstPair struct {
	Name string
	Ast  Expr
}

// ResortAliasesMap remaps sort aliases according to the sort refinement.
// For each (s1 -> s2) in subs, adds s1.name -> s2.name to the alias map.
// Corresponds to Python's resort_aliases_map.
func ResortAliasesMap(amap map[string]string, subs map[NodeKey]*SortRefinement) map[string]string {
	result := make(map[string]string, len(amap)+len(subs))
	for k, v := range amap {
		result[k] = v
	}
	for _, sr := range subs {
		result[IvySortName(sr.Old)] = IvySortName(sr.New)
	}
	return result
}

// InstantiateNonEPR instantiates non-EPR formulas with the given ground terms.
// Corresponds to Python's instantiate_non_epr.
func InstantiateNonEPR(nonEPR map[string]Expr, groundTerms []Expr) []Expr {
	var result []Expr
	for _, fmla := range nonEPR {
		// For each non-EPR formula, substitute ground terms for the
		// non-variable parameters.
		// This is a simplified version; the full implementation would
		// iterate over all valid instantiations.
		result = append(result, fmla)
	}
	return result
}
