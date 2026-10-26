package abc

// Concrete facade for src/sat/bsat/satSolver3.c and satSolver3.h.

import (
	"math"
	"slices"
	"time"
	"unsafe"
)

type SatSolver3Options struct {
	MaxVariables int
	Seed         float64
}

// SatSolver3Limits combines source-compatible relative and absolute budgets.
// Zero means unlimited. Global budgets are compared with cumulative statistics.
type SatSolver3Limits struct {
	Conflicts, Propagations             uint64
	GlobalConflicts, GlobalPropagations uint64
	Stop                                func() bool
}

type SatSolver3LexResult struct {
	Result   SatResult
	Literals []SatLit
}

type SatSolver3MinimizeResult struct {
	Literals, LearnedUnits []SatLit
}

type satSolver3Bookmark struct {
	core         SatSolver
	pushed       []SatLit
	pushConflict []SatLit
	pivots       []SatVar
	limits       SatSolver3Limits
	seed         float64
}

type SatSolver3 struct {
	core           *SatSolver
	maxVariables   int
	seed           float64
	pushed         []SatLit
	pushConflict   []SatLit
	pivots         []SatVar
	limits         SatSolver3Limits
	bookmark       *satSolver3Bookmark
	trace          *SatSolverTrace
	store          *SatProofStore
	userVariables  []SatVar
	userValues     []int
	runID          int
	stopForRun     func(int) bool
	printClauses   bool
	printedClauses [][]SatLit
	stopped        bool
}

func NewSatSolver3(options SatSolver3Options) (*SatSolver3, error) {
	if options.MaxVariables < 0 || math.IsNaN(options.Seed) || math.IsInf(options.Seed, 0) {
		return nil, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.Seed == 0 {
		options.Seed = 91648253
	}
	core, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return nil, err
	}
	if err = core.SetRandomSeed(options.Seed); err != nil {
		return nil, err
	}
	return &SatSolver3{core: core, maxVariables: options.MaxVariables, seed: options.Seed}, nil
}

func (s *SatSolver3) usable() bool {
	return s != nil && !s.stopped && s.core != nil && len(s.core.clauses) != 0
}

func (s *SatSolver3) SetVariableCount(count int) error {
	if !s.usable() || len(s.pushed) != 0 {
		return ErrInvariant
	}
	return s.core.SetVariableCount(count)
}

func (s *SatSolver3) VariableCount() int {
	if !s.usable() {
		return 0
	}
	return s.core.VariableCount()
}

func (s *SatSolver3) AddClause(literals []SatLit) (bool, error) {
	if !s.usable() || len(s.pushed) != 0 {
		return false, ErrInvariant
	}
	if s.store == nil {
		ok, err := s.core.AddClause(literals)
		if err == nil && s.printClauses {
			s.printedClauses = append(s.printedClauses, slices.Clone(literals))
		}
		return ok, err
	}
	trialCore := s.core.cloneState()
	trialStore := s.store.clone()
	if _, err := trialStore.AddClause(literals); err != nil {
		return false, err
	}
	ok, err := trialCore.AddClause(literals)
	if err == nil && s.printClauses {
		s.printedClauses = append(s.printedClauses, slices.Clone(literals))
	}
	if err == nil {
		s.core, s.store = &trialCore, trialStore
	}
	return ok, err
}

func (s *SatSolver3) AddVariable() (SatVar, error) {
	if !s.usable() || len(s.pushed) != 0 {
		return 0, ErrInvariant
	}
	variable := s.VariableCount()
	if err := s.SetVariableCount(variable + 1); err != nil {
		return 0, err
	}
	return SatVar(variable), nil
}

