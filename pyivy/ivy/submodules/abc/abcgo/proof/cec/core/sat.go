package core

import (
	"context"
	"fmt"
	"github.com/glycerine/abcgo/sat/satoko"
	"time"
)

type Status int8

const (
	Undecided  Status = -1
	Different  Status = 0
	Equivalent Status = 1
)

type Counterexample struct {
	Inputs      []bool
	Output      int
	Left, Right bool
}
type Result struct {
	Status         Status
	Counterexample *Counterexample
}
type Options struct {
	ConflictLimit                   int64
	MaxSATVariables, RecycleQueries int
	PolarFlip                       bool
	Deadline                        time.Time
	Context                         context.Context
	TreatRegistersAsInputs          bool
}

func DefaultSATOptions() Options {
	return Options{ConflictLimit: 100, MaxSATVariables: 2000, RecycleQueries: 200, PolarFlip: true}
}
func DefaultCECOptions() Options { o := DefaultSATOptions(); o.ConflictLimit = 1000; return o }

// Query counts include phase/trivial checks; result counters count CDCL calls.
type Statistics struct {
	Queries, SATCalls, Satisfiable, Unsatisfiable, Undecided, Recycles, Variables, Clauses int
	Conflicts                                                                              int64
	Elapsed                                                                                time.Duration
}

// Manager provides incremental cone-CNF and two-direction assumption checks
// using Satoko. This Go adapter is separate from the source-shaped CEC SAT
// implementation in the parent package. A Manager is not concurrent-safe.
type Manager struct {
	models       []*Model
	options      Options
	solver       *satoko.Solver
	literals     [][]int
	inputs       []int
	sinceRecycle int
	stats        Statistics
	oldConflicts int64
	started      time.Time
}

