// Ported to Go from ivy_solver.py.

package goivy

import (
	"fmt"
	"os"

	"github.com/glycerine/ivy/goivy/smt"
	"github.com/glycerine/ivy/goivy/xtracer"
)

func (s *Solver) decideZ3(z3solver *smt.Z3Solver) (smt.Z3CheckResult, error) {
	result := s.checkZ3(z3solver)
	if result == smt.Unknown {
		fmt.Println(z3solver.ToSMT2())
		return result, &IvyError{Msg: "Solver produced inconclusive result"}
	}
	return result, nil
}

// ModelResult holds a Z3 model and associated solver state.
type ModelResult struct {
	Solver  *smt.Z3Solver
	Model   *smt.Model
	Vocab   []*Const
	Context *smt.Z3Context
}

// SMTLIBBaseSolver is a persistent solver containing an SMT-LIB base
// assertion. Generated action generators use this to mirror ivy2cpp's
// constructor-loaded Z3 generator plus per-generate push/pop constraints.
type SMTLIBBaseSolver struct {
	Solver *smt.Z3Solver
}

// SoftAssumptionLogger observes the same soft-assumption events that the
// generated C++ ivy_z3_gen runtime writes to modelfile. Kind is "add" for a
// new soft predicate/assumption literal, "begin" for the solver state before
// the loop, "check" for each solver check, "sat" for the solver state before
// returning a SAT model, and "delete" for an unsat-core pruning decision.
type SoftAssumptionLogger func(kind, pred, alit string, core []string, toDelete string)

func clausesHaveHardContent(clauses *Clauses) bool {
	return clauses != nil && (len(clauses.Fmlas) != 0 || len(clauses.Defs) != 0)
}

func logSoftAssumptionAdd(log SoftAssumptionLogger, pred, alit smt.Z3Expr) {
	if log == nil {
		return
	}
	log("add", pred.String(), alit.String(), nil, "")
}

func logSoftAssumptionCheck(log SoftAssumptionLogger, z3solver *smt.Z3Solver, assumptions []smt.Z3Expr) {
	if log == nil {
		return
	}
	alits := make([]string, 0, len(assumptions))
	for _, expr := range assumptions {
		alits = append(alits, expr.String())
	}
	log("check", z3solver.String(), "", alits, "")
}

func logSoftAssumptionBegin(log SoftAssumptionLogger, z3solver *smt.Z3Solver) {
	if log == nil {
		return
	}
	log("begin", z3solver.String(), "", nil, "")
}

func logSoftAssumptionSat(log SoftAssumptionLogger, z3solver *smt.Z3Solver) {
	if log == nil {
		return
	}
	log("sat", z3solver.String(), "", nil, "")
}

func logSoftAssumptionDeletion(log SoftAssumptionLogger, core []smt.Z3Expr, toDelete smt.Z3Expr) {
	if log == nil {
		return
	}
	coreStrings := make([]string, 0, len(core))
	for _, expr := range core {
		coreStrings = append(coreStrings, expr.String())
	}
	log("delete", "", "", coreStrings, toDelete.String())
}