// AddLearnedClause is the checked external counterpart of
// sat_solver3_clause_new(..., learnt=1). The clause is accepted only after an
// independent solve of its negation proves it follows from the current formula.
func (s *SatSolver3) AddLearnedClause(literals []SatLit) (bool, error) {
	if !s.usable() || len(s.pushed) != 0 || len(literals) == 0 {
		return false, ErrInvariant
	}
	normalized, err := normalizeSatSolver2Clause(literals)
	if err != nil {
		return false, err
	}
	for index := 1; index < len(normalized); index++ {
		if normalized[index-1] == normalized[index].Not() {
			return true, nil
		}
	}
	trial := s.core.cloneState()
	negated := make([]SatLit, len(normalized))
	for index, literal := range normalized {
		negated[index] = literal.Not()
	}
	result, solveErr := trial.Solve(negated, SatLimits{})
	if solveErr != nil {
		return false, solveErr
	}
	if result.Status != SatUnsatisfiable {
		return false, ErrInvariant
	}
	before := len(trial.clauses)
	ok, addErr := trial.AddClause(normalized)
	if addErr != nil || !ok {
		return ok, addErr
	}
	if len(trial.clauses) == before+1 {
		id := len(trial.clauses) - 1
		trial.clauses[id].learnt = true
		trial.clauses[id].proofRoot = -1
		trial.learntCount++
		trial.stats.Learnt++
	}
	s.core = &trial
	return true, nil
}

func (s *SatSolver3) OriginalClauseCount() int {
	if !s.usable() {
		return 0
	}
	return s.core.OriginalClauseCount()
}

func (s *SatSolver3) LearnedClauseCount() int {
	if !s.usable() {
		return 0
	}
	return s.core.LearnedClauseCount()
}

func (s *SatSolver3) ConflictCount() uint64 {
	if !s.usable() {
		return 0
	}
	return s.core.ConflictCount()
}

func (s *SatSolver3) Statistics() SatStats {
	if !s.usable() {
		return SatStats{}
	}
	return s.core.Statistics()
}

func (s *SatSolver3) CountAssigned() int {
	if !s.usable() {
		return 0
	}
	count := 0
	for _, value := range s.core.assigns {
		if value >= 0 {
			count++
		}
	}
	return count
}

// VariableAssignment reports the live propagated value using SatUnknown for an
// unassigned variable, matching sat_solver3_get_var_value's three-valued API.
func (s *SatSolver3) VariableAssignment(variable SatVar) (SatStatus, error) {
	if !s.usable() || uint64(variable) >= uint64(s.VariableCount()) {
		return SatUnknown, ErrInvalidLiteral
	}
	switch s.core.assigns[variable] {
	case 0:
		return SatUnsatisfiable, nil
	case 1:
		return SatSatisfiable, nil
	default:
		return SatUnknown, nil
	}
}

func (s *SatSolver3) ModelValue(variable SatVar) (bool, error) {
	if !s.usable() {
		return false, ErrInvariant
	}
	return s.core.ModelValue(variable)
}

func (s *SatSolver3) ModelLiteral(variable SatVar) (SatLit, error) {
	if !s.usable() {
		return 0, ErrInvariant
	}
	return s.core.ModelLiteral(variable)
}

func (s *SatSolver3) FinalConflict() ([]SatLit, error) {
	if !s.usable() {
		return nil, ErrInvariant
	}
	return s.core.FinalConflict()
}

func (s *SatSolver3) PushConflict() []SatLit {
	if !s.usable() {
		return nil
	}
	return slices.Clone(s.pushConflict)
}

func satSolver3ApplyPushed(core *SatSolver, pushed []SatLit) ([]SatLit, error) {
	core.cancelUntil(0)
	ok, stopped := core.prepareRoot(nil)
	if stopped {
		return nil, ErrInvariant
	}
	if !ok {
		return []SatLit{}, nil
	}
	for index, literal := range pushed {
		if !literal.Valid() || uint64(literal.Var()) >= uint64(core.VariableCount()) {
			return nil, ErrInvalidLiteral
		}
		core.trailLim = append(core.trailLim, len(core.trail))
		if !core.enqueue(literal, 0) {
			return core.finalConflict([]SatLit{literal}, &literal), nil
		}
		conflict, _ := core.propagate(nil)
		if conflict != 0 {
			return core.finalConflict(core.clauses[conflict].lits, nil), nil
		}
		_ = index
	}
	return nil, nil
}

// Push performs only assumption propagation. A conflicting push is atomic and
// leaves the previously committed stack intact.
func (s *SatSolver3) Push(literal SatLit) (bool, error) {
	if !s.usable() {
		return false, ErrInvariant
	}
	trial := s.core.cloneState()
	stack := append(slices.Clone(s.pushed), literal)
	conflict, err := satSolver3ApplyPushed(&trial, stack)
	if err != nil {
		return false, err
	}
	if conflict != nil {
		s.pushConflict = slices.Clone(conflict)
		return false, nil
	}
	trial.invalidateResult()
	s.core = &trial
	s.pushed = stack
	s.pushConflict = nil
	return true, nil
}

