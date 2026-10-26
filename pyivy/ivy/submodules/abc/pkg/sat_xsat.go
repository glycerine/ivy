package abc

// Port of src/sat/xsat: the xSAT API, DIMACS reader, and concrete storage helpers.

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"time"
)

type XSatStatus int8

const (
	XSatFalse XSatStatus = -1
	XSatUndef XSatStatus = 0
	XSatTrue  XSatStatus = 1
)

const XSatCRefUndef = ^uint32(0)

func XSatVar2Lit(variable int, negative bool) (SatLit, error) {
	if variable < 0 || uint64(variable) > uint64(MaxLiteralID) {
		return 0, ErrInvalidLiteral
	}
	return MakeSatLit(SatVar(variable), negative)
}

func XSatNegLit(literal SatLit) (SatLit, error) {
	if !literal.Valid() {
		return 0, ErrInvalidLiteral
	}
	return literal.Not(), nil
}

type XSatOptions struct {
	Verbose                                        bool
	ConflictLimit, PropagationLimit                uint64
	RuntimeLimit                                   time.Duration
	RestartForce, RestartBlock                     float64
	FirstBlockRestart, LBDQueue, TrailQueue        int
	FirstReduce, ReduceIncrement, SpecialIncrement int
	FrozenLBD                                      uint32
}

func DefaultXSatOptions() XSatOptions {
	return XSatOptions{Verbose: true, RestartForce: .8, RestartBlock: 1.4, FirstBlockRestart: 10000,
		LBDQueue: 50, TrailQueue: 5000, FirstReduce: 2000, ReduceIncrement: 300,
		SpecialIncrement: 1000, FrozenLBD: 30}
}

type XSatLimits struct {
	MaxVariables, MaxClauses, MaxLiterals, MaxInputBytes int
	Stop                                                 func() bool
}

func normalizeXSatLimits(limits XSatLimits) (XSatLimits, error) {
	if limits.MaxVariables < 0 || limits.MaxClauses < 0 || limits.MaxLiterals < 0 || limits.MaxInputBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxVariables == 0 {
		limits.MaxVariables = 1 << 20
	}
	if limits.MaxClauses == 0 {
		limits.MaxClauses = 1 << 24
	}
	if limits.MaxLiterals == 0 {
		limits.MaxLiterals = 1 << 28
	}
	if limits.MaxInputBytes == 0 {
		limits.MaxInputBytes = 1 << 30
	}
	if uint64(limits.MaxVariables) > uint64(MaxLiteralID)+1 {
		return limits, ErrLimit
	}
	return limits, nil
}

type XSatStats struct {
	Starts, Reductions                              uint64
	Decisions, Propagations, Inspections, Conflicts uint64
	ClauseLiterals, LearnedLiterals                 uint64
}

type XSatResult struct {
	Status XSatStatus
	Reason SatStopReason
	Model  []bool
	Stats  XSatStats
}

func (r XSatResult) Clone() XSatResult { r.Model = slices.Clone(r.Model); return r }

type XSatSolver struct {
	core                 *SatSolver
	limits               XSatLimits
	options              XSatOptions
	clauses              [][]SatLit
	model                []bool
	last                 XSatResult
	literals             int
	starts               uint64
	inconsistent, closed bool
}

func NewXSatSolver() *XSatSolver {
	s, _ := NewXSatSolverWithLimits(XSatLimits{})
	return s
}

func NewXSatSolverWithLimits(limits XSatLimits) (*XSatSolver, error) {
	var err error
	limits, err = normalizeXSatLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	core, err := NewSatSolverWithLimit(limits.MaxVariables)
	if err != nil {
		return nil, err
	}
	return &XSatSolver{core: core, limits: limits, options: DefaultXSatOptions()}, nil
}

func (s *XSatSolver) check() error {
	if s == nil || s.closed || s.core == nil {
		return ErrInvariant
	}
	return nil
}

func (s *XSatSolver) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	s.core = nil
	s.clauses = nil
	s.model = nil
	return nil
}

