package abc

// Port of src/sat/bsat/satSolver.c, satSolver.h and satClause.h.
// The initial concrete solver implements the watched-clause CDCL path.

import (
	"fmt"
	"math"
	"slices"
	"time"
)

type SatStatus int8

const (
	SatUnsatisfiable SatStatus = -1
	SatUnknown       SatStatus = 0
	SatSatisfiable   SatStatus = 1
)

type SatStopReason uint8

const (
	SatFinished SatStopReason = iota
	SatConflictLimit
	SatPropagationLimit
	SatDecisionLimit
	SatRuntimeLimit
	SatCancelled
)

// SatLimits are per-call logical work budgets. Zero means unlimited. Stop is
// a synchronous cancellation callback; it must not reenter or mutate the solver.
type SatLimits struct {
	Conflicts, Propagations, Decisions uint64
	Stop                               func() bool
}
type SatStats struct {
	Conflicts, Propagations, Decisions, Inspections, Restarts, Learnt, Deleted uint64
}
type SatResult struct {
	Status SatStatus
	Reason SatStopReason
	// Model is a complete variable assignment, present only after SAT.
	Model []bool
	// Conflict is a clause of negated assumptions, entailed by the base formula.
	// It is not necessarily minimal. Empty means the base formula is UNSAT.
	Conflict []SatLit
	Stats    SatStats
}
type bsatClause struct {
	lits                       []SatLit
	learnt, deleted, unwatched bool
	activity                   uint32
	lbd                        uint8
	proofRoot, proofID         int
	proofNode                  int
}

type SatWatchPruneStats struct {
	OriginalClauses, LearnedClauses int
	WatchEntries                    int
}

// SatSolver ports ABC's incremental bsat CDCL solver. Variable zero is an
// ordinary SAT variable. Clause references are stable integer indexes and
// watches use typed slices; the C tagged binary-clause optimization is not yet
// used. The solver must not be copied or used concurrently.
type SatSolver struct {
	bookmark                  *bsatBookmark
	proof                     *satSolverProofInfo
	resolutionProof           *satSolverResolutionProof
	clauses                   []bsatClause // entry zero is the no-reason sentinel
	watches                   [][]int
	units                     []int
	assigns                   []int8 // -1 undefined, 0 false, 1 true
	levels, reasons           []int
	polarity                  []bool
	used                      []bool
	enqueueMask               []bool
	activity                  []uint64
	activityFactors           []float64
	activityFactorVariables   []SatVar
	globalActivityVariables   []bool
	order, orderpos           []int
	seen                      []bool
	trail                     []SatLit
	trailLim                  []int
	model                     []bool
	lastConflict              []SatLit
	qhead                     int
	inconsistent, solving     bool
	lastStatus                SatStatus
	runtimeDeadline           time.Time
	maxVariables              int
	varInc                    uint64
	clauseInc                 uint32
	randomSeed                float64
	randomDecisions           bool
	restartInterval           uint64
	activitySearches          uint64
	learntLimit, learntCount  int
	learntStart, learntDelta  int
	learntRatio, dbReductions int
	proofRootConflict         []int
	stats                     SatStats
}

