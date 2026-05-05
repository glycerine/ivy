// Ported to Go from ivy_solver.py.

package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// ClausesImplyList checks whether clauses1 implies each element of clauses2List.
// Returns a bool slice with the result for each element.
// Corresponds to Python's clauses_imply_list.
func (s *Solver) ClausesImplyList(clauses1 *Clauses, clauses2List []*Clauses) ([]bool, error) {
	xtracer.Trace("ivy_solver.py:1036 clauses_imply_list() ENTER")
	z3solver := s.tr.Ctx.NewZ3Solver()

	z1, err := s.ClausesToZ3(clauses1)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(z1)

	results := make([]bool, len(clauses2List))
	for i, clauses2 := range clauses2List {
		dual := NegateClauses(clauses2)
		z2, err := s.ClausesToZ3(dual)
		if err != nil {
			return nil, err
		}
		z3solver.Push()
		z3solver.Assert(z2)
		results[i] = z3solver.Check() == Unsat
		z3solver.Pop()
	}
	return results, nil
}

// ConditionClauses wraps each formula in clauses with an implication from fmla.
// Returns new Clauses where each formula is (Not(fmla) OR formula).
// This is a convenience wrapper around module.ConditionClauses.
func Z3ConditionClauses(clauses *Clauses, fmla Expr) *Clauses {
	return ConditionClauses(clauses, fmla)
}

// Z3TrueClauses returns a trivially true clause set.
func Z3TrueClauses() *Clauses {
	return TrueClauses(nil)
}

// FalseClauses returns a trivially false clause set.
func Z3FalseClauses() *Clauses {
	return FalseClauses(nil)
}

// AndClauses computes the conjunction of multiple clause sets.
func Z3AndClauses(args ...*Clauses) *Clauses {
	return AndClausesTyped(args...)
}

// FormulaToClauses wraps a formula as a Clauses set.
func Z3FormulaToClauses(fmla Expr) *Clauses {
	return FormulaToClauses(fmla, nil)
}