func validateXSatOptions(o XSatOptions) error {
	if math.IsNaN(o.RestartForce) || math.IsInf(o.RestartForce, 0) || math.IsNaN(o.RestartBlock) || math.IsInf(o.RestartBlock, 0) ||
		o.RestartForce < 0 || o.RestartBlock < 0 || o.RuntimeLimit < 0 || o.FirstBlockRestart < 0 || o.LBDQueue <= 0 || o.TrailQueue <= 0 ||
		o.FirstReduce < 0 || o.ReduceIncrement < 0 || o.SpecialIncrement < 0 {
		return ErrDimension
	}
	return nil
}

func (s *XSatSolver) SetOptions(options XSatOptions) error {
	if err := s.check(); err != nil {
		return err
	}
	if err := validateXSatOptions(options); err != nil {
		return err
	}
	s.options = options
	return nil
}

func (s *XSatSolver) Options() (XSatOptions, error) {
	if err := s.check(); err != nil {
		return XSatOptions{}, err
	}
	return s.options, nil
}

func (s *XSatSolver) AddVariable(_ bool) error {
	if err := s.check(); err != nil {
		return err
	}
	if s.limits.Stop != nil && s.limits.Stop() {
		return ErrCancelled
	}
	if s.core.VariableCount() >= s.limits.MaxVariables {
		return ErrLimit
	}
	if err := s.core.SetVariableCount(s.core.VariableCount() + 1); err != nil {
		return err
	}
	s.last = XSatResult{}
	s.model = nil
	return nil
}

func normalizeXSatClause(literals []SatLit) ([]SatLit, bool, int, error) {
	result := slices.Clone(literals)
	maximum := -1
	for _, literal := range result {
		if !literal.Valid() {
			return nil, false, -1, ErrInvalidLiteral
		}
		maximum = max(maximum, int(literal.Var()))
	}
	slices.Sort(result)
	write := 0
	for _, literal := range result {
		if write != 0 {
			previous := result[write-1]
			if literal == previous {
				continue
			}
			if literal == previous.Not() {
				return nil, true, maximum, nil
			}
		}
		result[write] = literal
		write++
	}
	return result[:write], false, maximum, nil
}

func (s *XSatSolver) AddClause(literals []SatLit) (bool, error) {
	if err := s.check(); err != nil {
		return false, err
	}
	if s.limits.Stop != nil && s.limits.Stop() {
		return false, ErrCancelled
	}
	normalized, tautology, maximum, err := normalizeXSatClause(literals)
	if err != nil {
		return false, err
	}
	if maximum >= s.limits.MaxVariables {
		return false, ErrLimit
	}
	if tautology {
		if maximum >= s.core.VariableCount() {
			if err := s.core.SetVariableCount(maximum + 1); err != nil {
				return false, err
			}
		}
		s.last = XSatResult{}
		s.model = nil
		return true, nil
	}
	if len(s.clauses) >= s.limits.MaxClauses || len(normalized) > s.limits.MaxLiterals-s.literals {
		return false, ErrLimit
	}
	ok, err := s.core.AddClause(normalized)
	if err != nil {
		return false, err
	}
	s.clauses = append(s.clauses, slices.Clone(normalized))
	s.literals += len(normalized)
	s.inconsistent = !ok
	s.model = nil
	s.last = XSatResult{}
	return ok, nil
}

func (s *XSatSolver) Simplify() (bool, error) {
	if err := s.check(); err != nil {
		return false, err
	}
	if s.limits.Stop != nil && s.limits.Stop() {
		return false, ErrCancelled
	}
	if s.inconsistent {
		return false, nil
	}
	ok, err := s.core.Simplify()
	if err != nil {
		return false, err
	}
	s.inconsistent = !ok
	return ok, nil
}

func minNonzeroXSat(a, b uint64) uint64 {
	if a == 0 {
		return b
	}
	if b == 0 || a < b {
		return a
	}
	return b
}

func xSatStats(stats SatStats, starts uint64, clauseLiterals int) XSatStats {
	return XSatStats{Starts: starts, Decisions: stats.Decisions, Propagations: stats.Propagations,
		Inspections: stats.Inspections, Conflicts: stats.Conflicts, ClauseLiterals: uint64(clauseLiterals),
		LearnedLiterals: stats.Learnt}
}