func NewSatSolver() *SatSolver {
	s, _ := NewSatSolverWithLimit(1 << 20)
	return s
}
func NewSatSolverWithLimit(maxVariables int) (*SatSolver, error) {
	if maxVariables < 0 {
		return nil, ErrDimension
	}
	if uint64(maxVariables) > uint64(MaxLiteralID)+1 {
		return nil, ErrLimit
	}
	return &SatSolver{clauses: make([]bsatClause, 1), maxVariables: maxVariables, varInc: 32, clauseInc: 2048, randomSeed: 91648253, randomDecisions: true, restartInterval: 100, learntLimit: SatLearntMaxStartDefault, learntStart: SatLearntMaxStartDefault, learntDelta: SatLearntMaxIncreDefault, learntRatio: SatLearntMaxRatioDefault}, nil
}
func (s *SatSolver) VariableCount() int   { return len(s.assigns) }
func (s *SatSolver) Statistics() SatStats { return s.stats }
func (s *SatSolver) OriginalClauseCount() int {
	count := 0
	for id := 1; id < len(s.clauses); id++ {
		if !s.clauses[id].learnt && !s.clauses[id].deleted {
			count++
		}
	}
	return count
}
func (s *SatSolver) LearnedClauseCount() int { return s.learntCount }
func (s *SatSolver) ConflictCount() uint64   { return s.stats.Conflicts }
func (s *SatSolver) LastStatus() SatStatus   { return s.lastStatus }
func (s *SatSolver) ModelValue(variable SatVar) (bool, error) {
	if s == nil || s.lastStatus != SatSatisfiable || uint64(variable) >= uint64(len(s.model)) {
		return false, ErrInvalidLiteral
	}
	return s.model[variable], nil
}
func (s *SatSolver) ModelLiteral(variable SatVar) (SatLit, error) {
	value, err := s.ModelValue(variable)
	if err != nil {
		return 0, err
	}
	return MakeSatLit(variable, !value)
}
func (s *SatSolver) FinalConflict() ([]SatLit, error) {
	if s == nil || s.lastStatus != SatUnsatisfiable {
		return nil, ErrInvariant
	}
	return slices.Clone(s.lastConflict), nil
}
func (s *SatSolver) SetRuntimeDeadline(deadline time.Time) (time.Time, error) {
	if s == nil {
		return time.Time{}, ErrDimension
	}
	if s.solving {
		return s.runtimeDeadline, ErrInvariant
	}
	previous := s.runtimeDeadline
	s.runtimeDeadline = deadline
	return previous, nil
}
func (s *SatSolver) SetLearnedClauseLimit(limit int) (int, error) {
	if s == nil || limit < 0 {
		return 0, ErrDimension
	}
	if s.solving {
		return s.learntLimit, ErrInvariant
	}
	previous := s.learntLimit
	s.learntLimit = limit
	return previous, nil
}
func (s *SatSolver) SetLearnedReductionPolicy(start, delta, ratio int) error {
	if s == nil || start < 0 || delta < 0 || ratio <= 0 || ratio > 100 {
		return ErrDimension
	}
	if s.solving {
		return ErrInvariant
	}
	s.learntStart, s.learntDelta, s.learntRatio = start, delta, ratio
	s.learntLimit = start
	return nil
}

// SetActivityFactors installs the legacy solver's per-variable search bias.
// The factors slice has one entry per solver variable; variables selects the
// entries that are bumped at odd-numbered search starts. Inputs are copied and
// rejected atomically.
func (s *SatSolver) SetActivityFactors(factors []float64, variables []SatVar) error {
	if s == nil || len(factors) != len(s.assigns) {
		return ErrDimension
	}
	if s.solving {
		return ErrInvariant
	}
	seen := make([]bool, len(s.assigns))
	for _, factor := range factors {
		if math.IsNaN(factor) || math.IsInf(factor, 0) || factor < 0 {
			return ErrDimension
		}
	}
	for _, variable := range variables {
		if uint64(variable) >= uint64(len(seen)) || seen[variable] {
			return ErrDimension
		}
		seen[variable] = true
	}
	s.activityFactors = slices.Clone(factors)
	s.activityFactorVariables = slices.Clone(variables)
	return nil
}

func (s *SatSolver) ClearActivityFactors() error {
	if s == nil {
		return ErrDimension
	}
	if s.solving {
		return ErrInvariant
	}
	s.activityFactors = nil
	s.activityFactorVariables = nil
	return nil
}

// SetGlobalActivityVariables installs the experimental interpolation bias.
// Marked variables receive a factor-three bump at every search start when
// they are also present in the activity-factor variable list.
func (s *SatSolver) SetGlobalActivityVariables(variables []SatVar) error {
	if s == nil {
		return ErrDimension
	}
	if s.solving {
		return ErrInvariant
	}
	marked := make([]bool, len(s.assigns))
	for _, variable := range variables {
		if uint64(variable) >= uint64(len(marked)) || marked[variable] {
			return ErrDimension
		}
		marked[variable] = true
	}
	if len(variables) == 0 {
		marked = nil
	}
	s.globalActivityVariables = marked
	return nil
}
func (s *SatSolver) ClearVariableActivity() error {
	if s == nil {
		return ErrDimension
	}
	if s.solving {
		return ErrInvariant
	}
	for variable := range s.activity {
		s.activity[variable] = 0
	}
	s.varInc = 1
	return nil
}

