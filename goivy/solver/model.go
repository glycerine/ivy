// Ported to Go from ivy_solver.py.

package solver

import (
	"fmt"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

// ModelResult holds a Z3 model and associated solver state.
type ModelResult struct {
	Solver  *z3bridge.Solver
	Model   *z3bridge.Model
	Vocab   []*lg.Const
	Context *z3bridge.Z3Context
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
func (s *Solver) GetModelClauses(clauses *module.Clauses) (*ModelResult, error) {
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
	for _, sym := range symSet {
		vocab = append(vocab, sym)
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
	Cond() *module.Clauses
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
	clauses *module.Clauses,
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
	clauses *module.Clauses,
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

	// Process final conditions (checkers).
	// Supports both incremental (push/pop) and non-incremental (fresh solver)
	// modes, matching Python's opt_incremental parameter.
	// Python: ivy_solver.py:1221-1265
	overallResult := z3bridge.Unsat
	var assumes []*module.Clauses // track assumed conditions for non-incremental replay
	if len(finalCond) > 0 {
		for _, fc := range finalCond {
			// NON-INCREMENTAL: create fresh solver before each non-assumed check.
			// Python (ivy_solver.py:1226-1230):
			//   if not opt_incremental.get():
			//       s = z3.Solver()
			//       s.add(clauses_to_z3(clauses))
			//       for fmla in assumes: s.add(clauses_to_z3(fmla))
			if !s.opts.Incremental && !fc.Assume() {
				z3solver = s.tr.Ctx.NewSolver()
				zc, err = s.ClausesToZ3(clauses)
				if err != nil {
					return nil, err
				}
				z3solver.Assert(zc)
				for _, afmla := range assumes {
					af, aerr := s.ClausesToZ3(afmla)
					if aerr != nil {
						continue
					}
					z3solver.Assert(af)
				}
			}

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
				assumes = append(assumes, cond) // track for non-incremental replay
			} else {
				// Checked condition.
				// Python (ivy_solver.py:1240-1260): pop happens AFTER Sat()/Unsat()
				// because callbacks may inspect the solver/model state.
				zCond, err := s.ClausesToZ3(cond)
				if err != nil {
					continue
				}
				if s.opts.Incremental {
					z3solver.Push()
				}
				z3solver.Assert(zCond)
				res := z3solver.Check()

				if res != z3bridge.Unsat {
					overallResult = res
					// SAT: the check condition is satisfiable (property fails)
					if fc.Sat() {
						// Checker says to continue (ignore this failure)
						overallResult = z3bridge.Unsat
						if s.opts.Incremental {
							z3solver.Pop()
						}
						continue
					}
					if s.opts.Incremental {
						z3solver.Pop()
					}
					break // stop checking
				} else {
					overallResult = z3bridge.Unsat
					fc.Unsat()
				}
				if s.opts.Incremental {
					z3solver.Pop()
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
				zsc, err := s.formulaToZ3(sc)
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
				zsc, err := s.formulaToZ3(sc)
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
	for _, sym := range symSet {
		vocab = append(vocab, sym)
	}

	return &ModelResult{
		Solver:  z3solver,
		Model:   m,
		Vocab:   vocab,
		Context: s.tr.Ctx,
	}, nil
}

// EvalFormula evaluates a formula in a model, returning true/false/unknown.
func (s *Solver) EvalFormula(model *z3bridge.Model, fmla lg.Expr) (bool, error) {
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

// CubeMemoEntry stores a cached check_cube result.
// Keeps a reference to the Z3 expression to preserve the AST ID from GC.
// Corresponds to Python's memo[fid] = (f, res) in check_cube.
type CubeMemoEntry struct {
	Expr   z3bridge.Expr // prevent GC so AST ID stays valid
	Result bool
}

// CheckCube checks if a cube (conjunction of literals) is consistent with
// the solver state. Returns true if sat.
//
// If memo is non-nil, results are cached by Z3 AST ID. When memoUnsatOnly
// is true, only UNSAT results are returned from cache (SAT results are
// rechecked). Pass nil for memo to disable caching.
//
// Corresponds to Python's check_cube (ivy_solver.py:714-733).
func (s *Solver) CheckCube(
	z3solver *z3bridge.Solver,
	cube []*il.Literal,
	memo map[uint]*CubeMemoEntry,
	memoUnsatOnly bool,
) (bool, error) {
	z3solver.Push()
	defer z3solver.Pop()

	zcube, err := s.CubeToZ3(cube)
	if err != nil {
		return false, err
	}

	// Check memo by Z3 AST ID
	// Python: fid = get_id(f); if memo is not None and fid in memo: ...
	if memo != nil {
		fid := zcube.GetId()
		if entry, ok := memo[fid]; ok {
			// Python: if (not res) or (not memo_unsat_only): return memo[fid][1]
			if !entry.Result || !memoUnsatOnly {
				return entry.Result, nil
			}
		}
	}

	z3solver.Assert(zcube)
	result := z3solver.Check()
	sat := result != z3bridge.Unsat

	// Store in memo
	// Python: memo[fid] = (f, res) -- keep reference to f to preserve id
	if memo != nil {
		fid := zcube.GetId()
		memo[fid] = &CubeMemoEntry{Expr: zcube, Result: sat}
	}

	return sat, nil
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
	clauses *module.Clauses,
	ignore func(*lg.Const) bool,
) (*module.Clauses, error) {
	return s.ClausesModelToClausesWithModel(clauses, nil, ignore, false)
}

// ClausesModelToClausesWithModel is like ClausesModelToClauses but accepts
// an existing ModelResult (if nil, one is created from the clauses).
// If numerals is true, universe elements are assigned numeral names;
// otherwise they get "__" prefix.
// Uses ModelFacts for full extraction, then applies numeral/prefix renaming
// and constant substitution.
// Corresponds to Python's clauses_model_to_clauses (ivy_solver.py:1373-1395).
func (s *Solver) ClausesModelToClausesWithModel(
	clauses *module.Clauses,
	model *ModelResult,
	ignore func(*lg.Const) bool,
	numerals bool,
) (*module.Clauses, error) {
	if ignore == nil {
		ignore = func(*lg.Const) bool { return false }
	}

	// Get a HerbrandModel
	var h *HerbrandModel
	if model != nil {
		// Build HerbrandModel from existing ModelResult
		symSet := clauses.Symbols()
		vocab := make([]*lg.Const, 0, len(symSet))
		for _, sym := range symSet {
			vocab = append(vocab, sym)
		}
		h = NewHerbrandModel(s, model.Solver, model.Model, vocab)
	} else {
		h = s.ModelIfNone(clauses, nil, nil)
	}
	if h == nil {
		return nil, nil // unsat
	}

	// Extract model facts
	res := ModelFacts(h, ignore, clauses, false)

	// Build substitution map using structural keys
	subs := make(map[lg.NodeKey]lg.Expr)
	if numerals {
		na := NumeralAssignWithClauses(h, res)
		for elemName, numName := range na {
			for _, sort := range h.Sorts() {
				for _, c := range h.SortUniverse(sort) {
					if c.Name == elemName {
						subs[lg.Key(c)] = lg.NewConst(numName, c.CSort)
					}
				}
			}
		}
	} else {
		// Prefix with "__"
		for _, sort := range h.Sorts() {
			for _, c := range h.SortUniverse(sort) {
				subs[lg.Key(c)] = lg.NewConst("__"+c.Name, c.CSort)
			}
		}
	}

	res = module.SubstituteConstantsClauses(res, subs)
	return res, nil
}

// FilterRedundantFacts removes redundant negative formulas from clauses,
// given axioms.
// Corresponds to Python's filter_redundant_facts.
func (s *Solver) FilterRedundantFacts(clauses *module.Clauses, axioms *module.Clauses) (*module.Clauses, error) {
	// Separate positive and negative formulas.
	// Python: pos_fmlas = [f for f in fmlas if not isinstance(f, ivy_logic.Not)]
	var posFmlas, negFmlas []lg.Expr
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

	ctx := s.tr.Ctx
	z3solver := ctx.NewSolver()

	// Add axioms
	za, err := s.ClausesToZ3(axioms)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(za)

	// Add definitions
	// Python filter_redundant_facts line 1494: formula_to_z3(d.to_constraint())
	for _, d := range clauses.Defs {
		constraint := il.DefinitionToConstraint(d)
		zd, err := s.formulaToZ3(constraint)
		if err != nil {
			return nil, err
		}
		z3solver.Assert(zd)
	}

	// Add positive formulas
	for _, f := range posFmlas {
		zf, err := s.formulaToZ3(f)
		if err != nil {
			return nil, err
		}
		z3solver.Assert(zf)
	}

	// Create activation literals and gated negatives.
	// Python: alits = [z3.Const("__c%s" % n, z3.BoolSort()) for n,c in enumerate(neg_fmlas)]
	//         cc = [z3.Or(z3.Not(a), z3.Not(formula_to_z3(c))) for a,c in zip(alits,neg_fmlas)]
	alits := make([]z3bridge.Expr, len(negFmlas))
	for i, nf := range negFmlas {
		alit := ctx.Const(fmt.Sprintf("__c%d", i), ctx.BoolSort())
		alits[i] = alit
		zn, err := s.formulaToZ3(nf)
		if err != nil {
			continue
		}
		// Or(Not(alit), Not(neg_fmla)) means: if alit is true, neg_fmla must be false
		z3solver.Assert(ctx.Or(ctx.Not(alit), ctx.Not(zn)))
	}

	// Test each negative formula via assumptions.
	// Python: if decide(s2, [alit]) == z3.sat: keep.append(fmla)
	var keep []lg.Expr
	for i, fmla := range negFmlas {
		if z3solver.CheckAssumptions([]z3bridge.Expr{alits[i]}) == z3bridge.Sat {
			keep = append(keep, fmla)
		}
	}

	allFmlas := append(posFmlas, keep...)
	defs := make([]*il.Definition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return module.NewClauses(allFmlas, defs, clauses.Annot), nil
}
