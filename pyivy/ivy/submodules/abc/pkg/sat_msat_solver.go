package abc

// Concrete solver surface from src/sat/msat/msatSolverCore.c.

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
)

type MsatStatus int8

const (
	MsatFalse   MsatStatus = -1
	MsatUnknown MsatStatus = 0
	MsatTrue    MsatStatus = 1
)

type MsatSolverConfig struct {
	ClauseIncrement, ClauseDecay     float64
	VariableIncrement, VariableDecay float64
	Verbose                          bool
}

type MsatSolverStats struct {
	Starts, Decisions, Propagations, Inspections, Conflicts, Successes uint64
}

type MsatSolveResult struct {
	Status     MsatStatus
	StopReason SatStopReason
	Model      []int
	Conflict   []SatLit
	Stats      MsatSolverStats
}

type MsatSolver struct {
	core           *SatSolver
	memory         *MsatStepArena
	config         MsatSolverConfig
	capacity       int
	variableLevels []int
	factors        []float32
	original       []*MsatClause
	clausesStart   int
	model          []int
	last           MsatSolveResult
	backtracks     int
	nextClause     int
	starts         uint64
	successes      uint64
	seen           []int
	seenID         int
	adjacents      [][]SatVar
	coneVariables  []SatVar
	variablesUsed  []bool
	clausesAlloc   int
	learnedAlloc   int
	closed         bool
}

func normalizeMsatSolverConfig(config MsatSolverConfig) (MsatSolverConfig, error) {
	values := []float64{config.ClauseIncrement, config.ClauseDecay, config.VariableIncrement, config.VariableDecay}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return config, ErrDimension
		}
	}
	if config.ClauseIncrement < 0 || config.ClauseDecay < 0 || config.VariableIncrement < 0 {
		return config, ErrDimension
	}
	return config, nil
}

func NewMsatSolver(capacity int, config MsatSolverConfig) (*MsatSolver, error) {
	if capacity < 0 || uint64(capacity) > uint64(MaxLiteralID)+1 {
		return nil, ErrDimension
	}
	config, err := normalizeMsatSolverConfig(config)
	if err != nil {
		return nil, err
	}
	core, err := NewSatSolverWithLimit(capacity)
	if err != nil {
		return nil, err
	}
	memory, err := NewMsatStepArena(10, MsatMemoryLimits{})
	if err != nil {
		return nil, err
	}
	used := make([]bool, capacity)
	for index := range used {
		used[index] = true
	}
	core.enqueueMask = slices.Clone(used)
	return &MsatSolver{
		core:          core,
		memory:        memory,
		config:        config,
		capacity:      capacity,
		factors:       makeMsatFactors(capacity),
		seen:          make([]int, capacity),
		seenID:        1,
		adjacents:     make([][]SatVar, capacity),
		variablesUsed: used,
	}, nil
}

func makeMsatFactors(size int) []float32 {
	result := make([]float32, size)
	for index := range result {
		result[index] = 1
	}
	return result
}

func (s *MsatSolver) checkLive() error {
	if s == nil {
		return ErrDimension
	}
	if s.closed || s.core == nil || s.memory == nil || s.memory.closed || s.capacity < 0 || s.core.maxVariables != s.capacity ||
		len(s.factors) != s.capacity || len(s.seen) != s.capacity || len(s.adjacents) != s.capacity || len(s.variablesUsed) != s.capacity ||
		len(s.variableLevels) != s.core.VariableCount() {
		return ErrInvariant
	}
	return nil
}

func (s *MsatSolver) VariableCount() int {
	if s == nil || s.closed || s.core == nil {
		return 0
	}
	return s.core.VariableCount()
}

func (s *MsatSolver) VariableCapacity() int {
	if s == nil || s.closed {
		return 0
	}
	return s.capacity
}

func (s *MsatSolver) ClauseCount() int {
	if s == nil || s.closed || s.core == nil {
		return 0
	}
	return len(s.original) + s.core.LearnedClauseCount()
}

func (s *MsatSolver) OriginalClauses() []*MsatClause {
	if s == nil || s.closed {
		return nil
	}
	return slices.Clone(s.original)
}

func (s *MsatSolver) Resize(capacity int) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if capacity < s.VariableCount() {
		return ErrDimension
	}
	if uint64(capacity) > uint64(MaxLiteralID)+1 {
		return ErrLimit
	}
	if capacity <= s.capacity {
		return nil
	}
	factors := makeMsatFactors(capacity)
	copy(factors, s.factors)
	seen := make([]int, capacity)
	copy(seen, s.seen)
	adjacents := make([][]SatVar, capacity)
	copy(adjacents, s.adjacents)
	used := make([]bool, capacity)
	for index := range used {
		used[index] = true
	}
	s.capacity = capacity
	s.factors = factors
	s.seen = seen
	s.adjacents = adjacents
	s.variablesUsed = used
	s.core.maxVariables = capacity
	s.core.enqueueMask = slices.Clone(used)
	return nil
}

