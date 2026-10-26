package abc

// Typed solving from sat/cnf/cnfUtil.c:Cnf_DataSolveFromFile.

import (
	"errors"
	"slices"
)

type CNFSolveOptions struct {
	SAT          SatLimits
	Resources    CNFTransformOptions
	ModelSuffix  int // zero: no suffix; otherwise last N model variables, as native nPis
	LearnedStart int // zero retains the solver default
	LearnedDelta int // zero retains the solver default
	LearnedRatio int // zero retains the solver default; otherwise 1..100
	Verbose      bool
	ShowPattern  bool
}

type CNFSolveDiagnostics struct {
	Variables, Clauses, Literals int
	SolverStatistics             string
}

type CNFSolveResult struct {
	SatResult
	SuffixModel []bool
	Pattern     []bool
	Diagnostics *CNFSolveDiagnostics
}

func (r CNFSolveResult) NativeCode() int {
	switch r.Status {
	case SatSatisfiable:
		return 0
	case SatUnsatisfiable:
		return 1
	default:
		return -1
	}
}
func cnfSolveCancelled(stats SatStats) CNFSolveResult {
	return CNFSolveResult{SatResult: SatResult{Status: SatUnknown, Reason: SatCancelled, Stats: stats}}
}

func cnfConfigureSolver(solver *SatSolver, options CNFSolveOptions) error {
	if options.LearnedStart < 0 || options.LearnedDelta < 0 || options.LearnedRatio < 0 || options.LearnedRatio > 100 {
		return ErrDimension
	}
	if options.LearnedStart == 0 && options.LearnedDelta == 0 && options.LearnedRatio == 0 {
		return nil
	}
	start, delta, ratio := solver.learntStart, solver.learntDelta, solver.learntRatio
	if options.LearnedStart != 0 {
		start = options.LearnedStart
	}
	if options.LearnedDelta != 0 {
		delta = options.LearnedDelta
	}
	if options.LearnedRatio != 0 {
		ratio = options.LearnedRatio
	}
	return solver.SetLearnedReductionPolicy(start, delta, ratio)
}

func cnfSolveDecorate(result CNFSolveResult, formula *CNFFormula, solver *SatSolver, options CNFSolveOptions) CNFSolveResult {
	if options.ShowPattern && result.Status == SatSatisfiable {
		result.Pattern = slices.Clone(result.Model)
	}
	if options.Verbose {
		literals := 0
		for _, clause := range formula.clauses {
			literals += len(clause)
		}
		result.Diagnostics = &CNFSolveDiagnostics{
			Variables: formula.variables, Clauses: len(formula.clauses), Literals: literals,
			SolverStatistics: FormatSatSolverStatistics(solver.Statistics(), true),
		}
	}
	return result
}

// Solve owns a fresh concrete solver, verifies every SAT clause before exposing
// a model, and extracts the optional final-variable suffix. Cancellation during
// checking/loading/verification returns UNKNOWN without partial model storage.
// SAT work budgets apply to search, matching native load-then-solve ordering.
func (f *CNFFormula) Solve(o CNFSolveOptions) (CNFSolveResult, error) {
	if f == nil || o.ModelSuffix < 0 || o.ModelSuffix > f.variables || o.LearnedStart < 0 || o.LearnedDelta < 0 || o.LearnedRatio < 0 || o.LearnedRatio > 100 {
		return CNFSolveResult{}, ErrDimension
	}
	a, b := o.Resources.Stop, o.SAT.Stop
	stop := func() bool { return a != nil && a() || b != nil && b() }
	o.Resources.Stop = stop
	o.SAT.Stop = stop
	if e := checkCNFFormula(f, o.Resources); e != nil {
		if errors.Is(e, ErrCancelled) {
			return cnfSolveCancelled(SatStats{}), nil
		}
		return CNFSolveResult{}, e
	}
	maximum := o.Resources.MaxVariables
	if maximum == 0 {
		maximum = 1 << 20
	}
	solver, e := NewSatSolverWithLimit(maximum)
	if e != nil {
		return CNFSolveResult{}, e
	}
	if e = solver.SetVariableCount(f.variables); e != nil {
		return CNFSolveResult{}, e
	}
	if e = cnfConfigureSolver(solver, o); e != nil {
		return CNFSolveResult{}, e
	}
	for _, clause := range f.clauses {
		if stop() {
			return cnfSolveCancelled(SatStats{}), nil
		}
		ok, e := solver.AddClause(clause)
		if e != nil {
			return CNFSolveResult{}, e
		}
		if !ok {
			return cnfSolveDecorate(CNFSolveResult{SatResult: SatResult{Status: SatUnsatisfiable}}, f, solver, o), nil
		}
	}
	answer, e := solver.Solve(nil, o.SAT)
	if e != nil {
		return CNFSolveResult{}, e
	}
	if answer.Status != SatSatisfiable {
		return cnfSolveDecorate(CNFSolveResult{SatResult: answer}, f, solver, o), nil
	}
	if len(answer.Model) != f.variables {
		return CNFSolveResult{}, ErrInvariant
	}
	for _, clause := range f.clauses {
		if stop() {
			return cnfSolveCancelled(answer.Stats), nil
		}
		satisfied := false
		for _, l := range clause {
			satisfied = satisfied || (answer.Model[l.Var()] != (l&1 != 0))
		}
		if !satisfied {
			return CNFSolveResult{}, ErrInvariant
		}
	}
	if stop() {
		return cnfSolveCancelled(answer.Stats), nil
	}
	result := CNFSolveResult{SatResult: answer}
	if o.ModelSuffix > 0 {
		result.SuffixModel = slices.Clone(answer.Model[len(answer.Model)-o.ModelSuffix:])
	}
	return cnfSolveDecorate(result, f, solver, o), nil
}
func cnfSolveStops(read CNFDIMACSReadOptions, solve CNFSolveOptions) (CNFDIMACSReadOptions, CNFSolveOptions) {
	a, b, c := read.Limits.Stop, solve.Resources.Stop, solve.SAT.Stop
	stop := func() bool { return a != nil && a() || b != nil && b() || c != nil && c() }
	read.Limits.Stop = stop
	solve.Resources.Stop = stop
	solve.SAT.Stop = nil
	return read, solve
}
func SolveCNFDIMACS(data []byte, read CNFDIMACSReadOptions, solve CNFSolveOptions) (CNFSolveResult, error) {
	read, solve = cnfSolveStops(read, solve)
	formula, e := ParseCNFDIMACS(data, read)
	if e != nil {
		if errors.Is(e, ErrCancelled) {
			return cnfSolveCancelled(SatStats{}), nil
		}
		return CNFSolveResult{}, e
	}
	return formula.Solve(solve)
}
func SolveCNFDIMACSFile(path string, read CNFDIMACSReadOptions, solve CNFSolveOptions) (CNFSolveResult, error) {
	read, solve = cnfSolveStops(read, solve)
	formula, e := ReadCNFDIMACSFile(path, read)
	if e != nil {
		if errors.Is(e, ErrCancelled) {
			return cnfSolveCancelled(SatStats{}), nil
		}
		return CNFSolveResult{}, e
	}
	return formula.Solve(solve)
}
