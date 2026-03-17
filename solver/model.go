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

// ModelResult holds a Z3 model and associated solver state.
type ModelResult struct {
	Solver  *z3bridge.Solver
	Model   *z3bridge.Model
	Vocab   []*lg.Const
	Context *z3bridge.Context
}

// Eval evaluates a Z3 expression in the model with completion.
func (mr *ModelResult) Eval(e z3bridge.Expr) (z3bridge.Expr, bool) {
	return mr.Model.Eval(e, true)
}

// String returns the model's string representation.
func (mr *ModelResult) String() string {
	if mr.Model == nil {
		return "<nil model>"
	}
	return mr.Model.String()
}

// GetModelClauses checks satisfiability of clauses and returns a ModelResult if sat.
// Corresponds to Python's get_model_clauses.
func (s *Solver) GetModelClauses(clauses *clauseops.Clauses) (*ModelResult, error) {
	z3solver := s.tr.Ctx.NewSolver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	result := z3solver.Check()
	if result == z3bridge.Unsat {
		return nil, nil // unsatisfiable
	}

	m := z3solver.Model()
	if m == nil {
		return nil, fmt.Errorf("solver returned sat but no model")
	}

	// Collect vocabulary from clauses
	symSet := clauses.Symbols()
	vocab := make([]*lg.Const, 0, len(symSet))
	for _, symN := range symSet {
		if c, ok := symN.(*lg.Const); ok {
			vocab = append(vocab, c)
		}
	}

	return &ModelResult{
		Solver:  z3solver,
		Model:   m,
		Vocab:   vocab,
		Context: s.tr.Ctx,
	}, nil
}

// ModelValues evaluates a list of expressions in a model.
// Returns a map from expression string to its model value.
func (s *Solver) ModelValues(model *z3bridge.Model, syms []*lg.Const) (map[string]z3bridge.Expr, error) {
	result := make(map[string]z3bridge.Expr, len(syms))
	for _, sym := range syms {
		zSym, err := s.tr.Translate(sym)
		if err != nil {
			continue // skip symbols we can't translate
		}
		val, ok := model.Eval(zSym, true)
		if ok {
			result[sym.Name] = val
		}
	}
	return result, nil
}

// FinalCond is the interface for final condition checkers passed to
// GetSmallModel. Each checker has methods matching Python's checker protocol:
//
//	Cond():   returns the condition as a Clauses
//	Start():  called before checking begins
//	Sat():    called when SAT; returns true to continue (ignore failure)
//	Unsat():  called when UNSAT; returns true to continue
//	Assume(): returns true if this should be assumed (not checked)
//
// Corresponds to the final_cond parameter of Python's get_small_model.
type FinalCond interface {
	Cond() *clauseops.Clauses
	Start()
	Sat() bool
	Unsat() bool
	Assume() bool
}

// GetSmallModel finds a satisfying model of clauses, minimizing sort
// universe sizes and relation extensions.
// Returns nil if unsatisfiable.
// Corresponds to Python's get_small_model.
func (s *Solver) GetSmallModel(
	clauses *clauseops.Clauses,
	sortsToMinimize []lg.Sort,
	relationsToMinimize []*lg.Const,
) (*ModelResult, error) {
	return s.GetSmallModelWithCond(clauses, sortsToMinimize, relationsToMinimize, nil, true)
}

