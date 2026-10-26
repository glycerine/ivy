package abc

// Checked search-state operations from src/sat/msat/msatSolverSearch.c and
// the remaining state accessors in msatSolverApi.c.

import "slices"

type MsatSearchConflict struct {
	ClauseID int
	Literals []SatLit
}

type MsatLearnedClauseSnapshot struct {
	ClauseID int
	Literals []SatLit
	Activity uint32
	Locked   bool
}

func (s *MsatSolver) DecisionLevel() int {
	if s == nil || s.closed || s.core == nil {
		return 0
	}
	return len(s.core.trailLim)
}

func (s *MsatSolver) DecisionLevels() []int {
	if s == nil || s.closed || s.core == nil {
		return nil
	}
	result := make([]int, len(s.core.levels))
	for variable := range result {
		result[variable] = -1
		if s.core.assigns[variable] >= 0 {
			result[variable] = s.core.levels[variable]
		}
	}
	return result
}

func (s *MsatSolver) ReasonClauseIDs() []int {
	if s == nil || s.closed || s.core == nil {
		return nil
	}
	return slices.Clone(s.core.reasons)
}

func (s *MsatSolver) Trail() []SatLit {
	if s == nil || s.closed || s.core == nil {
		return nil
	}
	return slices.Clone(s.core.trail)
}

func (s *MsatSolver) Seen() []int {
	if s == nil || s.closed {
		return nil
	}
	return slices.Clone(s.seen)
}

func (s *MsatSolver) IncrementSeenID() (int, error) {
	if err := s.checkLive(); err != nil {
		return 0, err
	}
	if s.seenID == int(^uint(0)>>1) {
		clear(s.seen)
		s.seenID = 1
		return s.seenID, nil
	}
	s.seenID++
	return s.seenID, nil
}

func (s *MsatSolver) LearnedClauses() []MsatLearnedClauseSnapshot {
	if s == nil || s.closed || s.core == nil {
		return nil
	}
	locked := make([]bool, len(s.core.clauses))
	for _, reason := range s.core.reasons {
		if reason > 0 && reason < len(locked) {
			locked[reason] = true
		}
	}
	result := make([]MsatLearnedClauseSnapshot, 0, s.core.learntCount)
	for identifier := 1; identifier < len(s.core.clauses); identifier++ {
		clause := &s.core.clauses[identifier]
		if clause.learnt && !clause.deleted {
			result = append(result, MsatLearnedClauseSnapshot{identifier, slices.Clone(clause.lits), clause.activity, locked[identifier]})
		}
	}
	return result
}

func (s *MsatSolver) WatchedClauseIDs() [][]int {
	if s == nil || s.closed || s.core == nil {
		return nil
	}
	result := make([][]int, len(s.core.watches))
	for literal := range result {
		result[literal] = slices.Clone(s.core.watches[literal])
	}
	return result
}

func (s *MsatSolver) MemoryStats() MsatStepMemoryStats {
	if s == nil || s.closed || s.memory == nil {
		return MsatStepMemoryStats{}
	}
	return s.memory.Stats()
}

func (s *MsatSolver) ConeVariables() []SatVar {
	if s == nil || s.closed {
		return nil
	}
	return slices.Clone(s.coneVariables)
}

func (s *MsatSolver) VariablesUsed() []bool {
	if s == nil || s.closed {
		return nil
	}
	return slices.Clone(s.variablesUsed)
}

func (s *MsatSolver) SetVariablesUsed(used []bool) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if len(used) != s.capacity {
		return ErrDimension
	}
	s.variablesUsed = slices.Clone(used)
	s.core.enqueueMask = slices.Clone(used)
	return nil
}

func (s *MsatSolver) Adjacencies() [][]SatVar {
	if s == nil || s.closed {
		return nil
	}
	result := make([][]SatVar, len(s.adjacents))
	for variable := range result {
		result[variable] = slices.Clone(s.adjacents[variable])
	}
	return result
}

func (s *MsatSolver) SetAdjacency(variable SatVar, adjacent []SatVar) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if uint64(variable) >= uint64(s.capacity) {
		return ErrDimension
	}
	seen := make([]bool, s.capacity)
	for _, other := range adjacent {
		if uint64(other) >= uint64(s.capacity) || seen[other] {
			return ErrDimension
		}
		seen[other] = true
	}
	s.adjacents[variable] = slices.Clone(adjacent)
	return nil
}

