package abc

// Shared session facade for src/sat/bsat/satSolver.c and satSolver3.c.

import "slices"

type SatSolverSession = SatSolver3
type SatSolverSessionOptions = SatSolver3Options
type SatSolverSessionLimits = SatSolver3Limits
type SatSolverSessionLexResult = SatSolver3LexResult
type SatSolverSessionMinimizeResult = SatSolver3MinimizeResult

func NewSatSolverSession(options SatSolverSessionOptions) (*SatSolverSession, error) {
	return NewSatSolver3(options)
}

func (s *SatSolver3) SetLearnedReductionPolicy(start, delta, ratio int) error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.SetLearnedReductionPolicy(start, delta, ratio)
}

// SetRestartInterval configures periodic Luby restarts. Zero implements the
// legacy fNoRestarts setting.
func (s *SatSolver3) SetRestartInterval(interval uint64) error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.SetRestartInterval(interval)
}

func (s *SatSolver3) SetActivityFactors(factors []float64, variables []SatVar) error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.SetActivityFactors(factors, variables)
}

func (s *SatSolver3) ClearActivityFactors() error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.ClearActivityFactors()
}

func (s *SatSolver3) SetGlobalActivityVariables(variables []SatVar) error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.SetGlobalActivityVariables(variables)
}

// Compress completes pending root propagation. The Go core does not perform
// the C solver's destructive clause compaction, so this is the checked form of
// the source inline helper.
func (s *SatSolver3) Compress() (bool, error) {
	return s.Simplify()
}

func (s *SatSolver3) SetVariableActivityPrefix(count int) error {
	if !s.usable() || count < 0 || count > s.VariableCount() {
		return ErrDimension
	}
	variables := make([]SatVar, count)
	for index := range variables {
		variables[index] = SatVar(index)
	}
	return s.SetVariableActivity(variables)
}

// SetRunStop installs the source's run-ID termination callback. The callback
// is synchronous and must not reenter or mutate the session.
func (s *SatSolver3) SetRunStop(runID int, stop func(int) bool) error {
	if !s.usable() {
		return ErrInvariant
	}
	s.runID, s.stopForRun = runID, stop
	return nil
}

func (s *SatSolver3) SetModelProjection(variables []SatVar) error {
	if !s.usable() {
		return ErrInvariant
	}
	for _, variable := range variables {
		if uint64(variable) >= uint64(s.VariableCount()) {
			return ErrInvalidLiteral
		}
	}
	s.userVariables = slices.Clone(variables)
	s.userValues = s.userValues[:0]
	return nil
}

func (s *SatSolver3) ModelProjectionValues() []int {
	if !s.usable() {
		return nil
	}
	return slices.Clone(s.userValues)
}

func (s *SatSolver3) ClearModelProjectionValues() error {
	if !s.usable() {
		return ErrInvariant
	}
	s.userValues = s.userValues[:0]
	return nil
}

// ToggleClauseDiagnostics replaces sat_solver_flip_print_clause. Clause
// snapshots are retained for the caller instead of being printed to stdout.
func (s *SatSolver3) ToggleClauseDiagnostics() (bool, error) {
	if !s.usable() {
		return false, ErrInvariant
	}
	s.printClauses = !s.printClauses
	return s.printClauses, nil
}

func (s *SatSolver3) ClauseDiagnostics() [][]SatLit {
	if !s.usable() {
		return nil
	}
	return CloneSatSolver3Clauses(s.printedClauses)
}

func (s *SatSolver3) ClearClauseDiagnostics() error {
	if !s.usable() {
		return ErrInvariant
	}
	s.printedClauses = s.printedClauses[:0]
	return nil
}

func (s *SatSolver3) ChangeLastProofStoreClause() (SatVar, error) {
	if !s.usable() || s.store == nil {
		return 0, ErrInvariant
	}
	return s.store.ChangeLastClause()
}
