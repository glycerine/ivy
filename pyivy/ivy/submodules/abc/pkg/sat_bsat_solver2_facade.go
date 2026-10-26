package abc

// Concrete public solver2 facade for src/sat/bsat/satSolver2.h.

import (
	"slices"
	"time"
)

type SatSolver2Options struct {
	MaxVariables int
	MaxClauses   int
}

type SatSolver2Clause struct {
	Literals []SatLit
	PartA    bool
	ID       int
}

type satSolver2Bookmark struct {
	clauses      []SatSolver2Clause
	variablePart []bool
	called       bool
}

type SatSolver2 struct {
	core         *SatSolver
	clauses      []SatSolver2Clause // slot zero is an invalid handle
	variablePart []bool
	maxClauses   int
	called       bool
	stopped      bool
	bookmark     *satSolver2Bookmark
}

func NewSatSolver2(options SatSolver2Options) (*SatSolver2, error) {
	if options.MaxVariables < 0 || options.MaxClauses < 0 {
		return nil, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 16 << 20
	}
	core, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return nil, err
	}
	return &SatSolver2{core: core, clauses: make([]SatSolver2Clause, 1), maxClauses: options.MaxClauses}, nil
}

func cloneSatSolver2Clauses(clauses []SatSolver2Clause) []SatSolver2Clause {
	result := make([]SatSolver2Clause, len(clauses))
	for index, clause := range clauses {
		result[index] = clause
		result[index].Literals = slices.Clone(clause.Literals)
	}
	return result
}

func normalizeSatSolver2Clause(literals []SatLit) ([]SatLit, error) {
	if len(literals) == 0 {
		return nil, ErrDimension
	}
	result := slices.Clone(literals)
	for _, literal := range result {
		if !literal.Valid() {
			return nil, ErrInvalidLiteral
		}
	}
	slices.Sort(result)
	for index := 1; index < len(result); index++ {
		if result[index-1] == result[index].Not() {
			return result, nil
		}
	}
	write := 0
	for _, literal := range result {
		if write == 0 || result[write-1] != literal {
			result[write] = literal
			write++
		}
	}
	return result[:write:write], nil
}

func (s *SatSolver2) usable() bool {
	return s != nil && !s.stopped && s.core != nil && len(s.clauses) > 0
}

func (s *SatSolver2) SetVariableCount(count int) error {
	if !s.usable() {
		return ErrInvariant
	}
	if s.core.proof != nil || s.core.resolutionProof != nil {
		return ErrInvariant
	}
	if err := s.core.SetVariableCount(count); err != nil {
		return err
	}
	if count > len(s.variablePart) {
		s.variablePart = append(s.variablePart, make([]bool, count-len(s.variablePart))...)
	}
	return nil
}

func (s *SatSolver2) AddClause(literals []SatLit, id int) (int, error) {
	if !s.usable() {
		return 0, ErrInvariant
	}
	if s.core.proof != nil || s.core.resolutionProof != nil {
		return 0, ErrInvariant
	}
	if id == 0 {
		return 0, ErrDimension
	}
	normalized, err := normalizeSatSolver2Clause(literals)
	if err != nil {
		return 0, err
	}
	if len(s.clauses)-1 >= s.maxClauses {
		return 0, ErrLimit
	}
	maximum := int(normalized[len(normalized)-1].Var()) + 1
	handle := len(s.clauses)
	previousClauses := len(s.core.clauses)
	trial := s.core.cloneState()
	trial.bookmark = s.core.bookmark
	ok, err := trial.AddClause(normalized)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrInvariant
	}
	if len(trial.clauses) == previousClauses+1 {
		trial.clauses[len(trial.clauses)-1].proofRoot = handle
	} else if len(trial.clauses) != previousClauses {
		return 0, ErrInvariant
	}
	s.core = &trial
	if maximum > len(s.variablePart) {
		s.variablePart = append(s.variablePart, make([]bool, maximum-len(s.variablePart))...)
	}
	s.clauses = append(s.clauses, SatSolver2Clause{Literals: normalized, ID: id})
	return handle, nil
}

