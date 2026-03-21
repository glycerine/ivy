// Conversion layer between ivylogic.Literal (lg.Expr atoms) and
// unitres.Literal (resolution.Atom). Needed because Go has two
// separate Literal types; Python has only one.

package solver

import (
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/resolution"
	"github.com/glycerine/goivy/unitres"
)

// ivyLitToUnitResLit converts an ivylogic.Literal to a unitres.Literal.
// Records the original *lg.Symbol in symMap so the reverse conversion
// can reconstruct the full AST.
func ivyLitToUnitResLit(lit *il.Literal, symMap map[string]*lg.Symbol) *unitres.Literal {
	var atom *resolution.Atom
	switch a := lit.Atom.(type) {
	case *lg.Apply:
		atom = resolution.AtomFromApply(a)
		if atom != nil {
			if sym, ok := a.Func.(*lg.Symbol); ok {
				if _, exists := symMap[sym.Name]; !exists {
					symMap[sym.Name] = sym
				}
			}
		}
	case *lg.Eq:
		atom = resolution.NewAtom("=", a.T1, a.T2)
	case *lg.Symbol:
		atom = resolution.NewAtom(a.Name)
		if _, exists := symMap[a.Name]; !exists {
			symMap[a.Name] = a
		}
	default:
		// Fallback: treat as nullary with string name
		atom = resolution.NewAtom(lit.Atom.String())
	}
	if atom == nil {
		atom = resolution.NewAtom(lit.Atom.String())
	}
	return unitres.NewLiteral(lit.Polarity, atom)
}

// unitResLitToIvyLit converts a unitres.Literal back to an ivylogic.Literal.
// Uses symMap to reconstruct the *lg.Symbol for non-equality atoms.
func unitResLitToIvyLit(lit *unitres.Literal, symMap map[string]*lg.Symbol) *il.Literal {
	var atom lg.Expr
	if lit.Atom.RelName == "=" && len(lit.Atom.Args) == 2 {
		atom = &lg.Eq{T1: lit.Atom.Args[0], T2: lit.Atom.Args[1]}
	} else if sym, ok := symMap[lit.Atom.RelName]; ok {
		if len(lit.Atom.Args) == 0 {
			atom = sym
		} else {
			atom = &lg.Apply{Func: sym, Terms: lit.Atom.Args}
		}
	} else {
		// Fallback: create a boolean symbol
		sym := lg.NewSymbol(lit.Atom.RelName, lg.Boolean)
		if len(lit.Atom.Args) == 0 {
			atom = sym
		} else {
			atom = &lg.Apply{Func: sym, Terms: lit.Atom.Args}
		}
	}
	return il.NewLiteral(lit.Polarity, atom)
}

// ivyLitsToUnitResClauses converts a CNF clause set from ivylogic
// literal lists to unitres literal lists, building a symbol map for
// reverse conversion.
func ivyLitsToUnitResClauses(cnf [][]*il.Literal) ([][]*unitres.Literal, map[string]*lg.Symbol) {
	symMap := make(map[string]*lg.Symbol)
	result := make([][]*unitres.Literal, len(cnf))
	for i, clause := range cnf {
		urClause := make([]*unitres.Literal, len(clause))
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
func extractUnitResResults(r *unitres.UnitRes, symMap map[string]*lg.Symbol) [][]*il.Literal {
	var result [][]*il.Literal

	// Unit queue: each unit literal becomes a single-literal clause
	for _, ulit := range r.UnitQueue {
		ivyLit := unitResLitToIvyLit(ulit, symMap)
		result = append(result, []*il.Literal{ivyLit})
	}

	// Remaining multi-literal clauses
	for _, ucl := range r.Clauses {
		litClause := make([]*il.Literal, len(ucl))
		for j, ulit := range ucl {
			litClause[j] = unitResLitToIvyLit(ulit, symMap)
		}
		result = append(result, litClause)
	}

	return result
}
