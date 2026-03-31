package clauseops

// Literal-level clause operations.
// These work with ivylogic.Literal (polarity+atom) representation,
// as opposed to the Clauses struct operations in ops.go.
// Corresponds to Python's ivy_logic_utils.py clause manipulation functions.

import (
	"fmt"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
)

// CoerceClauseToFormula converts a clause (list of Literal) or a formula
// to a formula. Corresponds to Python's coerce_clause_to_formula
// (ivy_logic_utils.py:35-38).
func CoerceClauseToFormula(c interface{}) lg.Expr {
	switch v := c.(type) {
	case []*il.Literal:
		return ClauseToFormula(v)
	case lg.Expr:
		return DropUniversals(v)
	}
	return nil
}

// ConditionConj returns a list of formulas [Or(c,q) for q in conjuncts of p].
// Corresponds to Python's condition_conj (ivy_logic_utils.py:877-879).
func ConditionConj(c, p lg.Expr) []lg.Expr {
	var ps []lg.Expr
	if a, ok := p.(*lg.And); ok {
		ps = a.Terms
	} else {
		ps = []lg.Expr{p}
	}
	result := make([]lg.Expr, len(ps))
	for i, q := range ps {
		result[i] = &lg.Or{Terms: []lg.Expr{c, q}}
	}
	return result
}

// FormulaToCube converts a formula to a cube (list of Literal).
// Corresponds to Python's formula_to_cube (ivy_logic_utils.py:919-924).
func FormulaToCube(f lg.Expr) []*il.Literal {
	f = lu.ExpandAbbrevs(f)
	if not, ok := f.(*lg.Not); ok {
		// De Morgan: ~(A|B|...) => ~A & ~B & ...
		clause := formulaToClauseLits(not.Body)
		result := make([]*il.Literal, len(clause))
		for i, lit := range clause {
			result[i] = lit.Invert()
		}
		return result
	}
	var lits []lg.Expr
	if a, ok := f.(*lg.And); ok {
		lits = a.Terms
	} else {
		lits = []lg.Expr{f}
	}
	result := make([]*il.Literal, len(lits))
	for i, x := range lits {
		result[i] = formulaToLitLiteral(x)
	}
	return result
}

// formulaToLitLiteral converts a formula to an il.Literal.
func formulaToLitLiteral(f lg.Expr) *il.Literal {
	f = lu.ExpandAbbrevs(f)
	if not, ok := f.(*lg.Not); ok {
		inner := formulaToLitLiteral(not.Body)
		return inner.Invert()
	}
	return il.NewLiteral(1, f)
}

// formulaToClauseLits converts a formula to a clause as a list of il.Literal.
func formulaToClauseLits(f lg.Expr) []*il.Literal {
	f = lu.ExpandAbbrevs(f)
	f = lu.DeMorgan(f)
	if lu.IsTrue(f) {
		return []*il.Literal{il.NewLiteral(1, f)}
	}
	if or, ok := f.(*lg.Or); ok {
		var result []*il.Literal
		for _, t := range or.Terms {
			result = append(result, formulaToClauseLits(t)...)
		}
		return result
	}
	return []*il.Literal{formulaToLitLiteral(f)}
}

// FormulaToClausesAux clausifies a formula into a list of clauses
// (each clause is a list of Literal). Requires the formula to be in CNF.
// Corresponds to Python's formula_to_clauses_aux (ivy_logic_utils.py:953-966).
func FormulaToClausesAux(f lg.Expr) [][]*il.Literal {
	f = DropUniversals(f)
	f = lu.DeMorgan(f)
	if il.IsFalse(f) {
		return [][]*il.Literal{{}} // empty clause = false
	}
	if _, isAnd := f.(*lg.And); !isAnd {
		cls := formulaToClauseLits(f)
		for _, lit := range cls {
			if isTautLiteral(lit) {
				return nil // tautological clause removed
			}
		}
		return [][]*il.Literal{cls}
	}
	a := f.(*lg.And)
	var result [][]*il.Literal
	for _, x := range a.Terms {
		result = append(result, FormulaToClausesAux(x)...)
	}
	return result
}

// isTautLiteral checks if an il.Literal is a tautology.
func isTautLiteral(lit *il.Literal) bool {
	if lit.Polarity == 1 {
		return lu.IsTautLit(lit.Atom)
	}
	return lu.IsTautLit(&lg.Not{Body: lit.Atom})
}

// isVacLiteral checks if an il.Literal is vacuously false.
func isVacLiteral(lit *il.Literal) bool {
	if lit.Polarity == 1 {
		return lu.IsVacLit(lit.Atom)
	}
	return lu.IsVacLit(&lg.Not{Body: lit.Atom})
}

// LitToFormula converts a Literal to a formula.
// Corresponds to Python's lit_to_formula (ivy_logic_utils.py:988-989).
func LitToFormula(lit *il.Literal) lg.Expr {
	if lit.Polarity == 1 {
		return lit.Atom
	}
	return &lg.Not{Body: lit.Atom}
}