func (s *XSatSolver) Solve(limits SatLimits) (XSatResult, error) {
	if err := s.check(); err != nil {
		return XSatResult{}, err
	}
	if s.limits.Stop != nil && s.limits.Stop() {
		return XSatResult{}, ErrCancelled
	}
	if len(s.clauses) == 0 {
		if limits.Stop != nil && limits.Stop() {
			return XSatResult{Status: XSatUndef, Reason: SatCancelled}, nil
		}
		s.starts++
		result := XSatResult{Status: XSatTrue, Model: make([]bool, s.core.VariableCount()), Stats: xSatStats(s.core.Statistics(), s.starts, 0)}
		s.last = result.Clone()
		s.model = slices.Clone(result.Model)
		return result, nil
	}
	limits.Conflicts = minNonzeroXSat(limits.Conflicts, s.options.ConflictLimit)
	limits.Propagations = minNonzeroXSat(limits.Propagations, s.options.PropagationLimit)
	outerStop := limits.Stop
	limits.Stop = func() bool {
		return (s.limits.Stop != nil && s.limits.Stop()) || (outerStop != nil && outerStop())
	}
	var oldDeadline time.Time
	var err error
	if s.options.RuntimeLimit != 0 {
		oldDeadline, err = s.core.SetRuntimeDeadline(time.Now().Add(s.options.RuntimeLimit))
		if err != nil {
			return XSatResult{}, err
		}
		defer s.core.SetRuntimeDeadline(oldDeadline)
	}
	answer, err := s.core.Solve(nil, limits)
	if err != nil {
		return XSatResult{}, err
	}
	status := XSatUndef
	if answer.Status == SatSatisfiable {
		status = XSatTrue
	}
	if answer.Status == SatUnsatisfiable {
		status = XSatFalse
	}
	s.starts++
	result := XSatResult{Status: status, Reason: answer.Reason, Model: slices.Clone(answer.Model),
		Stats: xSatStats(s.core.Statistics(), s.starts, s.literals)}
	s.last = result.Clone()
	s.model = slices.Clone(result.Model)
	return result, nil
}

func (s *XSatSolver) LastResult() (XSatResult, error) {
	if err := s.check(); err != nil {
		return XSatResult{}, err
	}
	return s.last.Clone(), nil
}

func (s *XSatSolver) Statistics() (XSatStats, error) {
	if err := s.check(); err != nil {
		return XSatStats{}, err
	}
	return xSatStats(s.core.Statistics(), s.starts, s.literals), nil
}

func (s *XSatSolver) ModelValue(variable int) (bool, error) {
	if err := s.check(); err != nil {
		return false, err
	}
	if s.last.Status != XSatTrue || variable < 0 || variable >= len(s.model) {
		return false, ErrInvalidLiteral
	}
	return s.model[variable], nil
}

func (s *XSatSolver) VariableCount() int {
	if s == nil || s.closed || s.core == nil {
		return 0
	}
	return s.core.VariableCount()
}
func (s *XSatSolver) ClauseCount() int {
	if s == nil || s.closed {
		return 0
	}
	return len(s.clauses)
}
func (s *XSatSolver) Clauses() [][]SatLit {
	if s == nil || s.closed {
		return nil
	}
	r := make([][]SatLit, len(s.clauses))
	for i := range r {
		r[i] = slices.Clone(s.clauses[i])
	}
	return r
}

func (s *XSatSolver) StatsText() (string, error) {
	if err := s.check(); err != nil {
		return "", err
	}
	stats, err := s.Statistics()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("starts        : %10d\nconflicts     : %10d\ndecisions     : %10d\npropagations  : %10d\n",
		stats.Starts, stats.Conflicts, stats.Decisions, stats.Propagations), nil
}

type xsatDIMACSParser struct {
	data   []byte
	at     int
	limits XSatLimits
}

func xsatDIMACSSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\r' || value == '\n' || value == '\v' || value == '\f'
}