func chooseOptions(defaults Options, provided []Options) (Options, error) {
	if len(provided) > 1 {
		return Options{}, fmt.Errorf("cec: expected at most one options value")
	}
	if len(provided) == 1 {
		defaults = provided[0]
	}
	if defaults.ConflictLimit < 0 || defaults.MaxSATVariables < 0 || defaults.RecycleQueries < 0 {
		return Options{}, fmt.Errorf("cec: negative resource limit")
	}
	return defaults, nil
}
func NewManager(n Network, provided ...Options) (*Manager, error) {
	options, err := chooseOptions(DefaultSATOptions(), provided)
	if err != nil {
		return nil, err
	}
	model, err := NewModel(n)
	if err != nil {
		return nil, err
	}
	return newManager([]*Model{model}, options)
}
func newManager(models []*Model, options Options) (*Manager, error) {
	for _, model := range models {
		if !options.TreatRegistersAsInputs && model.registers != 0 {
			return nil, fmt.Errorf("cec: sequential network requires explicit combinational register treatment")
		}
		if len(model.inputs) != len(models[0].inputs) {
			return nil, fmt.Errorf("cec: input count mismatch")
		}
	}
	m := &Manager{models: models, options: options, started: time.Now()}
	m.reset()
	return m, nil
}
func (m *Manager) reset() {
	if m.solver != nil {
		m.oldConflicts += m.solver.Stats().AllConflicts
		m.stats.Recycles++
	}
	m.solver = satoko.New()
	m.solver.AddVariable(false)
	if !m.solver.AddClause([]int{1}) {
		panic("cec: contradictory constant encoding")
	}
	m.inputs = make([]int, len(m.models[0].inputs))
	for i := range m.inputs {
		m.inputs[i] = -1
	}
	m.literals = make([][]int, len(m.models))
	for i, model := range m.models {
		cache := make([]int, len(model.gates))
		for j := range cache {
			cache[j] = -1
		}
		cache[0] = 0
		m.literals[i] = cache
	}
	m.sinceRecycle = 0
	m.solver.SetDeadline(m.options.Deadline)
	if m.options.Context != nil {
		m.solver.SetStopFunc(func(int) int {
			if m.options.Context.Err() != nil {
				return 1
			}
			return 0
		})
	}
}
func (m *Manager) Statistics() Statistics {
	stats := m.stats
	stats.Variables = m.solver.VariableCount()
	stats.Clauses = m.solver.ClauseCount()
	stats.Conflicts = m.oldConflicts + m.solver.Stats().AllConflicts
	stats.Elapsed = time.Since(m.started)
	return stats
}
func (m *Manager) interrupted() bool {
	return (!m.options.Deadline.IsZero() && time.Now().After(m.options.Deadline)) || (m.options.Context != nil && m.options.Context.Err() != nil)
}
func (m *Manager) begin() {
	m.stats.Queries++
	m.sinceRecycle++
	if m.options.MaxSATVariables > 0 && m.solver.VariableCount() > m.options.MaxSATVariables && m.sinceRecycle > m.options.RecycleQueries {
		m.reset()
	}
}
func (m *Manager) clause(lits ...int) {
	if !m.solver.AddClause(lits) {
		panic("cec: inconsistent circuit CNF")
	}
}
func (m *Manager) inputLiteral(index int) int {
	if m.inputs[index] < 0 {
		m.inputs[index] = m.solver.AddVariable(false) << 1
	}
	return m.inputs[index]
}
func (m *Manager) encode(which int, lit int32) int {
	id := lit >> 1
	cache := m.literals[which]
	if cache[id] >= 0 {
		return cache[id] ^ int(lit&1)
	}
	model := m.models[which]
	node := model.gates[id]
	if node.kind == Input {
		value := m.inputLiteral(model.inputIndex[id<<1])
		cache[id] = value
		return value ^ int(lit&1)
	}
	var args [3]int
	for i, fanin := range node.fanins[:gateArity(node.kind)] {
		args[i] = m.encode(which, fanin)
	}
	if node.kind == Buffer || node.kind == Output {
		cache[id] = args[0]
		return args[0] ^ int(lit&1)
	}
	z := m.solver.AddVariable(false) << 1
	if m.options.PolarFlip && model.phase[id] {
		z ^= 1
	}
	cache[id] = z
	switch node.kind {
	case And:
		a, b := args[0], args[1]
		m.clause(a^1, b^1, z)
		m.clause(a, z^1)
		m.clause(b, z^1)
	case Xor:
		a, b := args[0], args[1]
		m.clause(a, b, z^1)
		m.clause(a^1, b^1, z^1)
		m.clause(a, b^1, z)
		m.clause(a^1, b, z)
	case Mux:
		i, t, e := args[0], args[1], args[2]
		m.clause(i^1, t^1, z)
		m.clause(i^1, t, z^1)
		m.clause(i, e^1, z)
		m.clause(i, e, z^1)
		m.clause(t, e, z^1)
		m.clause(t^1, e^1, z)
	default:
		panic("cec: invalid encoded gate")
	}
	return z ^ int(lit&1)
}
func (m *Manager) solve(assumptions ...int) int {
	m.stats.SATCalls++
	status := m.solver.SolveAssumptionsLimit(assumptions, m.options.ConflictLimit)
	switch status {
	case satoko.SATOKO_SAT:
		m.stats.Satisfiable++
	case satoko.SATOKO_UNSAT:
		m.stats.Unsatisfiable++
	default:
		m.stats.Undecided++
	}
	return status
}
func (m *Manager) assignment() []bool {
	inputs := make([]bool, len(m.inputs))
	for i, lit := range m.inputs {
		if lit >= 0 {
			inputs[i] = m.solver.ModelValue(lit>>1) == 1
		}
	}
	return inputs
}
func (m *Manager) witness(inputs []bool, leftModel int, left int32, rightModel int, right int32, output int) (Result, error) {
	leftValues, err := m.models[leftModel].Values(inputs)
	if err != nil {
		return Result{}, err
	}
	rightValues := leftValues
	if rightModel != leftModel {
		rightValues, err = m.models[rightModel].Values(inputs)
		if err != nil {
			return Result{}, err
		}
	}
	a := leftValues[left>>1] != (left&1 != 0)
	b := rightValues[right>>1] != (right&1 != 0)
	if a == b {
		return Result{}, fmt.Errorf("cec: SAT counterexample failed independent simulation")
	}
	return Result{Different, &Counterexample{inputs, output, a, b}}, nil
}