func (s *SatSolver2) AddClauseSpecs(specifications []SatSolver2ClauseSpec) ([]int, error) {
	if !s.usable() {
		return nil, ErrInvariant
	}
	trial := *s
	core := s.core.cloneState()
	core.bookmark = s.core.bookmark
	trial.core = &core
	trial.clauses = cloneSatSolver2Clauses(s.clauses)
	trial.variablePart = slices.Clone(s.variablePart)
	handles := make([]int, 0, len(specifications))
	for _, specification := range specifications {
		handle, err := trial.AddClause(specification.Literals, specification.ID)
		if err != nil {
			return nil, err
		}
		trial.clauses[handle].PartA = specification.PartA
		handles = append(handles, handle)
	}
	*s = trial
	return handles, nil
}

func (s *SatSolver2) Clause(handle int) (SatSolver2Clause, error) {
	if !s.usable() || handle <= 0 || handle >= len(s.clauses) {
		return SatSolver2Clause{}, ErrInvalidLiteral
	}
	result := s.clauses[handle]
	result.Literals = slices.Clone(result.Literals)
	return result, nil
}

func (s *SatSolver2) SetClausePartA(handle int, partA bool) error {
	if !s.usable() || s.called || s.core.resolutionProof != nil {
		return ErrInvariant
	}
	if handle <= 0 || handle >= len(s.clauses) {
		return ErrInvalidLiteral
	}
	s.clauses[handle].PartA = partA
	return nil
}

func (s *SatSolver2) SetClauseID(handle, id int) error {
	if !s.usable() || s.called {
		return ErrInvariant
	}
	if handle <= 0 || handle >= len(s.clauses) {
		return ErrInvalidLiteral
	}
	if id == 0 {
		return ErrDimension
	}
	s.clauses[handle].ID = id
	return nil
}

func (s *SatSolver2) SetVariablePartA(variable SatVar, partA bool) error {
	if !s.usable() || s.called || s.core.resolutionProof != nil {
		return ErrInvariant
	}
	if uint64(variable) >= uint64(len(s.variablePart)) {
		return ErrInvalidLiteral
	}
	s.variablePart[variable] = partA
	return nil
}

func (s *SatSolver2) VariablePartA(variable SatVar) (bool, error) {
	if !s.usable() || uint64(variable) >= uint64(len(s.variablePart)) {
		return false, ErrInvalidLiteral
	}
	return s.variablePart[variable], nil
}

func (s *SatSolver2) VariableCount() int {
	if !s.usable() {
		return 0
	}
	return s.core.VariableCount()
}
func (s *SatSolver2) ClauseCount() int {
	if !s.usable() {
		return 0
	}
	return len(s.clauses) - 1
}
func (s *SatSolver2) LearnedClauseCount() int {
	if !s.usable() {
		return 0
	}
	return s.core.LearnedClauseCount()
}
func (s *SatSolver2) ConflictCount() uint64 {
	if !s.usable() {
		return 0
	}
	return s.core.ConflictCount()
}

func (s *SatSolver2) Solve(assumptions []SatLit, limits SatLimits) (SatResult, error) {
	if !s.usable() {
		return SatResult{}, ErrInvariant
	}
	s.called = true
	return s.core.Solve(assumptions, limits)
}

