package abc

// Port of the concrete solver surface in src/sat/satoko. The search engine is
// backed by the package's owned watched-clause CDCL core; Satoko's storage,
// lifecycle, assumptions, budgets, DIMACS and bookkeeping remain concrete.

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"time"
)

type SatokoScaledDouble uint64

const (
	SatokoScaledOne SatokoScaledDouble = 0x0000800000000000
	SatokoScaledMax SatokoScaledDouble = ^SatokoScaledDouble(0)
)

func (v SatokoScaledDouble) Exponent() uint16 { return uint16(uint64(v) >> 48) }
func (v SatokoScaledDouble) Mantissa() uint64 { return uint64(v) & ((uint64(1) << 48) - 1) }

func (v SatokoScaledDouble) Float64() float64 {
	return math.Ldexp(float64(v.Mantissa()), int(v.Exponent())-47)
}

func MakeSatokoScaledDouble(value float64) (SatokoScaledDouble, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 1 {
		return 0, ErrDimension
	}
	fraction, exponent := math.Frexp(value)
	exponent--
	if exponent < 0 || exponent > math.MaxUint16 {
		return 0, ErrLimit
	}
	mantissa := uint64(math.Ldexp(fraction, 48))
	if mantissa < uint64(1)<<47 {
		return 0, ErrInvariant
	}
	return SatokoScaledDouble(uint64(exponent)<<48 | mantissa), nil
}

func (v SatokoScaledDouble) Add(other SatokoScaledDouble) SatokoScaledDouble {
	a, b := uint64(v), uint64(other)
	if a < b {
		a, b = b, a
	}
	exponent := a >> 48
	difference := exponent - (b >> 48)
	shifted := uint64(0)
	if difference < 64 {
		shifted = (b & ((uint64(1) << 48) - 1)) >> difference
	}
	mantissa := (a & ((uint64(1) << 48) - 1)) + shifted
	if mantissa>>48 != 0 {
		exponent++
		mantissa >>= 1
	}
	if exponent > math.MaxUint16 {
		return SatokoScaledMax
	}
	return SatokoScaledDouble(exponent<<48 | mantissa)
}

func (v SatokoScaledDouble) Multiply(other SatokoScaledDouble) SatokoScaledDouble {
	a, b := uint64(v), uint64(other)
	if a < b {
		a, b = b, a
	}
	am, bm := a&((uint64(1)<<48)-1), b&((uint64(1)<<48)-1)
	ahi, bhi := am>>32, bm>>32
	alo, blo := uint32(am), uint32(bm)
	mantissa := ((ahi * bhi) << 17) + ((uint64(alo) * uint64(blo)) >> 47) + ((uint64(alo) * bhi) >> 15) + ((ahi * uint64(blo)) >> 15)
	exponent := (a >> 48) + (b >> 48)
	if mantissa>>48 != 0 {
		exponent++
		mantissa >>= 1
	}
	if exponent > math.MaxUint16 {
		return SatokoScaledMax
	}
	return SatokoScaledDouble(exponent<<48 | mantissa)
}

func (v SatokoScaledDouble) DividePowerOfTwo(power uint16) SatokoScaledDouble {
	if v.Exponent() < power {
		return SatokoScaledOne
	}
	return SatokoScaledDouble(uint64(v.Exponent()-power)<<48 | v.Mantissa())
}

type SatokoOptions struct {
	ConflictLimit, PropagationLimit uint64
	RestartForce, RestartBlock      float64
	FirstBlockedRestart             uint32
	LBDQueueSize, TrailQueueSize    uint32
	FirstReduction, ReductionStep   uint32
	SpecialReductionStep            uint32
	FreezeLBD                       uint32
	LearnedRatio                    float32
	VariableDecay                   float64
	ClauseDecay                     float32
	VariableActivityRescale         uint32
	VariableActivityLimit           uint64
	BinaryResolutionMaxSize         uint32
	BinaryResolutionMinLBD          uint32
	GarbageRatio                    float32
	Verbose, NoSimplify             bool
}

