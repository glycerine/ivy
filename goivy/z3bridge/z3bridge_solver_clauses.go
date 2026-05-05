// Ported to Go from ivy_solver.py.

package z3bridge

import (
	"fmt"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// ClausesImplyList checks whether clauses1 implies each element of clauses2List.
// Returns a bool slice with the result for each element.
// Corresponds to Python's clauses_imply_list.
func (s *Solver) ClausesImplyList(clauses1 *module.Clauses, clauses2List []*module.Clauses) ([]bool, error) {
	xtracer.Trace("ivy_solver.py:1036 clauses_imply_list() ENTER")
	z3solver := s.tr.Ctx.NewZ3Solver()

	z1, err := s.ClausesToZ3(clauses1)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(z1)

	results := make([]bool, len(clauses2List))
	for i, clauses2 := range clauses2List {
		dual := module.NegateClauses(clauses2)
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
func ConditionClauses(clauses *module.Clauses, fmla lg.Expr) *module.Clauses {
	return module.ConditionClauses(clauses, fmla)
}

// TrueClauses returns a trivially true clause set.
func TrueClauses() *module.Clauses {
	return module.TrueClauses(nil)
}

// FalseClauses returns a trivially false clause set.
func FalseClauses() *module.Clauses {
	return module.FalseClauses(nil)
}

// AndClauses computes the conjunction of multiple clause sets.
func AndClauses(args ...*module.Clauses) *module.Clauses {
	return module.AndClausesTyped(args...)
}

// FormulaToClauses wraps a formula as a Clauses set.
func FormulaToClauses(fmla lg.Expr) *module.Clauses {
	return module.FormulaToClauses(fmla, nil)
}

// BoundQuantifiersClauses bounds universal quantifiers in clauses to
// the given set of representative terms per sort.
// Corresponds to Python's bound_quantifiers_clauses.
func (s *Solver) BoundQuantifiersClauses(
	clauses *module.Clauses,
	reps map[string][]lg.Expr,
	uninterpretedSorts map[string]bool,
) *module.Clauses {
	xtracer.Trace("ivy_solver.py:1541 bound_quantifiers_clauses() ENTER")
	bq := func(fmla lg.Expr) lg.Expr {
		vars := module.VariablesAST(fmla)
		if len(vars) == 0 {
			return fmla
		}
		var constraints []lg.Expr
		for _, v := range vars {
			sortName := il.SortName(v.VSort)
			if uninterpretedSorts != nil && !uninterpretedSorts[sortName] {
				continue
			}
			terms, ok := reps[sortName]
			if !ok || len(terms) == 0 {
				continue
			}
			eqs := make([]lg.Expr, len(terms))
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

	newFmlas := make([]lg.Expr, len(clauses.Fmlas))
	for i, f := range clauses.Fmlas {
		newFmlas[i] = bq(f)
	}
	defs := make([]*il.Definition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return module.NewClauses(newFmlas, defs, clauses.Annot)
}

// RemoveDuplicatesClauses removes duplicate formulas from a clause set.
// Uses Z3 expression identity for deduplication.
// Corresponds to Python's remove_duplicates_clauses.
func (s *Solver) RemoveDuplicatesClauses(clauses *module.Clauses) (*module.Clauses, error) {
	xtracer.Trace("ivy_solver.py:1122 remove_duplicates_clauses() ENTER")
	seen := make(map[string]bool)
	var unique []lg.Expr
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
	defs := make([]*il.Definition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return module.NewClauses(unique, defs, clauses.Annot), nil
}

// ClausesModelToDiagram returns a diagram (clause set) of a model of clauses.
// Uses ModelIfNone to get a HerbrandModel, then ModelFacts + NumeralAssign +
// SubstituteConstantsClauses + FilterRedundantFacts, matching Python's
// clauses_model_to_diagram (ivy_solver.py:1448-1516).
func (s *Solver) ClausesModelToDiagram(
	clauses *module.Clauses,
	ignore func(*lg.Const) bool,
	axioms *module.Clauses,
) (*module.Clauses, error) {
	return s.ClausesModelToDiagramFull(clauses, ignore, nil, nil, axioms, true, true, true)
}

// ClausesModelToDiagramFull is the full-featured version matching all Python parameters.
func (s *Solver) ClausesModelToDiagramFull(
	clauses *module.Clauses,
	ignore func(*lg.Const) bool,
	implied *module.Clauses,
	model *HerbrandModel,
	axioms *module.Clauses,
	weaken bool,
	numerals bool,
	upwardClose bool,
) (*module.Clauses, error) {
	xtracer.Trace("ivy_solver.py:1594 clauses_model_to_diagram() ENTER")
	if axioms == nil {
		axioms = module.TrueClauses(nil)
	}
	if ignore == nil {
		ignore = func(*lg.Const) bool { return false }
	}

	// Get model
	combined := module.AndClausesTyped(clauses, axioms)
	h := s.ModelIfNone(combined, implied, model)
	if h == nil {
		return nil, nil
	}

	// Extract model facts (always use upclose=true for diagrams, matching Python)
	noIgnore := func(*lg.Const) bool { return false }
	res := ModelFacts(h, noIgnore, clauses, true)

	// Find representative elements via numeral assignment or skolem prefix
	var reps map[lg.NodeKey]lg.Expr
	if numerals {
		na := NumeralAssignWithClauses(h, res)
		reps = make(map[lg.NodeKey]lg.Expr, len(na))
		for elemName, numName := range na {
			// Find the sort from the model
			for _, sort := range h.Sorts() {
				for _, c := range h.SortUniverse(sort) {
					if c.Name == elemName {
						reps[lg.Key(c)] = lg.NewConst(numName, c.CSort)
					}
				}
			}
		}
	} else {
		reps = make(map[lg.NodeKey]lg.Expr)
		// Use constants from clauses as reps where possible
		usedConsts := module.ConstantsClauses(clauses)
		for _, c := range usedConsts {
			mc := h.EvalConstant(c)
			if mc != nil {
				if existing, ok := reps[lg.Key(mc)]; ok {
					// Prefer non-skolem reps
					if existSym, ok2 := existing.(*lg.Const); ok2 {
						if isSkolem(existSym.Name) && !isSkolem(c.Name) {
							reps[lg.Key(mc)] = c
						}
					}
				} else {
					reps[lg.Key(mc)] = c
				}
			}
		}
		// Skolemize any remaining unassigned elements
		for _, sort := range h.Sorts() {
			for _, e := range h.SortUniverse(sort) {
				if _, ok := reps[lg.Key(e)]; !ok {
					reps[lg.Key(e)] = lg.NewConst("__"+e.Name, e.CSort)
				}
			}
		}
	}

	// Substitute constants
	res = module.SubstituteConstantsClauses(res, reps)

	// Filter defined skolems
	if len(clauses.DefIdx) > 0 {
		var filtered []lg.Expr
		for _, f := range res.Fmlas {
			syms := module.UsedSymbolsAST(f)
			hasSkolemDef := false
			for _, c := range syms.All() {
				if isSkolem(lg.ExprName(c)) {
					if _, inIdx := clauses.DefIdx[lg.Key(c)]; inIdx {
						hasSkolemDef = true
						break
					}
				}
			}
			if !hasSkolemDef {
				filtered = append(filtered, f)
			}
		}
		res = module.NewClauses(filtered, res.Defs, res.Annot)
	}

	// Filter redundant facts
	var err error
	res, err = s.FilterRedundantFacts(res, axioms)
	if err != nil {
		return res, nil
	}

	// Weakening via unsat core
	if weaken {
		unlikely := func(fmla lg.Expr) bool {
			if eq, ok := fmla.(*lg.Eq); ok {
				if _, isSym := eq.T1.(*lg.Const); isSym {
					return true
				}
			}
			return false
		}
		repTerms := make(map[string][]lg.Expr)
		for _, sort := range h.Sorts() {
			sortName := il.SortName(sort)
			for _, c := range h.SortUniverse(sort) {
				if rep, ok := reps[lg.Key(c)]; ok {
					repTerms[sortName] = append(repTerms[sortName], rep)
				}
			}
		}
		clauses1Weak := s.BoundQuantifiersClauses(clauses, repTerms, nil)
		core, err := s.UnsatCore(res, module.AndClausesTyped(module.TrueClauses(nil), axioms), clauses1Weak, unlikely)
		if err == nil && core != nil {
			res = core
		}
	}

	// Filter out non-rep skolems
	repSet := make(map[string]bool)
	for _, v := range reps {
		if sym, ok := v.(*lg.Const); ok {
			repSet[sym.Name] = true
		}
	}
	var finalFmlas []lg.Expr
	for _, f := range res.Fmlas {
		syms := module.UsedSymbolsAST(f)
		hasIgnored := false
		for _, c := range syms.All() {
			if cc, ok := c.(*lg.Const); ok && ignore(cc) && !repSet[lg.ExprName(c)] {
				hasIgnored = true
				break
			}
		}
		if !hasIgnored {
			finalFmlas = append(finalFmlas, f)
		}
	}
	res = module.NewClauses(finalFmlas, res.Defs, res.Annot)

	// If not upward-closing, add universe closure constraints
	if !upwardClose {
		var ucFmlas []lg.Expr
		for _, sort := range h.Sorts() {
			x, _ := lg.NewVariable("X", sort)
			var eqs []lg.Expr
			for _, c := range h.SortUniverse(sort) {
				if rep, ok := reps[lg.Key(c)]; ok {
					eqs = append(eqs, &lg.Eq{T1: x, T2: rep})
				}
			}
			if len(eqs) > 0 {
				ucFmlas = append(ucFmlas, &lg.Or{Terms: eqs})
			}
		}
		if len(ucFmlas) > 0 {
			ucClauses := module.NewClauses(ucFmlas, nil, nil)
			res = module.AndClausesTyped(res, ucClauses)
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
func (s *Solver) AddClauses(z3solver *Z3Solver, clauses *module.Clauses) error {
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
func (s *Solver) SolverAdd(z3solver *Z3Solver, fmla lg.Expr) error {
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
func Decide(z3solver *Z3Solver, assumptions ...Expr) (CheckResult, error) {
	xtracer.Trace("ivy_solver.py:1302 decide() ENTER")
	var result CheckResult
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