func (s *SatSolver2) ModelValue(variable SatVar) (bool, error) {
	if !s.usable() {
		return false, ErrInvariant
	}
	return s.core.ModelValue(variable)
}
func (s *SatSolver2) ModelLiteral(variable SatVar) (SatLit, error) {
	if !s.usable() {
		return 0, ErrInvariant
	}
	return s.core.ModelLiteral(variable)
}
func (s *SatSolver2) FinalConflict() ([]SatLit, error) {
	if !s.usable() {
		return nil, ErrInvariant
	}
	return s.core.FinalConflict()
}
func (s *SatSolver2) LastStatus() SatStatus {
	if !s.usable() {
		return SatUnknown
	}
	return s.core.LastStatus()
}
func (s *SatSolver2) Statistics() SatStats {
	if !s.usable() {
		return SatStats{}
	}
	return s.core.Statistics()
}
func (s *SatSolver2) SetRuntimeDeadline(deadline time.Time) (time.Time, error) {
	if !s.usable() {
		return time.Time{}, ErrInvariant
	}
	return s.core.SetRuntimeDeadline(deadline)
}
func (s *SatSolver2) SetLearnedClauseLimit(limit int) (int, error) {
	if !s.usable() {
		return 0, ErrInvariant
	}
	return s.core.SetLearnedClauseLimit(limit)
}
func (s *SatSolver2) SetLearnedReductionPolicy(start, delta, ratio int) error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.SetLearnedReductionPolicy(start, delta, ratio)
}
func (s *SatSolver2) ClearVariableActivity() error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.ClearVariableActivity()
}
func (s *SatSolver2) Simplify() (bool, error) {
	if !s.usable() {
		return false, ErrInvariant
	}
	return s.core.Simplify()
}
func (s *SatSolver2) VerifyModel() error {
	if !s.usable() || s.core.lastStatus != SatSatisfiable || len(s.core.model) != s.core.VariableCount() {
		return ErrInvariant
	}
	for handle := 1; handle < len(s.clauses); handle++ {
		satisfied := false
		for _, literal := range s.clauses[handle].Literals {
			if s.core.model[literal.Var()] != literal.Negative() {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return ErrInvariant
		}
	}
	return s.core.VerifyModel()
}
func (s *SatSolver2) PruneRootSatisfiedWatches() (SatWatchPruneStats, error) {
	if !s.usable() {
		return SatWatchPruneStats{}, ErrInvariant
	}
	return s.core.PruneRootSatisfiedWatches()
}
func (s *SatSolver2) ReduceLearnedDatabase() (SatLearnedReductionStats, error) {
	if !s.usable() {
		return SatLearnedReductionStats{}, ErrInvariant
	}
	return s.core.ReduceLearnedDatabase()
}

func (s *SatSolver2) Bookmark() error {
	if !s.usable() {
		return ErrInvariant
	}
	if err := s.core.Bookmark(); err != nil {
		return err
	}
	s.bookmark = &satSolver2Bookmark{
		clauses:      cloneSatSolver2Clauses(s.clauses),
		variablePart: slices.Clone(s.variablePart),
		called:       s.called,
	}
	return nil
}

func (s *SatSolver2) Rollback() error {
	if !s.usable() || s.bookmark == nil {
		return ErrInvariant
	}
	if err := s.core.Rollback(); err != nil {
		return err
	}
	s.clauses = cloneSatSolver2Clauses(s.bookmark.clauses)
	s.variablePart = slices.Clone(s.bookmark.variablePart)
	s.called = s.called || s.bookmark.called
	return nil
}

func (s *SatSolver2) Check() error {
	if !s.usable() || len(s.variablePart) != s.core.VariableCount() || len(s.clauses)-1 > s.maxClauses {
		return ErrInvariant
	}
	if err := s.core.Check(); err != nil {
		return err
	}
	for handle := 1; handle < len(s.clauses); handle++ {
		clause := s.clauses[handle]
		if clause.ID == 0 || len(clause.Literals) == 0 {
			return ErrInvariant
		}
		for index, literal := range clause.Literals {
			if !literal.Valid() || uint64(literal.Var()) >= uint64(len(s.variablePart)) || (index > 0 && clause.Literals[index-1] > literal) {
				return ErrInvariant
			}
		}
	}
	return nil
}

func (s *SatSolver2) Stop() {
	if s == nil || s.stopped {
		return
	}
	if s.core != nil && s.core.proof != nil && s.core.proof.manager != nil {
		s.core.proof.manager.Stop()
	}
	if s.core != nil && s.core.resolutionProof != nil && s.core.resolutionProof.dag != nil {
		s.core.resolutionProof.dag.Stop()
	}
	s.core = nil
	s.clauses = nil
	s.variablePart = nil
	s.bookmark = nil
	s.stopped = true
}
