// Ported to Go from ivy_solver.py.

package solver

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/clauseops"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/z3bridge"
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
func ConditionClauses(clauses *clauseops.Clauses, fmla lg.Expr) *clauseops.Clauses {
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
func FormulaToClauses(fmla lg.Expr) *clauseops.Clauses {
	return clauseops.FormulaToClauses(fmla, nil)
}

// BoundQuantifiersClauses bounds universal quantifiers in clauses to
// the given set of representative terms per sort.
// Corresponds to Python's bound_quantifiers_clauses.
func (s *Solver) BoundQuantifiersClauses(
	clauses *clauseops.Clauses,
	reps map[string][]lg.Expr,
	uninterpretedSorts map[string]bool,
) *clauseops.Clauses {
	bq := func(fmla lg.Expr) lg.Expr {
		vars := clauseops.VariablesAST(fmla)
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
	return clauseops.NewClauses(newFmlas, defs, clauses.Annot)
}

// RemoveDuplicatesClauses removes duplicate formulas from a clause set.
// Uses Z3 expression identity for deduplication.
// Corresponds to Python's remove_duplicates_clauses.
func (s *Solver) RemoveDuplicatesClauses(clauses *clauseops.Clauses) (*clauseops.Clauses, error) {
	seen := make(map[string]bool)
	var unique []lg.Expr
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
// Uses ModelIfNone to get a HerbrandModel, then ModelFacts + NumeralAssign +
// SubstituteConstantsClauses + FilterRedundantFacts, matching Python's
// clauses_model_to_diagram (ivy_solver.py:1448-1516).
func (s *Solver) ClausesModelToDiagram(
	clauses *clauseops.Clauses,
	ignore func(*lg.Const) bool,
	axioms *clauseops.Clauses,
) (*clauseops.Clauses, error) {
	return s.ClausesModelToDiagramFull(clauses, ignore, nil, nil, axioms, true, true, true)
}

// ClausesModelToDiagramFull is the full-featured version matching all Python parameters.
func (s *Solver) ClausesModelToDiagramFull(
	clauses *clauseops.Clauses,
	ignore func(*lg.Const) bool,
	implied *clauseops.Clauses,
	model *HerbrandModel,
	axioms *clauseops.Clauses,
	weaken bool,
	numerals bool,
	upwardClose bool,
) (*clauseops.Clauses, error) {
	if axioms == nil {
		axioms = clauseops.TrueClauses(nil)
	}
	if ignore == nil {
		ignore = func(*lg.Const) bool { return false }
	}

	// Get model
	combined := clauseops.AndClausesTyped(clauses, axioms)
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
		usedConsts := clauseops.ConstantsClauses(clauses)
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
	res = clauseops.SubstituteConstantsClauses(res, reps)

	// Filter defined skolems
	if len(clauses.DefIdx) > 0 {
		var filtered []lg.Expr
		for _, f := range res.Fmlas {
			syms := clauseops.UsedSymbolsAST(f)
			hasSkolemDef := false
			for _, c := range syms {
				if isSkolem(c.Name) {
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
		res = clauseops.NewClauses(filtered, res.Defs, res.Annot)
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
		core, err := s.UnsatCore(res, clauseops.AndClausesTyped(clauseops.TrueClauses(nil), axioms), clauses1Weak, unlikely)
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
		syms := clauseops.UsedSymbolsAST(f)
		hasIgnored := false
		for _, c := range syms {
			if ignore(c) && !repSet[c.Name] {
				hasIgnored = true
				break
			}
		}
		if !hasIgnored {
			finalFmlas = append(finalFmlas, f)
		}
	}
	res = clauseops.NewClauses(finalFmlas, res.Defs, res.Annot)

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
			ucClauses := clauseops.NewClauses(ucFmlas, nil, nil)
			res = clauseops.AndClausesTyped(res, ucClauses)
		}
	}

	return res, nil
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
func (s *Solver) SolverAdd(z3solver *z3bridge.Solver, fmla lg.Expr) error {
	zf, err := s.translateClosed(fmla)
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
func Decide(z3solver *z3bridge.Solver, assumptions ...z3bridge.Expr) (z3bridge.CheckResult, error) {
	var result z3bridge.CheckResult
	if len(assumptions) > 0 {
		result = z3solver.CheckAssumptions(assumptions)
	} else {
		result = z3solver.Check()
	}
	if result == z3bridge.Unknown {
		return result, fmt.Errorf("solver produced inconclusive result")
	}
	return result, nil
}