// Eval evaluates a Z3 expression in the model with completion.
func (mr *ModelResult) Eval(e smt.Z3Expr) (smt.Z3Expr, bool) {
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
func (s *Solver) GetModelClauses(clauses *Clauses) (*ModelResult, error) {
	xtracer.Trace("ivy_solver.py:1177 get_model_clauses() ENTER")
	z3solver := s.newZ3Solver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	result := s.checkZ3(z3solver)
	if result == smt.Unsat {
		return nil, nil // unsatisfiable
	}

	m := z3solver.Model()
	if m == nil {
		return nil, fmt.Errorf("solver returned sat but no model")
	}

	// Collect vocabulary from clauses
	symSet := clauses.Symbols()
	vocab := make([]*Const, 0, symSet.Len())
	for _, sym := range symSet.All() {
		if c, ok := sym.(*Const); ok {
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

// GetModelClausesWithSoftAssumptions checks clauses together with soft formulas.
// Each soft formula is guarded by a fresh assumption literal. If the assumptions
// are unsatisfiable, one literal from the unsat core is removed using choose,
// mirroring the generated C++ ivy_z3_gen::solve loop.
func (s *Solver) GetModelClausesWithSoftAssumptions(clauses *Clauses, soft []Expr, choose func(int) int) (*ModelResult, error) {
	return s.GetModelClausesWithSoftAssumptionsLogged(clauses, soft, choose, nil)
}

// GetModelClausesWithSoftAssumptionsLogged is
// GetModelClausesWithSoftAssumptions plus an optional modelfile logger for
// C++-style unsat-core pruning transcript lines.
func (s *Solver) GetModelClausesWithSoftAssumptionsLogged(clauses *Clauses, soft []Expr, choose func(int) int, log SoftAssumptionLogger) (*ModelResult, error) {
	xtracer.Trace("ivy_solver.py:get_model_clauses_with_soft_assumptions ENTER")
	debugSoft := os.Getenv("GOIVY_DEBUG_SOFT_SOLVER") != ""
	z3solver := s.newZ3Solver()
	if clausesHaveHardContent(clauses) {
		zc, err := s.ClausesToZ3(clauses)
		if err != nil {
			return nil, err
		}
		z3solver.Assert(zc)
	}
	var assumptions []smt.Z3Expr
	ctx := s.tr.Ctx
	for i, f := range soft {
		if f == nil {
			continue
		}
		zf, err := s.tr.Translate(f)
		if err != nil {
			return nil, err
		}
		alit := ctx.Const(fmt.Sprintf("alit:%d", i), ctx.BoolSort())
		z3solver.Assert(ctx.Or(ctx.Not(alit), zf))
		logSoftAssumptionAdd(log, zf, alit)
		assumptions = append(assumptions, alit)
	}
	if debugSoft {
		fmt.Fprintf(os.Stderr, "soft-solver clauses-start soft=%d\n", len(assumptions))
	}
	logSoftAssumptionBegin(log, z3solver)
	for {
		logSoftAssumptionCheck(log, z3solver, assumptions)
		result := s.checkZ3Assumptions(z3solver, assumptions)
		if result == smt.Sat {
			if debugSoft {
				fmt.Fprintf(os.Stderr, "soft-solver clauses-sat remaining=%d\n", len(assumptions))
			}
			break
		}
		if result == smt.Unknown {
			return nil, &IvyError{Msg: "Solver produced inconclusive result"}
		}
		core := z3solver.UnsatCore()
		if len(core) == 0 {
			if debugSoft {
				fmt.Fprintf(os.Stderr, "soft-solver clauses-unsat-empty-core remaining=%d\n", len(assumptions))
			}
			return nil, nil
		}
		idx := 0
		if choose != nil {
			idx = choose(len(core))
		}
		if idx < 0 || idx >= len(core) {
			idx = 0
		}
		toDelete := core[idx]
		if debugSoft && len(soft)-len(assumptions) < 20 {
			fmt.Fprintf(os.Stderr, "soft-solver clauses-delete[%d] core=%d idx=%d alit=%s\n", len(soft)-len(assumptions), len(core), idx, toDelete.String())
		}
		logSoftAssumptionDeletion(log, core, toDelete)
		for i, alit := range assumptions {
			if alit.Equal(toDelete) {
				assumptions[i] = assumptions[len(assumptions)-1]
				assumptions = assumptions[:len(assumptions)-1]
				break
			}
		}
	}
	logSoftAssumptionSat(log, z3solver)
	m := z3solver.Model()
	if m == nil {
		return nil, fmt.Errorf("solver returned sat but no model")
	}
	symSet := clauses.Symbols()
	for _, f := range soft {
		for k, sym := range UsedSymbolsAst(f).All() {
			symSet.Set(k, sym)
		}
	}
	vocab := make([]*Const, 0, symSet.Len())
	for _, sym := range symSet.All() {
		if c, ok := sym.(*Const); ok {
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

// NewSMTLIBBaseSolver parses and asserts an SMT-LIB2 base assertion into a
// reusable solver. Callers can then push temporary hard/soft constraints on top
// of this base without reparsing the static assertion.
func (s *Solver) NewSMTLIBBaseSolver(smtlib string) (*SMTLIBBaseSolver, error) {
	z3solver := s.newZ3Solver()
	if smtlib != "" {
		base, err := s.ParseSMTLIB2Assertion(smtlib)
		if err != nil {
			return nil, err
		}
		z3solver.Assert(base)
	}
	return &SMTLIBBaseSolver{Solver: z3solver}, nil
}

// GetModelSMTLIBWithSoftAssumptions checks an SMT-LIB2 assertion together
// with soft formulas. The SMT-LIB path mirrors the generated C++ tester
// runtime, which preloads declarations and calls Z3_parse_smtlib2_string.
func (s *Solver) GetModelSMTLIBWithSoftAssumptions(smtlib string, soft []Expr, choose func(int) int) (*ModelResult, error) {
	return s.GetModelSMTLIBWithSoftAssumptionsLogged(smtlib, soft, choose, nil)
}

// GetModelSMTLIBWithSoftAssumptionsLogged is
// GetModelSMTLIBWithSoftAssumptions plus an optional modelfile logger for
// C++-style unsat-core pruning transcript lines.
func (s *Solver) GetModelSMTLIBWithSoftAssumptionsLogged(smtlib string, soft []Expr, choose func(int) int, log SoftAssumptionLogger) (*ModelResult, error) {
	xtracer.Trace("ivy_solver.py:get_model_smtlib_with_soft_assumptions ENTER")
	debugSoft := os.Getenv("GOIVY_DEBUG_SOFT_SOLVER") != ""
	z3solver := s.newZ3Solver()
	if smtlib != "" {
		base, err := s.ParseSMTLIB2Assertion(smtlib)
		if err != nil {
			if debugSoft {
				fmt.Fprintf(os.Stderr, "soft-solver smt-parse-error: %v\n", err)
			}
			return nil, err
		}
		z3solver.Assert(base)
	}
	if debugSoft {
		fmt.Fprintf(os.Stderr, "soft-solver smtlib-start soft=%d\n", len(soft))
	}
	var assumptions []smt.Z3Expr
	ctx := s.tr.Ctx
	for i, f := range soft {
		if f == nil {
			continue
		}
		zf, err := s.tr.Translate(f)
		if err != nil {
			return nil, err
		}
		alit := ctx.Const(fmt.Sprintf("alit:%d", i), ctx.BoolSort())
		z3solver.Assert(ctx.Or(ctx.Not(alit), zf))
		logSoftAssumptionAdd(log, zf, alit)
		assumptions = append(assumptions, alit)
	}
	logSoftAssumptionBegin(log, z3solver)
	for {
		logSoftAssumptionCheck(log, z3solver, assumptions)
		result := s.checkZ3Assumptions(z3solver, assumptions)
		if result == smt.Sat {
			if debugSoft {
				fmt.Fprintf(os.Stderr, "soft-solver smtlib-sat remaining=%d\n", len(assumptions))
			}
			break
		}
		if result == smt.Unknown {
			return nil, &IvyError{Msg: "Solver produced inconclusive result"}
		}
		core := z3solver.UnsatCore()
		if len(core) == 0 {
			return nil, nil
		}
		idx := 0
		if choose != nil {
			idx = choose(len(core))
		}
		if idx < 0 || idx >= len(core) {
			idx = 0
		}
		toDelete := core[idx]
		if debugSoft && len(soft)-len(assumptions) < 20 {
			fmt.Fprintf(os.Stderr, "soft-solver smtlib-delete[%d] core=%d idx=%d alit=%s\n", len(soft)-len(assumptions), len(core), idx, toDelete.String())
		}
		logSoftAssumptionDeletion(log, core, toDelete)
		for i, alit := range assumptions {
			if alit.Equal(toDelete) {
				assumptions[i] = assumptions[len(assumptions)-1]
				assumptions = assumptions[:len(assumptions)-1]
				break
			}
		}
	}
	logSoftAssumptionSat(log, z3solver)
	m := z3solver.Model()
	if m == nil {
		return nil, fmt.Errorf("solver returned sat but no model")
	}
	symSet := make(map[NodeKey]Expr)
	if s.sig != nil {
		for _, sym := range s.sig.AllSymbols() {
			if sym != nil {
				symSet[Key(sym)] = sym
			}
		}
	}
	for _, f := range soft {
		for k, sym := range UsedSymbolsAst(f).All() {
			symSet[k] = sym
		}
	}
	vocab := make([]*Const, 0, len(symSet))
	for _, sym := range symSet {
		if c, ok := sym.(*Const); ok {
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

// GetModelSMTLIBBaseClausesWithSoftAssumptions checks a reusable SMT-LIB base
// solver under a temporary push frame containing additional hard clauses and
// soft formulas.
func (s *Solver) GetModelSMTLIBBaseClausesWithSoftAssumptions(base *SMTLIBBaseSolver, clauses *Clauses, soft []Expr, choose func(int) int) (*ModelResult, error) {
	return s.GetModelSMTLIBBaseClausesWithSoftAssumptionsLogged(base, clauses, soft, choose, nil)
}

// GetModelSMTLIBBaseClausesWithSoftAssumptionsLogged is
// GetModelSMTLIBBaseClausesWithSoftAssumptions plus an optional modelfile
// logger for C++-style unsat-core pruning transcript lines.
func (s *Solver) GetModelSMTLIBBaseClausesWithSoftAssumptionsLogged(base *SMTLIBBaseSolver, clauses *Clauses, soft []Expr, choose func(int) int, log SoftAssumptionLogger) (*ModelResult, error) {
	xtracer.Trace("ivy_solver.py:get_model_smtlib_base_clauses_with_soft_assumptions ENTER")
	debugSoft := os.Getenv("GOIVY_DEBUG_SOFT_SOLVER") != ""
	if base == nil || base.Solver == nil {
		return nil, fmt.Errorf("nil SMT-LIB base solver")
	}
	z3solver := base.Solver
	z3solver.Push()
	defer z3solver.Pop()
	if clausesHaveHardContent(clauses) {
		zc, err := s.ClausesToZ3(clauses)
		if err != nil {
			return nil, err
		}
		z3solver.Assert(zc)
	}
	if debugSoft {
		fmt.Fprintf(os.Stderr, "soft-solver smtlib-base-clauses-start soft=%d\n", len(soft))
	}
	var assumptions []smt.Z3Expr
	ctx := s.tr.Ctx
	for i, f := range soft {
		if f == nil {
			continue
		}
		zf, err := s.tr.Translate(f)
		if err != nil {
			return nil, err
		}
		alit := ctx.Const(fmt.Sprintf("alit:%d", i), ctx.BoolSort())
		z3solver.Assert(ctx.Or(ctx.Not(alit), zf))
		logSoftAssumptionAdd(log, zf, alit)
		assumptions = append(assumptions, alit)
	}
	logSoftAssumptionBegin(log, z3solver)
	for {
		logSoftAssumptionCheck(log, z3solver, assumptions)
		result := s.checkZ3Assumptions(z3solver, assumptions)
		if result == smt.Sat {
			if debugSoft {
				fmt.Fprintf(os.Stderr, "soft-solver smtlib-base-clauses-sat remaining=%d\n", len(assumptions))
			}
			break
		}
		if result == smt.Unknown {
			return nil, &IvyError{Msg: "Solver produced inconclusive result"}
		}
		core := z3solver.UnsatCore()
		if len(core) == 0 {
			return nil, nil
		}
		idx := 0
		if choose != nil {
			idx = choose(len(core))
		}
		if idx < 0 || idx >= len(core) {
			idx = 0
		}
		toDelete := core[idx]
		if debugSoft && len(soft)-len(assumptions) < 20 {
			fmt.Fprintf(os.Stderr, "soft-solver smtlib-base-clauses-delete[%d] core=%d idx=%d alit=%s\n", len(soft)-len(assumptions), len(core), idx, toDelete.String())
		}
		logSoftAssumptionDeletion(log, core, toDelete)
		for i, alit := range assumptions {
			if alit.Equal(toDelete) {
				assumptions[i] = assumptions[len(assumptions)-1]
				assumptions = assumptions[:len(assumptions)-1]
				break
			}
		}
	}
	logSoftAssumptionSat(log, z3solver)
	m := z3solver.Model()
	if m == nil {
		return nil, fmt.Errorf("solver returned sat but no model")
	}
	symSet := make(map[NodeKey]Expr)
	if s.sig != nil {
		for _, sym := range s.sig.AllSymbols() {
			if sym != nil {
				symSet[Key(sym)] = sym
			}
		}
	}
	if clausesHaveHardContent(clauses) {
		for _, sym := range clauses.Symbols().All() {
			symSet[Key(sym)] = sym
		}
	}
	for _, f := range soft {
		for k, sym := range UsedSymbolsAst(f).All() {
			symSet[k] = sym
		}
	}
	vocab := make([]*Const, 0, len(symSet))
	for _, sym := range symSet {
		if c, ok := sym.(*Const); ok {
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

// GetModelSMTLIBClausesWithSoftAssumptions checks an SMT-LIB2 assertion and
// additional hard clauses together with soft formulas. This is the generated
// action-generator analogue of Python/C++'s preloaded SMT-LIB precondition plus
// per-generate emit_set constraints.
func (s *Solver) GetModelSMTLIBClausesWithSoftAssumptions(smtlib string, clauses *Clauses, soft []Expr, choose func(int) int) (*ModelResult, error) {
	return s.GetModelSMTLIBClausesWithSoftAssumptionsLogged(smtlib, clauses, soft, choose, nil)
}

// GetModelSMTLIBClausesWithSoftAssumptionsLogged is
// GetModelSMTLIBClausesWithSoftAssumptions plus an optional modelfile logger
// for C++-style unsat-core pruning transcript lines.
func (s *Solver) GetModelSMTLIBClausesWithSoftAssumptionsLogged(smtlib string, clauses *Clauses, soft []Expr, choose func(int) int, log SoftAssumptionLogger) (*ModelResult, error) {
	xtracer.Trace("ivy_solver.py:get_model_smtlib_clauses_with_soft_assumptions ENTER")
	debugSoft := os.Getenv("GOIVY_DEBUG_SOFT_SOLVER") != ""
	z3solver := s.newZ3Solver()
	if smtlib != "" {
		base, err := s.ParseSMTLIB2Assertion(smtlib)
		if err != nil {
			if debugSoft {
				fmt.Fprintf(os.Stderr, "soft-solver smt-clauses-parse-error: %v\n", err)
			}
			return nil, err
		}
		z3solver.Assert(base)
	}
	if clausesHaveHardContent(clauses) {
		zc, err := s.ClausesToZ3(clauses)
		if err != nil {
			return nil, err
		}
		z3solver.Assert(zc)
	}
	if debugSoft {
		fmt.Fprintf(os.Stderr, "soft-solver smtlib-clauses-start soft=%d\n", len(soft))
	}
	var assumptions []smt.Z3Expr
	ctx := s.tr.Ctx
	for i, f := range soft {
		if f == nil {
			continue
		}
		zf, err := s.tr.Translate(f)
		if err != nil {
			return nil, err
		}
		alit := ctx.Const(fmt.Sprintf("alit:%d", i), ctx.BoolSort())
		z3solver.Assert(ctx.Or(ctx.Not(alit), zf))
		logSoftAssumptionAdd(log, zf, alit)
		assumptions = append(assumptions, alit)
	}
	logSoftAssumptionBegin(log, z3solver)
	for {
		logSoftAssumptionCheck(log, z3solver, assumptions)
		result := s.checkZ3Assumptions(z3solver, assumptions)
		if result == smt.Sat {
			if debugSoft {
				fmt.Fprintf(os.Stderr, "soft-solver smtlib-clauses-sat remaining=%d\n", len(assumptions))
			}
			break
		}
		if result == smt.Unknown {
			return nil, &IvyError{Msg: "Solver produced inconclusive result"}
		}
		core := z3solver.UnsatCore()
		if len(core) == 0 {
			return nil, nil
		}
		idx := 0
		if choose != nil {
			idx = choose(len(core))
		}
		if idx < 0 || idx >= len(core) {
			idx = 0
		}
		toDelete := core[idx]
		if debugSoft && len(soft)-len(assumptions) < 20 {
			fmt.Fprintf(os.Stderr, "soft-solver smtlib-clauses-delete[%d] core=%d idx=%d alit=%s\n", len(soft)-len(assumptions), len(core), idx, toDelete.String())
		}
		logSoftAssumptionDeletion(log, core, toDelete)
		for i, alit := range assumptions {
			if alit.Equal(toDelete) {
				assumptions[i] = assumptions[len(assumptions)-1]
				assumptions = assumptions[:len(assumptions)-1]
				break
			}
		}
	}
	logSoftAssumptionSat(log, z3solver)
	m := z3solver.Model()
	if m == nil {
		return nil, fmt.Errorf("solver returned sat but no model")
	}
	symSet := make(map[NodeKey]Expr)
	if s.sig != nil {
		for _, sym := range s.sig.AllSymbols() {
			if sym != nil {
				symSet[Key(sym)] = sym
			}
		}
	}
	if clauses != nil {
		for _, sym := range clauses.Symbols().All() {
			symSet[Key(sym)] = sym
		}
	}
	for _, f := range soft {
		for k, sym := range UsedSymbolsAst(f).All() {
			symSet[k] = sym
		}
	}
	vocab := make([]*Const, 0, len(symSet))
	for _, sym := range symSet {
		if c, ok := sym.(*Const); ok {
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
func (s *Solver) ModelValues(model *smt.Model, syms []*Const) (map[string]smt.Z3Expr, error) {
	result := make(map[string]smt.Z3Expr, len(syms))
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
	Cond() *Clauses
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
	clauses *Clauses,
	sortsToMinimize []Sort, relationsToMinimize []*Const,
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
	clauses *Clauses,
	sortsToMinimize []Sort, relationsToMinimize []*Const,
	finalCond []FinalCond,
	shrink bool,
) (*ModelResult, error) {
	xtracer.Trace("ivy_solver.py:1339 get_small_model() ENTER shrink=%v", shrink)
	s.showVCsBase(clauses)

	z3solver := s.newZ3Solver()
	zc, err := s.ClausesToZ3(clauses)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(zc)

	// Process final conditions (checkers).
	// Supports both incremental (push/pop) and non-incremental (fresh solver)
	// modes, matching Python's opt_incremental parameter.
	// Python: ivy_solver.py:1221-1265
	overallResult := smt.Unsat
	var assumes []*Clauses // track assumed conditions for non-incremental replay
	// Python ivy_solver.py:1469-1512:
	//   if final_cond is not None:    Go: finalCond != nil
	//       for fc in final_cond: ... (empty list means no-op, returns nil)
	//   else:                          Go: finalCond == nil
	//       res = decide(s)
	// nil finalCond matches Python None: caller did not request checker-driven
	// verification, so we ask the solver if the state is satisfiable.
	// Non-nil but empty finalCond matches Python []: caller had no checkers
	// after filtering - there is nothing to check, so we return nil
	// without calling decide(), exactly like Python.
	if finalCond != nil {
		for _, fc := range finalCond {
			// NON-INCREMENTAL: create a fresh solver before each final
			// condition, replaying prior assumptions.
			// Python (ivy_solver.py:1226-1230):
			//   if not opt_incremental.get():
			//       s = z3.Solver()
			//       s.add(clauses_to_z3(clauses))
			//       for fmla in assumes: s.add(clauses_to_z3(fmla))
			if !s.opts.Incremental {
				z3solver = s.newZ3Solver()
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
			if fc.Assume() {
				// Assumed condition: add permanently to solver
				cond := fc.Cond()
				if cond == nil {
					continue
				}
				s.showVCsFinalCond("assume", cond)
				zCond, err := s.ClausesToZ3(cond)
				if err != nil {
					continue
				}
				z3solver.Assert(zCond)
				assumes = append(assumes, fc.Cond()) // track for non-incremental replay
			} else {
				// Checked condition.
				// Python (ivy_solver.py:1240-1260): pop happens AFTER Sat()/Unsat()
				// because callbacks may inspect the solver/model state.
				cond := fc.Cond()
				if cond == nil {
					continue
				}
				s.showVCsFinalCond("assert", cond)
				zCond, err := s.ClausesToZ3(cond)
				if err != nil {
					continue
				}
				if s.opts.Incremental {
					z3solver.Push()
				}
				z3solver.Assert(zCond)
				// Python ivy_solver.py:1437 calls decide(s) which emits the
				// decide() ENTER trace before invoking solver.check.
				xtracer.Trace("ivy_solver.py:1302 decide() ENTER")
				res, err := s.decideZ3(z3solver)
				if err != nil {
					return nil, err
				}

				if res != smt.Unsat {
					overallResult = res
					// SAT: the check condition is satisfiable (property fails)
					if fc.Sat() {
						// Checker says to continue (ignore this failure)
						overallResult = smt.Unsat
						if s.opts.Incremental {
							z3solver.Pop()
						}
						continue
					}
					// Python breaks before popping here, so the diagnostic
					// model remains constrained by the failing final condition.
					break // stop checking
				} else {
					overallResult = smt.Unsat
					fc.Unsat()
				}
				if s.opts.Incremental {
					z3solver.Pop()
				}
			}
		}
	} else {
		// No final conditions: just check satisfiability.
		// Python ivy_solver.py:1451 calls decide(s) here.
		xtracer.Trace("ivy_solver.py:1302 decide() ENTER")
		var err error
		overallResult, err = s.decideZ3(z3solver)
		if err != nil {
			return nil, err
		}
	}

	if overallResult == smt.Unsat {
		return nil, nil
	}

	if shrink {
		// Minimize sorts. Python ivy_solver.py:1463 calls decide(s) here.
		for _, sort := range sortsToMinimize {
			for n := 1; ; n++ {
				sc := SortSizeConstraint(sort, n)
				zsc, err := s.formulaToZ3(sc)
				if err != nil {
					break
				}
				z3solver.Push()
				z3solver.Assert(zsc)
				xtracer.Trace("ivy_solver.py:1302 decide() ENTER")
				res, err := s.decideZ3(z3solver)
				if err != nil {
					return nil, err
				}
				if res == smt.Sat {
					break
				}
				z3solver.Pop()
			}
		}

		// Minimize relations. Python ivy_solver.py:1463 calls decide(s) here.
		for _, rel := range relationsToMinimize {
			for n := 1; ; n++ {
				sc := RelationSizeConstraint(rel, n)
				zsc, err := s.formulaToZ3(sc)
				if err != nil {
					break
				}
				z3solver.Push()
				z3solver.Assert(zsc)
				xtracer.Trace("ivy_solver.py:1302 decide() ENTER")
				res, err := s.decideZ3(z3solver)
				if err != nil {
					return nil, err
				}
				if res == smt.Sat {
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
	vocab := make([]*Const, 0, symSet.Len())
	for _, sym := range symSet.All() {
		if c, ok := sym.(*Const); ok {
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

func (s *Solver) showVCsEnabled() bool {
	return s != nil && s.opts != nil && s.opts.ShowVCs
}

func (s *Solver) showVCsBase(clauses *Clauses) {
	if !s.showVCsEnabled() || clauses == nil {
		return
	}
	fmt.Println()
	fmt.Println("definitions:")
	for _, df := range clauses.Defs {
		fmt.Println(df)
		fmt.Println()
	}
	fmt.Println("axioms:")
	for _, fmla := range clauses.Fmlas {
		fmt.Println(fmla)
		fmt.Println()
	}
}

func (s *Solver) showVCsFinalCond(kind string, clauses *Clauses) {
	if !s.showVCsEnabled() || clauses == nil {
		return
	}
	fmt.Printf("\n%s: %v\n", kind, clauses)
}

// EvalFormula evaluates a formula in a model, returning true/false/unknown.
func (s *Solver) EvalFormula(model *smt.Model, fmla Expr) (bool, error) {
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
func (s *Solver) CubeToZ3(cube []*LogicLiteral) (smt.Z3Expr, error) {
	xtracer.Trace("ivy_solver.py:774 cube_to_z3() ENTER nlits=%d", len(cube))
	if len(cube) == 0 {
		return s.tr.Ctx.BoolVal(true), nil
	}
	exprs := make([]smt.Z3Expr, len(cube))
	for i, lit := range cube {
		zlit, err := s.LiteralToZ3(lit)
		if err != nil {
			return smt.Z3Expr{}, err
		}
		exprs[i] = zlit
	}
	if len(exprs) == 1 {
		return exprs[0], nil
	}
	return s.tr.Ctx.And(exprs...), nil
}

// LiteralToZ3 converts a single literal to a Z3 expression.
func (s *Solver) LiteralToZ3(lit *LogicLiteral) (smt.Z3Expr, error) {
	xtracer.Trace("ivy_solver.py:537 literal_to_z3() ENTER polarity=%v", lit.Polarity)
	zAtom, err := s.tr.Translate(lit.Atom)
	if err != nil {
		return smt.Z3Expr{}, err
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
	Expr   smt.Z3Expr // prevent GC so AST ID stays valid
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
	z3solver *smt.Z3Solver,
	cube []*LogicLiteral,
	memo map[uint]*CubeMemoEntry,
	memoUnsatOnly bool,
) (bool, error) {
	xtracer.Trace("ivy_solver.py:785 check_cube() ENTER")
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
	result := s.checkZ3(z3solver)
	sat := result != smt.Unsat

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
	clauses *Clauses,
	ignore func(*Const) bool,
) (*Clauses, error) {
	xtracer.Trace("ivy_solver.py:1516 clauses_model_to_clauses() ENTER")
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
	clauses *Clauses,
	model *ModelResult,
	ignore func(*Const) bool,
	numerals bool,
) (*Clauses, error) {
	res, _, err := s.ClausesModelToClausesWithModelAndHerbrand(clauses, model, ignore, numerals)
	return res, err
}

func (s *Solver) ClausesModelToClausesWithModelAndHerbrand(
	clauses *Clauses,
	model *ModelResult,
	ignore func(*Const) bool,
	numerals bool,
) (*Clauses, *HerbrandModel, error) {
	if ignore == nil {
		ignore = func(*Const) bool { return false }
	}

	// Get a HerbrandModel
	var h *HerbrandModel
	if model != nil {
		// Build HerbrandModel from existing ModelResult
		symSet := clauses.Symbols()
		vocab := make([]*Const, 0, symSet.Len())
		for _, sym := range symSet.All() {
			if c, ok := sym.(*Const); ok {
				vocab = append(vocab, c)
			}
		}
		h = NewHerbrandModel(s, model.Solver, model.Model, vocab)
	} else {
		h = s.ModelIfNone(clauses, nil, nil)
	}
	if h == nil {
		return nil, nil, nil // unsat
	}

	// Extract model facts
	res := ModelFacts(h, ignore, clauses, false)

	// Build substitution map using structural keys
	subs := make(map[NodeKey]Expr)
	if numerals {
		na := NumeralAssignWithClauses(h, res)
		for elemName, numName := range na {
			for _, sort := range h.Sorts() {
				for _, c := range h.SortUniverse(sort) {
					if c.Name == elemName {
						subs[Key(c)] = NewConst(numName, c.CSort)
					}
				}
			}
		}
	} else {
		// Prefix with "__"
		for _, sort := range h.Sorts() {
			for _, c := range h.SortUniverse(sort) {
				subs[Key(c)] = NewConst("__"+c.Name, c.CSort)
			}
		}
	}

	res = SubstituteConstantsClauses(res, subs)
	return res, h, nil
}

// FilterRedundantFacts removes redundant negative formulas from clauses,
// given axioms.
// Corresponds to Python's filter_redundant_facts.
func (s *Solver) FilterRedundantFacts(clauses *Clauses, axioms *Clauses) (*Clauses, error) {
	xtracer.Trace("ivy_solver.py:1564 filter_redundant_facts() ENTER")
	// Separate positive and negative formulas.
	// Python: pos_fmlas = [f for f in fmlas if not isinstance(f, ivy_logic.Not)]
	var posFmlas, negFmlas []Expr
	for _, f := range clauses.Fmlas {
		if _, isNot := f.(*LogicNot); isNot {
			negFmlas = append(negFmlas, f)
		} else {
			posFmlas = append(posFmlas, f)
		}
	}

	if len(negFmlas) == 0 {
		return clauses, nil
	}

	ctx := s.tr.Ctx
	z3solver := s.newZ3Solver()

	// Add axioms
	za, err := s.ClausesToZ3(axioms)
	if err != nil {
		return nil, err
	}
	z3solver.Assert(za)

	// Add definitions
	// Python filter_redundant_facts line 1494: formula_to_z3(d.to_constraint())
	for _, d := range clauses.Defs {
		constraint := DefinitionToConstraint(d)
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
	alits := make([]smt.Z3Expr, len(negFmlas))
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
	var keep []Expr
	for i, fmla := range negFmlas {
		if s.checkZ3Assumptions(z3solver, []smt.Z3Expr{alits[i]}) == smt.Sat {
			keep = append(keep, fmla)
		}
	}

	allFmlas := append(posFmlas, keep...)
	defs := make([]*IvyDefinition, len(clauses.Defs))
	copy(defs, clauses.Defs)
	return NewClauses(allFmlas, defs, clauses.Annot), nil
}