func (s *SatSolver3) Pop() error {
	if !s.usable() || len(s.pushed) == 0 {
		return ErrInvariant
	}
	s.pushed = s.pushed[:len(s.pushed)-1]
	trial := s.core.cloneState()
	conflict, err := satSolver3ApplyPushed(&trial, s.pushed)
	if err != nil || conflict != nil {
		return ErrInvariant
	}
	trial.invalidateResult()
	s.core = &trial
	s.pushConflict = nil
	return nil
}

func (s *SatSolver3) PushedAssumptions() []SatLit {
	if !s.usable() {
		return nil
	}
	return slices.Clone(s.pushed)
}

func (s *SatSolver3) SetResourceLimits(limits SatSolver3Limits) error {
	if !s.usable() {
		return ErrInvariant
	}
	s.limits = limits
	return nil
}

func satSolver3EffectiveLimits(stats SatStats, limits SatSolver3Limits) (SatLimits, SatStopReason) {
	result := SatLimits{Conflicts: limits.Conflicts, Propagations: limits.Propagations, Stop: limits.Stop}
	if limits.GlobalConflicts != 0 {
		if stats.Conflicts >= limits.GlobalConflicts {
			return result, SatConflictLimit
		}
		remaining := limits.GlobalConflicts - stats.Conflicts
		if result.Conflicts == 0 || remaining < result.Conflicts {
			result.Conflicts = remaining
		}
	}
	if limits.GlobalPropagations != 0 {
		if stats.Propagations >= limits.GlobalPropagations {
			return result, SatPropagationLimit
		}
		remaining := limits.GlobalPropagations - stats.Propagations
		if result.Propagations == 0 || remaining < result.Propagations {
			result.Propagations = remaining
		}
	}
	return result, SatFinished
}

func (s *SatSolver3) solveWith(assumptions []SatLit, limits SatSolver3Limits) (SatResult, error) {
	if !s.usable() {
		return SatResult{}, ErrInvariant
	}
	all := make([]SatLit, 0, len(s.pushed)+len(assumptions))
	all = append(all, s.pushed...)
	all = append(all, assumptions...)
	for _, variable := range s.userVariables {
		if uint64(variable) >= uint64(s.VariableCount()) {
			return SatResult{}, ErrInvalidLiteral
		}
	}
	if s.stopForRun != nil {
		configured := limits.Stop
		limits.Stop = func() bool {
			return (configured != nil && configured()) || s.stopForRun(s.runID)
		}
	}
	effective, exhausted := satSolver3EffectiveLimits(s.core.Statistics(), limits)
	if exhausted != SatFinished {
		return SatResult{Reason: exhausted}, nil
	}
	result, err := s.core.Solve(all, effective)
	if err == nil && result.Status == SatSatisfiable && len(s.userVariables) != 0 {
		for _, variable := range s.userVariables {
			if result.Model[variable] {
				s.userValues = append(s.userValues, 1)
			} else {
				s.userValues = append(s.userValues, 0)
			}
		}
	}
	return result, err
}

func (s *SatSolver3) Solve(assumptions []SatLit, limits SatSolver3Limits) (SatResult, error) {
	if err := s.SetResourceLimits(limits); err != nil {
		return SatResult{}, err
	}
	result, err := s.solveWith(assumptions, limits)
	if err != nil || result.Status != SatUnsatisfiable || s.store == nil || s.store.empty >= 0 {
		return result, err
	}
	trial := s.store.clone()
	if _, err = trial.AddClause(nil); err != nil {
		return SatResult{}, err
	}
	s.store = trial
	return result, nil
}

func (s *SatSolver3) SolveInternal() (SatResult, error) {
	return s.solveWith(nil, s.limits)
}

func (s *SatSolver3) Simplify() (bool, error) {
	if !s.usable() {
		return false, ErrInvariant
	}
	trial := s.core.cloneState()
	conflict, err := satSolver3ApplyPushed(&trial, s.pushed)
	if err != nil {
		return false, err
	}
	if conflict != nil {
		s.pushConflict = slices.Clone(conflict)
		return false, nil
	}
	s.core = &trial
	return true, nil
}