// Simplify completes level-zero propagation without starting a search. False
// reports a logical root conflict; checked API failures are returned as errors.
func (s *SatSolver) Simplify() (bool, error) {
	if s == nil {
		return false, ErrDimension
	}
	if s.solving || len(s.clauses) == 0 {
		return false, ErrInvariant
	}
	ok, stopped := s.prepareRoot(nil)
	if stopped {
		return false, ErrInvariant
	}
	return ok, nil
}

// VerifyModel checks the retained SAT assignment against every live original
// logical clause. It returns an error instead of printing a failed clause.
func (s *SatSolver) VerifyModel() error {
	if s == nil || s.lastStatus != SatSatisfiable || len(s.model) != len(s.assigns) {
		return ErrInvariant
	}
	for id := 1; id < len(s.clauses); id++ {
		clause := s.clauses[id]
		if clause.deleted || clause.learnt {
			continue
		}
		satisfied := false
		for _, literal := range clause.lits {
			if s.model[literal.Var()] != literal.Negative() {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return ErrInvariant
		}
	}
	return nil
}

// PruneRootSatisfiedWatches implements sat_solver2_check_watched without
// process-global diagnostics. Clauses made permanently true by level-zero
// assignments no longer need either watch entry. The update is atomic.
func (s *SatSolver) PruneRootSatisfiedWatches() (SatWatchPruneStats, error) {
	if s == nil {
		return SatWatchPruneStats{}, ErrDimension
	}
	if s.solving || len(s.clauses) == 0 {
		return SatWatchPruneStats{}, ErrInvariant
	}
	trial := s.cloneState()
	trial.bookmark = s.bookmark
	stats, err := trial.pruneRootSatisfiedWatches()
	if err != nil {
		return SatWatchPruneStats{}, err
	}
	*s = trial
	return stats, nil
}

func (s *SatSolver) pruneRootSatisfiedWatches() (SatWatchPruneStats, error) {
	if ok, _ := s.prepareRoot(nil); !ok {
		return SatWatchPruneStats{}, ErrInvariant
	}
	pruned := make([]bool, len(s.clauses))
	watches := make([][]int, len(s.watches))
	stats := SatWatchPruneStats{}
	for literal, bucket := range s.watches {
		watches[literal] = make([]int, 0, len(bucket))
		for _, id := range bucket {
			if id <= 0 || id >= len(s.clauses) {
				return SatWatchPruneStats{}, ErrInvariant
			}
			clause := &s.clauses[id]
			if clause.deleted || clause.unwatched {
				continue
			}
			satisfied := false
			for _, candidate := range clause.lits {
				if s.literalValue(candidate) == 1 && s.levels[candidate.Var()] == 0 {
					satisfied = true
					break
				}
			}
			if satisfied {
				pruned[id] = true
				stats.WatchEntries++
				continue
			}
			watches[literal] = append(watches[literal], id)
		}
	}
	for id, yes := range pruned {
		if !yes {
			continue
		}
		s.clauses[id].unwatched = true
		if s.clauses[id].learnt {
			stats.LearnedClauses++
		} else {
			stats.OriginalClauses++
		}
	}
	s.watches = watches
	return stats, nil
}
func (s *SatSolver) invalidateResult() {
	s.lastStatus = SatUnknown
	s.model = nil
	s.lastConflict = nil
}
func (s *SatSolver) Reset() error {
	if s.solving {
		return ErrInvariant
	}
	*s = *NewSatSolver()
	return nil
}
func (s *SatSolver) SetRandomSeed(seed float64) error {
	if s.solving {
		return ErrInvariant
	}
	if math.IsNaN(seed) || math.IsInf(seed, 0) || seed < 1 || seed >= 2147483647 {
		return ErrDimension
	}
	s.randomSeed = seed
	return nil
}
func (s *SatSolver) SetRandomDecisions(enabled bool) error {
	if s.solving {
		return ErrInvariant
	}
	s.randomDecisions = enabled
	return nil
}

// SetRestartInterval uses the Luby sequence with this initial conflict count.
// Zero disables periodic restarts. It does not change external work budgets.
func (s *SatSolver) SetRestartInterval(interval uint64) error {
	if s.solving {
		return ErrInvariant
	}
	s.restartInterval = interval
	return nil
}
func (s *SatSolver) SetVariableCount(count int) error {
	if s.solving {
		return ErrInvariant
	}
	return s.setVariableCount(count)
}
func (s *SatSolver) setVariableCount(count int) error {
	if count < 0 {
		return ErrDimension
	}
	if count > s.maxVariables {
		return ErrLimit
	}
	if len(s.clauses) == 0 {
		return fmt.Errorf("%w: uninitialized SAT solver", ErrInvariant)
	}
	for len(s.assigns) < count {
		v := len(s.assigns)
		s.assigns = append(s.assigns, -1)
		s.levels = append(s.levels, 0)
		s.reasons = append(s.reasons, 0)
		s.polarity = append(s.polarity, false)
		s.used = append(s.used, false)
		s.activity = append(s.activity, 1024)
		if s.activityFactors != nil {
			s.activityFactors = append(s.activityFactors, 0)
		}
		if s.globalActivityVariables != nil {
			s.globalActivityVariables = append(s.globalActivityVariables, false)
		}
		s.orderpos = append(s.orderpos, -1)
		s.seen = append(s.seen, false)
		s.watches = append(s.watches, nil, nil)
		s.heapInsert(v)
	}
	return nil
}
func (s *SatSolver) literalValue(l SatLit) int8 {
	value := s.assigns[l.Var()]
	if value < 0 {
		return -1
	}
	if l.Negative() {
		return 1 - value
	}
	return value
}
func (s *SatSolver) enqueue(l SatLit, reason int) bool {
	if s.enqueueMask != nil && (uint64(l.Var()) >= uint64(len(s.enqueueMask)) || !s.enqueueMask[l.Var()]) {
		return true
	}
	s.used[l.Var()] = true
	if value := s.literalValue(l); value >= 0 {
		return value == 1
	}
	v := int(l.Var())
	value := int8(1)
	if l.Negative() {
		value = 0
	}
	s.assigns[v] = value
	s.levels[v] = len(s.trailLim)
	s.reasons[v] = reason
	s.trail = append(s.trail, l)
	return true
}
func (s *SatSolver) decision(l SatLit) {
	s.trailLim = append(s.trailLim, len(s.trail))
	s.enqueue(l, 0)
}
func (s *SatSolver) cancelUntil(level int) {
	if len(s.trailLim) <= level {
		return
	}
	bound := s.trailLim[level]
	last := s.trailLim[len(s.trailLim)-1]
	for i := len(s.trail) - 1; i >= bound; i-- {
		l := s.trail[i]
		v := int(l.Var())
		s.assigns[v] = -1
		s.reasons[v] = 0
		s.levels[v] = 0
		if i < last {
			s.polarity[v] = !l.Negative()
		}
		s.heapInsert(v)
	}
	s.trail = s.trail[:bound]
	s.qhead = min(s.qhead, bound)
	s.trailLim = s.trailLim[:level]
}
func (s *SatSolver) newClause(lits []SatLit, learnt bool) int {
	id := len(s.clauses)
	c := bsatClause{lits: slices.Clone(lits), learnt: learnt, proofRoot: -1, proofID: -1, proofNode: -1}
	if learnt {
		c.lbd = s.computeLBD(lits)
		s.learntCount++
		s.stats.Learnt++
	}
	s.clauses = append(s.clauses, c)
	if !learnt {
		s.clauses[id].proofRoot = id
	}
	if len(lits) == 1 {
		s.units = append(s.units, id)
	} else {
		s.watches[lits[0].Not()] = append(s.watches[lits[0].Not()], id)
		s.watches[lits[1].Not()] = append(s.watches[lits[1].Not()], id)
	}
	return id
}
func (s *SatSolver) prepareRoot(run *bsatRun) (ok, stopped bool) {
	s.cancelUntil(0)
	if s.inconsistent {
		return false, false
	}
	for _, id := range s.units {
		if run.stopped(s) {
			return false, true
		}
		if !s.enqueue(s.clauses[id].lits[0], id) {
			s.inconsistent = true
			s.proofRootConflict = []int{id, s.reasons[s.clauses[id].lits[0].Var()]}
			return false, false
		}
	}
	conflict, stopped := s.propagate(run)
	if stopped {
		return false, true
	}
	if conflict != 0 {
		s.inconsistent = true
		s.proofRootConflict = []int{conflict}
		return false, false
	}
	return true, false
}

// AddClause implements sat_solver_addclause with checked literals and an owned
// copy. It normalizes duplicates/tautologies and propagates root consequences.
// An empty clause is accepted and makes the formula inconsistent. False means
// the accumulated formula is already UNSAT, not an API error.
func (s *SatSolver) AddClause(literals []SatLit) (bool, error) {
	if s.solving {
		return false, ErrInvariant
	}
	maximum := -1
	for _, l := range literals {
		if !l.Valid() {
			return false, ErrInvalidLiteral
		}
		maximum = max(maximum, int(l.Var()))
	}
	if maximum >= s.maxVariables {
		return false, ErrLimit
	}
	s.invalidateResult()
	if err := s.setVariableCount(maximum + 1); err != nil {
		return false, err
	}
	if ok, _ := s.prepareRoot(nil); !ok {
		return false, nil
	}
	sorted := slices.Clone(literals)
	slices.Sort(sorted)
	normalized := sorted[:0]
	for _, l := range sorted {
		if len(normalized) > 0 {
			last := normalized[len(normalized)-1]
			if l == last {
				continue
			}
			if l == last.Not() {
				return true, nil
			}
		}
		value := s.literalValue(l)
		if value == 1 {
			return true, nil
		}
		if value == 0 {
			continue
		}
		normalized = append(normalized, l)
	}
	if len(normalized) == 0 {
		s.inconsistent = true
		s.proofRootConflict = nil
		return false, nil
	}
	id := s.newClause(normalized, false)
	if len(normalized) == 1 {
		if !s.enqueue(normalized[0], id) {
			s.inconsistent = true
			s.proofRootConflict = []int{id, s.reasons[normalized[0].Var()]}
			return false, nil
		}
		if conflict, _ := s.propagate(nil); conflict != 0 {
			s.inconsistent = true
			s.proofRootConflict = []int{conflict}
			return false, nil
		}
	}
	return true, nil
}

type bsatRun struct {
	limits   SatLimits
	baseline SatStats
	reason   SatStopReason
}

func (r *bsatRun) stopped(s *SatSolver) bool {
	if r == nil {
		return false
	}
	switch {
	case r.limits.Stop != nil && r.limits.Stop():
		r.reason = SatCancelled
	case !s.runtimeDeadline.IsZero() && !time.Now().Before(s.runtimeDeadline):
		r.reason = SatRuntimeLimit
	case r.limits.Conflicts != 0 && s.stats.Conflicts-r.baseline.Conflicts >= r.limits.Conflicts:
		r.reason = SatConflictLimit
	case r.limits.Propagations != 0 && s.stats.Propagations-r.baseline.Propagations >= r.limits.Propagations:
		r.reason = SatPropagationLimit
	case r.limits.Decisions != 0 && s.stats.Decisions-r.baseline.Decisions >= r.limits.Decisions:
		r.reason = SatDecisionLimit
	default:
		return false
	}
	return true
}

// propagate mirrors sat_solver_propagate: watches[p] contains clauses whose
// watched literal becomes false when p is enqueued. A bucket is updated
// atomically before returning a conflict; a budget stops between buckets.
func (s *SatSolver) propagate(run *bsatRun) (conflict int, stopped bool) {
	for s.qhead < len(s.trail) {
		if run.stopped(s) {
			return 0, true
		}
		p := s.trail[s.qhead]
		s.qhead++
		s.stats.Propagations++
		bucket := s.watches[p]
		write := 0
		for read, id := range bucket {
			c := &s.clauses[id]
			if c.deleted {
				continue
			}
			s.stats.Inspections++
			falseLiteral := p.Not()
			if c.lits[0] == falseLiteral {
				c.lits[0], c.lits[1] = c.lits[1], c.lits[0]
			}
			if s.literalValue(c.lits[0]) == 1 {
				bucket[write] = id
				write++
				continue
			}
			replacement := -1
			for k := 2; k < len(c.lits); k++ {
				if s.literalValue(c.lits[k]) != 0 {
					replacement = k
					break
				}
			}
			if replacement >= 0 {
				c.lits[1], c.lits[replacement] = c.lits[replacement], c.lits[1]
				target := c.lits[1].Not()
				s.watches[target] = append(s.watches[target], id)
				continue
			}
			bucket[write] = id
			write++
			if c.learnt {
				c.lbd = s.computeLBD(c.lits)
			}
			if !s.enqueue(c.lits[0], id) {
				write += copy(bucket[write:], bucket[read+1:])
				s.watches[p] = bucket[:write]
				return id, false
			}
		}
		s.watches[p] = bucket[:write]
	}
	return 0, false
}
func (s *SatSolver) computeLBD(lits []SatLit) uint8 {
	var mask uint32
	for _, l := range lits {
		mask |= uint32(1) << uint(s.levels[l.Var()]&31)
	}
	count := uint8(0)
	for mask != 0 {
		mask &= mask - 1
		count++
	}
	return count
}
func (s *SatSolver) bumpVariable(v int) {
	s.activity[v] += s.varInc
	if uint32(s.activity[v])&0x80000000 != 0 {
		for i := range s.activity {
			s.activity[i] >>= 19
		}
		s.varInc = max(s.varInc>>19, 16)
	}
	if s.orderpos[v] >= 0 {
		s.heapUp(s.orderpos[v])
	}
}

func (s *SatSolver) bumpVariableFactor(v int, factor float64) {
	product := float64(s.varInc) * factor
	var increment uint64
	if product >= float64(math.MaxUint64) {
		increment = math.MaxUint64
	} else {
		increment = uint64(product)
	}
	if math.MaxUint64-s.activity[v] < increment {
		s.activity[v] = math.MaxUint64
	} else {
		s.activity[v] += increment
	}
	if uint32(s.activity[v])&0x80000000 != 0 {
		for i := range s.activity {
			s.activity[i] >>= 19
		}
		s.varInc = max(s.varInc>>19, 16)
	}
	if s.orderpos[v] >= 0 {
		s.heapUp(s.orderpos[v])
	}
}

func (s *SatSolver) startActivitySearch() {
	s.activitySearches++
	if s.activitySearches&1 != 0 {
		for _, variable := range s.activityFactorVariables {
			s.bumpVariableFactor(int(variable), s.activityFactors[variable])
		}
	}
	if s.globalActivityVariables != nil {
		for _, variable := range s.activityFactorVariables {
			if s.globalActivityVariables[variable] {
				s.bumpVariableFactor(int(variable), 3)
			}
		}
	}
}
func (s *SatSolver) bumpClause(id int) {
	s.clauses[id].activity += s.clauseInc
	if s.clauses[id].activity&0x80000000 != 0 {
		for i := range s.clauses {
			if s.clauses[i].learnt {
				s.clauses[i].activity >>= 14
			}
		}
		s.clauseInc = max(s.clauseInc>>14, 1024)
	}
}

// analyze resolves the conflict along the implication trail until its first
// unique implication point, then applies ABC's reason-based minimization.
func (s *SatSolver) analyze(conflict int) ([]SatLit, int, error) {
	learnt := []SatLit{0}
	tagged := make([]int, 0)
	path := 0
	index := len(s.trail) - 1
	pivot := SatLit(math.MaxUint32)
	defer func() {
		for _, v := range tagged {
			s.seen[v] = false
		}
	}()
	for {
		if conflict <= 0 || conflict >= len(s.clauses) {
			return nil, 0, ErrInvariant
		}
		clause := &s.clauses[conflict]
		if clause.learnt {
			s.bumpClause(conflict)
		}
		for _, l := range clause.lits {
			if pivot.Valid() && l.Var() == pivot.Var() {
				continue
			}
			v := int(l.Var())
			if s.seen[v] || s.levels[v] == 0 {
				continue
			}
			s.seen[v] = true
			tagged = append(tagged, v)
			s.bumpVariable(v)
			if s.levels[v] == len(s.trailLim) {
				path++
			} else {
				learnt = append(learnt, l)
			}
		}
		for index >= 0 && !s.seen[s.trail[index].Var()] {
			index--
		}
		if index < 0 || path < 1 {
			return nil, 0, ErrInvariant
		}
		pivot = s.trail[index]
		index--
		s.seen[pivot.Var()] = false
		path--
		if path == 0 {
			break
		}
		conflict = s.reasons[pivot.Var()]
	}
	learnt[0] = pivot.Not()
	var levelMask uint32
	for _, l := range learnt[1:] {
		levelMask |= 1 << uint(s.levels[l.Var()]&31)
	}
	write := 1
	for _, l := range learnt[1:] {
		if s.reasons[l.Var()] == 0 || !s.literalRemovable(int(l.Var()), levelMask, &tagged) {
			learnt[write] = l
			write++
		}
	}
	learnt = learnt[:write]
	backtrack := 0
	if len(learnt) > 1 {
		largest := 1
		for i := 2; i < len(learnt); i++ {
			if s.levels[learnt[i].Var()] > s.levels[learnt[largest].Var()] {
				largest = i
			}
		}
		learnt[1], learnt[largest] = learnt[largest], learnt[1]
		backtrack = s.levels[learnt[1].Var()]
	}
	return learnt, backtrack, nil
}
func (s *SatSolver) literalRemovable(variable int, mask uint32, tagged *[]int) bool {
	start := len(*tagged)
	stack := []int{variable}
	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		reason := s.reasons[v]
		if reason == 0 {
			return false
		}
		for _, l := range s.clauses[reason].lits {
			w := int(l.Var())
			if w == v || s.seen[w] || s.levels[w] == 0 {
				continue
			}
			if s.reasons[w] == 0 || mask&(1<<uint(s.levels[w]&31)) == 0 {
				for _, x := range (*tagged)[start:] {
					s.seen[x] = false
				}
				*tagged = (*tagged)[:start]
				return false
			}
			s.seen[w] = true
			*tagged = append(*tagged, w)
			stack = append(stack, w)
		}
	}
	return true
}