// BoundQuantifiersClauses bounds universal quantifiers in clauses to
// the given set of representative terms per sort.
// Corresponds to Python's bound_quantifiers_clauses.
func (s *Solver) BoundQuantifiersClauses(
	clauses *Clauses,
	reps map[string][]Expr, uninterpretedSorts map[string]bool,
) *Clauses {
	xtracer.Trace("ivy_solver.py:1541 bound_quantifiers_clauses() ENTER")
	bq := func(fmla Expr) Expr {
		vars := VariablesAST(fmla)
		if len(vars) == 0 {
			return fmla
		}
		var constraints []Expr
		for _, v := range vars {
			sortName := IvySortName(v.VSort)
			if uninterpretedSorts != nil && !uninterpretedSorts[sortName] {
				continue
			}
			terms, ok := reps[sortName]
			if !ok || len(terms) == 0 {
				continue
			}
			eqs := make([]Expr, len(terms))
			for i, t := range terms {
				eqs[i] = &Eq{T1: v, T2: t}
			}
			constraints = append(constraints, &Or{Terms: eqs})
		}
		if len(constraints) == 0 {
			return fmla
		}
		ante := &And{Terms: constraints}
		return &Implies{T1: ante, T2: fmla}
	}

	newFmlas := make([]Expr, len(clauses.Fmlas))
	for i, f := range clauses.Fmlas {
		newFmlas[i] = bq(f)
	}
	defs := make([]*IvyDefinition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return NewClauses(newFmlas, defs, clauses.Annot)
}

// RemoveDuplicatesClauses removes duplicate formulas from a clause set.
// Uses Z3 expression identity for deduplication.
// Corresponds to Python's remove_duplicates_clauses.
func (s *Solver) RemoveDuplicatesClauses(clauses *Clauses) (*Clauses, error) {
	xtracer.Trace("ivy_solver.py:1122 remove_duplicates_clauses() ENTER")
	seen := make(map[string]bool)
	var unique []Expr
	for _, f := range clauses.Fmlas {
		zf, err := s.formulaToZ3(f)
		if err != nil {
			unique = append(unique, f)
			continue
		}
		key := zf.String()
		if !seen[key] {
			seen[key] = true
			unique = append(unique, f)
		}
	}
	defs := make([]*IvyDefinition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return NewClauses(unique, defs, clauses.Annot), nil
}

// ClausesModelToDiagram returns a diagram (clause set) of a model of clauses.
// Uses ModelIfNone to get a HerbrandModel, then ModelFacts + NumeralAssign +
// SubstituteConstantsClauses + FilterRedundantFacts, matching Python's
// clauses_model_to_diagram (ivy_solver.py:1448-1516).
func (s *Solver) ClausesModelToDiagram(
	clauses *Clauses,
	ignore func(*Const) bool,
	axioms *Clauses,
) (*Clauses, error) {
	return s.ClausesModelToDiagramFull(clauses, ignore, nil, nil, axioms, true, true, true)
}

// ClausesModelToDiagramFull is the full-featured version matching all Python parameters.
func (s *Solver) ClausesModelToDiagramFull(
	clauses *Clauses,
	ignore func(*Const) bool,
	implied *Clauses,
	model *HerbrandModel,
	axioms *Clauses,
	weaken bool,
	numerals bool,
	upwardClose bool,
) (*Clauses, error) {
	xtracer.Trace("ivy_solver.py:1594 clauses_model_to_diagram() ENTER")
	if axioms == nil {
		axioms = TrueClauses(nil)
	}
	if ignore == nil {
		ignore = func(*Const) bool { return false }
	}

	// Get model
	combined := AndClausesTyped(clauses, axioms)
	h := s.ModelIfNone(combined, implied, model)
	if h == nil {
		return nil, nil
	}

	// Extract model facts (always use upclose=true for diagrams, matching Python)
	noIgnore := func(*Const) bool { return false }
	res := ModelFacts(h, noIgnore, clauses, true)

	// Find representative elements via numeral assignment or skolem prefix
	var reps map[NodeKey]Expr
	if numerals {
		na := NumeralAssignWithClauses(h, res)
		reps = make(map[NodeKey]Expr, len(na))
		for elemName, numName := range na {
			// Find the sort from the model
			for _, sort := range h.Sorts() {
				for _, c := range h.SortUniverse(sort) {
					if c.Name == elemName {
						reps[Key(c)] = NewConst(numName, c.CSort)
					}
				}
			}
		}
	} else {
		reps = make(map[NodeKey]Expr)
		// Use constants from clauses as reps where possible
		usedConsts := ConstantsClauses(clauses)
		for _, c := range usedConsts {
			mc := h.EvalConstant(c)
			if mc != nil {
				if existing, ok := reps[Key(mc)]; ok {
					// Prefer non-skolem reps
					if existSym, ok2 := existing.(*Const); ok2 {
						if z3bridgeIsSkolem(existSym.Name) && !z3bridgeIsSkolem(c.Name) {
							reps[Key(mc)] = c
						}
					}
				} else {
					reps[Key(mc)] = c
				}
			}
		}
		// Skolemize any remaining unassigned elements
		for _, sort := range h.Sorts() {
			for _, e := range h.SortUniverse(sort) {
				if _, ok := reps[Key(e)]; !ok {
					reps[Key(e)] = NewConst("__"+e.Name, e.CSort)
				}
			}
		}
	}

	// Substitute constants
	res = SubstituteConstantsClauses(res, reps)

	// Filter defined skolems
	if len(clauses.DefIdx) > 0 {
		var filtered []Expr
		for _, f := range res.Fmlas {
			syms := UsedSymbolsAST(f)
			hasSkolemDef := false
			for _, c := range syms.All() {
				if z3bridgeIsSkolem(ExprName(c)) {
					if _, inIdx := clauses.DefIdx[Key(c)]; inIdx {
						hasSkolemDef = true
						break
					}
				}
			}
			if !hasSkolemDef {
				filtered = append(filtered, f)
			}
		}
		res = NewClauses(filtered, res.Defs, res.Annot)
	}

	// Filter redundant facts
	var err error
	res, err = s.FilterRedundantFacts(res, axioms)
	if err != nil {
		return res, nil
	}

	// Weakening via unsat core
	if weaken {
		unlikely := func(fmla Expr) bool {
			if eq, ok := fmla.(*Eq); ok {
				if _, isSym := eq.T1.(*Const); isSym {
					return true
				}
			}
			return false
		}
		repTerms := make(map[string][]Expr)
		for _, sort := range h.Sorts() {
			sortName := IvySortName(sort)
			for _, c := range h.SortUniverse(sort) {
				if rep, ok := reps[Key(c)]; ok {
					repTerms[sortName] = append(repTerms[sortName], rep)
				}
			}
		}
		clauses1Weak := s.BoundQuantifiersClauses(clauses, repTerms, nil)
		core, err := s.UnsatCore(res, AndClausesTyped(TrueClauses(nil), axioms), clauses1Weak, unlikely)
		if err == nil && core != nil {
			res = core
		}
	}

	// Filter out non-rep skolems
	repSet := make(map[string]bool)
	for _, v := range reps {
		if sym, ok := v.(*Const); ok {
			repSet[sym.Name] = true
		}
	}
	var finalFmlas []Expr
	for _, f := range res.Fmlas {
		syms := UsedSymbolsAST(f)
		hasIgnored := false
		for _, c := range syms.All() {
			if cc, ok := c.(*Const); ok && ignore(cc) && !repSet[ExprName(c)] {
				hasIgnored = true
				break
			}
		}
		if !hasIgnored {
			finalFmlas = append(finalFmlas, f)
		}
	}
	res = NewClauses(finalFmlas, res.Defs, res.Annot)

	// If not upward-closing, add universe closure constraints
	if !upwardClose {
		var ucFmlas []Expr
		for _, sort := range h.Sorts() {
			x, _ := NewVariable("X", sort)
			var eqs []Expr
			for _, c := range h.SortUniverse(sort) {
				if rep, ok := reps[Key(c)]; ok {
					eqs = append(eqs, &Eq{T1: x, T2: rep})
				}
			}
			if len(eqs) > 0 {
				ucFmlas = append(ucFmlas, &Or{Terms: eqs})
			}
		}
		if len(ucFmlas) > 0 {
			ucClauses := NewClauses(ucFmlas, nil, nil)
			res = AndClausesTyped(res, ucClauses)
		}
	}

	return res, nil
}

// NewZ3Solver creates a new Z3 solver on this solver's context.
func (s *Solver) NewZ3Solver() *Z3Solver {
	xtracer.Trace("ivy_solver.py:808 new_solver() ENTER")
	return s.tr.Ctx.NewZ3Solver()
}

// AddClauses adds a clauses set to a Z3 solver.
// Corresponds to Python's add_clauses.
func (s *Solver) AddClauses(z3solver *Z3Solver, clauses *Clauses) error {
	xtracer.Trace("ivy_solver.py:820 add_clauses() ENTER")
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return err
	}
	z3solver.Assert(zc)
	return nil
}

// SolverAdd adds a formula to a Z3 solver.
// Corresponds to Python's solver_add.
func (s *Solver) SolverAdd(z3solver *Z3Solver, fmla Expr) error {
	xtracer.Trace("ivy_solver.py:812 solver_add() ENTER")
	zf, err := s.formulaToZ3(fmla)
	if err != nil {
		return err
	}
	z3solver.Assert(zf)
	return nil
}

// Decide checks satisfiability and returns the result.
// If assumptions are provided, uses assumption-based checking.
// Returns an error if the result is unknown.
// Corresponds to Python's decide(s, atoms=None) (ivy_solver.py:1164).
func Decide(z3solver *Z3Solver, assumptions ...Z3Expr) (Z3CheckResult, error) {
	xtracer.Trace("ivy_solver.py:1302 decide() ENTER")
	var result Z3CheckResult
	if len(assumptions) > 0 {
		result = z3solver.CheckAssumptions(assumptions)
	} else {
		result = z3solver.Check()
	}
	if result == Unknown {
		return result, fmt.Errorf("solver produced inconclusive result")
	}
	return result, nil
}