// CubeToFormula converts a cube (list of Literal) to a conjunction formula.
// Corresponds to Python's cube_to_formula (ivy_logic_utils.py:991-992).
func CubeToFormula(c []*il.Literal) lg.Expr {
	terms := make([]lg.Expr, len(c))
	for i, lit := range c {
		terms[i] = LitToFormula(lit)
	}
	return &lg.And{Terms: terms}
}

// ClauseToFormula converts a clause (list of Literal) to a disjunction formula.
// Corresponds to Python's clause_to_formula (ivy_logic_utils.py:994-996).
func ClauseToFormula(c []*il.Literal) lg.Expr {
	lits := make([]lg.Expr, len(c))
	for i, lit := range c {
		lits[i] = LitToFormula(lit)
	}
	if len(lits) == 1 {
		return lits[0]
	}
	return &lg.Or{Terms: lits}
}

// CanonizeClause rewrites a clause so variables occur in order V0, V1, ...
// Corresponds to Python's canonize_clause (ivy_logic_utils.py:1010-1014).
func CanonizeClause(cl []*il.Literal) []*il.Literal {
	// Collect used variables in order
	seen := make(map[string]bool)
	var vars []*lg.Variable
	for _, lit := range cl {
		collectVarsOrderedLit(lit, seen, &vars)
	}
	// Build substitution
	subs := make(map[lg.NodeKey]lg.Expr, len(vars))
	for i, v := range vars {
		nv, _ := lg.NewVariable(fmt.Sprintf("V%d", i), v.VSort)
		subs[lg.Key(v)] = nv
	}
	return SubstituteLitClause(cl, subs)
}

func collectVarsOrderedLit(lit *il.Literal, seen map[string]bool, result *[]*lg.Variable) {
	collectVarsOrderedNode(lit.Atom, seen, result)
}

func collectVarsOrderedNode(node lg.Expr, seen map[string]bool, result *[]*lg.Variable) {
	if v, ok := node.(*lg.Variable); ok {
		if !seen[v.Name] {
			seen[v.Name] = true
			*result = append(*result, v)
		}
		return
	}
	for _, c := range node.Children() {
		collectVarsOrderedNode(c, seen, result)
	}
}

// SubstituteLitClause applies a substitution to a clause of Literals.
func SubstituteLitClause(cl []*il.Literal, subs map[lg.NodeKey]lg.Expr) []*il.Literal {
	result := make([]*il.Literal, len(cl))
	for i, lit := range cl {
		newAtom, err := lu.Substitute(lit.Atom, subs)
		if err != nil {
			result[i] = lit
		} else {
			result[i] = il.NewLiteral(lit.Polarity, newAtom)
		}
	}
	return result
}

// TrimClauses removes unused let bindings from a Clauses struct.
// Corresponds to Python's trim_clauses (ivy_logic_utils.py:1024-1037).
func TrimClauses(cls *Clauses) *Clauses {
	usedSyms := make(map[string]bool)
	var seeds []lg.Expr
	seeds = append(seeds, cls.Fmlas...)
	for _, d := range cls.Defs {
		rep := il.GetAppRep(d.Lhs)
		if rep != nil && !isSkolem(rep) {
			seeds = append(seeds, d.Rhs)
		}
	}
	for len(seeds) > 0 {
		seed := seeds[len(seeds)-1]
		seeds = seeds[:len(seeds)-1]
		syms := UsedSymbolsAST(seed)
		for _, sym := range syms {
			if isSkolem(sym) {
				if !usedSyms[sym.Name] {
					usedSyms[sym.Name] = true
					if idx, ok := cls.DefIdx[lg.Key(sym)]; ok && idx < len(cls.Defs) {
						seeds = append(seeds, cls.Defs[idx].Rhs)
					}
				}
			}
		}
	}
	var newDefs []*il.Definition
	for _, d := range cls.Defs {
		rep := il.GetAppRep(d.Lhs)
		if rep != nil && (!isSkolem(rep) || usedSyms[rep.Name]) {
			newDefs = append(newDefs, d)
		}
	}
	return NewClauses(cls.Fmlas, newDefs, cls.Annot)
}

// RewriteClause rewrites a clause by substituting a variable with a term.
// Corresponds to Python's rewrite_clause (ivy_logic_utils.py:1039-1042).
func RewriteClause(clause []*il.Literal, v lg.Expr, t lg.Expr) []*il.Literal {
	rep := il.GetAppRep(v)
	if rep == nil {
		return clause
	}
	subs := map[lg.NodeKey]lg.Expr{lg.Key(rep): t}
	return SubstituteLitClause(clause, subs)
}

// TrivFmlaToLit converts a formula to a Literal without full clausification.
// Corresponds to Python's triv_fmla_to_lit (ivy_logic_utils.py:1044-1047).
func TrivFmlaToLit(f lg.Expr) *il.Literal {
	if not, ok := f.(*lg.Not); ok {
		return il.NewLiteral(0, not.Body)
	}
	return il.NewLiteral(1, f)
}