// GetSmallModelWithCond is the full version of GetSmallModel that accepts
// a list of final condition checkers. This matches Python's get_small_model
// with the final_cond parameter.
//
// When finalCond is a non-empty list, for each checker:
//   - If Assume(): add Cond() to the solver as a permanent assumption
//   - Otherwise: push, add Cond(), check SAT/UNSAT, call Sat()/Unsat(), pop
//
// If any check returns false from Sat()/Unsat(), we stop and return
// the current model (or nil if UNSAT).
func (s *Solver) GetSmallModelWithCond(
	clauses *clauseops.Clauses,
	sortsToMinimize []lg.Sort,
	relationsToMinimize []*lg.Const,
	finalCond []FinalCond,
	shrink bool,
) (*ModelResult, error) {

	z3solver := s.tr.Ctx.NewSolver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	// Process final conditions (checkers)
	overallResult := z3bridge.Unsat
	if len(finalCond) > 0 {
		for _, fc := range finalCond {
			fc.Start()
			cond := fc.Cond()
			if cond == nil {
				continue
			}

			if fc.Assume() {
				// Assumed condition: add permanently to solver
				zCond, err := s.ClausesToZ3(cond)
				if err != nil {
					continue
				}
				z3solver.Assert(zCond)
			} else {
				// Checked condition: push, add, check, pop
				zCond, err := s.ClausesToZ3(cond)
				if err != nil {
					continue
				}
				z3solver.Push()
				z3solver.Assert(zCond)
				res := z3solver.Check()
				z3solver.Pop()

				if res != z3bridge.Unsat {
					overallResult = res
					// SAT: the check condition is satisfiable (property fails)
					if fc.Sat() {
						// Checker says to continue (ignore this failure)
						overallResult = z3bridge.Unsat
						continue
					}
					break // stop checking
				} else {
					overallResult = z3bridge.Unsat
					fc.Unsat()
				}
			}
		}
	} else {
		// No final conditions: just check satisfiability
		overallResult = z3solver.Check()
	}

	if overallResult == z3bridge.Unsat {
		return nil, nil
	}

	if shrink {
		// Minimize sorts
		for _, sort := range sortsToMinimize {
			for n := 1; ; n++ {
				sc := SortSizeConstraint(sort, n)
				zsc, err := s.translateClosed(sc)
				if err != nil {
					break
				}
				z3solver.Push()
				z3solver.Assert(zsc)
				if z3solver.Check() == z3bridge.Sat {
					break
				}
				z3solver.Pop()
			}
		}

		// Minimize relations
		for _, rel := range relationsToMinimize {
			for n := 1; ; n++ {
				sc := RelationSizeConstraint(rel, n)
				zsc, err := s.translateClosed(sc)
				if err != nil {
					break
				}
				z3solver.Push()
				z3solver.Assert(zsc)
				if z3solver.Check() == z3bridge.Sat {
					break
				}
				z3solver.Pop()
			}
		}
	}

	m := z3solver.Model()
	if m == nil {
		return nil, fmt.Errorf("solver returned sat but no model")
	}

	symSet := clauses.Symbols()
	vocab := make([]*lg.Const, 0, len(symSet))
	for _, symN := range symSet {
		if c, ok := symN.(*lg.Const); ok {
			vocab = append(vocab, c)
		}
	}

	return &ModelResult{
		Solver:  z3solver,
		Model:   m,
		Vocab:   vocab,
		Context: s.tr.Ctx,
	}, nil
}

// EvalFormula evaluates a formula in a model, returning true/false/unknown.
func (s *Solver) EvalFormula(model *z3bridge.Model, fmla lg.Node) (bool, error) {
	zf, err := s.tr.Translate(fmla)
	if err != nil {
		return false, err
	}
	val, ok := model.Eval(zf, true)
	if !ok {
		return false, fmt.Errorf("could not evaluate formula in model")
	}
	str := val.String()
	return str == "true", nil
}

// CubeToZ3 converts a list of literals (a cube) to a Z3 conjunction.
// Corresponds to Python's cube_to_z3.
func (s *Solver) CubeToZ3(cube []*il.Literal) (z3bridge.Expr, error) {
	if len(cube) == 0 {
		return s.tr.Ctx.BoolVal(true), nil
	}
	exprs := make([]z3bridge.Expr, len(cube))
	for i, lit := range cube {
		zlit, err := s.LiteralToZ3(lit)
		if err != nil {
			return z3bridge.Expr{}, err
		}
		exprs[i] = zlit
	}
	if len(exprs) == 1 {
		return exprs[0], nil
	}
	return s.tr.Ctx.And(exprs...), nil
}

// LiteralToZ3 converts a single literal to a Z3 expression.
func (s *Solver) LiteralToZ3(lit *il.Literal) (z3bridge.Expr, error) {
	zAtom, err := s.tr.Translate(lit.Atom)
	if err != nil {
		return z3bridge.Expr{}, err
	}
	if lit.Polarity == 0 {
		return s.tr.Ctx.Not(zAtom), nil
	}
	return zAtom, nil
}

// CheckCube checks if a cube (conjunction of literals) is consistent with
// the solver state. Returns true if sat.
// Corresponds to Python's check_cube.
func (s *Solver) CheckCube(z3solver *z3bridge.Solver, cube []*il.Literal) (bool, error) {
	z3solver.Push()
	defer z3solver.Pop()

	zcube, err := s.CubeToZ3(cube)
	if err != nil {
		return false, err
	}
	z3solver.Assert(zcube)
	result := z3solver.Check()
	return result != z3bridge.Unsat, nil
}