func (s *MsatSolver) Clean(variables int) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if variables < 0 || variables > s.capacity {
		return ErrDimension
	}
	core, err := NewSatSolverWithLimit(s.capacity)
	if err != nil {
		return err
	}
	if err = core.SetVariableCount(variables); err != nil {
		return err
	}
	for _, clause := range s.original {
		clause.closed = true
		clause.literals = nil
	}
	s.core = core
	s.variableLevels = make([]int, variables)
	for index := range s.variableLevels {
		s.variableLevels[index] = -1
	}
	s.original = nil
	s.clausesStart = 0
	s.model = nil
	s.last = MsatSolveResult{}
	s.backtracks = 0
	s.nextClause = 0
	clear(s.seen)
	s.seenID = 1
	s.coneVariables = nil
	for index := range s.variablesUsed {
		s.variablesUsed[index] = true
	}
	s.core.enqueueMask = slices.Clone(s.variablesUsed)
	return nil
}

func (s *MsatSolver) AddVariable(level int) (SatVar, error) {
	if err := s.checkLive(); err != nil {
		return 0, err
	}
	if level < -1 {
		return 0, ErrDimension
	}
	if s.VariableCount() == s.capacity {
		next := max(1, 2*s.capacity)
		if uint64(next) > uint64(MaxLiteralID)+1 {
			return 0, ErrLimit
		}
		if err := s.Resize(next); err != nil {
			return 0, err
		}
	}
	variable := SatVar(s.VariableCount())
	if err := s.core.SetVariableCount(int(variable) + 1); err != nil {
		return 0, err
	}
	s.variableLevels = append(s.variableLevels, level)
	return variable, nil
}

func (s *MsatSolver) AddClause(literals []SatLit) (bool, error) {
	if err := s.checkLive(); err != nil {
		return false, err
	}
	for _, literal := range literals {
		if !literal.Valid() || uint64(literal.Var()) >= uint64(s.VariableCount()) {
			return false, ErrInvalidLiteral
		}
	}
	before := len(s.core.clauses)
	ok, err := s.core.AddClause(literals)
	if err != nil {
		return false, err
	}
	for identifier := before; identifier < len(s.core.clauses); identifier++ {
		stored := &s.core.clauses[identifier]
		if stored.learnt || stored.deleted || len(stored.lits) < 2 {
			continue
		}
		clause := &MsatClause{number: s.nextClause, coreID: identifier, literals: slices.Clone(stored.lits)}
		s.nextClause++
		s.original = append(s.original, clause)
	}
	return ok, nil
}

func (s *MsatSolver) Simplify() (bool, error) {
	if err := s.checkLive(); err != nil {
		return false, err
	}
	return s.core.Simplify()
}

func (s *MsatSolver) Solve(assumptions []SatLit, backtrackLimit, timeLimitSeconds int) (MsatSolveResult, error) {
	if err := s.checkLive(); err != nil {
		return MsatSolveResult{}, err
	}
	if backtrackLimit < 0 || timeLimitSeconds < 0 {
		return MsatSolveResult{}, ErrDimension
	}
	limits := SatLimits{}
	if backtrackLimit > 0 {
		limits.Conflicts = uint64(backtrackLimit)
	}
	previousDeadline := s.core.runtimeDeadline
	if timeLimitSeconds > 0 {
		s.core.runtimeDeadline = time.Now().Add(time.Duration(timeLimitSeconds) * time.Second)
	}
	result, err := s.core.Solve(assumptions, limits)
	s.core.runtimeDeadline = previousDeadline
	if err != nil {
		return MsatSolveResult{}, err
	}
	status := MsatStatus(result.Status)
	model := make([]int, 0)
	if status == MsatTrue {
		model = make([]int, len(result.Model))
		for variable, value := range result.Model {
			literal, _ := MakeSatLit(SatVar(variable), !value)
			model[variable] = int(literal)
		}
	}
	stats := MsatSolverStats{
		Starts:       result.Stats.Restarts + 1,
		Decisions:    result.Stats.Decisions,
		Propagations: result.Stats.Propagations,
		Inspections:  result.Stats.Inspections,
		Conflicts:    result.Stats.Conflicts,
	}
	if status == MsatTrue {
		stats.Successes = 1
	}
	s.starts += stats.Starts
	s.successes += stats.Successes
	s.backtracks = int(result.Stats.Conflicts)
	s.model = slices.Clone(model)
	s.last = MsatSolveResult{Status: status, StopReason: result.Reason, Model: slices.Clone(model), Conflict: slices.Clone(result.Conflict), Stats: stats}
	return s.LastResult(), nil
}

func (s *MsatSolver) LastResult() MsatSolveResult {
	if s == nil || s.closed {
		return MsatSolveResult{}
	}
	result := s.last
	result.Model = slices.Clone(result.Model)
	result.Conflict = slices.Clone(result.Conflict)
	return result
}