// finalConflict translates an implication conflict into negated root
// assumptions. A globally valid learned unit has a clause reason, so it can
// never be mistaken for an assumption after a later incremental solve.
func (s *SatSolver) finalConflict(lits []SatLit, extra *SatLit) []SatLit {
	seen := make([]bool, len(s.assigns))
	for _, l := range lits {
		if s.levels[l.Var()] > 0 {
			seen[l.Var()] = true
		}
	}
	result := make([]SatLit, 0)
	if extra != nil {
		result = append(result, extra.Not())
	}
	for i := len(s.trail) - 1; i >= 0; i-- {
		l := s.trail[i]
		v := int(l.Var())
		if !seen[v] {
			continue
		}
		reason := s.reasons[v]
		if reason == 0 {
			result = append(result, l.Not())
		} else {
			for _, q := range s.clauses[reason].lits {
				if q.Var() != l.Var() && s.levels[q.Var()] > 0 {
					seen[q.Var()] = true
				}
			}
		}
	}
	return result
}
func satStatsDelta(a, b SatStats) SatStats {
	return SatStats{a.Conflicts - b.Conflicts, a.Propagations - b.Propagations, a.Decisions - b.Decisions, a.Inspections - b.Inspections, a.Restarts - b.Restarts, a.Learnt - b.Learnt, a.Deleted - b.Deleted}
}