func (p *xsatDIMACSParser) skipSpaceAndComments() {
	for p.at < len(p.data) {
		if p.data[p.at] == 0 {
			p.at = len(p.data)
			return
		}
		comment := p.data[p.at] == 'c'
		if comment {
			for index := p.at - 1; index >= 0 && p.data[index] != '\n'; index-- {
				if !xsatDIMACSSpace(p.data[index]) {
					comment = false
					break
				}
			}
		}
		if comment {
			for p.at < len(p.data) && p.data[p.at] != '\n' {
				p.at++
			}
			continue
		}
		if xsatDIMACSSpace(p.data[p.at]) {
			p.at++
			continue
		}
		break
	}
}

func (p *xsatDIMACSParser) word() (string, error) {
	p.skipSpaceAndComments()
	start := p.at
	for p.at < len(p.data) && !xsatDIMACSSpace(p.data[p.at]) && p.data[p.at] != 0 {
		p.at++
	}
	if start == p.at {
		return "", ErrDimension
	}
	return string(p.data[start:p.at]), nil
}

func (p *xsatDIMACSParser) integer() (int, error) {
	word, err := p.word()
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseInt(word, 10, 64)
	if err != nil || value < int64(math.MinInt32) || value > int64(math.MaxInt32) {
		return 0, ErrDimension
	}
	return int(value), nil
}

func ParseXSatDIMACS(data []byte, limits XSatLimits) (*XSatSolver, bool, error) {
	var err error
	limits, err = normalizeXSatLimits(limits)
	if err != nil {
		return nil, false, err
	}
	if len(data) > limits.MaxInputBytes {
		return nil, false, ErrLimit
	}
	p := xsatDIMACSParser{data: data, limits: limits}
	marker, err := p.word()
	if err != nil || marker != "p" {
		return nil, false, ErrDimension
	}
	_, err = p.word()
	if err != nil {
		return nil, false, ErrDimension
	}
	variables, err := p.integer()
	if err != nil || variables < 0 {
		return nil, false, ErrDimension
	}
	declaredClauses, err := p.integer()
	if err != nil || declaredClauses < 0 {
		return nil, false, ErrDimension
	}
	if variables > limits.MaxVariables || declaredClauses > limits.MaxClauses {
		return nil, false, ErrLimit
	}
	// The source reader ignores the rest of the parameter line, including an
	// unvalidated format token and trailing fields.
	for p.at < len(p.data) && p.data[p.at] != '\n' && p.data[p.at] != 0 {
		p.at++
	}
	solver, err := NewXSatSolverWithLimits(limits)
	if err != nil {
		return nil, false, err
	}
	// The source reader treats the header counts as allocation hints. It reads
	// clauses to EOF and grows variables to the largest literal encountered.
	for {
		p.skipSpaceAndComments()
		if p.at == len(p.data) {
			break
		}
		literals := make([]SatLit, 0)
		for {
			value, parseErr := p.integer()
			if parseErr != nil {
				solver.Close()
				return nil, false, parseErr
			}
			if value == 0 {
				break
			}
			variable := value
			if variable < 0 {
				variable = -variable
			}
			if variable == 0 {
				solver.Close()
				return nil, false, ErrInvalidLiteral
			}
			literal, literalErr := XSatVar2Lit(variable-1, value < 0)
			if literalErr != nil {
				solver.Close()
				return nil, false, literalErr
			}
			literals = append(literals, literal)
		}
		ok, addErr := solver.AddClause(literals)
		if addErr != nil {
			solver.Close()
			return nil, false, addErr
		}
		if !ok {
			return solver, false, nil
		}
	}
	ok, err := solver.Simplify()
	if err != nil {
		solver.Close()
		return nil, false, err
	}
	return solver, ok, nil
}

type xsatBoundedQueue struct {
	data        []uint32
	first, size int
	sum         uint64
}

