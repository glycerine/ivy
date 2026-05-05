package goivy

// MineConstants collects constants to use for quantifier instantiation.
// It looks at the invariant and module parameters, collecting non-function-sort
// symbols that are either Skolem or module parameters.
//
// Python: ivy_mc.py:778-784
func MineConstants(mod *Module, trans *Clauses, invariant Expr) *InsMap[string, []*Const] {
	res := NewInsMap[string, []*Const]()

	// Collect symbols from invariant and module params
	var fmlas []Expr
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
		syms := UsedSymbolsAST(fmla)
		for _, symExpr := range syms.All() {
			sym, ok := symExpr.(*Const)
			if !ok {
				continue
			}
			if seen[sym.Name] {
				continue
			}
			if IsFunctionSort(sym.CSort) {
				continue
			}
			if IsSkolem(sym.Name) || paramSet[sym.Name] {
				seen[sym.Name] = true
				sortKey := sortKeyStr(sym.CSort)
				res.Set(sortKey, append(res.Get(sortKey), sym))
			}
		}
	}
	return res
}

// MineConstants2 collects a broader set of constants for quantifier instantiation.
// It looks at all symbols in the invariant and transition relation.
//
// Python: ivy_mc.py:786-794
func MineConstants2(mod *Module, trans *Clauses, invariant Expr) *InsMap[string, []*Const] {
	res := NewInsMap[string, []*Const]()
	seen := make(map[string]bool)

	// Collect symbols from invariant
	syms := UsedSymbolsAST(invariant)
	for _, symExpr := range syms.All() {
		sym, ok := symExpr.(*Const)
		if !ok {
			continue
		}
		if seen[sym.Name] {
			continue
		}
		if !IsFunctionSort(sym.CSort) {
			seen[sym.Name] = true
			sortKey := sortKeyStr(sym.CSort)
			res.Set(sortKey, append(res.Get(sortKey), sym))
		}
	}

	// Collect symbols from transition clauses
	if trans != nil {
		for _, symExpr := range SymbolsClauses(trans) {
			sym, ok := symExpr.(*Const)
			if !ok {
				continue
			}
			if seen[sym.Name] {
				continue
			}
			if !IsFunctionSort(sym.CSort) {
				seen[sym.Name] = true
				sortKey := sortKeyStr(sym.CSort)
				res.Set(sortKey, append(res.Get(sortKey), sym))
			}
		}
	}
	return res
}

// sortKeyStr returns a string key for a sort, used as map key for sort→constants.
func sortKeyStr(s Sort) string {
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
func PrevExpr(stVarSet map[string]bool, expr Expr, sortConstants *InsMap[string, []*Const]) Expr {
	symsMap := UsedSymbolsAST(expr)

	// Check: expression must not contain current-state vars or non-constant Skolems
	for _, symExpr := range symsMap.All() {
		sym, ok := symExpr.(*Const)
		if !ok {
			continue
		}
		if stVarSet[sym.Name] {
			return nil
		}
		if IsSkolem(sym.Name) {
			sortKey := sortKeyStr(sym.CSort)
			inConstants := false
			for _, sc := range sortConstants.Get(sortKey) {
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
	var newSyms []*Const
	for _, symExpr := range symsMap.All() {
		if sym, ok := symExpr.(*Const); ok && IsNew(sym.Name) {
			newSyms = append(newSyms, sym)
		}
	}

	if len(newSyms) == 0 {
		return nil
	}

	// Build renaming: new_X → X
	renaming := make(map[NodeKey]*Const, len(newSyms))
	for _, sym := range newSyms {
		oldName := NewOf(sym.Name)
		renaming[Key(sym)] = NewConst(oldName, sym.CSort)
	}

	return RenameAST(expr, renaming)
}