func (s *MsatSolver) AdjustClauseAllocation(learned bool, delta int) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	target := &s.clausesAlloc
	if learned {
		target = &s.learnedAlloc
	}
	if *target+delta < 0 {
		return ErrInvariant
	}
	*target += delta
	return nil
}

func (s *MsatSolver) ClauseAllocationCounts() (original, learned int) {
	if s == nil || s.closed {
		return 0, 0
	}
	return s.clausesAlloc, s.learnedAlloc
}

// Prepare clears transient assignments and installs a concrete decision cone.
// Clauses and learned clauses remain owned by the solver, matching the source.
func (s *MsatSolver) Prepare(variables []SatVar) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	seen := make([]bool, s.VariableCount())
	for _, variable := range variables {
		if uint64(variable) >= uint64(len(seen)) || seen[variable] {
			return ErrDimension
		}
		seen[variable] = true
	}
	for variable := range s.core.assigns {
		s.core.assigns[variable] = -1
		s.core.levels[variable] = 0
		s.core.reasons[variable] = 0
		s.core.activity[variable] = 0
		s.core.orderpos[variable] = -1
		s.variableLevels[variable] = -1
	}
	s.core.order = make([]int, len(variables))
	for position, variable := range variables {
		s.core.order[position] = int(variable)
		s.core.orderpos[variable] = position
	}
	s.core.trail = nil
	s.core.trailLim = nil
	s.core.qhead = 0
	s.core.model = nil
	s.core.lastConflict = nil
	s.core.lastStatus = SatUnknown
	s.core.inconsistent = false
	s.model = nil
	s.last = MsatSolveResult{}
	s.coneVariables = slices.Clone(variables)
	return nil
}

func (s *MsatSolver) Assume(literal SatLit) (bool, error) {
	if err := s.checkLive(); err != nil {
		return false, err
	}
	if !literal.Valid() || uint64(literal.Var()) >= uint64(s.VariableCount()) {
		return false, ErrInvalidLiteral
	}
	if s.core.qhead != len(s.core.trail) {
		return false, ErrInvariant
	}
	s.core.trailLim = append(s.core.trailLim, len(s.core.trail))
	if !s.variablesUsed[literal.Var()] {
		return true, nil
	}
	return s.core.enqueue(literal, 0), nil
}

func (s *MsatSolver) Propagate() (MsatSearchConflict, bool, error) {
	if err := s.checkLive(); err != nil {
		return MsatSearchConflict{}, false, err
	}
	identifier, stopped := s.core.propagate(nil)
	if stopped {
		return MsatSearchConflict{}, false, ErrInvariant
	}
	if identifier == 0 {
		return MsatSearchConflict{}, false, nil
	}
	if identifier < 0 || identifier >= len(s.core.clauses) {
		return MsatSearchConflict{}, false, ErrInvariant
	}
	return MsatSearchConflict{identifier, slices.Clone(s.core.clauses[identifier].lits)}, true, nil
}

func (s *MsatSolver) CancelUntil(level int) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if level < 0 || level > len(s.core.trailLim) {
		return ErrDimension
	}
	s.core.cancelUntil(level)
	return nil
}

func (s *MsatSolver) CancelLast() error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if len(s.core.trailLim) == 0 {
		return ErrDimension
	}
	s.core.cancelUntil(len(s.core.trailLim) - 1)
	return nil
}

func (s *MsatSolver) AnalyzeConflict(conflict MsatSearchConflict) ([]SatLit, int, error) {
	if err := s.checkLive(); err != nil {
		return nil, 0, err
	}
	if conflict.ClauseID <= 0 || conflict.ClauseID >= len(s.core.clauses) || len(s.core.trailLim) == 0 {
		return nil, 0, ErrDimension
	}
	learned, level, err := s.core.analyze(conflict.ClauseID)
	return slices.Clone(learned), level, err
}