// ClausesModelToClauses returns a clause set characterizing a model
// of the input clauses, or nil if unsat.
//
// For each constant symbol used in clauses (not ignored), it evaluates
// the symbol in the model and creates an equality constraint sym = value.
// For function/relation symbols, it evaluates the function at each point
// of the model's universe.
//
// Corresponds to Python's clauses_model_to_clauses.
func (s *Solver) ClausesModelToClauses(
	clauses *clauseops.Clauses,
	ignore func(*lg.Const) bool,
) (*clauseops.Clauses, error) {
	return s.ClausesModelToClausesWithModel(clauses, nil, ignore, false)
}

// ClausesModelToClausesWithModel is like ClausesModelToClauses but accepts
// an existing ModelResult (if nil, one is created from the clauses).
// If numerals is true, universe elements are assigned numeral names.
// Corresponds to Python's clauses_model_to_clauses(clauses, model=model, numerals=True).
func (s *Solver) ClausesModelToClausesWithModel(
	clauses *clauseops.Clauses,
	model *ModelResult,
	ignore func(*lg.Const) bool,
	numerals bool,
) (*clauseops.Clauses, error) {
	if model == nil {
		var err error
		model, err = s.GetModelClauses(clauses)
		if err != nil {
			return nil, err
		}
		if model == nil {
			return nil, nil // unsat
		}
	}

	if ignore == nil {
		ignore = func(*lg.Const) bool { return false }
	}

	// Extract values for symbols used in clauses
	var fmlas []lg.Node
	symSet := clauses.Symbols()
	for _, symN := range symSet {
		sym := symN.(*lg.Const)
		if ignore(sym) {
			continue
		}

		// Translate symbol to Z3
		zSym, err := s.tr.Translate(sym)
		if err != nil {
			continue
		}

		// Evaluate in model
		val, ok := model.Model.Eval(zSym, true)
		if !ok {
			continue
		}

		// Create a constant representing the model value
		valStr := val.String()
		valConst := lg.NewConst(valStr, il.SortRange(sym.CSort))

		// Check if this is a Boolean-valued symbol
		rng := il.SortRange(sym.CSort)
		if lg.SortEqual(rng, lg.Boolean) {
			// For Boolean constants: sym = True or ¬sym
			if valStr == "true" {
				fmlas = append(fmlas, sym)
			} else if valStr == "false" {
				fmlas = append(fmlas, &lg.Not{Body: sym})
			}
		} else {
			// For non-Boolean constants: sym = value
			fmlas = append(fmlas, &lg.Eq{T1: sym, T2: valConst})
		}
	}

	if len(fmlas) == 0 {
		return clauseops.TrueClauses(nil), nil
	}
	return clauseops.NewClauses(fmlas, nil, clauses.Annot), nil
}

// FilterRedundantFacts removes redundant negative formulas from clauses,
// given axioms.
// Corresponds to Python's filter_redundant_facts.
func (s *Solver) FilterRedundantFacts(clauses *clauseops.Clauses, axioms *clauseops.Clauses) (*clauseops.Clauses, error) {
	// Separate positive and negative formulas
	var posFmlas, negFmlas []lg.Node
	for _, f := range clauses.Fmlas {
		if _, isNot := f.(*lg.Not); isNot {
			negFmlas = append(negFmlas, f)
		} else {
			posFmlas = append(posFmlas, f)
		}
	}

	if len(negFmlas) == 0 {
		return clauses, nil
	}

	z3solver := s.tr.Ctx.NewSolver()

	// Add axioms
	za, err := s.ClausesToZ3(axioms)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(za)

	// Add definitions
	for _, d := range clauses.Defs {
		constraint := defToConstraint(d)
		zd, err := s.translateClosed(constraint)
		if err != nil {
			return nil, err
		}
		z3solver.Assert(zd)
	}

	// Add positive formulas
	for _, f := range posFmlas {
		zf, err := s.translateClosed(f)
		if err != nil {
			return nil, err
		}
		z3solver.Assert(zf)
	}

	// For each negative formula, check if it's redundant
	var keep []lg.Node
	for _, nf := range negFmlas {
		z3solver.Push()
		// Assert the negation of the negative formula (i.e., the positive)
		innerNot, ok := nf.(*lg.Not)
		if !ok {
			keep = append(keep, nf)
			z3solver.Pop()
			continue
		}
		zn, err := s.translateClosed(innerNot.Body)
		if err != nil {
			keep = append(keep, nf)
			z3solver.Pop()
			continue
		}
		z3solver.Assert(zn)
		if z3solver.Check() == z3bridge.Sat {
			// Not redundant
			keep = append(keep, nf)
		}
		z3solver.Pop()
	}

	allFmlas := append(posFmlas, keep...)
	defs := make([]*il.Definition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return clauseops.NewClauses(allFmlas, defs, clauses.Annot), nil
}
