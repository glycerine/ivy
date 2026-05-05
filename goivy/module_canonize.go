// Type canonization for Module.
//
// This corresponds to the resort/canonize operations in Python's ivy_module.py:
// canonize_types, resort_ast, resort_labeled_asts, resort_symbols, etc.
package goivy

// CanonizeTypes removes implemented (refined) types from the module,
// replacing them with their refinements throughout all module formulas.
//
// Corresponds to Python's Module.canonize_types.
// SortRefinement maps an old sort to its replacement sort, keyed by structural identity.
type SortRefinement struct {
	Old Sort
	New Sort
}

func (m *Module) CanonizeTypes(sortRefinements []SortRefinement) {
	if len(sortRefinements) == 0 {
		return
	}

	// Build a SortKey-based lookup matching Python's structural equality on sorts.
	// Value is full SortRefinement so we always have access to both Old and New.
	rn := make(map[NodeKey]*SortRefinement, len(sortRefinements))
	for i := range sortRefinements {
		sr := &sortRefinements[i]
		rn[SortKey(sr.Old)] = sr
	}

	m.Definitions = resortLabeledFormulas(m.Definitions, rn)
	m.LabeledAxioms = resortLabeledFormulas(m.LabeledAxioms, rn)
	m.LabeledProps = resortLabeledFormulas(m.LabeledProps, rn)
	m.LabeledInits = resortLabeledFormulas(m.LabeledInits, rn)
	m.LabeledConjs = resortLabeledFormulas(m.LabeledConjs, rn)
	m.Assertions = resortLabeledFormulas(m.Assertions, rn)

	// Resort symbol lists.
	m.Params = resortSymbols(m.Params, rn)
	m.SymbolOrder = resortSymbols(m.SymbolOrder, rn)

	// Remove refined sort names from sets and lists.
	m.GhostSorts = removeRefinedSortNames(m.GhostSorts, rn)
	m.SortOrder = removeRefinedSortNamesList(m.SortOrder, rn)

	// Resort initializers.
	m.Initializers = resortNamedActions(m.Initializers, rn)

	// Resort aliases map.
	m.Aliases = resortAliases(m.Aliases, rn)

	// Resort ext_preconds map.
	m.ExtPreconds = resortMapAST(m.ExtPreconds, rn)

	// Resort InitCond (Python line 250: self.init_cond = resort_clauses(self.init_cond))
	m.InitCond = ResortClauses(m.InitCond, rn)

	// Resort ConceptSpaces (Python line 251: self.concept_spaces = resort_concept_spaces(self.concept_spaces))
	for i, cs := range m.ConceptSpaces {
		m.ConceptSpaces[i] = ConceptSpace{
			Label: ResortAST(cs.Label, rn),
			Body:  ResortAST(cs.Body, rn),
		}
	}

	// Resort Progress (Python line 254: self.progress = resort_asts(self.progress))
	for i, p := range m.Progress {
		if expr, ok := p.(Expr); ok {
			m.Progress[i] = ResortAST(expr, rn)
		}
	}

	// Resort BeforeExport (Python line 261: self.before_export = resort_map_any_ast(self.before_export))
	newBE := NewInsMap[string, Action]()
	for k, v := range m.BeforeExport.All() {
		if expr, ok := v.(Expr); ok {
			if resorted := ResortAST(expr, rn); resorted != nil {
				if act, ok2 := resorted.(Action); ok2 {
					newBE.Set(k, act)
					continue
				}
			}
		}
		newBE.Set(k, v)
	}
	m.BeforeExport = newBE

	// Resort signature (Python line 263: lu.resort_sig(sort_refinement))
	ResortSig(m.Sig, rn)
}

// ComputeSortRefinements computes sort refinements from the signature.
// Corresponds to Python's il.sort_refinement() (ivy_logic.py:1471).
func ComputeSortRefinements(sig *Sig) []SortRefinement {
	if sig == nil {
		return nil
	}
	raw := GetSortRefinement(sig)
	if len(raw) == 0 {
		return nil
	}
	result := make([]SortRefinement, 0, len(raw))
	for _, s := range sig.Sorts.All() {
		key := SortKey(s)
		if newSort, ok := raw[key]; ok {
			result = append(result, SortRefinement{Old: s, New: newSort})
		}
	}
	return result
}

// --- resort helpers ---

// ResortAST applies sort refinement to an AST node, replacing old sorts
// with their refinements throughout the tree.
func ResortAST(node Expr, rn map[NodeKey]*SortRefinement) Expr {
	if node == nil {
		return nil
	}
	return resortASTRec(node, rn)
}