// Solve implements incremental sat_solver_solve with assumptions, first-UIP
// learning, nonchronological backtracking, phase saving and Luby restarts.
// It always removes temporary assumptions before returning; learned clauses
// remain available to future solves and added clauses.
func (s *SatSolver) Solve(assumptions []SatLit, limits SatLimits) (result SatResult, err error) {
	if s.solving {
		return result, ErrInvariant
	}
	for _, l := range assumptions {
		if !l.Valid() || uint64(l.Var()) >= uint64(len(s.assigns)) {
			return result, ErrInvalidLiteral
		}
	}
	if len(s.clauses) == 0 {
		return result, ErrInvariant
	}
	if err = s.proofBeginSolve(assumptions); err != nil {
		return result, err
	}
	if err = s.resolutionProofBeginSolve(assumptions); err != nil {
		return result, err
	}
	s.invalidateResult()
	s.solving = true
	run := &bsatRun{limits: limits, baseline: s.stats}
	defer func() {
		s.cancelUntil(0)
		s.solving = false
		result.Stats = satStatsDelta(s.stats, run.baseline)
		if err != nil {
			s.invalidateResult()
			return
		}
		s.lastStatus = result.Status
		if result.Status == SatSatisfiable {
			s.model = append([]bool{}, result.Model...)
		} else if result.Status == SatUnsatisfiable {
			s.lastConflict = slices.Clone(result.Conflict)
		}
	}()
	if run.stopped(s) {
		result.Reason = run.reason
		return result, nil
	}
	if ok, stopped := s.prepareRoot(run); stopped {
		result.Reason = run.reason
		return result, nil
	} else if !ok {
		if err = s.proofRecordFinal(s.proofRootConflict...); err != nil {
			return result, err
		}
		conflict := 0
		if len(s.proofRootConflict) > 0 {
			conflict = s.proofRootConflict[0]
		}
		if _, err = s.resolutionProofRecord(nil, conflict, true); err != nil {
			return result, err
		}
		result.Status = SatUnsatisfiable
		return result, nil
	}
	for _, l := range assumptions {
		s.trailLim = append(s.trailLim, len(s.trail))
		if !s.enqueue(l, 0) {
			result.Conflict = s.finalConflict([]SatLit{l}, &l)
			reason := s.reasons[l.Var()]
			if reason := s.reasons[l.Var()]; reason != 0 {
				err = s.proofRecordFinal(reason)
			} else {
				err = s.proofRecordFinal()
			}
			if err != nil {
				return result, err
			}
			if _, err = s.resolutionProofRecord(result.Conflict, reason, true); err != nil {
				return result, err
			}
			result.Status = SatUnsatisfiable
			return result, nil
		}
		conflict, stopped := s.propagate(run)
		if stopped {
			result.Reason = run.reason
			return result, nil
		}
		if conflict != 0 {
			result.Conflict = s.finalConflict(s.clauses[conflict].lits, nil)
			if err = s.proofRecordFinal(conflict); err != nil {
				return result, err
			}
			if _, err = s.resolutionProofRecord(result.Conflict, conflict, true); err != nil {
				return result, err
			}
			result.Status = SatUnsatisfiable
			return result, nil
		}
	}
	root := len(s.trailLim)
	s.startActivitySearch()
	restart := uint64(0)
	sinceRestart := uint64(0)
	restartBound := s.restartInterval
	for {
		conflict, stopped := s.propagate(run)
		if stopped {
			result.Reason = run.reason
			return result, nil
		}
		if conflict != 0 {
			s.stats.Conflicts++
			sinceRestart++
			if len(s.trailLim) == root {
				result.Conflict = s.finalConflict(s.clauses[conflict].lits, nil)
				if err = s.proofRecordFinal(conflict); err != nil {
					return result, err
				}
				if _, err = s.resolutionProofRecord(result.Conflict, conflict, true); err != nil {
					return result, err
				}
				result.Status = SatUnsatisfiable
				if root == 0 {
					s.inconsistent = true
				}
				return result, nil
			}
			learnt, backtrack, analysisErr := s.analyze(conflict)
			if analysisErr != nil {
				return result, analysisErr
			}
			proofID, proofErr := s.proofRecordLearned(conflict)
			if proofErr != nil {
				return result, proofErr
			}
			proofNode, proofErr := s.resolutionProofRecord(learnt, conflict, false)
			if proofErr != nil {
				return result, proofErr
			}
			s.cancelUntil(max(root, backtrack))
			id := s.newClause(learnt, true)
			s.clauses[id].proofID = proofID
			s.clauses[id].proofNode = proofNode
			if !s.enqueue(learnt[0], id) {
				return result, ErrInvariant
			}
			s.varInc += s.varInc >> 4
			s.clauseInc += s.clauseInc >> 10
			continue
		}
		if run.stopped(s) {
			result.Reason = run.reason
			return result, nil
		}
		if s.restartInterval != 0 && sinceRestart >= restartBound {
			s.cancelUntil(root)
			s.stats.Restarts++
			s.startActivitySearch()
			sinceRestart = 0
			restart++
			factor := bsatLuby(restart)
			if factor > math.MaxUint64/s.restartInterval {
				restartBound = math.MaxUint64
			} else {
				restartBound = s.restartInterval * factor
			}
			continue
		}
		if s.learntLimit != 0 && s.learntCount >= s.learntLimit {
			s.reduceDB()
		}
		variable := s.selectVariable()
		if variable < 0 {
			result.Status = SatSatisfiable
			result.Model = make([]bool, len(s.assigns))
			for i, v := range s.assigns {
				result.Model[i] = v == 1
			}
			return result, nil
		}
		s.stats.Decisions++
		l := SatLit(variable) << 1
		if !s.polarity[variable] {
			l = l.Not()
		}
		s.decision(l)
	}
}
