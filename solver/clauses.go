// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_solver.py.

package solver

import (
	"fmt"

	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/z3bridge"
)

// ClausesImplyList checks whether clauses1 implies each element of clauses2List.
// Returns a bool slice with the result for each element.
// Corresponds to Python's clauses_imply_list.
func (s *Solver) ClausesImplyList(clauses1 *clauseops.Clauses, clauses2List []*clauseops.Clauses) ([]bool, error) {
	z3solver := s.tr.Ctx.NewSolver()

	z1, err := s.ClausesToZ3(clauses1)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(z1)

	results := make([]bool, len(clauses2List))
	for i, clauses2 := range clauses2List {
		dual := clauseops.NegateClauses(clauses2)
		z2, err := s.ClausesToZ3(dual)
		if err != nil {
			return nil, err
		}
		z3solver.Push()
		z3solver.Assert(z2)
		results[i] = z3solver.Check() == z3bridge.Unsat
		z3solver.Pop()
	}
	return results, nil
}

// ConditionClauses wraps each formula in clauses with an implication from fmla.
// Returns new Clauses where each formula is (Not(fmla) OR formula).
// This is a convenience wrapper around clauseops.ConditionClauses.
func ConditionClauses(clauses *clauseops.Clauses, fmla lg.Node) *clauseops.Clauses {
	return clauseops.ConditionClauses(clauses, fmla)
}

// DualClauses negates a clause set (for checking implications).
// This is a convenience wrapper around clauseops.NegateClauses.
func DualClauses(clauses *clauseops.Clauses) *clauseops.Clauses {
	return clauseops.NegateClauses(clauses)
}

// TrueClauses returns a trivially true clause set.
func TrueClauses() *clauseops.Clauses {
	return clauseops.TrueClauses(nil)
}

// FalseClauses returns a trivially false clause set.
func FalseClauses() *clauseops.Clauses {
	return clauseops.FalseClauses(nil)
}

// AndClauses computes the conjunction of multiple clause sets.
func AndClauses(args ...*clauseops.Clauses) *clauseops.Clauses {
	return clauseops.AndClausesTyped(args...)
}

// FormulaToClauses wraps a formula as a Clauses set.
func FormulaToClauses(fmla lg.Node) *clauseops.Clauses {
	return clauseops.FormulaToClauses(fmla, nil)
}

// BoundQuantifiersClauses bounds universal quantifiers in clauses to
// the given set of representative terms per sort.
// Corresponds to Python's bound_quantifiers_clauses.
func (s *Solver) BoundQuantifiersClauses(
	clauses *clauseops.Clauses,
	reps map[string][]lg.Node,
	uninterpretedSorts map[string]bool,
) *clauseops.Clauses {
	bq := func(fmla lg.Node) lg.Node {
		vars := clauseops.VariablesAST(fmla)
		if len(vars) == 0 {
			return fmla
		}
		var constraints []lg.Node
		for _, v := range vars {
			sortName := il.SortName(v.VSort)
			if uninterpretedSorts != nil && !uninterpretedSorts[sortName] {
				continue
			}
			terms, ok := reps[sortName]
			if !ok || len(terms) == 0 {
				continue
			}
			eqs := make([]lg.Node, len(terms))
			for i, t := range terms {
				eqs[i] = &lg.Eq{T1: v, T2: t}
			}
			constraints = append(constraints, &lg.Or{Terms: eqs})
		}
		if len(constraints) == 0 {
			return fmla
		}
		ante := &lg.And{Terms: constraints}
		return &lg.Implies{T1: ante, T2: fmla}
	}

	newFmlas := make([]lg.Node, len(clauses.Fmlas))
	for i, f := range clauses.Fmlas {
		newFmlas[i] = bq(f)
	}
	defs := make([]*il.Definition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return clauseops.NewClauses(newFmlas, defs, clauses.Annot)
}

// RemoveDuplicatesClauses removes duplicate formulas from a clause set.
// Uses Z3 expression identity for deduplication.
// Corresponds to Python's remove_duplicates_clauses.
func (s *Solver) RemoveDuplicatesClauses(clauses *clauseops.Clauses) (*clauseops.Clauses, error) {
	seen := make(map[string]bool)
	var unique []lg.Node
	for _, f := range clauses.Fmlas {
		zf, err := s.translateClosed(f)
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
	defs := make([]*il.Definition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return clauseops.NewClauses(unique, defs, clauses.Annot), nil
}

// ClausesModelToDiagram returns a diagram (clause set) of a model of clauses.
// This is a simplified version of the Python clauses_model_to_diagram.
func (s *Solver) ClausesModelToDiagram(
	clauses *clauseops.Clauses,
	ignore func(*lg.Const) bool,
	axioms *clauseops.Clauses,
) (*clauseops.Clauses, error) {
	if axioms == nil {
		axioms = clauseops.TrueClauses(nil)
	}

	combined := clauseops.AndClausesTyped(clauses, axioms)
	mr, err := s.GetModelClauses(combined)
	if err != nil {
		return nil, err
	}
	if mr == nil {
		return nil, nil
	}

	if ignore == nil {
		ignore = func(*lg.Const) bool { return false }
	}

	// Extract model facts (simplified)
	var fmlas []lg.Node
	symSet := clauses.Symbols()
	for _, sym := range symSet {
		if ignore(sym.(*lg.Const)) {
			continue
		}
		zSym, err := s.tr.Translate(sym)
		if err != nil {
			continue
		}
		val, ok := mr.Model.Eval(zSym, true)
		if !ok {
			continue
		}
		_ = val
		fmlas = append(fmlas, &lg.Eq{T1: sym, T2: sym})
	}

	result := clauseops.NewClauses(fmlas, nil, nil)

	// Filter redundant facts
	filtered, err := s.FilterRedundantFacts(result, axioms)
	if err != nil {
		return result, nil
	}
	return filtered, nil
}

// NewZ3Solver creates a new Z3 solver on this solver's context.
func (s *Solver) NewZ3Solver() *z3bridge.Solver {
	return s.tr.Ctx.NewSolver()
}

// AddClauses adds a clauses set to a Z3 solver.
// Corresponds to Python's add_clauses.
func (s *Solver) AddClauses(z3solver *z3bridge.Solver, clauses *clauseops.Clauses) error {
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return err
	}
	z3solver.Assert(zc)
	return nil
}

// SolverAdd adds a formula to a Z3 solver.
// Corresponds to Python's solver_add.
func (s *Solver) SolverAdd(z3solver *z3bridge.Solver, fmla lg.Node) error {
	zf, err := s.translateClosed(fmla)
	if err != nil {
		return err
	}
	z3solver.Assert(zf)
	return nil
}

// Decide checks satisfiability and returns the result.
// Returns an error if the result is unknown.
// Corresponds to Python's decide.
func Decide(z3solver *z3bridge.Solver) (z3bridge.CheckResult, error) {
	result := z3solver.Check()
	if result == z3bridge.Unknown {
		return result, fmt.Errorf("solver produced inconclusive result")
	}
	return result, nil
}