// CheckLiteral proves a signal is constant zero (1), finds a witness (0), or
// returns undecided (-1), matching Cec_ManSatCheckNode's result convention.
func (m *Manager) CheckLiteral(lit int32) (Result, error) {
	if !m.models[0].valid(lit) {
		return Result{}, fmt.Errorf("cec: invalid query literal")
	}
	if m.interrupted() {
		return Result{Status: Undecided}, nil
	}
	// Cec_ManSatCheckNode handles the two constant pointers before updating
	// counters or allocating a cone. The C assert on constant one is represented
	// as a direct witness so the general-purpose core remains usable by its callers.
	if lit == 0 {
		return Result{Status: Equivalent}, nil
	}
	if lit == 1 {
		return m.witness(make([]bool, len(m.inputs)), 0, lit, 0, 0, -1)
	}
	m.begin()
	if m.interrupted() {
		return Result{Status: Undecided}, nil
	}
	a := m.encode(0, lit)
	status := m.solve(a)
	switch status {
	case satoko.SATOKO_UNSAT:
		m.clause(a ^ 1)
		return Result{Status: Equivalent}, nil
	case satoko.SATOKO_SAT:
		return m.witness(m.assignment(), 0, lit, 0, 0, -1)
	default:
		return Result{Status: Undecided}, nil
	}
}
func (m *Manager) CheckPair(a, b int32) (Result, error) {
	if !m.models[0].valid(a) || !m.models[0].valid(b) {
		return Result{}, fmt.Errorf("cec: invalid equivalence literal")
	}
	return m.checkPair(0, a, 0, b, -1)
}

// CheckConjunction checks whether both literals can be true. Its status uses
// the CEC convention: Different means SAT and Equivalent means UNSAT.
func (m *Manager) CheckConjunction(a, b int32) (Result, error) {
	if !m.models[0].valid(a) || !m.models[0].valid(b) {
		return Result{}, fmt.Errorf("cec: invalid conjunction literal")
	}
	m.begin()
	if m.interrupted() {
		return Result{Status: Undecided}, nil
	}
	x, y := m.encode(0, a), m.encode(0, b)
	switch m.solve(x, y) {
	case satoko.SATOKO_UNSAT:
		m.clause(x^1, y^1)
		return Result{Status: Equivalent}, nil
	case satoko.SATOKO_SAT:
		inputs := m.assignment()
		values, err := m.models[0].Values(inputs)
		if err != nil {
			return Result{}, err
		}
		left := values[a>>1] != (a&1 != 0)
		right := values[b>>1] != (b&1 != 0)
		if !left || !right {
			return Result{}, fmt.Errorf("cec: SAT conjunction witness failed independent simulation")
		}
		return Result{Different, &Counterexample{Inputs: inputs, Output: -1, Left: left, Right: right}}, nil
	default:
		return Result{Status: Undecided}, nil
	}
}
func (m *Manager) checkPair(leftModel int, left int32, rightModel int, right int32, output int) (Result, error) {
	m.begin()
	if m.interrupted() {
		return Result{Status: Undecided}, nil
	}
	leftPhase := m.models[leftModel].phase[left>>1] != (left&1 != 0)
	rightPhase := m.models[rightModel].phase[right>>1] != (right&1 != 0)
	if leftPhase != rightPhase {
		return m.witness(make([]bool, len(m.inputs)), leftModel, left, rightModel, right, output)
	}
	a, b := m.encode(leftModel, left), m.encode(rightModel, right)
	if a == b {
		return Result{Status: Equivalent}, nil
	}
	for _, assumptions := range [][2]int{{a, b ^ 1}, {a ^ 1, b}} {
		status := m.solve(assumptions[:]...)
		switch status {
		case satoko.SATOKO_SAT:
			return m.witness(m.assignment(), leftModel, left, rightModel, right, output)
		case satoko.SATOKO_UNSAT:
			m.clause(assumptions[0]^1, assumptions[1]^1)
		default:
			return Result{Status: Undecided}, nil
		}
	}
	return Result{Status: Equivalent}, nil
}