func newXSatBoundedQueue(capacity int) (*xsatBoundedQueue, error) {
	if capacity <= 0 {
		return nil, ErrDimension
	}
	return &xsatBoundedQueue{data: make([]uint32, capacity)}, nil
}
func (q *xsatBoundedQueue) push(value uint32) {
	if q.size == len(q.data) {
		q.sum -= uint64(q.data[q.first])
		q.first = (q.first + 1) % len(q.data)
		q.size--
	}
	at := (q.first + q.size) % len(q.data)
	q.data[at] = value
	q.sum += uint64(value)
	q.size++
}
func (q *xsatBoundedQueue) pop() (uint32, error) {
	if q == nil || q.size == 0 {
		return 0, ErrInvariant
	}
	v := q.data[q.first]
	q.first = (q.first + 1) % len(q.data)
	q.size--
	q.sum -= uint64(v)
	return v, nil
}
func (q *xsatBoundedQueue) average() (uint32, error) {
	if q == nil || q.size == 0 {
		return 0, ErrInvariant
	}
	return uint32(q.sum / uint64(q.size)), nil
}
func (q *xsatBoundedQueue) clean() { q.first = 0; q.size = 0; q.sum = 0 }

type xsatHeap struct {
	activity       []uint32
	heap, position []int
}

func newXSatHeap(activity []uint32) *xsatHeap { return &xsatHeap{activity: activity} }
func (h *xsatHeap) in(variable int) bool {
	return variable >= 0 && variable < len(h.position) && h.position[variable] >= 0
}
func (h *xsatHeap) insert(variable int) error {
	if variable < 0 || variable >= len(h.activity) || h.in(variable) {
		return ErrDimension
	}
	for len(h.position) <= variable {
		h.position = append(h.position, -1)
	}
	h.position[variable] = len(h.heap)
	h.heap = append(h.heap, variable)
	h.up(len(h.heap) - 1)
	return nil
}
func (h *xsatHeap) up(i int) {
	x := h.heap[i]
	for i > 0 {
		p := (i - 1) / 2
		if h.activity[h.heap[p]] >= h.activity[x] {
			break
		}
		h.heap[i] = h.heap[p]
		h.position[h.heap[i]] = i
		i = p
	}
	h.heap[i] = x
	h.position[x] = i
}
func (h *xsatHeap) down(i int) {
	x := h.heap[i]
	for 2*i+1 < len(h.heap) {
		c := 2*i + 1
		if c+1 < len(h.heap) && h.activity[h.heap[c+1]] > h.activity[h.heap[c]] {
			c++
		}
		if h.activity[h.heap[c]] <= h.activity[x] {
			break
		}
		h.heap[i] = h.heap[c]
		h.position[h.heap[i]] = i
		i = c
	}
	h.heap[i] = x
	h.position[x] = i
}
func (h *xsatHeap) update(variable int) error {
	if !h.in(variable) {
		return h.insert(variable)
	}
	i := h.position[variable]
	h.up(i)
	h.down(h.position[variable])
	return nil
}
func (h *xsatHeap) removeMax() (int, error) {
	if h == nil || len(h.heap) == 0 {
		return -1, ErrInvariant
	}
	x := h.heap[0]
	last := h.heap[len(h.heap)-1]
	h.heap = h.heap[:len(h.heap)-1]
	h.position[x] = -1
	if len(h.heap) > 0 {
		h.heap[0] = last
		h.position[last] = 0
		h.down(0)
	}
	return x, nil
}

type xsatWatcher struct {
	clause  uint32
	blocker SatLit
}
type xsatWatchList []xsatWatcher

func (w *xsatWatchList) remove(clause uint32) bool {
	for i := range *w {
		if (*w)[i].clause == clause {
			copy((*w)[i:], (*w)[i+1:])
			*w = (*w)[:len(*w)-1]
			return true
		}
	}
	return false
}

type xsatClauseArena struct {
	clauses [][]SatLit
	wasted  int
}

func (a *xsatClauseArena) append(literals []SatLit) (uint32, error) {
	if a == nil {
		return XSatCRefUndef, ErrInvariant
	}
	if uint64(len(a.clauses)) >= uint64(XSatCRefUndef) {
		return XSatCRefUndef, ErrLimit
	}
	h := uint32(len(a.clauses))
	a.clauses = append(a.clauses, slices.Clone(literals))
	return h, nil
}
func (a *xsatClauseArena) clause(handle uint32) ([]SatLit, error) {
	if a == nil || uint64(handle) >= uint64(len(a.clauses)) {
		return nil, ErrInvalidLiteral
	}
	return slices.Clone(a.clauses[handle]), nil
}
func (a *xsatClauseArena) restart() { a.clauses = nil; a.wasted = 0 }