func resortASTRec(node Expr, rn map[NodeKey]*SortRefinement) Expr {
	switch t := node.(type) {
	case *Variable:
		newSort := ResortSort(t.VSort, rn)
		if SortEqual(newSort, t.VSort) {
			return node
		}
		v, err := NewVariable(t.Name, newSort)
		if err != nil {
			return node
		}
		return v

	case *Const:
		newSort := ResortSort(t.CSort, rn)
		if SortEqual(newSort, t.CSort) {
			return node
		}
		return NewConst(t.Name, newSort)

	case *Apply:
		newFunc := resortASTRec(t.Func, rn)
		newTerms := make([]Expr, len(t.Terms))
		for i, arg := range t.Terms {
			newTerms[i] = resortASTRec(arg, rn)
		}
		return MustApply(newFunc, newTerms...)

	case *ForAll:
		newVars := resortVars(t.Variables, rn)
		newBody := resortASTRec(t.Body, rn)
		return &ForAll{Variables: newVars, Body: newBody}

	case *Exists:
		newVars := resortVars(t.Variables, rn)
		newBody := resortASTRec(t.Body, rn)
		return &Exists{Variables: newVars, Body: newBody}

	case *Lambda:
		newVars := resortVars(t.Variables, rn)
		newBody := resortASTRec(t.Body, rn)
		return &Lambda{Variables: newVars, Body: newBody}

	case *NamedBinder:
		newVars := resortVars(t.Variables, rn)
		newBody := resortASTRec(t.Body, rn)
		return &NamedBinder{Name: t.Name, Variables: newVars, Environ: t.Environ, Body: newBody}

	case *IvyDefinition:
		newLhs := resortASTRec(t.Lhs, rn)
		newRhs := resortASTRec(t.Rhs, rn)
		return NewIvyDefinition(newLhs, newRhs)
	}

	// For other node types, recurse into children.
	children := node.Children()
	if len(children) == 0 {
		return node
	}
	newChildren := make([]Expr, len(children))
	for i, c := range children {
		newChildren[i] = resortASTRec(c, rn)
	}
	return CloneNode(node, newChildren)
}

// ResortSort applies sort refinement to a sort.
func ResortSort(s Sort, rn map[NodeKey]*SortRefinement) Sort {
	if s == nil {
		return nil
	}

	key := SortKey(s)
	if sr, ok := rn[key]; ok {
		return sr.New
	}

	if fs, ok := s.(*FunctionSort); ok {
		newSorts := make([]Sort, len(fs.Sorts))
		for i, sub := range fs.Sorts {
			newSorts[i] = ResortSort(sub, rn)
		}
		result, err := NewFunctionSort(newSorts...)
		if err != nil {
			return s
		}
		return result
	}

	return s
}

// ResortSymbol applies sort refinement to a symbol's sort.
func ResortSymbol(c *Const, rn map[NodeKey]*SortRefinement) *Const {
	newSort := ResortSort(c.CSort, rn)
	if SortEqual(newSort, c.CSort) {
		return c
	}
	return NewConst(c.Name, newSort)
}

// resortVars applies sort refinement to a slice of variables.
func resortVars(vars []*Variable, rn map[NodeKey]*SortRefinement) []*Variable {
	result := make([]*Variable, len(vars))
	for i, v := range vars {
		newSort := ResortSort(v.VSort, rn)
		if SortEqual(newSort, v.VSort) {
			result[i] = v
		} else {
			nv, err := NewVariable(v.Name, newSort)
			if err != nil {
				result[i] = v
			} else {
				result[i] = nv
			}
		}
	}
	return result
}

// resortLabeledFormulas applies sort refinement to a slice of labeled
// formulas, returning a new slice with resorted formulas.
func resortLabeledFormulas(lfs []*LabeledFormula, rn map[NodeKey]*SortRefinement) []*LabeledFormula {
	if len(lfs) == 0 {
		return lfs
	}
	result := make([]*LabeledFormula, len(lfs))
	for i, lf := range lfs {
		newFormula := ResortAST(lf.Formula.(Expr), rn)
		// Python: ast.clone([ast.args[0], lu.resort_ast(ast.args[1], sort_refinement)])
		cloned := lf.Clone([]Node{lf.Label, newFormula})
		result[i] = cloned.(*LabeledFormula)
	}
	return result
}

// resortSymbols applies sort refinement to a slice of constant symbols.
func resortSymbols(syms []*Const, rn map[NodeKey]*SortRefinement) []*Const {
	if len(syms) == 0 {
		return syms
	}
	result := make([]*Const, len(syms))
	for i, s := range syms {
		result[i] = ResortSymbol(s, rn)
	}
	return result
}