func VerifyTwo(left, right Network, provided ...Options) (Result, error) {
	options, err := chooseOptions(DefaultCECOptions(), provided)
	if err != nil {
		return Result{}, err
	}
	a, err := NewModel(left)
	if err != nil {
		return Result{}, err
	}
	b, err := NewModel(right)
	if err != nil {
		return Result{}, err
	}
	if len(a.outputs) != len(b.outputs) || a.registers != b.registers {
		return Result{}, fmt.Errorf("cec: output/register interface mismatch")
	}
	manager, err := newManager([]*Model{a, b}, options)
	if err != nil {
		return Result{}, err
	}
	unknown := false
	for i, lit := range a.outputs {
		result, err := manager.checkPair(0, lit, 1, b.outputs[i], i)
		if err != nil || result.Status == Different {
			return result, err
		}
		if result.Status == Undecided {
			unknown = true
		}
	}
	if unknown {
		return Result{Status: Undecided}, nil
	}
	return Result{Status: Equivalent}, nil
}

// VerifyPairs checks a dual-output miter: consecutive outputs are paired.
func VerifyPairs(n Network, provided ...Options) (Result, error) {
	options, err := chooseOptions(DefaultCECOptions(), provided)
	if err != nil {
		return Result{}, err
	}
	model, err := NewModel(n)
	if err != nil {
		return Result{}, err
	}
	if len(model.outputs)%2 != 0 {
		return Result{}, fmt.Errorf("cec: dual-output miter has odd output count")
	}
	manager, err := newManager([]*Model{model}, options)
	if err != nil {
		return Result{}, err
	}
	unknown := false
	for i := 0; i < len(model.outputs); i += 2 {
		result, err := manager.checkPair(0, model.outputs[i], 0, model.outputs[i+1], i/2)
		if err != nil || result.Status == Different {
			return result, err
		}
		if result.Status == Undecided {
			unknown = true
		}
	}
	if unknown {
		return Result{Status: Undecided}, nil
	}
	return Result{Status: Equivalent}, nil
}

// VerifyMiter proves every output false, for an XOR/OR difference miter.
func VerifyMiter(n Network, provided ...Options) (Result, error) {
	options, err := chooseOptions(DefaultCECOptions(), provided)
	if err != nil {
		return Result{}, err
	}
	manager, err := NewManager(n, options)
	if err != nil {
		return Result{}, err
	}
	unknown := false
	for i, lit := range manager.models[0].outputs {
		result, err := manager.CheckLiteral(lit)
		if result.Counterexample != nil {
			result.Counterexample.Output = i
		}
		if err != nil || result.Status == Different {
			return result, err
		}
		if result.Status == Undecided {
			unknown = true
		}
	}
	if unknown {
		return Result{Status: Undecided}, nil
	}
	return Result{Status: Equivalent}, nil
}

// Model returns the manager's immutable circuit snapshot for resimulation and
// regular-object value extraction without taking a second network snapshot.
func (m *Manager) Model() *Model { return m.models[0] }

// SetConflictLimit updates the next query's conflict budget without discarding
// learned clauses. Zero allows unlimited conflicts.
func (m *Manager) SetConflictLimit(limit int64) error {
	if limit < 0 {
		return fmt.Errorf("cec: negative conflict limit")
	}
	m.options.ConflictLimit = limit
	return nil
}