func (s *SatSolver3) SetVariableActivity(variables []SatVar) error {
	if !s.usable() {
		return ErrInvariant
	}
	seen := make([]bool, s.core.VariableCount())
	for _, variable := range variables {
		if uint64(variable) >= uint64(len(seen)) || seen[variable] {
			return ErrDimension
		}
		seen[variable] = true
	}
	clear(s.core.activity)
	for index, variable := range variables {
		s.core.activity[variable] = uint64(len(variables) - index)
	}
	s.core.varInc = 1
	s.core.order = s.core.order[:0]
	for index := range s.core.orderpos {
		s.core.orderpos[index] = -1
	}
	for variable, value := range s.core.assigns {
		if value < 0 {
			s.core.heapInsert(variable)
		}
	}
	return nil
}

func (s *SatSolver3) SetPreferredVariables(variables []SatVar) error {
	if !s.usable() {
		return ErrInvariant
	}
	return s.core.SetPreferredVariables(variables)
}

func (s *SatSolver3) CleanPolarity(variables []SatVar) error {
	if !s.usable() {
		return ErrInvariant
	}
	for _, variable := range variables {
		if uint64(variable) >= uint64(s.VariableCount()) {
			return ErrInvalidLiteral
		}
	}
	for _, variable := range variables {
		s.core.polarity[variable] = false
	}
	return nil
}

func (s *SatSolver3) SetLiteralPolarity(literals []SatLit) error {
	if !s.usable() {
		return ErrInvariant
	}
	for _, literal := range literals {
		if !literal.Valid() || uint64(literal.Var()) >= uint64(s.core.VariableCount()) {
			return ErrInvalidLiteral
		}
	}
	for _, literal := range literals {
		s.core.polarity[literal.Var()] = !literal.Negative()
	}
	return nil
}

func (s *SatSolver3) SetRandomDecisions(enabled bool) (bool, error) {
	if !s.usable() {
		return false, ErrInvariant
	}
	previous := s.core.randomDecisions
	return previous, s.core.SetRandomDecisions(enabled)
}

func (s *SatSolver3) SetRuntimeDeadline(deadline time.Time) (time.Time, error) {
	if !s.usable() {
		return time.Time{}, ErrInvariant
	}
	return s.core.SetRuntimeDeadline(deadline)
}

func (s *SatSolver3) Restart(seed float64) error {
	if !s.usable() || math.IsNaN(seed) || math.IsInf(seed, 0) {
		return ErrDimension
	}
	if seed == 0 {
		seed = 91648253
	}
	core, err := NewSatSolverWithLimit(s.maxVariables)
	if err != nil {
		return err
	}
	if err = core.SetRandomSeed(seed); err != nil {
		return err
	}
	s.core = core
	s.seed = seed
	s.pushed = nil
	s.pushConflict = nil
	s.pivots = nil
	s.limits = SatSolver3Limits{}
	s.bookmark = nil
	return nil
}

func (s *SatSolver3) Bookmark() error {
	if !s.usable() {
		return ErrInvariant
	}
	state := s.core.cloneState()
	s.bookmark = &satSolver3Bookmark{
		core: state, pushed: slices.Clone(s.pushed),
		pushConflict: slices.Clone(s.pushConflict), pivots: slices.Clone(s.pivots),
		limits: s.limits, seed: s.seed,
	}
	return nil
}

func (s *SatSolver3) Rollback() error {
	if !s.usable() || s.bookmark == nil {
		return ErrInvariant
	}
	b := s.bookmark
	state := b.core.cloneState()
	s.core = &state
	s.pushed = slices.Clone(b.pushed)
	s.pushConflict = slices.Clone(b.pushConflict)
	s.pivots = slices.Clone(b.pivots)
	s.limits = b.limits
	s.seed = b.seed
	return nil
}