// removeRefinedSortNames removes sort names that are in the refinement map
// from a set of sort names. Probes the SortKey-keyed rn map by constructing
// an UninterpretedSort with the candidate name — matching how Python's
// sort_refinement keys are simple named sorts from sig.sorts.
func removeRefinedSortNames(sorts map[string]bool, rn map[NodeKey]*SortRefinement) map[string]bool {
	result := make(map[string]bool, len(sorts))
	for name, val := range sorts {
		probe := &UninterpretedSort{Name: name}
		if _, refined := rn[SortKey(probe)]; !refined {
			result[name] = val
		}
	}
	return result
}

// removeRefinedSortNamesList removes sort names that are in the refinement
// map from an ordered list.
func removeRefinedSortNamesList(sorts []string, rn map[NodeKey]*SortRefinement) []string {
	var result []string
	for _, name := range sorts {
		probe := &UninterpretedSort{Name: name}
		if _, refined := rn[SortKey(probe)]; !refined {
			result = append(result, name)
		}
	}
	return result
}

// resortNamedActions applies sort refinement to named action pairs.
// Since the action is interface{}, we only resort it if it implements lg.Expr.
func resortNamedActions(pairs []NamedAction, rn map[NodeKey]*SortRefinement) []NamedAction {
	if len(pairs) == 0 {
		return pairs
	}
	result := make([]NamedAction, len(pairs))
	for i, p := range pairs {
		result[i] = p
		if node, ok := p.Action.(Expr); ok {
			resorted := ResortAST(node, rn)
			if act, ok2 := resorted.(Action); ok2 {
				result[i].Action = act
			}
		}
	}
	return result
}

// resortAliases applies sort refinement to an aliases map.
func resortAliases(aliases map[string]string, rn map[NodeKey]*SortRefinement) map[string]string {
	result := make(map[string]string, len(aliases))
	for k, v := range aliases {
		result[k] = v
	}
	for _, sr := range rn {
		result[IvySortName(sr.Old)] = IvySortName(sr.New)
	}
	return result
}

// resortMapAST applies sort refinement to a map of name -> lg.Expr.
func resortMapAST(m map[string]Expr, rn map[NodeKey]*SortRefinement) map[string]Expr {
	result := make(map[string]Expr, len(m))
	for k, v := range m {
		result[k] = ResortAST(v, rn)
	}
	return result
}

// ResortClauses applies sort refinement to a Clauses object,
// transforming all formulas and definitions within it.
// Corresponds to Python's resort_clauses (via apply_func_to_clauses(resort_ast)).
func ResortClauses(cls *Clauses, rn map[NodeKey]*SortRefinement) *Clauses {
	if cls == nil {
		return nil
	}
	newFmlas := make([]Expr, len(cls.Fmlas))
	for i, f := range cls.Fmlas {
		newFmlas[i] = ResortAST(f, rn)
	}
	newDefs := make([]*IvyDefinition, len(cls.Defs))
	for i, d := range cls.Defs {
		nd := ResortAST(d, rn)
		if def, ok := nd.(*IvyDefinition); ok {
			newDefs[i] = def
		} else {
			newDefs[i] = d
		}
	}
	return NewClauses(newFmlas, newDefs, cls.Annot)
}

// ResortAsts applies sort refinement to a slice of AST nodes.
// Corresponds to Python's resort_asts.
func ResortAsts(asts []Expr, rn map[NodeKey]*SortRefinement) []Expr {
	if len(asts) == 0 {
		return asts
	}
	result := make([]Expr, len(asts))
	for i, a := range asts {
		result[i] = ResortAST(a, rn)
	}
	return result
}

// ResortMapAnyAST applies sort refinement to values in a map of name -> lg.Expr.
// Corresponds to Python's resort_map_any_ast. This is the public version of resortMapAST.
func ResortMapAnyAST(m map[string]Expr, rn map[NodeKey]*SortRefinement) map[string]Expr {
	return resortMapAST(m, rn)
}

// RemoveRefinedSortnamesFromSet removes sort names that are in the refinement map
// from a set of sort names. Public wrapper for removeRefinedSortNames.
// Corresponds to Python's remove_refined_sortnames_from_set.
func RemoveRefinedSortnamesFromSet(sorts map[string]bool, rn map[NodeKey]*SortRefinement) map[string]bool {
	return removeRefinedSortNames(sorts, rn)
}

// RemoveRefinedSortnamesFromList removes sort names that are in the refinement
// map from an ordered list. Public wrapper for removeRefinedSortNamesList.
// Corresponds to Python's remove_refined_sortnames_from_list.
func RemoveRefinedSortnamesFromList(sorts []string, rn map[NodeKey]*SortRefinement) []string {
	return removeRefinedSortNamesList(sorts, rn)
}