func (s *MsatSolver) Assignments() []int {
	if s == nil || s.closed || s.core == nil {
		return nil
	}
	result := make([]int, len(s.core.assigns))
	for variable, value := range s.core.assigns {
		result[variable] = -1
		if value >= 0 {
			literal, _ := MakeSatLit(SatVar(variable), value == 0)
			result[variable] = int(literal)
		}
	}
	return result
}

func (s *MsatSolver) Model() []int {
	if s == nil || s.closed {
		return nil
	}
	return slices.Clone(s.model)
}

func (s *MsatSolver) Backtracks() int {
	if s == nil || s.closed {
		return 0
	}
	return s.backtracks
}

func (s *MsatSolver) Inspections() uint64 {
	if s == nil || s.closed || s.core == nil {
		return 0
	}
	return s.core.stats.Inspections
}

func (s *MsatSolver) SetVerbose(verbose bool) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	s.config.Verbose = verbose
	return nil
}

func (s *MsatSolver) Factors() []float32 {
	if s == nil || s.closed {
		return nil
	}
	return slices.Clone(s.factors)
}

func (s *MsatSolver) SetFactors(factors []float32) error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if len(factors) != s.capacity {
		return ErrDimension
	}
	for _, factor := range factors {
		if math.IsNaN(float64(factor)) || math.IsInf(float64(factor), 0) || factor < 0 {
			return ErrDimension
		}
	}
	s.factors = slices.Clone(factors)
	return nil
}

func (s *MsatSolver) MarkLastClauseTypeA() error {
	if err := s.checkLive(); err != nil {
		return err
	}
	if len(s.original) == 0 {
		return ErrDimension
	}
	return s.original[len(s.original)-1].SetTypeA(true)
}

func (s *MsatSolver) MarkClausesStart() error {
	if err := s.checkLive(); err != nil {
		return err
	}
	s.clausesStart = len(s.original)
	return nil
}

func (s *MsatSolver) Clause(number int) (*MsatClause, error) {
	if err := s.checkLive(); err != nil {
		return nil, err
	}
	if number < 0 || number >= len(s.original) {
		return nil, ErrDimension
	}
	return s.original[number], nil
}

func (s *MsatSolver) ProgressEstimate() (float64, error) {
	if err := s.checkLive(); err != nil {
		return 0, err
	}
	variables := s.VariableCount()
	if variables == 0 {
		return 0, nil
	}
	factor := 1 / float64(variables)
	progress := 0.0
	for variable, assignment := range s.core.assigns {
		if assignment >= 0 {
			progress += math.Pow(factor, float64(s.core.levels[variable]))
		}
	}
	return progress / float64(variables), nil
}

func (s *MsatSolver) Statistics() MsatSolverStats {
	if s == nil || s.closed || s.core == nil {
		return MsatSolverStats{}
	}
	stats := s.core.Statistics()
	result := MsatSolverStats{Starts: s.starts, Decisions: stats.Decisions, Propagations: stats.Propagations, Inspections: stats.Inspections, Conflicts: stats.Conflicts, Successes: s.successes}
	return result
}

func (s *MsatSolver) StatisticsDiagnostic() string {
	if s == nil || s.checkLive() != nil {
		return ""
	}
	stats := s.Statistics()
	return fmt.Sprintf("C solver (%d vars; %d clauses; %d learned):\nstarts        : %d\nconflicts     : %d\ndecisions     : %d\npropagations  : %d\ninspects      : %d\n",
		s.VariableCount(), len(s.original), s.core.LearnedClauseCount(), stats.Starts, stats.Conflicts, stats.Decisions, stats.Propagations, stats.Inspections)
}

func MsatTruthTables() [6][2]uint32 {
	var result [6][2]uint32
	for minterm := 0; minterm < 32; minterm++ {
		for variable := 0; variable < 5; variable++ {
			if minterm&(1<<variable) != 0 {
				result[variable][0] |= 1 << minterm
			}
		}
	}
	for variable := 0; variable < 5; variable++ {
		result[variable][1] = result[variable][0]
	}
	result[5][1] = math.MaxUint32
	return result
}

func (s *MsatSolver) Close() {
	if s == nil || s.closed {
		return
	}
	for _, clause := range s.original {
		clause.closed = true
		clause.literals = nil
	}
	s.core = nil
	s.memory.Close()
	s.memory = nil
	s.variableLevels = nil
	s.factors = nil
	s.original = nil
	s.model = nil
	s.seen = nil
	s.adjacents = nil
	s.coneVariables = nil
	s.variablesUsed = nil
	s.closed = true
}

func msatClauseText(literals []SatLit) string {
	var builder strings.Builder
	for _, literal := range literals {
		if literal.Negative() {
			fmt.Fprintf(&builder, " -%d", literal.Var()+1)
		} else {
			fmt.Fprintf(&builder, " %d", literal.Var()+1)
		}
	}
	builder.WriteByte('\n')
	return builder.String()
}