func (s *SatSolver3) MemoryUsage(includeWatchEntries bool) (SatSolverMemoryUsage, error) {
	if !s.usable() {
		return SatSolverMemoryUsage{}, ErrInvariant
	}
	usage, err := s.core.MemoryUsage(includeWatchEntries)
	if err != nil {
		return SatSolverMemoryUsage{}, err
	}
	usage.FacadeStruct = unsafe.Sizeof(*s)
	usage.FacadeVariables = satCapacityBytes(s.pushed) + satCapacityBytes(s.pushConflict) + satCapacityBytes(s.pivots) + satCapacityBytes(s.userVariables) + satCapacityBytes(s.userValues)
	usage.FacadeArena = satCapacityBytes(s.printedClauses)
	for _, clause := range s.printedClauses {
		usage.FacadeLiterals += satCapacityBytes(clause)
	}
	if s.bookmark != nil {
		bookmark := satSolverMemoryUsage(&s.bookmark.core, includeWatchEntries, false)
		usage.FacadeBookmark = unsafe.Sizeof(*s.bookmark) + bookmark.Total - bookmark.SolverStruct +
			satCapacityBytes(s.bookmark.pushed) + satCapacityBytes(s.bookmark.pushConflict) + satCapacityBytes(s.bookmark.pivots)
	}
	usage.setTotal()
	return usage, nil
}

func (s *SatSolver3) Stop() {
	if s == nil {
		return
	}
	variables := s.VariableCount()
	if s.trace != nil {
		_ = s.trace.Stop(variables)
	}
	if s.store != nil {
		s.store.Stop()
	}
	s.stopped = true
	s.core = nil
	s.pushed = nil
	s.pushConflict = nil
	s.pivots = nil
	s.userVariables = nil
	s.userValues = nil
	s.stopForRun = nil
	s.printedClauses = nil
	s.bookmark = nil
}

func (s *SatSolver3) SetPivotVariables(variables []SatVar) error {
	if !s.usable() {
		return ErrInvariant
	}
	for _, variable := range variables {
		if uint64(variable) >= uint64(s.VariableCount()) {
			return ErrInvalidLiteral
		}
	}
	s.pivots = slices.Clone(variables)
	return nil
}

func (s *SatSolver3) PivotVariables() []SatVar {
	if !s.usable() {
		return nil
	}
	return slices.Clone(s.pivots)
}

// CountUsedVariables reports and clears the variables touched by enqueue since
// the preceding call, matching solver3's pFreqs bookkeeping.
func (s *SatSolver3) CountUsedVariables() (int, error) {
	if !s.usable() {
		return 0, ErrInvariant
	}
	count := 0
	for index, used := range s.core.used {
		if used {
			count++
			s.core.used[index] = false
		}
	}
	return count, nil
}

func (s *SatSolver3) SolveLexicographic(literals []SatLit, limits SatSolver3Limits) (SatSolver3LexResult, error) {
	if !s.usable() || len(literals) == 0 {
		return SatSolver3LexResult{}, ErrDimension
	}
	seen := make([]bool, s.VariableCount())
	for _, literal := range literals {
		if !literal.Valid() || uint64(literal.Var()) >= uint64(len(seen)) || seen[literal.Var()] {
			return SatSolver3LexResult{}, ErrInvalidLiteral
		}
		seen[literal.Var()] = true
	}
	query := func(prefix []SatLit) (SatResult, error) { return s.solveWith(prefix, limits) }
	result, err := query(literals)
	if err != nil || result.Status == SatSatisfiable || result.Status == SatUnknown {
		return SatSolver3LexResult{result, slices.Clone(literals)}, err
	}
	for pivot := len(literals) - 1; pivot >= 0; pivot-- {
		if !literals[pivot].Negative() {
			continue
		}
		prefix := slices.Clone(literals[:pivot])
		prefix = append(prefix, literals[pivot].Not())
		result, err = query(prefix)
		if err != nil {
			return SatSolver3LexResult{}, err
		}
		if result.Status == SatUnknown {
			return SatSolver3LexResult{Result: result}, nil
		}
		if result.Status != SatSatisfiable {
			continue
		}
		for suffix := pivot + 1; suffix < len(literals); suffix++ {
			zero, _ := MakeSatLit(literals[suffix].Var(), true)
			candidate := append(slices.Clone(prefix), zero)
			probe, probeErr := query(candidate)
			if probeErr != nil {
				return SatSolver3LexResult{}, probeErr
			}
			if probe.Status == SatUnknown {
				return SatSolver3LexResult{Result: probe}, nil
			}
			if probe.Status == SatSatisfiable {
				prefix, result = candidate, probe
			} else {
				one := zero.Not()
				prefix = append(prefix, one)
				result, err = query(prefix)
				if err != nil || result.Status != SatSatisfiable {
					return SatSolver3LexResult{Result: result}, err
				}
			}
		}
		return SatSolver3LexResult{result, prefix}, nil
	}
	return SatSolver3LexResult{Result: result, Literals: slices.Clone(literals)}, nil
}