func DefaultSatokoOptions() SatokoOptions {
	return SatokoOptions{
		RestartForce: 0.8, RestartBlock: 1.4, FirstBlockedRestart: 10000,
		LBDQueueSize: 50, TrailQueueSize: 5000,
		FirstReduction: 2000, ReductionStep: 300, SpecialReductionStep: 1000,
		FreezeLBD: 30, LearnedRatio: 0.5,
		VariableDecay: 0.95, ClauseDecay: 0.995,
		VariableActivityRescale: 200, VariableActivityLimit: 0x014c924d692ca61b,
		BinaryResolutionMaxSize: 30, BinaryResolutionMinLBD: 6,
		GarbageRatio: 0.3,
	}
}

type SatokoLimits struct {
	MaxVariables, MaxClauses, MaxLiterals int
	MaxDIMACSBytes                        int
	Runtime                               time.Duration
	Stop                                  func() bool
}

func normalizeSatokoLimits(l SatokoLimits) (SatokoLimits, error) {
	if l.MaxVariables < 0 || l.MaxClauses < 0 || l.MaxLiterals < 0 || l.MaxDIMACSBytes < 0 || l.Runtime < 0 {
		return l, ErrDimension
	}
	if l.MaxVariables == 0 {
		l.MaxVariables = 1 << 20
	}
	if l.MaxClauses == 0 {
		l.MaxClauses = 1 << 20
	}
	if l.MaxLiterals == 0 {
		l.MaxLiterals = 4 << 20
	}
	if l.MaxDIMACSBytes == 0 {
		l.MaxDIMACSBytes = 64 << 20
	}
	if uint64(l.MaxVariables) > uint64(MaxLiteralID)+1 {
		return l, ErrLimit
	}
	return l, nil
}

func checkSatokoOptions(o SatokoOptions) error {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	if !finite(o.RestartForce) || !finite(o.RestartBlock) || !finite(o.VariableDecay) ||
		!finite(float64(o.ClauseDecay)) || !finite(float64(o.LearnedRatio)) || !finite(float64(o.GarbageRatio)) ||
		o.RestartForce <= 0 || o.RestartBlock <= 0 || o.VariableDecay <= 0 || o.VariableDecay >= 1 ||
		o.ClauseDecay <= 0 || o.ClauseDecay >= 1 || o.LearnedRatio <= 0 || o.LearnedRatio > 1 ||
		o.GarbageRatio < 0 || o.GarbageRatio > 1 || o.LBDQueueSize == 0 || o.TrailQueueSize == 0 {
		return ErrDimension
	}
	return nil
}

type SatokoStats struct {
	Starts, Reductions uint64
	Decisions          uint64
	Propagations       uint64
	AllPropagations    uint64
	Inspections        uint64
	Conflicts          uint64
	AllConflicts       uint64
	OriginalLiterals   int
	LearnedLiterals    int
}

type SatokoResult struct {
	Status   SatStatus
	Reason   SatStopReason
	Model    []bool
	Conflict []SatLit
	Stats    SatokoStats
}

type satokoBookmark struct {
	assumptions []SatLit
	marks       []bool
	marked      bool
	clauses     int
	literals    int
}

type SatokoSolver struct {
	core        *SatSolver
	options     SatokoOptions
	limits      SatokoLimits
	assumptions []SatLit
	marks       []bool
	marked      bool
	bookmark    *satokoBookmark
	clauses     int
	literals    int
	runID       int
	stopRun     func(int) bool
	closed      bool
}