func (s *MsatSolver) RecordLearned(literals []SatLit) (MsatLearnedClauseSnapshot, error) {
	if err := s.checkLive(); err != nil {
		return MsatLearnedClauseSnapshot{}, err
	}
	if len(literals) == 0 {
		return MsatLearnedClauseSnapshot{}, ErrDimension
	}
	work := slices.Clone(literals)
	for _, literal := range work {
		if !literal.Valid() || uint64(literal.Var()) >= uint64(s.VariableCount()) {
			return MsatLearnedClauseSnapshot{}, ErrInvalidLiteral
		}
	}
	if len(work) == 1 {
		if !s.core.enqueue(work[0], 0) {
			return MsatLearnedClauseSnapshot{}, ErrInvariant
		}
		return MsatLearnedClauseSnapshot{}, nil
	}
	best := 1
	for index := 2; index < len(work); index++ {
		if s.core.levels[work[best].Var()] < s.core.levels[work[index].Var()] {
			best = index
		}
	}
	work[1], work[best] = work[best], work[1]
	identifier := s.core.newClause(work, true)
	if !s.core.enqueue(work[0], identifier) {
		return MsatLearnedClauseSnapshot{}, ErrInvariant
	}
	clause := &s.core.clauses[identifier]
	locked := s.core.reasons[work[0].Var()] == identifier
	return MsatLearnedClauseSnapshot{identifier, slices.Clone(clause.lits), clause.activity, locked}, nil
}

func (s *MsatSolver) ReduceLearnedDatabase() (SatLearnedReductionStats, error) {
	if err := s.checkLive(); err != nil {
		return SatLearnedReductionStats{}, err
	}
	stats, err := s.core.ReduceLearnedDatabase()
	if err == nil {
		err = s.syncOriginalCoreIDs()
	}
	return stats, err
}

func (s *MsatSolver) syncOriginalCoreIDs() error {
	position := 0
	for identifier := 1; identifier < len(s.core.clauses); identifier++ {
		clause := &s.core.clauses[identifier]
		if clause.deleted || clause.learnt || len(clause.lits) < 2 {
			continue
		}
		if position >= len(s.original) {
			return ErrInvariant
		}
		s.original[position].coreID = identifier
		position++
	}
	if position != len(s.original) {
		return ErrInvariant
	}
	return nil
}

func (s *MsatSolver) rebuildWithoutLearned(removeOriginalFrom int) error {
	remove := make([]bool, len(s.core.clauses))
	if removeOriginalFrom >= 0 {
		for _, clause := range s.original[removeOriginalFrom:] {
			if clause.coreID > 0 && clause.coreID < len(remove) {
				remove[clause.coreID] = true
			}
		}
	}
	type retainedClause struct{ literals []SatLit }
	retained := make([]retainedClause, 0)
	for identifier := 1; identifier < len(s.core.clauses); identifier++ {
		clause := &s.core.clauses[identifier]
		if clause.deleted || clause.learnt || remove[identifier] {
			continue
		}
		retained = append(retained, retainedClause{slices.Clone(clause.lits)})
	}
	old := s.core
	rebuilt, err := NewSatSolverWithLimit(s.capacity)
	if err != nil {
		return err
	}
	if err = rebuilt.SetVariableCount(s.VariableCount()); err != nil {
		return err
	}
	for _, clause := range retained {
		if _, err = rebuilt.AddClause(clause.literals); err != nil {
			return err
		}
	}
	rebuilt.stats = old.stats
	rebuilt.randomSeed = old.randomSeed
	rebuilt.randomDecisions = old.randomDecisions
	rebuilt.restartInterval = old.restartInterval
	rebuilt.learntStart, rebuilt.learntDelta, rebuilt.learntRatio, rebuilt.learntLimit = old.learntStart, old.learntDelta, old.learntRatio, old.learntLimit
	rebuilt.enqueueMask = slices.Clone(s.variablesUsed)
	s.core = rebuilt
	if removeOriginalFrom >= 0 {
		for _, clause := range s.original[removeOriginalFrom:] {
			clause.closed = true
			clause.literals = nil
		}
		s.original = s.original[:removeOriginalFrom]
	}
	s.nextClause = len(s.original)
	s.clausesStart = min(s.clausesStart, len(s.original))
	s.model = nil
	s.last = MsatSolveResult{}
	return s.syncOriginalCoreIDs()
}

func (s *MsatSolver) RemoveLearned() error {
	if err := s.checkLive(); err != nil {
		return err
	}
	return s.rebuildWithoutLearned(-1)
}

func (s *MsatSolver) RemoveMarked() error {
	if err := s.checkLive(); err != nil {
		return err
	}
	return s.rebuildWithoutLearned(s.clausesStart)
}