// TrivFmlaToClause converts a formula to a clause without full clausification.
// Corresponds to Python's triv_fmla_to_clause (ivy_logic_utils.py:1062-1063).
func TrivFmlaToClause(fmla lg.Expr) []*il.Literal {
	ors := CollectOr(fmla)
	result := make([]*il.Literal, len(ors))
	for i, f := range ors {
		result[i] = TrivFmlaToLit(f)
	}
	return result
}

// SimplifyClauseFmla simplifies a formula by converting to a clause,
// simplifying, and converting back.
// Corresponds to Python's simplify_clause_fmla (ivy_logic_utils.py:1065-1066).
func SimplifyClauseFmla(fmla lg.Expr) lg.Expr {
	return ClauseToFormula(SimplifyClause(TrivFmlaToClause(fmla)))
}

// SimplifyClause simplifies a clause by rewriting with negative equalities
// and removing tautologies/vacuous literals.
// Corresponds to Python's simplify_clause (ivy_logic_utils.py:1069-1078).
func SimplifyClause(clause []*il.Literal) []*il.Literal {
	// Rewrite using negative equalities: ~(X=t) => substitute X->t
	for _, lit := range clause {
		if lit.Polarity == 0 {
			if eq, ok := lit.Atom.(*lg.Eq); ok {
				for _, idx := range []int{0, 1} {
					var lhs, rhs lg.Expr
					if idx == 0 {
						lhs, rhs = eq.T1, eq.T2
					} else {
						lhs, rhs = eq.T2, eq.T1
					}
					if _, ok := lhs.(*lg.Variable); ok {
						clause = RewriteClause(clause, lhs, rhs)
						break
					}
				}
			}
		}
	}
	// Check for tautology
	for _, lit := range clause {
		if isTautLiteral(lit) {
			return []*il.Literal{il.NewLiteral(1, &lg.And{})} // [true]
		}
	}
	// Remove vacuous literals and duplicates
	var filtered []*il.Literal
	for _, lit := range clause {
		if !isVacLiteral(lit) {
			filtered = append(filtered, lit)
		}
	}
	return RemoveDuplicatesLit(filtered)
}

// IsTautology checks if a literal clause is a tautology.
// Corresponds to Python's is_tautology (ivy_logic_utils.py:1080-1087).
func IsTautology(clause []*il.Literal) bool {
	for _, lit := range clause {
		if isTautLiteral(lit) {
			return true
		}
		for _, lit2 := range clause {
			if lit.Polarity != lit2.Polarity && lit.Atom.Equal(lit2.Atom) {
				return true
			}
		}
	}
	return false
}

// IsTautologyFmla checks if a formula is a tautology when treated as a clause.
// Corresponds to Python's is_tautology_fmla (ivy_logic_utils.py:1089-1090).
func IsTautologyFmla(fmla lg.Expr) bool {
	return IsTautology(TrivFmlaToClause(fmla))
}

// LitInClause checks if a literal is in a clause (by structural equality).
// Corresponds to Python's lit_in_clause (ivy_logic_utils.py:1092-1096).
func LitInClause(lit1 *il.Literal, clause []*il.Literal) bool {
	for _, lit2 := range clause {
		if lit1.Polarity == lit2.Polarity && lit1.Atom.Equal(lit2.Atom) {
			return true
		}
	}
	return false
}

// RemoveDuplicatesLit removes duplicate literals from a clause.
// Corresponds to Python's remove_duplicates (ivy_logic_utils.py:1098-1103).
func RemoveDuplicatesLit(clause []*il.Literal) []*il.Literal {
	var res []*il.Literal
	for _, lit1 := range clause {
		if !LitInClause(lit1, res) {
			res = append(res, lit1)
		}
	}
	return res
}

// ReduceClauses reduces a clause set by eliminating subsumed clauses.
// Corresponds to Python's reduce_clauses (ivy_logic_utils.py:1105-1115).
// Note: This requires subsume() which is in Batch 1.7; for now uses
// a simplified version that only removes exact duplicates.
func ReduceClauses(clauses [][]*il.Literal) [][]*il.Literal {
	var used [][]*il.Literal
	for _, cl := range clauses {
		subsumed := false
		for _, u := range used {
			if clauseEqual(u, cl) {
				subsumed = true
				break
			}
		}
		if !subsumed {
			used = append(used, cl)
		}
	}
	return used
}

// clauseEqual checks if two literal clauses are equal (same literals in order).
func clauseEqual(c1, c2 []*il.Literal) bool {
	if len(c1) != len(c2) {
		return false
	}
	for i := range c1 {
		if c1[i].Polarity != c2[i].Polarity || !c1[i].Atom.Equal(c2[i].Atom) {
			return false
		}
	}
	return true
}

// BoolConst creates a boolean constant (0-arity relation).
// Corresponds to Python's bool_const (ivy_logic_utils.py:1388-1389).
func BoolConst(name string) lg.Expr {
	sym := lg.NewConst(name, il.RelationSort(nil))
	return il.Atom(sym, nil)
}
