// Conversion layer between ivylogic.Literal (lg.Expr atoms) and
// unitres.Literal (resolution.Atom). Needed because Go has two
// separate Literal types; Python has only one.

package goivy

// ivyLitToUnitResLit converts an ivylogic.Literal to a unitres.Literal.
// Records the original *lg.Const in symMap so the reverse conversion
// can reconstruct the full AST.
func ivyLitToUnitResLit(lit *Literal, symMap map[string]*Const) *UnitResLiteral {
	var atom *ResolutionAtom
	switch a := lit.Atom.(type) {
	case *Apply:
		atom = AtomFromApply(a)
		if atom != nil {
			if sym, ok := a.Func.(*Const); ok {
				if _, exists := symMap[sym.Name]; !exists {
					symMap[sym.Name] = sym
				}
			}
		}
	case *Eq:
		atom = NewAtom("=", a.T1, a.T2)
	case *Const:
		atom = NewAtom(a.Name)
		if _, exists := symMap[a.Name]; !exists {
			symMap[a.Name] = a
		}
	default:
		// Fallback: treat as nullary with string name
		atom = NewAtom(lit.Atom.String())
	}
	if atom == nil {
		atom = NewAtom(lit.Atom.String())
	}
	return NewUnitResLiteral(lit.Polarity, atom)
}

// unitResLitToIvyLit converts a unitres.Literal back to an ivylogic.Literal.
// Uses symMap to reconstruct the *lg.Const for non-equality atoms.
func unitResLitToIvyLit(lit *UnitResLiteral, symMap map[string]*Const) *Literal {
	var atom Expr
	if lit.Atom.RelName == "=" && len(lit.Atom.Args) == 2 {
		atom = &Eq{T1: lit.Atom.Args[0], T2: lit.Atom.Args[1]}
	} else if sym, ok := symMap[lit.Atom.RelName]; ok {
		if len(lit.Atom.Args) == 0 {
			atom = sym
		} else {
			atom = MustApply(sym, lit.Atom.Args...)
		}
	} else {
		// Fallback: create a boolean symbol
		sym := NewConst(lit.Atom.RelName, Boolean)
		if len(lit.Atom.Args) == 0 {
			atom = sym
		} else {
			atom = MustApply(sym, lit.Atom.Args...)
		}
	}
	return NewLiteral(lit.Polarity, atom)
}

// ivyLitsToUnitResClauses converts a CNF clause set from ivylogic
// literal lists to unitres literal lists, building a symbol map for
// reverse conversion.
func ivyLitsToUnitResClauses(cnf [][]*Literal) ([][]*UnitResLiteral, map[string]*Const) {
	symMap := make(map[string]*Const)
	result := make([][]*UnitResLiteral, len(cnf))
	for i, clause := range cnf {
		urClause := make([]*UnitResLiteral, len(clause))
		for j, lit := range clause {
			urClause[j] = ivyLitToUnitResLit(lit, symMap)
		}
		result[i] = urClause
	}
	return result, symMap
}

// extractUnitResResults extracts the propagation results from a UnitRes
// engine. Returns [[l] for l in r.UnitQueue] + r.Clauses converted to
// ivylogic literal lists, matching Python's clauses_case output assembly.
func extractUnitResResults(r *UnitRes, symMap map[string]*Const) [][]*Literal {
	var result [][]*Literal

	// Unit queue: each unit literal becomes a single-literal clause
	for _, ulit := range r.UnitQueue {
		ivyLit := unitResLitToIvyLit(ulit, symMap)
		result = append(result, []*Literal{ivyLit})
	}

	// Remaining multi-literal clauses
	for _, ucl := range r.Clauses {
		litClause := make([]*Literal, len(ucl))
		for j, ulit := range ucl {
			litClause[j] = unitResLitToIvyLit(ulit, symMap)
		}
		result = append(result, litClause)
	}

	return result
}