func NewSatokoSolver(limits SatokoLimits) (*SatokoSolver, error) {
	limits, err := normalizeSatokoLimits(limits)
	if err != nil {
		return nil, err
	}
	core, err := NewSatSolverWithLimit(limits.MaxVariables)
	if err != nil {
		return nil, err
	}
	s := &SatokoSolver{core: core, options: DefaultSatokoOptions(), limits: limits}
	if err = s.applyOptions(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SatokoSolver) live() error {
	if s == nil || s.closed || s.core == nil {
		return ErrInvariant
	}
	return nil
}

func (s *SatokoSolver) applyOptions() error {
	if err := checkSatokoOptions(s.options); err != nil {
		return err
	}
	ratio := int(math.Round(float64(s.options.LearnedRatio) * 100))
	if err := s.core.SetLearnedReductionPolicy(int(s.options.FirstReduction), int(s.options.ReductionStep), ratio); err != nil {
		return err
	}
	return nil
}

func (s *SatokoSolver) Configure(options SatokoOptions) error {
	if err := s.live(); err != nil {
		return err
	}
	if err := checkSatokoOptions(options); err != nil {
		return err
	}
	old := s.options
	s.options = options
	if err := s.applyOptions(); err != nil {
		s.options = old
		_ = s.applyOptions()
		return err
	}
	return nil
}

func (s *SatokoSolver) Options() (SatokoOptions, error) {
	if err := s.live(); err != nil {
		return SatokoOptions{}, err
	}
	return s.options, nil
}

func (s *SatokoSolver) VariableCount() int {
	if s == nil || s.core == nil || s.closed {
		return 0
	}
	return s.core.VariableCount()
}

func (s *SatokoSolver) ClauseCount() int {
	if s == nil || s.closed || s.core == nil {
		return 0
	}
	return s.clauses
}

func (s *SatokoSolver) LearnedClauseCount() int {
	if s == nil || s.core == nil || s.closed {
		return 0
	}
	return s.core.LearnedClauseCount()
}

func (s *SatokoSolver) ConflictCount() uint64 {
	if s == nil || s.core == nil || s.closed {
		return 0
	}
	return s.core.ConflictCount()
}

func (s *SatokoSolver) Statistics() (SatokoStats, error) {
	if err := s.live(); err != nil {
		return SatokoStats{}, err
	}
	return satokoStatsFromCore(s.core.stats, s.core), nil
}

func (s *SatokoSolver) SetVariableCount(count int) error {
	if err := s.live(); err != nil {
		return err
	}
	if count < 0 {
		return ErrDimension
	}
	if count > s.limits.MaxVariables {
		return ErrLimit
	}
	before := s.core.VariableCount()
	if err := s.core.SetVariableCount(count); err != nil {
		return err
	}
	for i := before; i < s.core.VariableCount(); i++ {
		s.core.polarity[i] = true
	}
	if len(s.marks) < s.core.VariableCount() {
		s.marks = append(s.marks, make([]bool, s.core.VariableCount()-len(s.marks))...)
	}
	return nil
}

func (s *SatokoSolver) AddVariable(preferNegative bool) (SatVar, error) {
	if err := s.live(); err != nil {
		return 0, err
	}
	variable := s.core.VariableCount()
	if err := s.SetVariableCount(variable + 1); err != nil {
		return 0, err
	}
	s.core.polarity[variable] = !preferNegative
	return SatVar(variable), nil
}

func satokoClauseTotals(core *SatSolver) (int, int) {
	clauses, literals := 0, 0
	for id := 1; id < len(core.clauses); id++ {
		clause := &core.clauses[id]
		if clause.deleted || clause.learnt {
			continue
		}
		clauses++
		literals += len(clause.lits)
	}
	return clauses, literals
}

func (s *SatokoSolver) AddClause(literals []SatLit) (bool, error) {
	if err := s.live(); err != nil {
		return false, err
	}
	if s.limits.Stop != nil && s.limits.Stop() {
		return false, ErrCancelled
	}
	for _, literal := range literals {
		if !literal.Valid() {
			return false, ErrInvalidLiteral
		}
		if uint64(literal.Var()) >= uint64(s.limits.MaxVariables) {
			return false, ErrLimit
		}
	}
	work := s.core.cloneState()
	work.bookmark = s.core.bookmark
	before := work.VariableCount()
	ok, err := work.AddClause(literals)
	if err != nil {
		return false, err
	}
	for variable := before; variable < work.VariableCount(); variable++ {
		work.polarity[variable] = true
	}
	clauses, count := satokoClauseTotals(&work)
	if clauses > s.limits.MaxClauses || count > s.limits.MaxLiterals {
		return false, ErrLimit
	}
	s.core = &work
	s.clauses, s.literals = clauses, count
	if len(s.marks) < s.core.VariableCount() {
		s.marks = append(s.marks, make([]bool, s.core.VariableCount()-len(s.marks))...)
	}
	return ok, nil
}

func (s *SatokoSolver) PushAssumption(literal SatLit) error {
	if err := s.live(); err != nil {
		return err
	}
	if !literal.ValidForVariables(s.core.VariableCount()) {
		return ErrInvalidLiteral
	}
	s.assumptions = append(s.assumptions, literal)
	s.core.polarity[literal.Var()] = !literal.Negative()
	return nil
}

func (s *SatokoSolver) PopAssumption() (SatLit, error) {
	if err := s.live(); err != nil {
		return 0, err
	}
	if len(s.assumptions) == 0 {
		return 0, ErrDimension
	}
	last := s.assumptions[len(s.assumptions)-1]
	s.assumptions = s.assumptions[:len(s.assumptions)-1]
	return last, nil
}

func minSatokoBudget(a, b uint64) uint64 {
	if a == 0 {
		return b
	}
	if b == 0 || a < b {
		return a
	}
	return b
}

func (s *SatokoSolver) Solve() (SatokoResult, error) {
	return s.SolveAssumptions(nil, SatLimits{})
}

func (s *SatokoSolver) SolveAssumptions(extra []SatLit, limits SatLimits) (SatokoResult, error) {
	if err := s.live(); err != nil {
		return SatokoResult{}, err
	}
	assumptions := append(slices.Clone(s.assumptions), extra...)
	for _, literal := range assumptions {
		if !literal.ValidForVariables(s.core.VariableCount()) {
			return SatokoResult{}, ErrInvalidLiteral
		}
	}
	limits.Conflicts = minSatokoBudget(limits.Conflicts, s.options.ConflictLimit)
	limits.Propagations = minSatokoBudget(limits.Propagations, s.options.PropagationLimit)
	configuredStop := limits.Stop
	limits.Stop = func() bool {
		return s.limits.Stop != nil && s.limits.Stop() || configuredStop != nil && configuredStop() || s.stopRun != nil && s.stopRun(s.runID)
	}
	previous := s.core.runtimeDeadline
	if s.limits.Runtime > 0 {
		s.core.runtimeDeadline = time.Now().Add(s.limits.Runtime)
	}
	result, err := s.core.Solve(assumptions, limits)
	s.core.runtimeDeadline = previous
	if err != nil {
		return SatokoResult{}, err
	}
	return SatokoResult{
		Status: result.Status, Reason: result.Reason,
		Model: slices.Clone(result.Model), Conflict: slices.Clone(result.Conflict),
		Stats: satokoStatsFromCore(result.Stats, s.core),
	}, nil
}

func (s *SatokoSolver) SolveAssumptionsLimit(extra []SatLit, conflicts uint64) (SatokoResult, error) {
	return s.SolveAssumptions(extra, SatLimits{Conflicts: conflicts})
}

func (s *SatokoSolver) SetStop(stop func() bool) error {
	if err := s.live(); err != nil {
		return err
	}
	s.limits.Stop = stop
	return nil
}

func (s *SatokoSolver) SetStopForRun(stop func(int) bool) error {
	if err := s.live(); err != nil {
		return err
	}
	s.stopRun = stop
	return nil
}

func (s *SatokoSolver) SetRunID(id int) error {
	if err := s.live(); err != nil {
		return err
	}
	s.runID = id
	return nil
}

func (s *SatokoSolver) SetRuntimeLimit(limit time.Duration) (time.Duration, error) {
	if err := s.live(); err != nil {
		return 0, err
	}
	if limit < 0 {
		return s.limits.Runtime, ErrDimension
	}
	previous := s.limits.Runtime
	s.limits.Runtime = limit
	return previous, nil
}

func satokoStatsFromCore(delta SatStats, core *SatSolver) SatokoStats {
	original, learned := 0, 0
	for id := 1; id < len(core.clauses); id++ {
		c := &core.clauses[id]
		if c.deleted {
			continue
		}
		if c.learnt {
			learned += len(c.lits)
		} else {
			original += len(c.lits)
		}
	}
	return SatokoStats{Starts: delta.Restarts, Reductions: delta.Deleted, Decisions: delta.Decisions,
		Propagations: delta.Propagations, AllPropagations: core.stats.Propagations,
		Inspections: delta.Inspections, Conflicts: delta.Conflicts, AllConflicts: core.stats.Conflicts,
		OriginalLiterals: original, LearnedLiterals: learned}
}

func (s *SatokoSolver) Simplify() (bool, error) {
	if err := s.live(); err != nil {
		return false, err
	}
	return s.core.Simplify()
}

func (s *SatokoSolver) FinalConflict() ([]SatLit, error) {
	if err := s.live(); err != nil {
		return nil, err
	}
	return s.core.FinalConflict()
}

func (s *SatokoSolver) ModelValue(variable SatVar) (bool, error) {
	if err := s.live(); err != nil {
		return false, err
	}
	return s.core.ModelValue(variable)
}

func (s *SatokoSolver) VariablePolarity(variable SatVar) (preferNegative bool, err error) {
	if err = s.live(); err != nil {
		return false, err
	}
	if uint64(variable) >= uint64(s.core.VariableCount()) {
		return false, ErrDimension
	}
	return !s.core.polarity[variable], nil
}

func (s *SatokoSolver) CounterexampleValue(variable SatVar) (bool, error) {
	if err := s.live(); err != nil {
		return false, err
	}
	if uint64(variable) >= uint64(s.core.VariableCount()) {
		return false, ErrDimension
	}
	return s.core.polarity[variable], nil
}

func (s *SatokoSolver) MinimizeAssumptions(assumptions []SatLit, limits SatLimits) ([]SatLit, error) {
	if err := s.live(); err != nil {
		return nil, err
	}
	kept := slices.Clone(assumptions)
	result, err := s.SolveAssumptions(kept, limits)
	if err != nil {
		return nil, err
	}
	if result.Status == SatUnknown {
		if result.Reason == SatCancelled {
			return nil, ErrCancelled
		}
		return nil, ErrLimit
	}
	if result.Status != SatUnsatisfiable {
		return nil, ErrInvariant
	}
	for i := 0; i < len(kept); {
		trial := append(slices.Clone(kept[:i]), kept[i+1:]...)
		result, err = s.SolveAssumptions(trial, limits)
		if err != nil {
			return nil, err
		}
		if result.Status == SatUnknown {
			if result.Reason == SatCancelled {
				return nil, ErrCancelled
			}
			return nil, ErrLimit
		}
		if result.Status == SatUnsatisfiable {
			kept = trial
			continue
		}
		i++
	}
	return kept, nil
}

func (s *SatokoSolver) Bookmark() error {
	if err := s.live(); err != nil {
		return err
	}
	if err := s.core.Bookmark(); err != nil {
		return err
	}
	s.bookmark = &satokoBookmark{assumptions: slices.Clone(s.assumptions), marks: slices.Clone(s.marks), marked: s.marked, clauses: s.clauses, literals: s.literals}
	return nil
}

func (s *SatokoSolver) Rollback() error {
	if err := s.live(); err != nil {
		return err
	}
	if s.bookmark == nil {
		return ErrInvariant
	}
	if err := s.core.Rollback(); err != nil {
		return err
	}
	s.assumptions = slices.Clone(s.bookmark.assumptions)
	s.marks = slices.Clone(s.bookmark.marks)
	s.marked = s.bookmark.marked
	s.clauses, s.literals = s.bookmark.clauses, s.bookmark.literals
	return nil
}

func (s *SatokoSolver) Unbookmark() error {
	if err := s.live(); err != nil {
		return err
	}
	s.bookmark = nil
	s.core.bookmark = nil
	return nil
}

func (s *SatokoSolver) MarkCone(variables []SatVar) error {
	if err := s.live(); err != nil {
		return err
	}
	next := slices.Clone(s.marks)
	for _, variable := range variables {
		if uint64(variable) >= uint64(len(next)) {
			return ErrDimension
		}
		next[variable] = true
	}
	s.marks, s.marked = next, true
	preferred := make([]SatVar, 0, len(next))
	for variable, marked := range next {
		if marked {
			preferred = append(preferred, SatVar(variable))
		}
	}
	return s.core.SetGlobalActivityVariables(preferred)
}

func (s *SatokoSolver) UnmarkCone(variables []SatVar) error {
	if err := s.live(); err != nil {
		return err
	}
	if !s.marked {
		return ErrInvariant
	}
	next := slices.Clone(s.marks)
	for _, variable := range variables {
		if uint64(variable) >= uint64(len(next)) {
			return ErrDimension
		}
		next[variable] = false
	}
	s.marks = next
	preferred := make([]SatVar, 0, len(next))
	for variable, marked := range next {
		if marked {
			preferred = append(preferred, SatVar(variable))
		}
	}
	return s.core.SetGlobalActivityVariables(preferred)
}

func (s *SatokoSolver) Reset() error {
	if err := s.live(); err != nil {
		return err
	}
	core, err := NewSatSolverWithLimit(s.limits.MaxVariables)
	if err != nil {
		return err
	}
	s.core, s.assumptions, s.marks, s.bookmark = core, nil, nil, nil
	s.marked, s.clauses, s.literals = false, 0, 0
	return s.applyOptions()
}

func (s *SatokoSolver) Close() {
	if s == nil || s.closed {
		return
	}
	s.core, s.assumptions, s.marks, s.bookmark = nil, nil, nil, nil
	s.closed = true
}

type SatokoDIMACSOptions struct {
	IncludeLearned, ZeroBased bool
	MaxBytes                  int
	Stop                      func() bool
}

func (s *SatokoSolver) DIMACS(options SatokoDIMACSOptions) ([]byte, error) {
	if err := s.live(); err != nil {
		return nil, err
	}
	if options.MaxBytes < 0 {
		return nil, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = s.limits.MaxDIMACSBytes
	}
	work := s.core.cloneState()
	work.prepareRoot(nil)
	clauses := make([][]SatLit, 0, len(work.clauses)+len(work.assigns))
	for id := 1; id < len(work.clauses); id++ {
		clause := &work.clauses[id]
		if clause.deleted || clause.learnt && !options.IncludeLearned || len(clause.lits) <= 1 {
			continue
		}
		clauses = append(clauses, slices.Clone(clause.lits))
	}
	for variable, value := range work.assigns {
		if value < 0 || work.levels[variable] != 0 {
			continue
		}
		literal, err := MakeSatLit(SatVar(variable), value == 0)
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, []SatLit{literal})
	}
	if work.inconsistent {
		clauses = append(clauses, nil)
	}
	var out strings.Builder
	appendText := func(text string) error {
		if len(text) > options.MaxBytes-out.Len() {
			return ErrLimit
		}
		out.WriteString(text)
		return nil
	}
	if err := appendText(fmt.Sprintf("p cnf %d %d\n", work.VariableCount(), len(clauses))); err != nil {
		return nil, err
	}
	for _, clause := range clauses {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		text, err := FormatSatSolverClauseDIMACS(clause, !options.ZeroBased)
		if err != nil {
			return nil, err
		}
		if err = appendText(text); err != nil {
			return nil, err
		}
	}
	return []byte(out.String()), nil
}