// MinimizeAssumptions returns a deterministic order-preserving irredundant
// UNSAT subset. Unknown deletion queries conservatively retain the literal.
func (s *SatSolver3) MinimizeAssumptions(literals []SatLit, limits SatSolver3Limits) ([]SatLit, error) {
	if !s.usable() {
		return nil, ErrInvariant
	}
	current := slices.Clone(literals)
	initial, err := s.solveWith(current, limits)
	if err != nil {
		return nil, err
	}
	if initial.Status != SatUnsatisfiable {
		return nil, ErrInvariant
	}
	for index := 0; index < len(current); {
		candidate := make([]SatLit, 0, len(current)-1)
		candidate = append(candidate, current[:index]...)
		candidate = append(candidate, current[index+1:]...)
		result, solveErr := s.solveWith(candidate, limits)
		if solveErr != nil {
			return nil, solveErr
		}
		if result.Status == SatUnsatisfiable {
			current = candidate
			continue
		}
		index++
	}
	return current, nil
}

// MinimizeAssumptionsWithUnits is the sound counterpart of the specialized
// source minimizer. When dropping a literal, it may prove that literal is a
// base-formula consequence and commit it as a unit. The source instead adds
// the complemented assumption that was just proved inconsistent, which can
// make a satisfiable formula inconsistent; this port intentionally corrects it.
func (s *SatSolver3) MinimizeAssumptionsWithUnits(literals []SatLit, limits SatSolver3Limits) (SatSolver3MinimizeResult, error) {
	if !s.usable() || len(s.pushed) != 0 {
		return SatSolver3MinimizeResult{}, ErrInvariant
	}
	for _, literal := range literals {
		if !literal.Valid() || uint64(literal.Var()) >= uint64(s.VariableCount()) {
			return SatSolver3MinimizeResult{}, ErrInvalidLiteral
		}
	}
	base, err := s.solveWith(nil, limits)
	if err != nil {
		return SatSolver3MinimizeResult{}, err
	}
	if base.Status == SatUnknown {
		return SatSolver3MinimizeResult{}, ErrLimit
	}
	if base.Status != SatSatisfiable {
		return SatSolver3MinimizeResult{}, ErrInvariant
	}
	initial, err := s.solveWith(literals, limits)
	if err != nil {
		return SatSolver3MinimizeResult{}, err
	}
	if initial.Status == SatUnknown {
		return SatSolver3MinimizeResult{}, ErrLimit
	}
	if initial.Status != SatUnsatisfiable {
		return SatSolver3MinimizeResult{}, ErrInvariant
	}
	current := slices.Clone(literals)
	units := make([]SatLit, 0)
	for index := 0; index < len(current); {
		removed := current[index]
		candidate := make([]SatLit, 0, len(current)-1)
		candidate = append(candidate, current[:index]...)
		candidate = append(candidate, current[index+1:]...)
		result, solveErr := s.solveWith(candidate, limits)
		if solveErr != nil {
			return SatSolver3MinimizeResult{}, solveErr
		}
		if result.Status != SatUnsatisfiable {
			index++
			continue
		}
		current = candidate
		probe, probeErr := s.solveWith([]SatLit{removed.Not()}, limits)
		if probeErr != nil {
			return SatSolver3MinimizeResult{}, probeErr
		}
		if probe.Status == SatUnsatisfiable && !slices.Contains(units, removed) {
			units = append(units, removed)
		}
	}
	if len(current) == 0 {
		return SatSolver3MinimizeResult{}, ErrInvariant
	}
	trial := s.core.cloneState()
	for _, unit := range units {
		ok, addErr := trial.AddClause([]SatLit{unit})
		if addErr != nil {
			return SatSolver3MinimizeResult{}, addErr
		}
		if !ok {
			return SatSolver3MinimizeResult{}, ErrInvariant
		}
	}
	s.core = &trial
	return SatSolver3MinimizeResult{Literals: current, LearnedUnits: units}, nil
}