func (s *SatokoSolver) WriteDIMACS(path string, options SatokoDIMACSOptions) error {
	if path == "" {
		return ErrDimension
	}
	data, err := s.DIMACS(options)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o666)
}

func ParseSatokoDIMACS(data []byte, limits SatokoLimits) (*SatokoSolver, error) {
	limits, err := normalizeSatokoLimits(limits)
	if err != nil {
		return nil, err
	}
	formula, err := ParseCNFDIMACS(data, CNFDIMACSReadOptions{MaxBytes: limits.MaxDIMACSBytes, Limits: CNFTransformOptions{
		MaxVariables: limits.MaxVariables, MaxClauses: limits.MaxClauses, MaxLiterals: limits.MaxLiterals, Stop: limits.Stop,
	}})
	if err != nil {
		return nil, err
	}
	return newSatokoFromFormula(formula, limits)
}

func newSatokoFromFormula(formula *CNFFormula, limits SatokoLimits) (*SatokoSolver, error) {
	if formula == nil {
		return nil, ErrDimension
	}
	solver, err := NewSatokoSolver(limits)
	if err != nil {
		return nil, err
	}
	if err = solver.SetVariableCount(formula.variables); err != nil {
		return nil, err
	}
	for _, clause := range formula.clauses {
		if _, err = solver.AddClause(clause); err != nil {
			return nil, err
		}
	}
	return solver, nil
}

func ReadSatokoDIMACSFile(path string, limits SatokoLimits) (*SatokoSolver, error) {
	if path == "" {
		return nil, ErrDimension
	}
	limits, err := normalizeSatokoLimits(limits)
	if err != nil {
		return nil, err
	}
	formula, err := ReadCNFDIMACSFile(path, CNFDIMACSReadOptions{MaxBytes: limits.MaxDIMACSBytes, Limits: CNFTransformOptions{
		MaxVariables: limits.MaxVariables, MaxClauses: limits.MaxClauses, MaxLiterals: limits.MaxLiterals, Stop: limits.Stop,
	}})
	if err != nil {
		return nil, err
	}
	return newSatokoFromFormula(formula, limits)
}
